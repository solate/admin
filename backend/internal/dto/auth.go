package dto

type LoginRequest struct {
	UserName  string `json:"username" binding:"required"`   // 用户名
	Password  string `json:"password" binding:"required"`   // 密码
	CaptchaID string `json:"captcha_id" binding:"required"` // 验证码ID
	Captcha   string `json:"captcha" binding:"required"`    // 验证码
	TenantID  string `json:"tenant_id" binding:"omitempty"` // 租户ID (可选) 多租户
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int64  `json:"expires_in"`    // 过期时间（秒）
	UserID       string `json:"user_id"`       // 用户ID
	TenantID     string `json:"tenant_id"`     // 租户ID
	Phone        string `json:"phone"`         // 手机号
	Email        string `json:"email"`         // 邮箱
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // 刷新令牌
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
