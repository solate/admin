package role

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/constants"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// AssignMenus 为角色分配菜单（已弃用，保留向后兼容）
// Deprecated: 使用 AssignPermissions 代替
func (s *Service) AssignMenus(ctx context.Context, roleID string, menuIDs []string) error {
	req := &dto.AssignPermissionsRequest{
		MenuPermIDs: menuIDs,
	}
	return s.AssignPermissions(ctx, roleID, req)
}

// GetRoleMenus 获取角色的菜单ID列表（已弃用，保留向后兼容）
// Deprecated: 使用 GetRolePermissions 代替
func (s *Service) GetRoleMenus(ctx context.Context, roleID string) ([]string, error) {
	permissions, err := s.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return permissions.MenuPermIDs, nil
}

// AssignPermissions 为角色分配权限（菜单+按钮）
// 权限存储在 role_permissions 表中，通过 PermissionCache 实时计算
//
// 重要变更：分配菜单权限时，会自动关联该菜单的 API 权限
// 这解决了"菜单权限粒度"问题：前端隐藏菜单时，后端 API 也会被拦截
func (s *Service) AssignPermissions(ctx context.Context, roleID string, req *dto.AssignPermissionsRequest) error {
	// 1. 查询角色信息
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	tenantID := xcontext.GetTenantID(ctx)

	// 2. 清除角色的所有现有权限（role_permissions 表）
	if err := s.rolePermRepo.DeleteByRole(ctx, roleID, tenantID); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("清除旧权限失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "清除旧权限失败", err)
	}

	// 3. 收集所有需要添加的权限ID
	var allPermIDs []string

	// 3.1 添加菜单权限
	for _, menuID := range req.MenuPermIDs {
		// 菜单权限的 permission_id 可以是 "menu:<menuID>" 形式
		// 或者直接使用 menuID 查找对应的权限记录
		allPermIDs = append(allPermIDs, menuID)
	}

	// 3.2 添加按钮权限
	for _, buttonID := range req.ButtonPermIDs {
		allPermIDs = append(allPermIDs, buttonID)
	}

	// 4. 批量添加权限到 role_permissions 表
	if len(allPermIDs) > 0 {
		items := make([]*model.RolePermission, 0, len(allPermIDs))
		for _, permID := range allPermIDs {
			items = append(items, &model.RolePermission{
				RoleID:       role.RoleID,
				PermissionID: permID,
				TenantID:     tenantID,
			})
		}
		if err := s.rolePermRepo.AddPermissions(ctx, items); err != nil {
			log.Error().Err(err).Str("role_id", roleID).Int("count", len(items)).Msg("添加权限失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "添加权限失败", err)
		}
	}

	// 5. 通知权限缓存刷新
	s.cache.NotifyRefresh()

	return nil
}

// GetRolePermissions 获取角色的所有权限（菜单+按钮）
func (s *Service) GetRolePermissions(ctx context.Context, roleID string) (*dto.RolePermissionsResponse, error) {
	// 1. 查询角色信息
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return nil, xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	tenantID := xcontext.GetTenantID(ctx)

	// 2. 从 role_permissions 表获取角色的所有权限ID
	permIDs, err := s.rolePermRepo.GetPermissionIDsByRole(ctx, role.RoleID, tenantID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色权限失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色权限失败", err)
	}

	// 3. 根据权限ID查询权限详情，区分菜单和按钮
	var menuPermIDs []string
	var buttonPermIDs []string

	if len(permIDs) > 0 {
		perms, err := s.permissionRepo.GetByIDs(ctx, permIDs)
		if err != nil {
			log.Error().Err(err).Str("role_id", roleID).Msg("查询权限详情失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询权限详情失败", err)
		}

		for _, perm := range perms {
			switch perm.Type {
			case constants.TypeMenu:
				menuPermIDs = append(menuPermIDs, perm.PermissionID)
			case constants.TypeButton:
				buttonPermIDs = append(buttonPermIDs, perm.PermissionID)
			}
		}
	}

	return &dto.RolePermissionsResponse{
		MenuPermIDs:   menuPermIDs,
		ButtonPermIDs: buttonPermIDs,
	}, nil
}

// getDefaultTenantID 获取 default 租户ID
func (s *Service) getDefaultTenantID(ctx context.Context) (string, error) {
	tenant, err := s.tenantRepo.GetByCode(ctx, constants.DefaultTenantCode)
	if err != nil {
		return "", xerr.Wrap(xerr.ErrInternal.Code, "查询默认租户失败", err)
	}
	return tenant.TenantID, nil
}

// filterSuperAdminRoles 过滤超级管理员角色（临时方案）
// 当调用者不是超级管理员时，过滤掉 super_admin 角色
// 判断依据：role_code = super_admin
func (s *Service) filterSuperAdminRoles(ctx context.Context, roles []*model.Role) []*model.Role {
	// 如果是超级管理员，返回所有角色
	if xcontext.HasRole(ctx, constants.SuperAdmin) {
		return roles
	}

	// 非超级管理员，过滤掉 super_admin 角色
	filtered := make([]*model.Role, 0, len(roles))
	for _, role := range roles {
		// 跳过 super_admin 角色
		if role.RoleCode == constants.SuperAdmin {
			continue
		}
		filtered = append(filtered, role)
	}

	return filtered
}
