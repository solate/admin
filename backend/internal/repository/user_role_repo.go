package repository

import (
	"admin/pkg/casbin"
	"context"
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
