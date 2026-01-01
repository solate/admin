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
	AccessToken  string `json:"access_token"`  // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int64  `json:"expires_in"`    // 过期时间（秒）
}

// SwitchTenantRequest 切换租户请求
type SwitchTenantRequest struct {
	TenantID string `json:"tenant_id" binding:"required"` // 切换租户ID
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

// AvailableTenantsResponse 可用租户列表响应
type AvailableTenantsResponse struct {
	Tenants []*TenantInfo `json:"tenants"` // 租户列表
}
