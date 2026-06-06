package user

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// ChangePassword 用户修改自己的密码
// @Summary 修改密码
// @Description 用户修改自己的登录密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.ChangePasswordRequest true "修改密码请求参数"
// @Success 200 {object} response.Response{data=dto.ChangePasswordResponse} "修改成功"
// @Router /api/v1/user/password/change [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, &dto.ChangePasswordResponse{
		Success: true,
		Message: "密码修改成功，请重新登录",
	})
}

// ResetPassword 超管重置用户密码
// @Summary 重置用户密码
// @Description 超级管理员重置指定用户的密码，密码自动生成并在响应中显示一次
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.ResetPasswordRequest true "重置密码请求参数"
// @Success 200 {object} response.Response{data=dto.ResetPasswordResponse} "重置成功"
// @Router /api/v1/users/password/reset [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ResetPassword(c.Request.Context(), req.UserID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
