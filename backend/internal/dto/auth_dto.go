package dto

// LoginRequest 登录请求
type LoginRequest struct {
	UserName  string `json:"username" binding:"required"`   // 用户名
	Password  string `json:"password" binding:"required"`   // 密码
	CaptchaID string `json:"captcha_id" binding:"required"` // 验证码ID
	Captcha   string `json:"captcha" binding:"required"`    // 验证码
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`  // 访问令牌
	RefreshToken string      `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int64       `json:"expires_in"`    // 过期时间（秒）
	User         *UserInfo   `json:"user"`          // 当前用户信息
	Tenant       *TenantInfo `json:"tenant"`        // 当前租户信息
	Roles        []*RoleInfo `json:"roles"`         // 用户角色列表
}

// SelectTenantRequest 选择租户请求
type SelectTenantRequest struct {
	TenantID string `json:"tenant_id" binding:"required"` // 要选择的租户ID
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // 刷新令牌
}

// RefreshResponse 刷新令牌响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
