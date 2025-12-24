package dto

import "admin/pkg/pagination"

// AddTenantMemberRequest 添加租户成员请求
type AddTenantMemberRequest struct {
	UserName string   `json:"username" binding:"required"`       // 用户名（全局唯一）
	Name     string   `json:"name" binding:"required"`           // 姓名/昵称
	Phone    string   `json:"phone" binding:"omitempty"`         // 手机号（可选）
	Email    string   `json:"email" binding:"omitempty,email"`   // 邮箱（可选）
	RoleIDs  []string `json:"role_ids" binding:"required,min=1"` // 角色ID列表（至少一个角色）
}

// AddTenantMemberResponse 添加租户成员响应
type AddTenantMemberResponse struct {
	UserID          string   `json:"user_id"`          // 用户ID
	UserName        string   `json:"username"`         // 用户名
	Name            string   `json:"name"`             // 姓名
	InitialPassword string   `json:"initial_password"` // 初始密码（仅返回一次）
	TenantID        string   `json:"tenant_id"`        // 租户ID
	RoleIDs         []string `json:"role_ids"`         // 分配的角色ID列表
}

// RemoveTenantMemberRequest 移除租户成员请求
type RemoveTenantMemberRequest struct {
	UserID string `json:"user_id" binding:"required"` // 用户ID
}

// UpdateMemberRolesRequest 更新成员角色请求
type UpdateMemberRolesRequest struct {
	UserID  string   `json:"user_id" binding:"required"`        // 用户ID
	RoleIDs []string `json:"role_ids" binding:"required,min=1"` // 新的角色ID列表（至少一个角色）
}

// UpdateMemberRolesResponse 更新成员角色响应
type UpdateMemberRolesResponse struct {
	UserID  string   `json:"user_id"`  // 用户ID
	RoleIDs []string `json:"role_ids"` // 更新后的角色ID列表
}

// ListTenantMembersRequest 获取租户成员列表请求
type ListTenantMembersRequest struct {
	*pagination.Request `json:",inline"`
	Keyword             string `form:"keyword" binding:"omitempty"`          // 关键词搜索（用户名/姓名）
	Status              int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
}

// ListTenantMembersResponse 获取租户成员列表响应
type ListTenantMembersResponse struct {
	*pagination.Response `json:",inline"`
	List                 []*TenantMemberResponse `json:"list"` // 列表数据
}

// TenantMemberResponse 租户成员响应
type TenantMemberResponse struct {
	UserID        string   `json:"user_id"`         // 用户ID
	UserName      string   `json:"username"`        // 用户名
	Name          string   `json:"name"`            // 姓名
	Phone         string   `json:"phone"`           // 手机号
	Email         string   `json:"email"`           // 邮箱
	Status        int      `json:"status"`          // 状态 1:正常 2:禁用
	RoleIDs       []string `json:"role_ids"`        // 角色ID列表
	FirstLogin    bool     `json:"first_login"`     // 是否首次登录
	LastLoginTime int64    `json:"last_login_time"` // 最后登录时间
	CreatedAt     int64    `json:"created_at"`      // 加入时间
}
