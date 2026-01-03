package constants

const (
	// DefaultTenantCode 默认租户编码（超管租户，用于 Casbin domain）
	DefaultTenantCode = "default"
	// SuperAdmin 超级管理员角色code
	SuperAdmin = "super_admin"
	// Admin 租户管理员角色code
	Admin = "admin"
	// Auditor 审计员/监管员角色code
	Auditor = "auditor"
	// User 普通用户角色code
	User = "user"
)

// 租户管理员权限范围定义
// 说明：租户管理员（Admin 角色）在租户内的权限边界
const (
	// 用户管理权限
	TenantAdminCanViewUsers   = true // 查看租户用户列表
	TenantAdminCanCreateUsers = true // 创建新用户
	TenantAdminCanUpdateUsers = true // 更新用户信息
	TenantAdminCanDeleteUsers = true // 删除用户
	TenantAdminCanAssignRoles = true // 分配角色给用户

	// 角色管理权限
	TenantAdminCanViewRoles   = true  // 查看角色列表
	TenantAdminCanCreateRoles = true  // 创建新角色（必须继承模板）
	TenantAdminCanDeleteRoles = false // 不能删除角色（防止误删系统角色）
	TenantAdminCanUpdateRoles = true  // 更新角色基本信息（名称、描述）

	// 权限管理
	TenantAdminCanModifyInheritedPermissions = false // 不能修改继承的权限（模板权限）
	TenantAdminCanAddCustomPermissions       = true  // 可以添加额外权限
	TenantAdminCanRemoveCustomPermissions    = true  // 可以移除额外权限
	TenantAdminCanAssignPermissions          = true  // 可以为角色分配权限

	// 菜单管理
	TenantAdminCanViewMenus              = true  // 查看菜单列表
	TenantAdminCannotModifySystemMenus   = true  // 不能修改系统菜单（default 租户的菜单）
	TenantAdminCanCreateCustomMenus      = true  // 可以创建自定义菜单
	TenantAdminCanUpdateCustomMenus      = true  // 可以更新自定义菜单
	TenantAdminCanDeleteCustomMenus      = true  // 可以删除自定义菜单
)

// 权限类型常量
const (
	TypeMenu   = "MENU"   // 菜单权限
	TypeButton = "BUTTON" // 按钮权限
	TypeAPI    = "API"    // 接口权限
	TypeData   = "DATA"   // 数据权限
)

// 菜单来源类型常量
const (
	MenuSourceTypeSystem = "SYSTEM" // 系统模板
	MenuSourceTypeCustom = "CUSTOM" // 租户自定义
)

// 菜单状态常量
const (
	MenuStatusShow   = 1 // 显示
	MenuStatusHidden = 2 // 隐藏
)

// PermissionTypeText 权限类型中文描述映射
var PermissionTypeText = map[string]string{
	TypeMenu:   "菜单",
	TypeButton: "按钮",
	TypeAPI:    "接口",
	TypeData:   "数据",
}

// MenuStatusText 菜单状态中文描述映射
var MenuStatusText = map[int16]string{
	MenuStatusShow:   "显示",
	MenuStatusHidden: "隐藏",
}
