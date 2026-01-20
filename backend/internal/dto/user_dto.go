package dto

import "admin/pkg/pagination"

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	UserName    string   `json:"username" binding:"omitempty"`         // 用户名（可选）
	Nickname    string   `json:"nickname" binding:"omitempty"`         // 昵称/显示名称
	Phone       string   `json:"phone" binding:"omitempty"`            // 手机号
	Email       string   `json:"email" binding:"omitempty"`            // 邮箱
	Description string   `json:"description" binding:"omitempty"`      // 用户描述
	Remark      string   `json:"remark" binding:"omitempty"`           // 备注信息
	Status      int      `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:正常 2:禁用
	TenantID    string   `json:"tenant_id"`                            // 租户ID（可选，从上下文获取）
	RoleCodes   []string `json:"role_codes" binding:"omitempty"`       // 角色编码列表（可选）
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	UserID      string   `json:"user_id" form:"user_id" binding:"required" example:"123456789012345678"` // 用户ID
	Phone       string   `json:"phone" binding:"omitempty"`                                              // 手机号
	Email       string   `json:"email" binding:"omitempty"`                                              // 邮箱
	Status      int      `json:"status" binding:"omitempty,oneof=1 2"`                                   // 状态 1:正常 2:禁用
	Nickname    string   `json:"nickname" binding:"omitempty"`                                           // 昵称/显示名称
	Description string   `json:"description" binding:"omitempty"`                                        // 用户描述
	Remark      string   `json:"remark" binding:"omitempty"`                                             // 备注信息
	TenantID    string   `json:"tenant_id" binding:"omitempty"`                                          // 租户ID（可选，用于迁移用户到其他租户）
	RoleCodes   []string `json:"role_codes" binding:"omitempty"`                                         // 角色编码列表（可选，不可传空数组）
}

// UserInfo 用户基础信息（可复用）
type UserInfo struct {
	UserID             string      `json:"user_id" example:"123456789012345678"`            // 用户ID
	UserName           string      `json:"username" example:"admin"`                        // 用户名
	Nickname           string      `json:"nickname" example:"系统管理员"`                        // 昵称/显示名称
	Avatar             string      `json:"avatar" example:"https://example.com/avatar.jpg"` // 头像URL
	Phone              string      `json:"phone" example:"13800138000"`                     // 手机号
	Email              string      `json:"email" example:"admin@example.com"`               // 邮箱
	Description        string      `json:"description" example:"用户描述"`                      // 用户描述
	Remark             string      `json:"remark" example:"备注信息"`                           // 备注信息
	Status             int         `json:"status" example:"1" enum:"1,2"`                   // 状态 1:正常 2:禁用
	TenantID           string      `json:"tenant_id" example:"123456789012345678"`          // 租户ID
	LastLoginTime      int64       `json:"last_login_time" example:"1735206400"`            // 最后登录时间（Unix时间戳）
	MustChangePassword int16       `json:"must_change_password" example:"1" enum:"1,2"`     // 是否必须修改密码 1:是 2:否
	CreatedAt          int64       `json:"created_at" example:"1735200000"`                 // 创建时间（Unix时间戳）
	UpdatedAt          int64       `json:"updated_at" example:"1735206400"`                 // 更新时间（Unix时间戳）
	Roles              []*RoleInfo `json:"roles"`                                           // 角色列表
	Tenant             *TenantInfo `json:"tenant"`                                          // 租户信息
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	pagination.Request `json:",inline"`
	Nickname           string `form:"nickname" binding:"omitempty"`         // 昵称模糊搜索
	Status             int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
	TenantID           string `form:"tenant_id" binding:"omitempty"`        // 租户ID筛选
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	pagination.Response `json:",inline"`
	List                []*UserInfo `json:"list"` // 列表数据
}

// ProfileResponse 用户档案响应（含角色和租户信息）
type ProfileResponse struct {
	User   *UserInfo   `json:"user"`   // 当前用户信息
	Tenant *TenantInfo `json:"tenant"` // 当前租户信息
	Roles  []*RoleInfo `json:"roles"`  // 用户角色列表
}

// AssignRolesRequest 为用户分配角色请求
type AssignRolesRequest struct {
	UserID    string   `json:"user_id" form:"user_id" binding:"required" example:"123456789012345678"` // 用户ID
	RoleCodes []string `json:"role_codes" binding:"required,min=1"`                                    // 角色编码列表
}

// UserRolesResponse 用户角色响应
type UserRolesResponse struct {
	UserID   string      `json:"user_id" example:"123456789012345678"` // 用户ID
	UserName string      `json:"username" example:"admin"`             // 用户名
	Roles    []*RoleInfo `json:"roles"`                                // 角色列表
}

// ChangePasswordRequest 用户修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`       // 原密码
	NewPassword string `json:"new_password" binding:"required,min=6"` // 新密码
}

// ChangePasswordResponse 修改密码响应
type ChangePasswordResponse struct {
	Success bool   `json:"success" example:"true"`         // 是否成功
	Message string `json:"message" example:"密码修改成功，请重新登录"` // 提示信息
}

// ResetPasswordRequest 超管重置密码请求
type ResetPasswordRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required" example:"123456789012345678"` // 用户ID
	// 密码由系统自动生成，前端不需要传入任何参数
}

// ResetPasswordResponse 重置密码响应（密码只显示一次）
type ResetPasswordResponse struct {
	Password string `json:"password" example:"Abc123456"` // 重置后的密码（仅重置时显示一次）
	Message  string `json:"message" example:"密码重置成功"`     // 提示信息
}

// CreateUserResponse 创建用户响应（包含初始密码）
type CreateUserResponse struct {
	User     *UserInfo `json:"user"`                         // 用户信息
	Password string    `json:"password" example:"Abc123456"` // 初始密码（仅创建时显示一次）
	Message  string    `json:"message" example:"用户创建成功"`     // 提示信息
}

// UserDetailRequest 获取用户详情请求
type UserDetailRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required" example:"123456789012345678"` // 用户ID
}

// UserDeleteRequest 删除用户请求
type UserDeleteRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required" example:"123456789012345678"` // 用户ID
}

// UserBatchDeleteRequest 批量删除用户请求
type UserBatchDeleteRequest struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1,dive,required" example:"[\"123456789012345678\", \"987654321098765432\"]"` // 用户ID列表
}

// UserStatusRequest 更新用户状态请求
type UserStatusRequest struct {
	UserID string `json:"user_id" binding:"required" example:"123456789012345678"` // 用户ID
	Status int    `json:"status" binding:"required" example:"1"`                   // 状态值
}
