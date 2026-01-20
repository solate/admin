package audit

// LogEntry 日志条目（内部使用）
type LogEntry struct {
	TenantID      string
	UserID        string
	UserName      string
	Module        string
	OperationType string
	ResourceType  string
	ResourceID    string // 单个资源ID或批量操作的JSON数组
	ResourceName  string // 单个资源名称或批量操作的汇总描述
	OldValue      any    // 任意类型，写入时转为 JSON 字符串
	NewValue      any    // 任意类型，写入时转为 JSON 字符串
	RequestMethod string
	RequestPath   string
	RequestParams string
	IPAddress     string
	UserAgent     string
	Status        int16
	ErrorMessage  string
	CreatedAt     int64
}
