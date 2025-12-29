package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/cache"
	"admin/pkg/casbin"
	"admin/pkg/constants"
	"admin/pkg/xerr"
	"context"
	"strings"
)

// UserMenuService 用户菜单服务
type UserMenuService struct {
	permissionRepo *repository.PermissionRepo
	tenantMenuRepo *repository.TenantMenuRepo
	enforcer       *casbin.Enforcer
}

// NewUserMenuService 创建用户菜单服务
func NewUserMenuService(permissionRepo *repository.PermissionRepo, tenantMenuRepo *repository.TenantMenuRepo, enforcer *casbin.Enforcer) *UserMenuService {
	return &UserMenuService{
		permissionRepo: permissionRepo,
		tenantMenuRepo: tenantMenuRepo,
		enforcer:       enforcer,
	}
}

// GetUserMenu 获取用户菜单树
func (s *UserMenuService) GetUserMenu(ctx context.Context, userName, tenantCode string) (*dto.UserMenuResponse, error) {
	// 1. 超管特殊处理：返回所有菜单
	if s.isSuperAdmin(userName, tenantCode) {
		return s.getAllMenusForTenant(ctx, tenantCode)
	}

	// 2. 获取用户在租户中的角色
	roles := s.enforcer.GetRolesForUserInDomain(userName, tenantCode)
	if len(roles) == 0 {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}

	// 3. 获取角色的所有 MENU 类型权限ID
	menuIDs, err := s.getMenuPermissionsForRoles(ctx, roles, tenantCode)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取菜单权限失败", err)
	}

	// 4. 与租户菜单边界取交集
	tenantMenuIDs, err := s.tenantMenuRepo.GetMenuIDsByTenant(ctx, cache.Get().Tenant.GetDefaultTenantID())
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取租户菜单边界失败", err)
	}

	validMenuIDs := s.intersectMenuIDs(menuIDs, tenantMenuIDs)
	if len(validMenuIDs) == 0 {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}

	// 5. 查询权限详情
	permissions, err := s.permissionRepo.GetByIDs(ctx, validMenuIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单详情失败", err)
	}

	// 6. 过滤状态=1(显示)的菜单，且只保留 MENU 类型
	visibleMenus := s.filterVisibleMenus(permissions)

	// 7. 构建树结构
	tree := s.buildPermissionTree(visibleMenus)

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

	// 3. 获取角色的所有 BUTTON 类型权限ID
	buttonIDs, err := s.getButtonPermissionsForRoles(ctx, roles, tenantCode)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取按钮权限失败", err)
	}

	// 4. 查询权限详情
	buttons, err := s.permissionRepo.GetByIDs(ctx, buttonIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询按钮详情失败", err)
	}

	// 5. 过滤出属于指定菜单的按钮（通过 resource 字段判断）
	var menuButtons []*dto.ButtonInfo
	for _, btn := range buttons {
		// 检查 resource 是否以 "btn:" 开头且属于指定菜单
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
func (s *UserMenuService) getAllMenusForTenant(ctx context.Context, tenantCode string) (*dto.UserMenuResponse, error) {
	defaultTenantID := cache.Get().Tenant.GetDefaultTenantID()

	// 获取租户分配的菜单ID
	menuIDs, err := s.tenantMenuRepo.GetMenuIDsByTenant(ctx, defaultTenantID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户菜单失败", err)
	}

	if len(menuIDs) == 0 {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}

	// 查询权限详情
	permissions, err := s.permissionRepo.GetByIDs(ctx, menuIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询权限详情失败", err)
	}

	// 过滤显示状态的菜单
	visibleMenus := s.filterVisibleMenus(permissions)

	// 构建树结构
	tree := s.buildPermissionTree(visibleMenus)

	return &dto.UserMenuResponse{List: tree}, nil
}

// getAllButtonsForMenu 获取菜单的所有按钮（超管使用）
func (s *UserMenuService) getAllButtonsForMenu(ctx context.Context, menuID string) (*dto.UserButtonsResponse, error) {
	// 获取所有 BUTTON 类型的权限
	buttons, err := s.permissionRepo.ListByType(ctx, constants.TypeButton)
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

// getMenuPermissionsForRoles 获取角色的 MENU 类型权限ID列表
func (s *UserMenuService) getMenuPermissionsForRoles(ctx context.Context, roles []string, tenantCode string) ([]string, error) {
	return s.getPermissionsForRoles(ctx, roles, tenantCode, constants.TypeMenu, "menu:")
}

// getButtonPermissionsForRoles 获取角色的 BUTTON 类型权限ID列表
func (s *UserMenuService) getButtonPermissionsForRoles(ctx context.Context, roles []string, tenantCode string) ([]string, error) {
	return s.getPermissionsForRoles(ctx, roles, tenantCode, constants.TypeButton, "btn:")
}

// getPermissionsForRoles 获取角色的指定类型权限ID列表
func (s *UserMenuService) getPermissionsForRoles(ctx context.Context, roles []string, tenantCode, permType, resourcePrefix string) ([]string, error) {
	permissionSet := make(map[string]bool)

	for _, role := range roles {
		// 使用 Casbin 获取角色的策略
		// 策略格式: p, role_code, tenant_code, resource, action
		policies, _ := s.enforcer.GetFilteredPolicy(0, role, tenantCode)

		for _, policy := range policies {
			if len(policy) >= 4 {
				resource := policy[2]
				action := policy[3]

				// 对于 MENU/BUTTON 类型，resource 格式为 "menu:xxx" 或 "btn:xxx:yyy"，action 是 "*"
				if action == "*" || action == "" {
					if strings.HasPrefix(resource, resourcePrefix) {
						// 提取权限ID（去掉前缀）
						permID := strings.TrimPrefix(resource, resourcePrefix)
						// 对于按钮，resource 可能是 "btn:menuID:action"，需要特殊处理
						if permType == constants.TypeButton {
							// 按钮的 resource 就是完整的标识符
							permissionSet[resource] = true
						} else {
							permissionSet[permID] = true
						}
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

// filterVisibleMenus 过滤显示状态的菜单，且只保留 MENU 类型
func (s *UserMenuService) filterVisibleMenus(permissions []*model.Permission) []*model.Permission {
	var result []*model.Permission
	for _, perm := range permissions {
		// 只保留 MENU 类型
		if perm.Type != constants.TypeMenu {
			continue
		}
		// Permission 模型没有 Status 字段，默认都返回
		result = append(result, perm)
	}
	return result
}

// buildPermissionTree 构建权限树
func (s *UserMenuService) buildPermissionTree(permissions []*model.Permission) []*dto.MenuTreeNode {
	// Permission 模型没有 ParentID，暂时返回平铺列表
	// TODO: 需要设计 Permission 的层级关系
	tree := make([]*dto.MenuTreeNode, 0, len(permissions))
	for _, perm := range permissions {
		tree = append(tree, &dto.MenuTreeNode{
			MenuInfo: s.toMenuInfo(perm),
			Children: []*dto.MenuTreeNode{},
		})
	}
	return tree
}

// toMenuInfo 转换为菜单信息格式
func (s *UserMenuService) toMenuInfo(perm *model.Permission) *dto.MenuInfo {
	return &dto.MenuInfo{
		PermissionID: perm.PermissionID,
		Name:         perm.Name,
		Type:         perm.Type,
		Resource:     &perm.Resource,
		Action:       &perm.Action,
		Description:  &perm.Description,
		CreatedAt:    perm.CreatedAt,
		UpdatedAt:    perm.UpdatedAt,
	}
}
