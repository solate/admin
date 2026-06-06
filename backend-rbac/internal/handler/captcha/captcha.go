package captcha

import (
	"admin/pkg/response"
	"admin/pkg/utils/captcha"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Handler 验证码处理器
type Handler struct {
	captchaMgr *captcha.Manager
}

// NewHandler 创建验证码处理器
func NewHandler(rdb redis.UniversalClient) *Handler {
	return &Handler{
		captchaMgr: captcha.NewManager(rdb),
	}
}

// Get 获取验证码
// @Summary 获取图形验证码
// @Description 生成图形验证码，返回验证码ID和Base64图片数据
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response{data=dto.CaptchaResponse}
// @Router /api/v1/auth/captcha [get]
func (h *Handler) Get(c *gin.Context) {
	id, b64s, _, err := h.captchaMgr.Generate()
	if err != nil {
		response.Error(c, xerr.ErrCaptchaInvalid)
		return
	}

	response.Success(c, gin.H{
		"captcha_id":   id,
		"captcha_data": b64s,
	})
}
