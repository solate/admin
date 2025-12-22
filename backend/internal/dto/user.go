package dto

import (
	"admin/pkg/common"
	"time"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	UserName string `json:"username" binding:"required"`  // 用户名
	Password string `json:"password" binding:"required"`  // 密码
	Phone    string `json:"phone" binding:"omitempty"`    // 手机号
	Email    string `json:"email" binding:"omitempty"`    // 邮箱
	Status   int32  `json:"status" binding:"omitempty"`   // 状态 1:正常 2:禁用
	TenantID string `json:"tenant_id"`                    // 租户ID（可选，从上下文获取）
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Password string `json:"password" binding:"omitempty"` // 密码
	Phone    string `json:"phone" binding:"omitempty"`    // 手机号
	Email    string `json:"email" binding:"omitempty"`    // 邮箱
	Status   int32  `json:"status" binding:"omitempty"`   // 状态 1:正常 2:禁用
	Name     string `json:"name" binding:"omitempty"`     // 姓名/昵称
	Remark   string `json:"remark" binding:"omitempty"`   // 备注信息
}

// UserResponse 用户响应
type UserResponse struct {
	UserID        string    `json:"user_id"`         // 用户ID
	UserName      string    `json:"username"`        // 用户名
	Name          string    `json:"name"`            // 姓名/昵称
	Avatar        string    `json:"avatar"`          // 头像URL
	Phone         string    `json:"phone"`           // 手机号
	Email         string    `json:"email"`           // 邮箱
	Status        int32     `json:"status"`          // 状态
	TenantID      string    `json:"tenant_id"`       // 租户ID
	RoleType      int32     `json:"role_type"`       // 角色类型
	LastLoginTime int64     `json:"last_login_time"` // 最后登录时间
	CreatedAt     time.Time `json:"created_at"`      // 创建时间
	UpdatedAt     time.Time `json:"updated_at"`      // 更新时间
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	common.PageRequest `json:",inline"` // 分页请求
	UserName           string           `form:"username" binding:"omitempty"` // 用户名模糊搜索
	Status             int32            `form:"status" binding:"omitempty"`   // 状态筛选
	TenantID           string           `form:"tenant_id" binding:"omitempty"` // 租户ID筛选
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	List      []*UserResponse `json:"list"`       // 用户列表
	Page      int             `json:"page"`       // 当前页码
	PageSize  int             `json:"page_size"`  // 每页大小
	Total     int64           `json:"total"`      // 总记录数
	TotalPage int64           `json:"total_page"` // 总页数
}
