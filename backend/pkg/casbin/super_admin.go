package casbin

import (
	"admin/pkg/constants"
	"github.com/casbin/casbin/v2"
)

// IsSuperAdmin 检查是否为超管（使用默认租户）
func IsSuperAdmin(enforcer *casbin.Enforcer, userID, tenantID string) bool {
	if tenantID != constants.DefaultTenant {
		return false
	}
	roles, err := enforcer.GetRolesForUser(userID, tenantID)
	if err != nil {
		return false
	}
	for _, role := range roles {
		if role == constants.SuperAdmin {
			return true
		}
	}
	return false
}

// AddSuperAdminPolicy 添加超管全局权限策略
// 这为超管用户添加通配符权限，允许访问所有资源
func AddSuperAdminPolicy(enforcer *casbin.Enforcer, userID string) error {
	// p, super_admin, default, *, *
	_, err := enforcer.AddPolicy(constants.SuperAdmin, constants.DefaultTenant, "*", "*")
	if err != nil {
		return err
	}
	// g, user_id, super_admin, default
	_, err = enforcer.AddRoleForUser(userID, constants.SuperAdmin, constants.DefaultTenant)
	return err
}

// RemoveSuperAdminPolicy 移除超管全局权限策略
func RemoveSuperAdminPolicy(enforcer *casbin.Enforcer, userID string) error {
	// 先移除用户角色关联
	_, err := enforcer.DeleteRoleForUser(userID, constants.SuperAdmin, constants.DefaultTenant)
	if err != nil {
		return err
	}
	// 移除超管策略
	_, err = enforcer.RemovePolicy(constants.SuperAdmin, constants.DefaultTenant, "*", "*")
	return err
}
