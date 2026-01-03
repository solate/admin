package dto

import "admin/pkg/pagination"

// LoginLogResponse 登录日志响应
type LoginLogResponse struct {
	LogID         string `json:"log_id"`
	TenantID      string `json:"tenant_id"`
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	LoginType     string `json:"login_type"`     // PASSWORD, SSO, OAUTH
	LoginIP       string `json:"login_ip"`
	LoginLocation string `json:"login_location"` // IP解析的地理位置
	UserAgent     string `json:"user_agent"`
	Status        int16  `json:"status"`         // 1:成功 0:失败
	FailReason    string `json:"fail_reason"`    // 失败原因
	CreatedAt     int64  `json:"created_at"`
}

// ListLoginLogsRequest 获取登录日志列表请求
type ListLoginLogsRequest struct {
	pagination.Request  `json:",inline"`
	UserID              string  `form:"user_id" binding:"omitempty"`
	UserName            string  `form:"user_name" binding:"omitempty"`
	LoginType           string  `form:"login_type" binding:"omitempty"` // PASSWORD, SSO, OAUTH
	Status              *int16  `form:"status" binding:"omitempty"`      // 1:成功 0:失败
	StartDate           *int64  `form:"start_date" binding:"omitempty"`  // 开始时间(毫秒时间戳)
	EndDate             *int64  `form:"end_date" binding:"omitempty"`    // 结束时间(毫秒时间戳)
	IPAddress           string  `form:"ip_address" binding:"omitempty"`
}

// ListLoginLogsResponse 获取登录日志列表响应
type ListLoginLogsResponse struct {
	pagination.Response `json:",inline"`
	List                []*LoginLogResponse `json:"list"` // 列表数据
}
