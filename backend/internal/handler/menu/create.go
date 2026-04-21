package menu

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreateMenu 创建菜单
// @Summary 创建菜单
// @Description 创建新的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateMenuRequest true "创建菜单请求参数"
// @Success 200 {object} response.Response{data=dto.MenuInfo} "创建成功"
// @Router /api/v1/menus [post]
func (h *Handler) CreateMenu(c *gin.Context) {
	var req dto.CreateMenuRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.CreateMenu(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
