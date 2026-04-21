package auth

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// Login 处理登录请求
// @Summary 用户登录
// @Description 用户通过邮箱、密码和验证码进行登录。邮箱全局唯一，系统自动识别用户所属租户。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求参数"
// @Success 200 {object} response.Response{data=dto.LoginResponse} "登录成功"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), c.Request, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
