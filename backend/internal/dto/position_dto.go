package dto

import "admin/pkg/pagination"

// CreatePositionRequest 创建岗位请求
type CreatePositionRequest struct {
	PositionCode string `json:"position_code" binding:"required"`     // 岗位编码（租户内唯一）
	PositionName string `json:"position_name" binding:"required"`     // 岗位名称
	Level        int    `json:"level" binding:"omitempty"`            // 职级
	Description  string `json:"description" binding:"omitempty"`      // 岗位描述
	Sort         int    `json:"sort" binding:"omitempty"`             // 排序权重
	Status       int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
}

// UpdatePositionRequest 更新岗位请求
type UpdatePositionRequest struct {
	PositionID   string `json:"position_id" form:"position_id" binding:"required" example:"123456789012345678"` // 岗位ID
	PositionCode string `json:"position_code" binding:"omitempty"`                                              // 岗位编码
	PositionName string `json:"position_name" binding:"omitempty"`                                              // 岗位名称
	Level        int    `json:"level" binding:"omitempty"`                                                      // 职级
	Description  string `json:"description" binding:"omitempty"`                                                // 岗位描述
	Sort         int    `json:"sort" binding:"omitempty"`                                                       // 排序权重
	Status       int    `json:"status" binding:"omitempty,oneof=1 2"`                                           // 状态 1:启用 2:禁用
}

// PositionInfo 岗位基础信息（可复用）
type PositionInfo struct {
	PositionID   string `json:"position_id" example:"123456789012345678"` // 岗位ID
	PositionCode string `json:"position_code" example:"SENIOR_ENGINEER"`  // 岗位编码
	PositionName string `json:"position_name" example:"高级工程师"`            // 岗位名称
	Level        int    `json:"level" example:"7"`                        // 职级
	Description  string `json:"description" example:"负责核心技术研发工作"`         // 岗位描述
	Sort         int    `json:"sort" example:"1"`                         // 排序权重
	Status       int    `json:"status" example:"1" enum:"1,2"`            // 状态 1:启用 2:禁用
	CreatedAt    int64  `json:"created_at" example:"1735200000"`          // 创建时间
	UpdatedAt    int64  `json:"updated_at" example:"1735206400"`          // 更新时间
}

// ListPositionsRequest 岗位列表请求
type ListPositionsRequest struct {
	pagination.Request `json:",inline"`
	PositionName       string `form:"position_name" binding:"omitempty,max=50"` // 岗位名称（可选，模糊匹配）
	PositionCode       string `form:"position_code" binding:"omitempty,max=50"` // 岗位编码（可选，模糊匹配）
	Status             int    `form:"status" binding:"omitempty,oneof=1 2"`     // 状态筛选
}

// ListPositionsResponse 岗位列表响应
type ListPositionsResponse struct {
	pagination.Response `json:",inline"`
	List                []*PositionInfo `json:"list"` // 列表数据
}

// PositionDetailRequest 获取岗位详情请求
type PositionDetailRequest struct {
	PositionID string `json:"position_id" form:"position_id" binding:"required" example:"123456789012345678"` // 岗位ID
}

// PositionDeleteRequest 删除岗位请求
type PositionDeleteRequest struct {
	PositionID string `json:"position_id" form:"position_id" binding:"required" example:"123456789012345678"` // 岗位ID
}

// PositionStatusRequest 更新岗位状态请求
type PositionStatusRequest struct {
	PositionID string `json:"position_id" binding:"required" example:"123456789012345678"` // 岗位ID
	Status     int    `json:"status" binding:"required" example:"1"`                       // 状态值
}

// PositionBatchDeleteRequest 批量删除岗位请求
type PositionBatchDeleteRequest struct {
	PositionIDs []string `json:"position_ids" binding:"required,min=1,dive,required" example:"[\"123456789012345678\", \"987654321098765432\"]"` // 岗位ID列表
}
