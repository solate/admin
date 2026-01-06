package dto

import "admin/pkg/pagination"

// LoginLogResponse 登录日志响应
type LoginLogResponse struct {
	LogID         string `json:"log_id" example:"123456789012345678"`
	TenantID      string `json:"tenant_id" example:"123456789012345678"`
	UserID        string `json:"user_id" example:"123456789012345678"`
	UserName      string `json:"user_name" example:"admin"`
	OperationType string `json:"operation_type" example:"LOGIN"` // LOGIN:登录, LOGOUT:登出
	LoginType     string `json:"login_type" example:"PASSWORD"`     // PASSWORD:密码, SSO:单点登录, OAUTH:第三方登录
	LoginIP       string `json:"login_ip" example:"192.168.1.100"`
	LoginLocation string `json:"login_location" example:"北京市朝阳区"` // IP解析的地理位置
	UserAgent     string `json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"`
	Status        int16  `json:"status" example:"1"`         // 1:成功 0:失败
	FailReason    string `json:"fail_reason" example:""`    // 失败原因
	CreatedAt     int64  `json:"created_at" example:"1735206400"`
}

// ListLoginLogsRequest 获取登录日志列表请求
type ListLoginLogsRequest struct {
	pagination.Request  `json:",inline"`
	UserID              string  `form:"user_id" binding:"omitempty"`
	UserName            string  `form:"user_name" binding:"omitempty"`
	OperationType       string  `form:"operation_type" binding:"omitempty"` // LOGIN:登录, LOGOUT:登出
	LoginType           string  `form:"login_type" binding:"omitempty"`     // PASSWORD:密码, SSO:单点登录, OAUTH:第三方登录
	Status              *int16  `form:"status" binding:"omitempty"`         // 1:成功 0:失败
	StartDate           *int64  `form:"start_date" binding:"omitempty"`     // 开始时间(毫秒时间戳)
	EndDate             *int64  `form:"end_date" binding:"omitempty"`       // 结束时间(毫秒时间戳)
	IPAddress           string  `form:"ip_address" binding:"omitempty"`
}

// ListLoginLogsResponse 获取登录日志列表响应
type ListLoginLogsResponse struct {
	pagination.Response `json:",inline"`
	List                []*LoginLogResponse `json:"list"` // 列表数据
}
