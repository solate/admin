package dto

import "admin/pkg/pagination"

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	RoleCode       string `json:"role_code" binding:"required"`         // 角色编码（租户内唯一）
	Name           string `json:"name" binding:"required"`              // 角色名称
	Description    string `json:"description" binding:"omitempty"`      // 角色描述
	Status         int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
	ParentRoleCode *string `json:"parent_role_code" binding:"omitempty"` // 父角色编码（继承 default 租户的角色模板）
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty"`             // 角色名称
	Description string `json:"description" binding:"omitempty"`      // 角色描述
	Status      int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
}

// RoleResponse 角色响应
type RoleResponse struct {
	RoleID        string  `json:"role_id"`         // 角色ID
	TenantID      string  `json:"tenant_id"`       // 租户ID
	RoleCode      string  `json:"role_code"`       // 角色编码
	Name          string  `json:"name"`            // 角色名称
	Description   string  `json:"description"`     // 角色描述
	Status        int     `json:"status"`          // 状态 1:启用 2:禁用
	ParentRoleCode *string `json:"parent_role_code"` // 父角色编码
	CreatedAt     int64   `json:"created_at"`      // 创建时间
	UpdatedAt     int64   `json:"updated_at"`      // 更新时间
}

// ListRolesRequest 角色列表请求
type ListRolesRequest struct {
	pagination.Request  `json:",inline"`
	Keyword             string `form:"keyword" binding:"omitempty"`          // 关键词搜索（角色名称/编码）
	Status              int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
}

// ListRolesResponse 角色列表响应
type ListRolesResponse struct {
	pagination.Response `json:",inline"`
	List                []*RoleResponse `json:"list"` // 列表数据
}

// RoleInfo 角色基础信息（可复用）
type RoleInfo struct {
	RoleID      string `json:"role_id"`     // 角色ID
	RoleCode    string `json:"role_code"`   // 角色编码
	Name        string `json:"name"`        // 角色名称
	Description string `json:"description"` // 角色描述
}

// AssignPermissionsRequest 分配权限请求（菜单+按钮）
type AssignPermissionsRequest struct {
	MenuIDs   []string `json:"menu_ids" binding:"required"`    // 菜单ID列表
	ButtonIDs []string `json:"button_ids" binding:"omitempty"` // 按钮权限ID列表
}

// RolePermissionsResponse 角色权限响应
type RolePermissionsResponse struct {
	MenuIDs   []string `json:"menu_ids"`   // 菜单ID列表
	ButtonIDs []string `json:"button_ids"` // 按钮权限ID列表
}

// AssignMenusRequest 分配菜单请求（保留向后兼容，已弃用）
// Deprecated: 使用 AssignPermissionsRequest 代替
type AssignMenusRequest = AssignPermissionsRequest

// RoleMenusResponse 角色菜单响应（保留向后兼容，已弃用）
// Deprecated: 使用 RolePermissionsResponse 代替
type RoleMenusResponse = RolePermissionsResponse
