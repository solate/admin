package user

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户（软删除）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UserDeleteRequest true "删除用户请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/users [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	var req dto.UserDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeleteUser(c.Request.Context(), req.UserID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// BatchDeleteUsers 批量删除用户
// @Summary 批量删除用户
// @Description 批量软删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UserBatchDeleteRequest true "批量删除请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/users/batch-delete [delete]
func (h *Handler) BatchDeleteUsers(c *gin.Context) {
	var req dto.UserBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.BatchDeleteUsers(c.Request.Context(), req.UserIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.UserIDs)})
}
