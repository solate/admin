package menu

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// UpdateMenu 更新菜单
// @Summary 更新菜单
// @Description 更新菜单信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdateMenuRequest true "更新菜单请求参数"
// @Success 200 {object} response.Response{data=dto.MenuInfo} "更新成功"
// @Router /api/v1/menus [put]
func (h *Handler) UpdateMenu(c *gin.Context) {
	var req dto.UpdateMenuRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.UpdateMenu(c.Request.Context(), req.MenuID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateMenuStatus 更新菜单状态
// @Summary 更新菜单状态
// @Description 更新菜单显示/隐藏状态
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.MenuStatusRequest true "更新状态请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/menus/status [put]
func (h *Handler) UpdateMenuStatus(c *gin.Context) {
	var req dto.MenuStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.UpdateMenuStatus(c.Request.Context(), req.MenuID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
