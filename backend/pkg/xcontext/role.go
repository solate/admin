package xcontext

import (
	"context"
)

const (
	// 角色相关
	RolesKey contextKey = "roles"
)

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

// HasRole 判断是否拥有指定角色
func HasRole(ctx context.Context, roleCode string) bool {
	roles := GetRoles(ctx)
	for _, role := range roles {
		if role == roleCode {
			return true
		}
	}
	return false
}
