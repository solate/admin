package dto

import "admin/pkg/pagination"

// TenantCreateRequest 创建租户请求
type TenantCreateRequest struct {
	Code        string `json:"code" binding:"required,min=2,max=50" example:"tenant_shanghai"` // 租户编码（全局唯一）
	Name        string `json:"name" binding:"required,min=2,max=200" example:"上海分公司"`          // 租户名称
	Description string `json:"description" example:"上海地区业务运营"`                                 // 租户描述
}

// TenantUpdateRequest 更新租户请求
type TenantUpdateRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=200" example:"上海分公司"` // 租户名称
	Description string `json:"description" example:"上海地区业务运营"`                         // 租户描述
	Status      int    `json:"status" binding:"omitempty,oneof=1 2" example:"1"`       // 状态：1-正常，2-禁用
}

// TenantResponse 租户响应
type TenantResponse struct {
	TenantID    string `json:"tenant_id" example:"123e4567-e89b-12d3-a456-426614174000"` // 租户ID
	Code        string `json:"code" example:"tenant_shanghai"`                           // 租户编码
	Name        string `json:"name" example:"上海分公司"`                                     // 租户名称
	Description string `json:"description" example:"上海地区业务运营"`                           // 租户描述
	Status      int    `json:"status" example:"1"`                                       // 状态：1-正常，2-禁用
	CreatedAt   int64  `json:"created_at" example:"1703123456789"`                       // 创建时间
	UpdatedAt   int64  `json:"updated_at" example:"1703123456789"`                       // 更新时间
}

// TenantListRequest 租户列表查询请求
type TenantListRequest struct {
	pagination.Request  `json:",inline"`
	Code                string `form:"code" example:"tenant_shanghai"`                   // 租户编码（模糊查询）
	Name                string `form:"name" example:"上海"`                                // 租户名称（模糊查询）
	Status              int    `form:"status" binding:"omitempty,oneof=1 2" example:"1"` // 状态筛选（指针表示可选）
}

// TenantListResponse 租户列表响应
type TenantListResponse struct {
	pagination.Response `json:",inline"`
	List                []*TenantInfo `json:"list"` // 列表数据
}

// TenantInfo 租户基础信息（可复用）
type TenantInfo struct {
	TenantID    string `json:"tenant_id" example:"123456789012345678"`   // 租户ID
	TenantCode  string `json:"tenant_code" example:"tenant_shanghai"`   // 租户编码
	Name        string `json:"name" example:"上海分公司"`                 // 租户名称
	Description string `json:"description" example:"上海地区业务运营"`     // 租户描述
}
