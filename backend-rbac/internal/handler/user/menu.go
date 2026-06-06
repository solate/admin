package user

import (
	"admin/internal/dto"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// GetUserMenu 获取用户菜单树
// @Summary 获取用户菜单
// @Description 获取当前登录用户的菜单树
// @Tags 用户菜单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.UserMenuResponse} "获取成功"
// @Router /api/v1/user/menus [get]
func (h *Handler) GetUserMenu(c *gin.Context) {
	// 从上下文获取用户名和租户编码
	// 这些信息由 Auth 中间件从 JWT token 中提取并设置到 context
	userName := xcontext.GetUserName(c.Request.Context())
	tenantCode := xcontext.GetTenantCode(c.Request.Context())

	if userName == "" || tenantCode == "" {
		response.Error(c, xerr.ErrUnauthorized)
		return
	}

	resp, err := h.menuSvc.GetUserMenu(c.Request.Context(), userName, tenantCode)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUserButtons 获取菜单按钮权限
// @Summary 获取菜单按钮权限
// @Description 获取指定菜单的按钮权限
// @Tags 用户菜单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param menu_id query string true "菜单ID"
// @Success 200 {object} response.Response{data=dto.UserButtonsResponse} "获取成功"
// @Router /api/v1/user/buttons [get]
func (h *Handler) GetUserButtons(c *gin.Context) {
	var req dto.MenuDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.menuSvc.GetUserButtons(c.Request.Context(), req.MenuID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
