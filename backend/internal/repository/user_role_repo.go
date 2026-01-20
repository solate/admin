package repository

import (
	"admin/pkg/casbin"
	"context"

	"github.com/rs/zerolog/log"
)

// UserRoleRepo 用户角色仓储（基于 Casbin）
type UserRoleRepo struct {
	enforcer *casbin.Enforcer
}

// NewUserRoleRepo 创建用户角色仓储
func NewUserRoleRepo(enforcer *casbin.Enforcer) *UserRoleRepo {
	return &UserRoleRepo{
		enforcer: enforcer,
	}
}

// GetUserRoles 获取用户在指定租户下的角色列表
// 返回角色编码列表 (如: ["super_admin", "sales"])
func (r *UserRoleRepo) GetUserRoles(ctx context.Context, userName, tenantCode string) ([]string, error) {
	// 从 Casbin 获取用户的角色
	// g 策略: g, username, rolecode, tenantcode
	roles := r.enforcer.GetRolesForUserInDomain(userName, tenantCode)
	return roles, nil
}

// AddUserRole 为用户添加角色（添加 g 策略）
// g, username, rolecode, tenantcode
func (r *UserRoleRepo) AddUserRole(ctx context.Context, userName, roleCode, tenantCode string) (bool, error) {
	return r.enforcer.AddRoleForUserInDomain(userName, roleCode, tenantCode)
}

// DeleteUserRole 删除用户的角色（删除 g 策略）
func (r *UserRoleRepo) DeleteUserRole(ctx context.Context, userName, roleCode, tenantCode string) (bool, error) {
	return r.enforcer.DeleteRoleForUserInDomain(userName, roleCode, tenantCode)
}

// DeleteUserRoles 删除用户在指定租户下的所有角色
func (r *UserRoleRepo) DeleteUserRoles(ctx context.Context, userName, tenantCode string) (bool, error) {
	// 获取用户所有角色
	roles, err := r.GetUserRoles(ctx, userName, tenantCode)
	if err != nil {
		return false, err
	}

	// 逐个删除
	for _, role := range roles {
		if _, err := r.enforcer.DeleteRoleForUserInDomain(userName, role, tenantCode); err != nil {
			return false, err
		}
	}

	return true, nil
}

// CheckUserRole 检查用户是否拥有指定角色
func (r *UserRoleRepo) CheckUserRole(ctx context.Context, userName, roleCode, tenantCode string) bool {
	// 直接查询用户的角色列表
	roles := r.enforcer.GetRolesForUserInDomain(userName, tenantCode)
	for _, role := range roles {
		if role == roleCode {
			return true
		}
	}
	return false
}

// GetRoleUsers 获取指定角色下的所有用户
func (r *UserRoleRepo) GetRoleUsers(ctx context.Context, roleCode, tenantCode string) ([]string, error) {
	users := r.enforcer.GetUsersForRoleInDomain(roleCode, tenantCode)
	return users, nil
}

// AssignRoles 为用户批量分配角色（覆盖式）
// 先删除用户在租户下的所有角色，再添加新角色
func (r *UserRoleRepo) AssignRoles(ctx context.Context, userName string, roleCodes []string, tenantCode string) error {
	// 1. 删除用户在该租户下的所有现有角色
	if _, err := r.DeleteUserRoles(ctx, userName, tenantCode); err != nil {
		return err
	}

	// 2. 添加新角色
	for _, roleCode := range roleCodes {
		if _, err := r.AddUserRole(ctx, userName, roleCode, tenantCode); err != nil {
			return err
		}
	}

	return nil
}

// AddRoles 为用户批量添加角色（增量式）
// 不删除现有角色，只添加新角色
func (r *UserRoleRepo) AddRoles(ctx context.Context, userName string, roleCodes []string, tenantCode string) error {
	// 获取用户现有角色
	existingRoles := r.enforcer.GetRolesForUserInDomain(userName, tenantCode)
	existingRoleMap := make(map[string]bool)
	for _, role := range existingRoles {
		existingRoleMap[role] = true
	}

	// 只添加不存在的角色
	for _, roleCode := range roleCodes {
		if !existingRoleMap[roleCode] {
			if _, err := r.AddUserRole(ctx, userName, roleCode, tenantCode); err != nil {
				return err
			}
		}
	}

	return nil
}

// RemoveRoles 为用户批量删除角色
func (r *UserRoleRepo) RemoveRoles(ctx context.Context, userName string, roleCodes []string, tenantCode string) error {
	for _, roleCode := range roleCodes {
		if _, err := r.DeleteUserRole(ctx, userName, roleCode, tenantCode); err != nil {
			return err
		}
	}
	return nil
}

// EnsureRoleInheritance 确保角色继承自 default 租户的同名角色
// 通过复制 default 租户的权限到当前租户来实现继承
func (r *UserRoleRepo) EnsureRoleInheritance(ctx context.Context, roleCode, tenantCode string) error {
	// 如果是 default 租户，不需要继承
	if tenantCode == "default" {
		return nil
	}

	// 获取 default 租户中该角色的所有权限策略
	defaultPolicies, _ := r.enforcer.GetFilteredPolicy(1, roleCode, "default")

	// 如果 default 租户没有该角色的权限，跳过
	if len(defaultPolicies) == 0 {
		return nil
	}

	// 为当前租户添加相同的权限策略
	addedCount := 0
	for _, policy := range defaultPolicies {
		// policy 格式: [roleCode, "default", resource, action]
		if len(policy) < 4 {
			continue
		}
		resource := policy[2]
		action := policy[3]

		// 检查当前租户是否已有该权限
		hasPolicy, _ := r.enforcer.HasPolicy(roleCode, tenantCode, resource, action)
		if hasPolicy {
			continue
		}

		// 添加权限到当前租户
		if _, err := r.enforcer.AddPolicy(roleCode, tenantCode, resource, action); err != nil {
			log.Error().Err(err).Str("role_code", roleCode).Str("tenant_code", tenantCode).
				Str("resource", resource).Str("action", action).Msg("添加角色权限失败")
			continue
		}
		addedCount++
	}

	if addedCount > 0 {
		log.Info().Str("role_code", roleCode).Str("tenant_code", tenantCode).
			Int("count", addedCount).Msg("角色权限继承成功")
	}

	return nil
}
