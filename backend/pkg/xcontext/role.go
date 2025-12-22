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

// GetRoleType 从context获取角色类型
func GetRoleType(ctx context.Context) (int32, bool) {
	value := ctx.Value(RoleTypeKey)
	if value == nil {
		return 0, false
	}
	roleType, ok := value.(int32)
	return roleType, ok
}

// SetRoles 设置角色列表到context
func SetRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesKey, roles)
}

// GetRoles 从context获取角色列表
func GetRoles(ctx context.Context) ([]string, bool) {
	value := ctx.Value(RolesKey)
	if value == nil {
		return nil, false
	}
	roles, ok := value.([]string)
	return roles, ok
}
