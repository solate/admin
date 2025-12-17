package constants

const (
	// DefaultTenant 默认租户名称
	DefaultTenant = "default"
	// SuperAdminRole 超级管理员角色code
	SuperAdminRole = "super_admin"
)

// roleType 管理员类型(1:普通用户, 2:租户管理员, 3:平台超级管理员)

const (
	RoleTypeUser        = 1 // 普通用户
	RoleTypeTenantAdmin = 2 // 租户管理员
	RoleTypeSuperAdmin  = 3 // 平台超级管理员
)
