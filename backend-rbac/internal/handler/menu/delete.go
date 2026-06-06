package menu

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeleteMenu 删除菜单
// @Summary 删除菜单
// @Description 删除菜单（软删除）
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.MenuDeleteRequest true "删除菜单请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/menus [delete]
func (h *Handler) DeleteMenu(c *gin.Context) {
	var req dto.MenuDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeleteMenu(c.Request.Context(), req.MenuID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// BatchDeleteMenus 批量删除菜单
// @Summary 批量删除菜单
// @Description 批量软删除菜单（无子菜单的菜单才能删除）
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.MenuBatchDeleteRequest true "批量删除请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/menus/batch-delete [delete]
func (h *Handler) BatchDeleteMenus(c *gin.Context) {
	var req dto.MenuBatchDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.BatchDeleteMenus(c.Request.Context(), req.MenuIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.MenuIDs)})
}
