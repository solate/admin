package dto

// LoginRequest 登录请求
type LoginRequest struct {
	Email     string `json:"email" binding:"required,email"` // 邮箱
	Password  string `json:"password" binding:"required"`    // 密码
	CaptchaID string `json:"captcha_id" binding:"required"`  // 验证码ID
	Captcha   string `json:"captcha" binding:"required"`     // 验证码
}

// PhoneLoginRequest 手机号登录请求
type PhoneLoginRequest struct {
	Phone    string `json:"phone" binding:"required"`    // 手机号
	Password string `json:"password" binding:"required"` // 密码
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`  // 访问令牌
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // 刷新令牌
	ExpiresIn    int64  `json:"expires_in" example:"3600"`                                       // 过期时间（秒）
}

// SwitchTenantRequest 切换租户请求
type SwitchTenantRequest struct {
	TenantID string `json:"tenant_id" binding:"required" example:"123456789012345678"` // 切换租户ID
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // 刷新令牌
}

// RefreshResponse 刷新令牌响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// AvailableTenantsResponse 可用租户列表响应
type AvailableTenantsResponse struct {
	Tenants []*TenantInfo `json:"tenants"` // 租户列表
}
