package dto

import "admin/pkg/pagination"

// TenantCreateRequest 创建租户请求
type TenantCreateRequest struct {
	TenantCode   string `json:"tenant_code" binding:"required,min=2,max=50" example:"tenant_shanghai"` // 租户编码（全局唯一）
	Name         string `json:"name" binding:"required,min=2,max=200" example:"上海分公司"`                 // 租户名称
	Description  string `json:"description" example:"上海地区业务运营"`                                        // 租户描述
	ContactName  string `json:"contact_name" binding:"required,max=100" example:"张三"`                  // 联系人姓名
	ContactPhone string `json:"contact_phone" binding:"required,max=20" example:"13800138000"`         // 联系人手机号
}

// TenantUpdateRequest 更新租户请求
type TenantUpdateRequest struct {
	TenantID     string `json:"tenant_id" form:"tenant_id" binding:"required" example:"123456789012345678"` // 租户ID
	Name         string `json:"name" binding:"omitempty,min=2,max=200" example:"上海分公司"`                     // 租户名称
	Description  string `json:"description" example:"上海地区业务运营"`                                             // 租户描述
	ContactName  string `json:"contact_name" binding:"omitempty,max=100" example:"张三"`                      // 联系人姓名
	ContactPhone string `json:"contact_phone" binding:"omitempty,max=20" example:"13800138000"`             // 联系人手机号
	Status       int    `json:"status" binding:"omitempty,oneof=1 2" example:"1"`                           // 状态：1-正常，2-禁用
}

// TenantDetailRequest 获取租户详情请求
type TenantDetailRequest struct {
	TenantID string `json:"tenant_id" form:"tenant_id" binding:"required" example:"123456789012345678"` // 租户ID
}

// TenantDeleteRequest 删除租户请求
type TenantDeleteRequest struct {
	TenantID string `json:"tenant_id" form:"tenant_id" binding:"required" example:"123456789012345678"` // 租户ID
}

// TenantStatusRequest 更新租户状态请求
type TenantStatusRequest struct {
	TenantID string `json:"tenant_id" binding:"required" example:"123456789012345678"` // 租户ID
	Status   int    `json:"status" binding:"required" example:"1"`                     // 状态值
}

// TenantBatchDeleteRequest 批量删除租户请求
type TenantBatchDeleteRequest struct {
	TenantIDs []string `json:"tenant_ids" binding:"required,min=1,dive,required" example:"[\"123456789012345678\", \"987654321098765432\"]"` // 租户ID列表
}

// TenantListRequest 租户列表查询请求
type TenantListRequest struct {
	pagination.Request `json:",inline"`
	TenantCode         string `form:"tenant_code" example:"tenant_shanghai"`            // 租户编码（模糊查询）
	Name               string `form:"name" example:"上海"`                                // 租户名称（模糊查询）
	Status             int    `form:"status" binding:"omitempty,oneof=1 2" example:"1"` // 状态筛选（指针表示可选）
}

// TenantListResponse 租户列表响应
type TenantListResponse struct {
	pagination.Response `json:",inline"`
	List                []*TenantInfo `json:"list"` // 列表数据
}

// TenantInfo 租户信息（可复用）
type TenantInfo struct {
	TenantID     string `json:"tenant_id" example:"123456789012345678"` // 租户ID
	TenantCode   string `json:"tenant_code" example:"tenant_shanghai"`  // 租户编码
	Name         string `json:"name" example:"上海分公司"`                   // 租户名称
	Description  string `json:"description" example:"上海地区业务运营"`         // 租户描述
	ContactName  string `json:"contact_name" example:"张三"`              // 联系人姓名
	ContactPhone string `json:"contact_phone" example:"13800138000"`    // 联系人手机号
	Status       int    `json:"status" example:"1"`                     // 状态：1-正常，2-禁用
	CreatedAt    int64  `json:"created_at" example:"1703123456789"`     // 创建时间
	UpdatedAt    int64  `json:"updated_at" example:"1703123456789"`     // 更新时间
}
