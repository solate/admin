package dto

import "admin/pkg/pagination"

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	ParentID       string `json:"parent_id" binding:"omitempty"`       // 父部门ID（根部门为空字符串）
	DepartmentName string `json:"department_name" binding:"required"`  // 部门名称
	Description    string `json:"description" binding:"omitempty"`     // 部门描述
	Sort           int    `json:"sort" binding:"omitempty"`            // 排序权重
	Status         int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
}

// UpdateDepartmentRequest 更新部门请求
type UpdateDepartmentRequest struct {
	ParentID       string `json:"parent_id" binding:"omitempty"`       // 父部门ID
	DepartmentName string `json:"department_name" binding:"omitempty"` // 部门名称
	Description    string `json:"description" binding:"omitempty"`     // 部门描述
	Sort           int    `json:"sort" binding:"omitempty"`            // 排序权重
	Status         int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
}

// DepartmentResponse 部门响应
type DepartmentResponse struct {
	DepartmentID   string `json:"department_id"`   // 部门ID
	ParentID       string `json:"parent_id"`       // 父部门ID
	DepartmentName string `json:"department_name"` // 部门名称
	Description    string `json:"description"`     // 部门描述
	Sort           int    `json:"sort"`            // 排序权重
	Status         int    `json:"status"`          // 状态 1:启用 2:禁用
	CreatedAt      int64  `json:"created_at"`      // 创建时间
	UpdatedAt      int64  `json:"updated_at"`      // 更新时间
}

// DepartmentTreeNode 部门树节点
type DepartmentTreeNode struct {
	*DepartmentResponse
	Children []*DepartmentTreeNode `json:"children"` // 子部门
}

// ListDepartmentsRequest 部门列表请求
type ListDepartmentsRequest struct {
	pagination.Request  `json:",inline"`
	Keyword             string `form:"keyword" binding:"omitempty"`          // 关键词搜索（部门名称）
	Status              int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
	ParentID            string `form:"parent_id" binding:"omitempty"`        // 父部门ID（为空则查询所有）
}

// ListDepartmentsResponse 部门列表响应
type ListDepartmentsResponse struct {
	pagination.Response `json:",inline"`
	List                []*DepartmentResponse `json:"list"` // 列表数据
}

// DepartmentTreeResponse 部门树响应
type DepartmentTreeResponse struct {
	Tree []*DepartmentTreeNode `json:"tree"` // 部门树
}

// DepartmentInfo 部门基础信息（可复用）
type DepartmentInfo struct {
	DepartmentID   string `json:"department_id"`   // 部门ID
	DepartmentName string `json:"department_name"` // 部门名称
}
