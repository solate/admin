package dto

import "admin/pkg/pagination"

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	UserName string `json:"username" binding:"required"`          // 用户名
	Nickname string `json:"nickname" binding:"omitempty"`         // 昵称/显示名称
	Password string `json:"password" binding:"required"`          // 密码
	Phone    string `json:"phone" binding:"omitempty"`            // 手机号
	Email    string `json:"email" binding:"omitempty"`            // 邮箱
	Remark   string `json:"remark" binding:"omitempty"`           // 备注信息
	Status   int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:正常 2:禁用
	TenantID string `json:"tenant_id"`                            // 租户ID（可选，从上下文获取）
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Password string `json:"password" binding:"omitempty"`         // 密码
	Phone    string `json:"phone" binding:"omitempty"`            // 手机号
	Email    string `json:"email" binding:"omitempty"`            // 邮箱
	Status   int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:正常 2:禁用
	Nickname string `json:"nickname" binding:"omitempty"`         // 昵称/显示名称
	Remark   string `json:"remark" binding:"omitempty"`           // 备注信息
}

// UserInfo 用户基础信息（可复用）
type UserInfo struct {
	UserID        string `json:"user_id"`         // 用户ID
	UserName      string `json:"username"`        // 用户名
	Nickname      string `json:"nickname"`        // 昵称/显示名称
	Avatar        string `json:"avatar"`          // 头像URL
	Phone         string `json:"phone"`           // 手机号
	Email         string `json:"email"`           // 邮箱
	Status        int    `json:"status"`          // 状态 1:正常 2:禁用
	TenantID      string `json:"tenant_id"`       // 租户ID
	LastLoginTime int64  `json:"last_login_time"` // 最后登录时间
	CreatedAt     int64  `json:"created_at"`      // 创建时间
	UpdatedAt     int64  `json:"updated_at"`      // 更新时间
}

// UserResponse 用户响应
type UserResponse struct {
	User *UserInfo `json:"user"` // 用户基础信息
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	*pagination.Request `json:",inline"`
	UserName            string `form:"username" binding:"omitempty"`         // 用户名模糊搜索
	Status              int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
	TenantID            string `form:"tenant_id" binding:"omitempty"`        // 租户ID筛选
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	*pagination.Response `json:",inline"`
	List                 []*UserResponse `json:"list"` // 列表数据
}

// ProfileResponse 用户档案响应（含角色和租户信息）
type ProfileResponse struct {
	User   *UserInfo   `json:"user"`   // 当前用户信息
	Tenant *TenantInfo `json:"tenant"` // 当前租户信息
	Roles  []*RoleInfo `json:"roles"`  // 用户角色列表
}
