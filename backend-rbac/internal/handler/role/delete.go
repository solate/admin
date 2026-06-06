package role

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除角色（软删除）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.RoleDeleteRequest true "删除角色请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/roles [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	var req dto.RoleDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeleteRole(c.Request.Context(), req.RoleID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// BatchDeleteRoles 批量删除角色
// @Summary 批量删除角色
// @Description 批量软删除角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.RoleBatchDeleteRequest true "批量删除请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/roles/batch-delete [delete]
func (h *Handler) BatchDeleteRoles(c *gin.Context) {
	var req dto.RoleBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.BatchDeleteRoles(c.Request.Context(), req.RoleIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.RoleIDs)})
}
