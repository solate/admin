package dto

import "admin/pkg/pagination"

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	ParentID       string `json:"parent_id" binding:"omitempty"`        // 父部门ID（根部门为空字符串）
	DepartmentName string `json:"department_name" binding:"required"`   // 部门名称
	Description    string `json:"description" binding:"omitempty"`      // 部门描述
	Sort           int    `json:"sort" binding:"omitempty"`             // 排序权重
	Status         int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
}

// UpdateDepartmentRequest 更新部门请求
type UpdateDepartmentRequest struct {
	DepartmentID   string `json:"department_id" form:"department_id" binding:"required" example:"123456789012345678"` // 部门ID
	ParentID       string `json:"parent_id" binding:"omitempty"`                                                      // 父部门ID
	DepartmentName string `json:"department_name" binding:"omitempty"`                                                // 部门名称
	Description    string `json:"description" binding:"omitempty"`                                                    // 部门描述
	Sort           int    `json:"sort" binding:"omitempty"`                                                           // 排序权重
	Status         int    `json:"status" binding:"omitempty,oneof=1 2"`                                               // 状态 1:启用 2:禁用
}

// DepartmentInfo 部门基础信息（可复用）
type DepartmentInfo struct {
	DepartmentID   string `json:"department_id" example:"123456789012345678"` // 部门ID
	ParentID       string `json:"parent_id" example:"123456789012345678"`     // 父部门ID
	DepartmentName string `json:"department_name" example:"技术部"`              // 部门名称
	Description    string `json:"description" example:"负责技术研发工作"`             // 部门描述
	Sort           int    `json:"sort" example:"1"`                           // 排序权重
	Status         int    `json:"status" example:"1" enum:"1,2"`              // 状态 1:启用 2:禁用
	CreatedAt      int64  `json:"created_at" example:"1735200000"`            // 创建时间
	UpdatedAt      int64  `json:"updated_at" example:"1735206400"`            // 更新时间
}

// DepartmentTreeNode 部门树节点
type DepartmentTreeNode struct {
	*DepartmentInfo
	Children []*DepartmentTreeNode `json:"children"` // 子部门
}

// ListDepartmentsRequest 部门列表请求
type ListDepartmentsRequest struct {
	pagination.Request `json:",inline"`
	DepartmentName     string `form:"department_name" binding:"omitempty,max=100"` // 部门名称（可选，模糊匹配）
	Status             int    `form:"status" binding:"omitempty,oneof=1 2"`        // 状态筛选
	ParentID           string `form:"parent_id" binding:"omitempty"`               // 父部门ID（为空则查询所有）
}

// ListDepartmentsResponse 部门列表响应
type ListDepartmentsResponse struct {
	pagination.Response `json:",inline"`
	List                []*DepartmentInfo `json:"list"` // 列表数据
}

// DepartmentTreeResponse 部门树响应
type DepartmentTreeResponse struct {
	Tree []*DepartmentTreeNode `json:"tree"` // 部门树
}

// DepartmentDetailRequest 获取部门详情请求
type DepartmentDetailRequest struct {
	DepartmentID string `json:"department_id" form:"department_id" binding:"required" example:"123456789012345678"` // 部门ID
}

// DepartmentDeleteRequest 删除部门请求
type DepartmentDeleteRequest struct {
	DepartmentID string `json:"department_id" form:"department_id" binding:"required" example:"123456789012345678"` // 部门ID
}

// DepartmentStatusRequest 更新部门状态请求
type DepartmentStatusRequest struct {
	DepartmentID string `json:"department_id" binding:"required" example:"123456789012345678"` // 部门ID
	Status       int    `json:"status" binding:"required" example:"1"`                         // 状态值
}

// DepartmentBatchDeleteRequest 批量删除部门请求
type DepartmentBatchDeleteRequest struct {
	DepartmentIDs []string `json:"department_ids" binding:"required,min=1,dive,required" example:"[\"123456789012345678\", \"987654321098765432\"]"` // 部门ID列表
}
