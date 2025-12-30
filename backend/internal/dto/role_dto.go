package dto

import "admin/pkg/pagination"

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	RoleCode    string `json:"role_code" binding:"required"`         // 角色编码（租户内唯一）
	Name        string `json:"name" binding:"required"`              // 角色名称
	Description string `json:"description" binding:"omitempty"`      // 角色描述
	Status      int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty"`             // 角色名称
	Description string `json:"description" binding:"omitempty"`      // 角色描述
	Status      int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
}

// RoleResponse 角色响应
type RoleResponse struct {
	RoleID      string `json:"role_id"`     // 角色ID
	TenantID    string `json:"tenant_id"`   // 租户ID
	RoleCode    string `json:"role_code"`   // 角色编码
	Name        string `json:"name"`        // 角色名称
	Description string `json:"description"` // 角色描述
	Status      int    `json:"status"`      // 状态 1:启用 2:禁用
	CreatedAt   int64  `json:"created_at"`  // 创建时间
	UpdatedAt   int64  `json:"updated_at"`  // 更新时间
}

// ListRolesRequest 角色列表请求
type ListRolesRequest struct {
	*pagination.Request `json:",inline"`
	Keyword             string `form:"keyword" binding:"omitempty"`          // 关键词搜索（角色名称/编码）
	Status              int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
}

// ListRolesResponse 角色列表响应
type ListRolesResponse struct {
	*pagination.Response `json:",inline"`
	List                 []*RoleResponse `json:"list"` // 列表数据
}

// RoleInfo 角色基础信息（可复用）
type RoleInfo struct {
	RoleID      string `json:"role_id"`     // 角色ID
	RoleCode    string `json:"role_code"`   // 角色编码
	Name        string `json:"name"`        // 角色名称
	Description string `json:"description"` // 角色描述
}
