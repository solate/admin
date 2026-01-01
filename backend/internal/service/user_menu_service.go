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
	"strings"
)

// UserMenuService 用户菜单服务
type UserMenuService struct {
	menuRepo       *repository.MenuRepo
	permissionRepo *repository.PermissionRepo
	tenantMenuRepo *repository.TenantMenuRepo
	enforcer       *casbin.Enforcer
}

// NewUserMenuService 创建用户菜单服务
func NewUserMenuService(menuRepo *repository.MenuRepo, permissionRepo *repository.PermissionRepo, tenantMenuRepo *repository.TenantMenuRepo, enforcer *casbin.Enforcer) *UserMenuService {
	return &UserMenuService{
		menuRepo:       menuRepo,
		permissionRepo: permissionRepo,
		tenantMenuRepo: tenantMenuRepo,
		enforcer:       enforcer,
	}
}

// GetUserMenu 获取用户菜单树
func (s *UserMenuService) GetUserMenu(ctx context.Context, userName, tenantCode string) (*dto.UserMenuResponse, error) {
	// 1. 超管特殊处理：返回所有菜单
	if s.isSuperAdmin(userName, tenantCode) {
		return s.getAllMenusForTenant(ctx)
	}

	// 2. 获取用户在租户中的角色
	roles := s.enforcer.GetRolesForUserInDomain(userName, tenantCode)
	if len(roles) == 0 {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}

	// 3. 递归获取所有角色（通过 g2 处理继承）
	allRoleCodes := s.getAllRoleCodes(roles)

	// 4. 从 Casbin 获取角色的菜单权限
	menuIDs := s.getMenuPermissionsForRoles(ctx, allRoleCodes, tenantCode)

	// 5. 【关键】与租户菜单边界取交集
	tenantID := xcontext.GetTenantID(ctx)
	tenantMenuIDs, err := s.tenantMenuRepo.GetMenuIDsByTenant(ctx, tenantID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取租户菜单边界失败", err)
	}

	validMenuIDs := s.intersectMenuIDs(menuIDs, tenantMenuIDs)
	if len(validMenuIDs) == 0 {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}

	// 6. 查询菜单详情
	menus, err := s.menuRepo.GetByIDs(ctx, validMenuIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单详情失败", err)
	}

	// 7. 过滤状态=1(启用)的菜单
	visibleMenus := s.filterVisibleMenus(menus)

	// 8. 构建树结构
	tree := s.buildMenuTree(visibleMenus)

	return &dto.UserMenuResponse{List: tree}, nil
}

// GetUserButtons 获取指定菜单的按钮权限
func (s *UserMenuService) GetUserButtons(ctx context.Context, userName, tenantCode, menuID string) (*dto.UserButtonsResponse, error) {
	// 1. 超管特殊处理：返回所有按钮
	if s.isSuperAdmin(userName, tenantCode) {
		return s.getAllButtonsForMenu(ctx, menuID)
	}

	// 2. 获取用户在租户中的角色
	roles := s.enforcer.GetRolesForUserInDomain(userName, tenantCode)
	if len(roles) == 0 {
		return &dto.UserButtonsResponse{Buttons: []*dto.ButtonInfo{}}, nil
	}

	// 3. 递归获取所有角色（通过 g2 处理继承）
	allRoleCodes := s.getAllRoleCodes(roles)

	// 4. 获取角色的所有 BUTTON 类型权限
	buttonPerms, err := s.getButtonPermissionsForRoles(ctx, allRoleCodes, tenantCode)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取按钮权限失败", err)
	}

	// 5. 过滤出属于指定菜单的按钮（通过 resource 字段判断）
	var menuButtons []*dto.ButtonInfo
	for _, btn := range buttonPerms {
		// 检查 resource 是否以 "btn:" 开头且属于指定菜单
		// resource 格式: btn:menuID:action
		if strings.HasPrefix(btn.Resource, "btn:"+menuID+":") {
			menuButtons = append(menuButtons, &dto.ButtonInfo{
				PermissionID: btn.PermissionID,
				Name:         btn.Name,
				Action:       &btn.Action,
				Resource:     &btn.Resource,
			})
		}
	}

	return &dto.UserButtonsResponse{Buttons: menuButtons}, nil
}

// isSuperAdmin 判断是否为超管（role_code=super_admin 且在 default 租户）
func (s *UserMenuService) isSuperAdmin(userName, tenantCode string) bool {
	defaultTenantCode := constants.DefaultTenantCode
	if tenantCode != defaultTenantCode {
		return false
	}
	roles := s.enforcer.GetRolesForUserInDomain(userName, tenantCode)
	for _, role := range roles {
		if role == constants.SuperAdmin {
			return true
		}
	}
	return false
}

// getAllMenusForTenant 获取租户的所有菜单（超管使用）
func (s *UserMenuService) getAllMenusForTenant(ctx context.Context) (*dto.UserMenuResponse, error) {
	// 获取所有菜单
	menus, err := s.menuRepo.List(ctx)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	// 过滤启用状态的菜单
	visibleMenus := s.filterVisibleMenus(menus)

	// 构建树结构
	tree := s.buildMenuTree(visibleMenus)

	return &dto.UserMenuResponse{List: tree}, nil
}

// getAllButtonsForMenu 获取菜单的所有按钮（超管使用）
func (s *UserMenuService) getAllButtonsForMenu(ctx context.Context, menuID string) (*dto.UserButtonsResponse, error) {
	// 获取所有 BUTTON 类型的权限
	buttons, err := s.permissionRepo.ListByType(ctx, "BUTTON")
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询按钮失败", err)
	}

	var buttonInfos []*dto.ButtonInfo
	for _, btn := range buttons {
		// 检查 resource 是否以 "btn:" 开头且属于指定菜单
		if strings.HasPrefix(btn.Resource, "btn:"+menuID+":") {
			buttonInfos = append(buttonInfos, &dto.ButtonInfo{
				PermissionID: btn.PermissionID,
				Name:         btn.Name,
				Action:       &btn.Action,
				Resource:     &btn.Resource,
			})
		}
	}

	return &dto.UserButtonsResponse{Buttons: buttonInfos}, nil
}

// getAllRoleCodes 递归获取所有角色（通过 g2 处理继承）
func (s *UserMenuService) getAllRoleCodes(roleCodes []string) []string {
	var allCodes []string
	visited := make(map[string]bool)

	var dfs func(roleCode string)
	dfs = func(roleCode string) {
		if visited[roleCode] {
			return
		}
		visited[roleCode] = true
		allCodes = append(allCodes, roleCode)

		// 通过 g2 获取继承的父角色
		// GetRolesForUser 在 g2 中返回 (child -> parent)
		parents, _ := s.enforcer.GetRolesForUser(roleCode)
		for _, parentCode := range parents {
			dfs(parentCode)
		}
	}

	for _, code := range roleCodes {
		dfs(code)
	}

	return allCodes
}

// getMenuPermissionsForRoles 获取角色的 MENU 类型权限ID列表
func (s *UserMenuService) getMenuPermissionsForRoles(ctx context.Context, roles []string, tenantCode string) []string {
	menuIDSet := make(map[string]bool)

	for _, role := range roles {
		// 使用 Casbin 获取角色的策略
		// 策略格式: p, role, domain, resource, action
		policies, _ := s.enforcer.GetFilteredPolicy(0, role, tenantCode)

		for _, policy := range policies {
			if len(policy) >= 4 {
				resource := policy[2]
				action := policy[3]

				// 对于 MENU 类型，resource 格式为 "menu:menuID"，action 是 "*"
				if (action == "*" || action == "") && strings.HasPrefix(resource, "menu:") {
					menuID := strings.TrimPrefix(resource, "menu:")
					menuIDSet[menuID] = true
				}
			}
		}
	}

	// 转换为切片
	menuIDs := make([]string, 0, len(menuIDSet))
	for menuID := range menuIDSet {
		menuIDs = append(menuIDs, menuID)
	}

	return menuIDs
}

// getButtonPermissionsForRoles 获取角色的 BUTTON 类型权限
func (s *UserMenuService) getButtonPermissionsForRoles(ctx context.Context, roles []string, tenantCode string) ([]*model.Permission, error) {
	var buttonResources []string

	for _, role := range roles {
		// 使用 Casbin 获取角色的策略
		policies, _ := s.enforcer.GetFilteredPolicy(0, role, tenantCode)

		for _, policy := range policies {
			if len(policy) >= 4 {
				resource := policy[2]
				action := policy[3]

				// 对于 BUTTON 类型，resource 格式为 "btn:menuID:action"，action 是 "*"
				if (action == "*" || action == "") && strings.HasPrefix(resource, "btn:") {
					buttonResources = append(buttonResources, resource)
				}
			}
		}
	}

	if len(buttonResources) == 0 {
		return []*model.Permission{}, nil
	}

	// 根据 resource 查询权限详情
	var permissions []*model.Permission
	for _, resource := range buttonResources {
		perm, err := s.permissionRepo.GetByResource(ctx, resource)
		if err == nil {
			permissions = append(permissions, perm)
		}
	}

	return permissions, nil
}

// intersectMenuIDs 计算两个菜单ID列表的交集
func (s *UserMenuService) intersectMenuIDs(menuIDs, tenantMenuIDs []string) []string {
	tenantMenuSet := make(map[string]bool, len(tenantMenuIDs))
	for _, id := range tenantMenuIDs {
		tenantMenuSet[id] = true
	}

	var result []string
	for _, id := range menuIDs {
		if tenantMenuSet[id] {
			result = append(result, id)
		}
	}
	return result
}

// filterVisibleMenus 过滤启用状态的菜单
func (s *UserMenuService) filterVisibleMenus(menus []*model.Menu) []*model.Menu {
	var result []*model.Menu
	for _, menu := range menus {
		if menu.Status == constants.StatusEnabled {
			result = append(result, menu)
		}
	}
	return result
}

// buildMenuTree 构建菜单树
func (s *UserMenuService) buildMenuTree(menus []*model.Menu) []*dto.MenuTreeNode {
	// 构建菜单映射
	menuMap := make(map[string]*dto.MenuTreeNode)
	for _, menu := range menus {
		menuMap[menu.MenuID] = &dto.MenuTreeNode{
			MenuInfo: s.toMenuInfo(menu),
			Children: []*dto.MenuTreeNode{},
		}
	}

	// 构建树结构
	var roots []*dto.MenuTreeNode
	for _, menu := range menus {
		node := menuMap[menu.MenuID]
		if menu.ParentID == "" || menu.ParentID == "0" {
			// 根节点
			roots = append(roots, node)
		} else if parent, exists := menuMap[menu.ParentID]; exists {
			// 添加到父节点的子节点
			parent.Children = append(parent.Children, node)
		}
	}

	return roots
}

// toMenuInfo 转换为菜单信息格式
func (s *UserMenuService) toMenuInfo(menu *model.Menu) *dto.MenuInfo {
	return &dto.MenuInfo{
		PermissionID: menu.MenuID,
		Name:         menu.Name,
		Type:         "MENU",
		ParentID:     stringPtr(menu.ParentID),
		Resource:     stringPtr("menu:" + menu.MenuID),
		Action:       stringPtr("*"),
		Path:         stringPtr(menu.Path),
		Component:    stringPtr(menu.Component),
		Redirect:     stringPtr(menu.Redirect),
		Icon:         stringPtr(menu.Icon),
		Sort:         int16Ptr(menu.Sort),
		Status:       menu.Status,
		Description:  stringPtr(menu.Description),
		CreatedAt:    menu.CreatedAt,
		UpdatedAt:    menu.UpdatedAt,
	}
}

// stringPtr 辅助函数：返回字符串指针
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// int16Ptr 辅助函数：返回 int16 指针
func int16Ptr(i int32) *int16 {
	v := int16(i)
	return &v
}
