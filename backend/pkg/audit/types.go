package audit

// OperationType 操作类型常量
const (
	OperationLogin  = "LOGIN"
	OperationLogout = "LOGOUT"
	OperationCreate = "CREATE"
	OperationUpdate = "UPDATE"
	OperationDelete = "DELETE"
	OperationQuery  = "QUERY"
	OperationExport = "EXPORT"
)

// Module 模块常量
const (
	ModuleAuth       = "auth"
	ModuleUser       = "user"
	ModuleRole       = "role"
	ModuleMenu       = "menu"
	ModuleTenant     = "tenant"
	ModuleDepartment = "department"
	ModulePosition   = "position"
	ModuleDict       = "dict"
	ModuleSystem     = "system"
)

// ResourceType 资源类型常量
const (
	ResourceTenant     = "tenant"
	ResourceUser       = "user"
	ResourceRole       = "role"
	ResourceMenu       = "menu"
	ResourceDepartment = "department"
	ResourcePosition   = "position"
	ResourceDict       = "dict"
	ResourceDictItem   = "dict_item"
)

// LogStatus 日志状态常量
const (
	StatusSuccess = 1
	StatusFailure = 2
)

// LogEntry 日志条目（内部使用）
type LogEntry struct {
	TenantID      string
	UserID        string
	UserName      string
	Module        string
	OperationType string
	ResourceType  string
	ResourceID    string
	ResourceName  string
	OldValue      any // 任意类型，写入时转为 JSON 字符串
	NewValue      any // 任意类型，写入时转为 JSON 字符串
	RequestMethod string
	RequestPath   string
	RequestParams string
	IPAddress     string
	UserAgent     string
	Status        int16
	ErrorMessage  string
	CreatedAt     int64
}
