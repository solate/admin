package constants

const (
	// DefaultTenantCode 默认租户编码（超管租户，用于 Casbin domain）
	DefaultTenantCode = "default"
	// SuperAdmin 超级管理员角色code
	SuperAdmin = "super_admin"
	// Admin 租户管理员角色code
	Admin = "admin"
	// Reviewer 审核员角色code
	Reviewer = "reviewer"
	// User 普通用户角色code
	User = "user"
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
