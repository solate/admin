package dto

import "admin/pkg/pagination"

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	RoleCode       string  `json:"role_code" binding:"required"`         // 角色编码（租户内唯一）
	Name           string  `json:"name" binding:"required"`              // 角色名称
	Description    string  `json:"description" binding:"omitempty"`      // 角色描述
	Status         int     `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
	ParentRoleCode *string `json:"parent_role_code" binding:"omitempty"` // 父角色编码（继承 default 租户的角色模板）
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	RoleID      string `json:"role_id" form:"role_id" binding:"required" example:"123456789012345678"` // 角色ID
	Name        string `json:"name" binding:"omitempty"`                                               // 角色名称
	Description string `json:"description" binding:"omitempty"`                                        // 角色描述
	Status      int    `json:"status" binding:"omitempty,oneof=1 2"`                                   // 状态 1:启用 2:禁用
}

// ListRolesRequest 角色列表请求
type ListRolesRequest struct {
	pagination.Request `json:",inline"`
	RoleName           string `form:"role_name" binding:"omitempty,max=50"` // 角色名称（可选，模糊匹配）
	RoleCode           string `form:"role_code" binding:"omitempty,max=50"` // 角色编码（可选，模糊匹配）
	Status             int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
}

// ListRolesResponse 角色列表响应
type ListRolesResponse struct {
	pagination.Response `json:",inline"`
	List                []*RoleInfo `json:"list"` // 列表数据
}

// RoleInfo 角色信息（可复用）
type RoleInfo struct {
	RoleID         string  `json:"role_id" example:"123456789012345678"`   // 角色ID
	TenantID       string  `json:"tenant_id" example:"123456789012345678"` // 租户ID
	TenantCode     string  `json:"tenant_code" example:"default"`          // 租户代码
	RoleCode       string  `json:"role_code" example:"admin"`              // 角色编码
	Name           string  `json:"name" example:"系统管理员"`                   // 角色名称
	Description    string  `json:"description" example:"系统管理员，拥有所有权限"`     // 角色描述
	Status         int     `json:"status" example:"1" enum:"1,2"`          // 状态 1:启用 2:禁用
	ParentRoleCode *string `json:"parent_role_code"`                       // 父角色编码
	CreatedAt      int64   `json:"created_at" example:"1735200000"`        // 创建时间
	UpdatedAt      int64   `json:"updated_at" example:"1735206400"`        // 更新时间
}

// AssignPermissionsRequest 分配权限请求（菜单+按钮）
type AssignPermissionsRequest struct {
	RoleID    string   `json:"role_id" form:"role_id" binding:"required" example:"123456789012345678"` // 角色ID
	MenuIDs   []string `json:"menu_ids" binding:"required"`                                            // 菜单ID列表
	ButtonIDs []string `json:"button_ids" binding:"omitempty"`                                         // 按钮权限ID列表
}

// RolePermissionsResponse 角色权限响应
type RolePermissionsResponse struct {
	MenuIDs   []string `json:"menu_ids"`   // 菜单ID列表
	ButtonIDs []string `json:"button_ids"` // 按钮权限ID列表
}

// GetAllRolesRequest 获取所有角色请求
type GetAllRolesRequest struct {
	RoleName string `form:"role_name" binding:"omitempty,max=50"` // 角色名称（可选，模糊匹配）
	RoleCode string `form:"role_code" binding:"omitempty,max=50"` // 角色编码（可选，模糊匹配）
	Status   int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
}

// GetAllRolesResponse 获取所有角色响应
type GetAllRolesResponse struct {
	List []*RoleInfo `json:"list"` // 角色列表
}

// RoleDetailRequest 获取角色详情请求
type RoleDetailRequest struct {
	RoleID string `json:"role_id" form:"role_id" binding:"required" example:"123456789012345678"` // 角色ID
}

// RoleDeleteRequest 删除角色请求
type RoleDeleteRequest struct {
	RoleID string `json:"role_id" form:"role_id" binding:"required" example:"123456789012345678"` // 角色ID
}

// RoleStatusRequest 更新角色状态请求
type RoleStatusRequest struct {
	RoleID string `json:"role_id" binding:"required" example:"123456789012345678"` // 角色ID
	Status int    `json:"status" binding:"required" example:"1"`                   // 状态值
}

// RoleBatchDeleteRequest 批量删除角色请求
type RoleBatchDeleteRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required,min=1,dive,required" example:"[\"123456789012345678\", \"987654321098765432\"]"` // 角色ID列表
}
