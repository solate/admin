package constants

// 模块名称常量
const (
	ModuleUser       = "user"       // 用户管理
	ModuleRole       = "role"       // 角色管理
	ModulePermission = "permission" // 权限管理
	ModuleTenant     = "tenant"     // 租户管理
	ModuleSystem     = "system"     // 系统设置
	ModuleMenu       = "menu"       // 菜单管理
	ModuleDict       = "dict"       // 字典管理
	ModuleLog        = "log"        // 日志管理
	ModuleFile       = "file"       // 文件管理
	ModuleAuth       = "auth"       // 认证相关 (登录、登出等)
	ModuleDept       = "dept"       // 部门管理
	ModulePosition   = "position"   // 岗位管理
)

// 资源类型常量（用于操作日志记录）
const (
	ResourceTypeUser       = "user"       // 用户资源
	ResourceTypeRole       = "role"       // 角色资源
	ResourceTypePermission = "permission" // 权限资源
	ResourceTypeTenant     = "tenant"     // 租户资源
	ResourceTypeMenu       = "menu"       // 菜单资源
	ResourceTypeDict       = "dict"       // 字典资源
	ResourceTypeDept       = "dept"       // 部门资源
	ResourceTypePosition   = "position"   // 岗位资源
)

// 操作类型常量
const (
	OperationCreate = "CREATE" // 创建
	OperationUpdate = "UPDATE" // 更新
	OperationDelete = "DELETE" // 删除
	OperationQuery  = "QUERY"  // 查询
	OperationExport = "EXPORT" // 导出
	OperationImport = "IMPORT" // 导入
	OperationLogin  = "LOGIN"  // 登录
	OperationLogout = "LOGOUT" // 登出
)

// 操作状态常量
const (
	OperationStatusSuccess = 1 // 成功
	OperationStatusFailed  = 2 // 失败
)

// OperationTypeText 操作类型中文描述映射
var OperationTypeText = map[string]string{
	OperationCreate: "创建",
	OperationUpdate: "更新",
	OperationDelete: "删除",
	OperationQuery:  "查询",
	OperationExport: "导出",
	OperationImport: "导入",
	OperationLogin:  "登录",
	OperationLogout: "登出",
}

// ModuleText 模块名称中文描述映射
var ModuleText = map[string]string{
	ModuleUser:       "用户管理",
	ModuleRole:       "角色管理",
	ModulePermission: "权限管理",
	ModuleTenant:     "租户管理",
	ModuleSystem:     "系统设置",
	ModuleMenu:       "菜单管理",
	ModuleDict:       "字典管理",
	ModuleLog:        "日志管理",
	ModuleFile:       "文件管理",
	ModuleAuth:       "认证管理",
	ModuleDept:       "部门管理",
	ModulePosition:   "岗位管理",
}
