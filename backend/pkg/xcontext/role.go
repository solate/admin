package xcontext

import "context"

const (
	// 角色相关
	RoleTypeKey contextKey = "role_type"
	RolesKey    contextKey = "roles"
)

// RoleContext 角色上下文信息
type RoleContext struct {
	RoleType int32
	Roles    []string
}

// SetRoleType 设置角色类型到context
func SetRoleType(ctx context.Context, roleType int32) context.Context {
	return context.WithValue(ctx, RoleTypeKey, roleType)
}

// GetRoleType 从context获取角色类型，如果不存在返回0
func GetRoleType(ctx context.Context) int32 {
	value := ctx.Value(RoleTypeKey)
	if value == nil {
		return 0
	}
	roleType, _ := value.(int32)
	return roleType
}

// SetRoles 设置角色列表到context
func SetRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesKey, roles)
}

// GetRoles 从context获取角色列表，如果不存在返回nil
func GetRoles(ctx context.Context) []string {
	value := ctx.Value(RolesKey)
	if value == nil {
		return nil
	}
	roles, _ := value.([]string)
	return roles
}
