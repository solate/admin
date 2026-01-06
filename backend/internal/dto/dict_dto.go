package dto

import "admin/pkg/pagination"

// CreateSystemDictRequest 创建系统字典请求（超管专用）
type CreateSystemDictRequest struct {
	TypeCode    string                `json:"type_code" binding:"required"` // 字典编码（如：order_status）
	TypeName    string                `json:"type_name" binding:"required"` // 字典名称（如：订单状态）
	Description string                `json:"description" binding:"omitempty"` // 字典描述
	Items       []CreateDictItemRequest `json:"items" binding:"required,min=1"` // 字典项列表
}

// CreateDictItemRequest 创建字典项请求
type CreateDictItemRequest struct {
	Label string `json:"label" binding:"required"` // 显示文本
	Value string `json:"value" binding:"required"` // 实际值
	Sort  int    `json:"sort" binding:"omitempty"` // 排序
}

// UpdateSystemDictRequest 更新系统字典请求
type UpdateSystemDictRequest struct {
	TypeName    string                `json:"type_name" binding:"omitempty"`    // 字典名称
	Description string                `json:"description" binding:"omitempty"` // 字典描述
	Items       []CreateDictItemRequest `json:"items" binding:"omitempty"`      // 字典项列表（更新时替换所有项）
}

// UpdateDictItemRequest 更新字典项请求（租户覆盖）
type UpdateDictItemRequest struct {
	Label string `json:"label" binding:"required"` // 新的显示文本
	Value string `json:"value" binding:"required"` // 实际值（用于匹配字典项）
	Sort  int    `json:"sort" binding:"omitempty"` // 排序
}

// BatchUpdateDictItemsRequest 批量更新字典项请求
type BatchUpdateDictItemsRequest struct {
	TypeCode string                      `json:"type_code" binding:"required"` // 字典编码
	Items    []UpdateDictItemRequest     `json:"items" binding:"required,min=1"` // 字典项列表
}

// DictItemResponse 字典项响应
type DictItemResponse struct {
	ItemID string `json:"item_id" example:"123456789012345678"` // 字典项ID
	Label  string `json:"label" example:"启用"`   // 显示文本
	Value  string `json:"value" example:"1"`   // 实际值
	Sort   int    `json:"sort" example:"1"`    // 排序
	Source string `json:"source" example:"system"`  // 来源（system:系统默认, custom:租户覆盖）
}

// DictResponse 字典响应
type DictResponse struct {
	TypeID   string                 `json:"type_id" example:"123456789012345678"`   // 字典类型ID
	TypeCode string                 `json:"type_code" example:"status"` // 字典编码
	TypeName string                 `json:"type_name" example:"状态"` // 字典名称
	Items    []*DictItemResponse `json:"items"`    // 字典项列表（已合并系统+覆盖）
}

// DictTypeResponse 字典类型响应
type DictTypeResponse struct {
	TypeID      string `json:"type_id" example:"123456789012345678"`      // 字典类型ID
	TenantID    string `json:"tenant_id" example:"123456789012345678"`    // 租户ID
	TypeCode    string `json:"type_code" example:"status"`    // 字典编码
	TypeName    string `json:"type_name" example:"状态"`    // 字典名称
	Description string `json:"description" example:"通用状态字典"` // 字典描述
	CreatedAt   int64  `json:"created_at" example:"1735200000"`  // 创建时间
	UpdatedAt   int64  `json:"updated_at" example:"1735206400"`  // 更新时间
}

// ListDictTypesRequest 字典类型列表请求
type ListDictTypesRequest struct {
	pagination.Request `json:",inline"`
	Keyword            string `form:"keyword" binding:"omitempty"` // 关键词搜索（字典名称/编码）
}

// ListDictTypesResponse 字典类型列表响应
type ListDictTypesResponse struct {
	pagination.Response `json:",inline"`
	List                []*DictTypeInfo `json:"list"` // 列表数据
}

// DictTypeInfo 字典类型基础信息（可复用）
type DictTypeInfo struct {
	TypeID   string `json:"type_id" example:"123456789012345678"`   // 字典类型ID
	TypeCode string `json:"type_code" example:"status"` // 字典编码
	TypeName string `json:"type_name" example:"状态"`    // 字典名称
}
