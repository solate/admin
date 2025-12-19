package dto

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaID   string `json:"captcha_id"`   // 验证码ID
	CaptchaData string `json:"captcha_data"` // Base64图片数据
}
