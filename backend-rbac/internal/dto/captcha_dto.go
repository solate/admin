package dto

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaID   string `json:"captcha_id" example:"123456789012345678"`   // 验证码ID
	CaptchaData string `json:"captcha_data" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."` // Base64图片数据
}
