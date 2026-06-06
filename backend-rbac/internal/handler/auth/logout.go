package auth

import (
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// Logout 处理登出请求
// @Summary 用户登出
// @Description 用户登出，将当前访问令牌加入黑名单
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response "登出成功"
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	if err := h.svc.Logout(c.Request.Context(), c.Request); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
