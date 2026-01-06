package dto

import "admin/pkg/pagination"

// ListOperationLogsRequest 操作日志列表请求
type ListOperationLogsRequest struct {
	pagination.Request  `json:",inline"`
	Module              string `form:"module" binding:"omitempty"`           // 模块筛选
	OperationType       string `form:"operation_type" binding:"omitempty"`   // 操作类型筛选
	ResourceType        string `form:"resource_type" binding:"omitempty"`    // 资源类型筛选
	Status              int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
	UserName            string `form:"user_name" binding:"omitempty"`        // 用户名筛选
	StartDate           int64  `form:"start_date" binding:"omitempty"`       // 开始时间
	EndDate             int64  `form:"end_date" binding:"omitempty"`         // 结束时间
}

// OperationLogResponse 操作日志响应
type OperationLogResponse struct {
	LogID         string `json:"log_id" example:"123456789012345678"`         // 日志ID
	TenantID      string `json:"tenant_id" example:"123456789012345678"`      // 租户ID
	UserID        string `json:"user_id" example:"123456789012345678"`        // 用户ID
	UserName      string `json:"user_name" example:"admin"`      // 用户名
	Module        string `json:"module" example:"用户管理"`         // 模块名
	OperationType string `json:"operation_type" example:"CREATE"` // 操作类型
	ResourceType  string `json:"resource_type" example:"用户"`  // 资源类型
	ResourceID    string `json:"resource_id" example:"123456789012345678"`    // 资源ID
	ResourceName  string `json:"resource_name" example:"admin"`  // 资源名称
	RequestMethod string `json:"request_method" example:"POST"` // 请求方法
	RequestPath   string `json:"request_path" example:"/api/v1/users"`   // 请求路径
	RequestParams string `json:"request_params" example:"{\"username\":\"admin\"}"` // 请求参数
	OldValue      string `json:"old_value" example:""`      // 操作前数据
	NewValue      string `json:"new_value" example:"{\"id\":\"123\",\"username\":\"admin\"}"`      // 操作后数据
	Status        int    `json:"status" example:"1"`         // 操作状态
	ErrorMessage  string `json:"error_message" example:""`  // 错误信息
	IPAddress     string `json:"ip_address" example:"192.168.1.100"`     // IP地址
	UserAgent     string `json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"`     // 用户代理
	CreatedAt     int64  `json:"created_at" example:"1735206400"`     // 创建时间
}

// ListOperationLogsResponse 操作日志列表响应
type ListOperationLogsResponse struct {
	pagination.Response `json:",inline"`
	List                []*OperationLogInfo `json:"list"` // 列表数据
}

// OperationLogInfo 操作日志基础信息（可复用）
type OperationLogInfo struct {
	LogID         string `json:"log_id" example:"123456789012345678"`         // 日志ID
	UserName      string `json:"user_name" example:"admin"`      // 用户名
	Module        string `json:"module" example:"用户管理"`         // 模块名
	OperationType string `json:"operation_type" example:"CREATE"` // 操作类型
	ResourceType  string `json:"resource_type" example:"用户"`  // 资源类型
	ResourceName  string `json:"resource_name" example:"admin"`  // 资源名称
	Status        int    `json:"status" example:"1"`         // 操作状态
	CreatedAt     int64  `json:"created_at" example:"1735206400"`     // 创建时间
}
