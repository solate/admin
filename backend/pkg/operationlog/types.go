package operationlog

// LogContext 日志上下文
type LogContext struct {
	TenantID      string
	Module        string
	OperationType string // LOGIN, LOGOUT, CREATE, UPDATE, DELETE, QUERY
	ResourceType  string
	ResourceID    string
	ResourceName  string
	OldValue      any
	NewValue      any
	Status        int16  // 1:成功 2:失败
	ErrorMessage  string
	CreatedAt     int64
}

// LogEntry 日志条目
type LogEntry struct {
	TenantID        string
	UserID          string
	UserName        string
	UserDisplayName string
	RequestMethod   string
	RequestPath     string
	RequestParams   string
	IPAddress       string
	UserAgent       string
	LogContext      *LogContext
}
