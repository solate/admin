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
	LogID         string `json:"log_id"`         // 日志ID
	TenantID      string `json:"tenant_id"`      // 租户ID
	UserID        string `json:"user_id"`        // 用户ID
	UserName      string `json:"user_name"`      // 用户名
	Module        string `json:"module"`         // 模块名
	OperationType string `json:"operation_type"` // 操作类型
	ResourceType  string `json:"resource_type"`  // 资源类型
	ResourceID    string `json:"resource_id"`    // 资源ID
	ResourceName  string `json:"resource_name"`  // 资源名称
	RequestMethod string `json:"request_method"` // 请求方法
	RequestPath   string `json:"request_path"`   // 请求路径
	RequestParams string `json:"request_params"` // 请求参数
	OldValue      string `json:"old_value"`      // 操作前数据
	NewValue      string `json:"new_value"`      // 操作后数据
	Status        int    `json:"status"`         // 操作状态
	ErrorMessage  string `json:"error_message"`  // 错误信息
	IPAddress     string `json:"ip_address"`     // IP地址
	UserAgent     string `json:"user_agent"`     // 用户代理
	CreatedAt     int64  `json:"created_at"`     // 创建时间
}

// ListOperationLogsResponse 操作日志列表响应
type ListOperationLogsResponse struct {
	pagination.Response `json:",inline"`
	List                []*OperationLogResponse `json:"list"` // 列表数据
}
