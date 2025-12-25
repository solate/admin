package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/casbin"
	"admin/pkg/constants"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
)

// UserMenuService 用户菜单服务
type UserMenuService struct {
	userMenuRepo *repository.UserMenuRepo
	enforcer     *casbin.Enforcer
}

// NewUserMenuService 创建用户菜单服务
func NewUserMenuService(userMenuRepo *repository.UserMenuRepo, enforcer *casbin.Enforcer) *UserMenuService {
	return &UserMenuService{
		userMenuRepo: userMenuRepo,
		enforcer:     enforcer,
	}
}

// GetUserMenu 获取用户菜单树
func (s *UserMenuService) GetUserMenu(ctx context.Context) (*dto.UserMenuResponse, error) {
	// 1. 获取用户ID和租户ID
	userID := xcontext.GetUserID(ctx)
	tenantID := xcontext.GetTenantID(ctx)
	if userID == "" || tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 2. 超管特殊处理：返回所有菜单
	if s.isSuperAdmin(ctx, userID, tenantID) {
		return s.getAllMenusForTenant(ctx, tenantID)
	}

	// 3. 获取用户在租户中的角色
	roles, err := s.enforcer.GetRolesForUser(userID, tenantID)
	if err != nil {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}
	if len(roles) == 0 {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}

	// 4. 获取角色的所有 MENU 类型权限ID
	menuIDs, err := s.getPermissionsForRoles(ctx, tenantID, roles, constants.TypeMenu)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取菜单权限失败", err)
	}

	// 5. 查询权限详情
	menus, err := s.userMenuRepo.GetByPermissionIDs(ctx, menuIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单详情失败", err)
	}

	// 6. 过滤状态=1(显示)的菜单，且只保留 MENU 类型
	visibleMenus := s.filterVisibleMenus(menus)

	// 7. 构建树结构
	tree := s.buildMenuTree(visibleMenus)

	return &dto.UserMenuResponse{List: tree}, nil
}

// GetUserButtons 获取指定菜单的按钮权限
func (s *UserMenuService) GetUserButtons(ctx context.Context, menuID string) (*dto.UserButtonsResponse, error) {
	// 1. 获取用户ID和租户ID
	userID := xcontext.GetUserID(ctx)
	tenantID := xcontext.GetTenantID(ctx)
	if userID == "" || tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 2. 超管特殊处理：返回所有按钮
	if s.isSuperAdmin(ctx, userID, tenantID) {
		return s.getAllButtonsForMenu(ctx, tenantID, menuID)
	}

	// 3. 获取用户在租户中的角色
	roles, err := s.enforcer.GetRolesForUser(userID, tenantID)
	if err != nil || len(roles) == 0 {
		return &dto.UserButtonsResponse{Buttons: []*dto.ButtonInfo{}}, nil
	}

	// 4. 获取角色的所有 BUTTON 类型权限ID
	buttonIDs, err := s.getPermissionsForRoles(ctx, tenantID, roles, constants.TypeButton)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取按钮权限失败", err)
	}

	// 5. 查询权限详情
	buttons, err := s.userMenuRepo.GetByPermissionIDs(ctx, buttonIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询按钮详情失败", err)
	}

	// 6. 过滤出属于指定菜单的按钮
	var menuButtons []*dto.ButtonInfo
	for _, btn := range buttons {
		if btn.ParentID != nil && *btn.ParentID == menuID {
			menuButtons = append(menuButtons, &dto.ButtonInfo{
				PermissionID: btn.PermissionID,
				Name:         btn.Name,
				Action:       btn.Action,
				Resource:     btn.Resource,
			})
		}
	}

	return &dto.UserButtonsResponse{Buttons: menuButtons}, nil
}

// isSuperAdmin 判断是否为超管（role_code=super_admin 且 tenant_id=default）
func (s *UserMenuService) isSuperAdmin(ctx context.Context, userID, tenantID string) bool {
	if tenantID != constants.DefaultTenant {
		return false
	}
	roles, err := s.enforcer.GetRolesForUser(userID, tenantID)
	if err != nil {
		return false
	}
	for _, role := range roles {
		// 检查角色编码是否为 super_admin
		if role == constants.SuperAdmin {
			return true
		}
	}
	return false
}

// getAllMenusForTenant 获取租户的所有菜单（超管使用）
func (s *UserMenuService) getAllMenusForTenant(ctx context.Context, tenantID string) (*dto.UserMenuResponse, error) {
	menus, err := s.userMenuRepo.GetTenantAvailableMenus(ctx, tenantID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	// 过滤显示状态的菜单
	visibleMenus := s.filterVisibleMenus(menus)

	// 构建树结构
	tree := s.buildMenuTree(visibleMenus)

	return &dto.UserMenuResponse{List: tree}, nil
}

// getAllButtonsForMenu 获取菜单的所有按钮（超管使用）
func (s *UserMenuService) getAllButtonsForMenu(ctx context.Context, tenantID, menuID string) (*dto.UserButtonsResponse, error) {
	buttons, err := s.userMenuRepo.GetButtonsByMenuID(ctx, tenantID, menuID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询按钮失败", err)
	}

	var buttonInfos []*dto.ButtonInfo
	for _, btn := range buttons {
		buttonInfos = append(buttonInfos, &dto.ButtonInfo{
			PermissionID: btn.PermissionID,
			Name:         btn.Name,
			Action:       btn.Action,
			Resource:     btn.Resource,
		})
	}

	return &dto.UserButtonsResponse{Buttons: buttonInfos}, nil
}

// getPermissionsForRoles 获取角色的指定类型权限ID列表
func (s *UserMenuService) getPermissionsForRoles(ctx context.Context, tenantID string, roles []string, permType string) ([]string, error) {
	permissionSet := make(map[string]bool)

	for _, role := range roles {
		// 使用 Casbin 获取角色的策略
		// 策略格式: p, role_id, tenant_id, resource, action
		policies, _ := s.enforcer.GetFilteredPolicy(0, role, tenantID)

		for _, policy := range policies {
			// policy[2] 是 resource (permission_id 或 API 路径)
			// policy[3] 是 action (操作方法)
			if len(policy) >= 4 {
				resource := policy[2]
				action := policy[3]

				// 对于 MENU/BUTTON 类型，resource 是 permission_id，action 是 "*"
				// 对于 API 类型，resource 是 API 路径，action 是 HTTP 方法
				if permType == constants.TypeMenu || permType == constants.TypeButton {
					// 如果 action 是通配符，认为是菜单/按钮权限
					if action == "*" || action == "" {
						// 查询权限详情验证类型
						// 这里先简单处理，将所有可能的 permission_id 收集起来
						// 后续在查询时再过滤类型
						permissionSet[resource] = true
					}
				}
			}
		}
	}

	// 转换为切片
	permissionIDs := make([]string, 0, len(permissionSet))
	for permID := range permissionSet {
		permissionIDs = append(permissionIDs, permID)
	}

	return permissionIDs, nil
}

// filterVisibleMenus 过滤显示状态的菜单，且只保留 MENU 类型
func (s *UserMenuService) filterVisibleMenus(menus []*model.Permission) []*model.Permission {
	var result []*model.Permission
	for _, menu := range menus {
		// 只保留 MENU 类型
		if menu.Type != constants.TypeMenu {
			continue
		}
		// 只保留显示状态
		if menu.Status != nil && *menu.Status == constants.MenuStatusShow {
			result = append(result, menu)
		}
	}
	return result
}

// buildMenuTree 构建菜单树
func (s *UserMenuService) buildMenuTree(menus []*model.Permission) []*dto.MenuTreeNode {
	// 创建节点映射
	nodeMap := make(map[string]*dto.MenuTreeNode)
	for _, menu := range menus {
		nodeMap[menu.PermissionID] = &dto.MenuTreeNode{
			MenuInfo: s.toMenuInfo(menu),
			Children: []*dto.MenuTreeNode{},
		}
	}

	// 构建树结构
	var roots []*dto.MenuTreeNode
	for _, menu := range menus {
		node := nodeMap[menu.PermissionID]
		if menu.ParentID == nil || *menu.ParentID == "" {
			// 顶级菜单
			roots = append(roots, node)
		} else if parent, exists := nodeMap[*menu.ParentID]; exists {
			// 添加到父菜单的子节点
			parent.Children = append(parent.Children, node)
		}
	}

	return roots
}

// toMenuInfo 转换为菜单信息格式
func (s *UserMenuService) toMenuInfo(menu *model.Permission) *dto.MenuInfo {
	status := int16(constants.MenuStatusShow) // 默认显示
	if menu.Status != nil {
		status = *menu.Status
	}

	info := &dto.MenuInfo{
		PermissionID: menu.PermissionID,
		TenantID:     menu.TenantID,
		Name:         menu.Name,
		Type:         menu.Type,
		ParentID:     menu.ParentID,
		Resource:     menu.Resource,
		Action:       menu.Action,
		Path:         menu.Path,
		Component:    menu.Component,
		Redirect:     menu.Redirect,
		Icon:         menu.Icon,
		Sort:         menu.Sort,
		Status:       status,
		Description:  menu.Description,
		CreatedAt:    menu.CreatedAt,
		UpdatedAt:    menu.UpdatedAt,
	}
	return info
}
