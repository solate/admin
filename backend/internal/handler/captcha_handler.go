package handler

import (
	"admin/internal/dto"
	"admin/pkg/captcha"
	"admin/pkg/response"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// CaptchaHandler 验证码处理器
type CaptchaHandler struct {
	captchaMgr *captcha.Manager
}

// NewCaptchaHandler 创建验证码处理器
func NewCaptchaHandler(rdb redis.UniversalClient) *CaptchaHandler {
	return &CaptchaHandler{
		captchaMgr: captcha.NewManager(rdb),
	}
}

// Get 获取验证码
// @Summary 获取图形验证码
// @Description 生成图形验证码，返回验证码ID和Base64图片数据
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response{data=dto.CaptchaResponse}
// @Failure 200 {object} response.Response
// @Router /api/v1/captcha [get]
func (h *CaptchaHandler) Get(c *gin.Context) {
	id, b64s, _, err := h.captchaMgr.Generate()
	if err != nil {
		response.Error(c, xerr.ErrCaptchaInvalid)
		return
	}

	response.Success(c, dto.CaptchaResponse{
		CaptchaID:   id,
		CaptchaData: b64s,
	})
}
