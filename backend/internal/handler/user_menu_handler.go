package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xerr"
	"github.com/gin-gonic/gin"
)

// UserMenuHandler 用户菜单处理器
type UserMenuHandler struct {
	userMenuService *service.UserMenuService
}

// NewUserMenuHandler 创建用户菜单处理器
func NewUserMenuHandler(userMenuService *service.UserMenuService) *UserMenuHandler {
	return &UserMenuHandler{
		userMenuService: userMenuService,
	}
}

// GetUserMenu 获取用户菜单树
// @Summary 获取用户菜单树
// @Description 获取当前登录用户可见的菜单树（基于角色权限）
// @Tags 用户菜单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.UserMenuResponse} "获取成功"
// @Router /user/menu [get]
func (h *UserMenuHandler) GetUserMenu(c *gin.Context) {
	resp, err := h.userMenuService.GetUserMenu(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUserButtons 获取菜单按钮权限
// @Summary 获取菜单按钮权限
// @Description 获取指定菜单下的按钮权限列表
// @Tags 用户菜单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param menu_id query string true "菜单ID"
// @Success 200 {object} response.Response{data=dto.UserButtonsResponse} "获取成功"
// @Router /user/buttons [get]
func (h *UserMenuHandler) GetUserButtons(c *gin.Context) {
	var req dto.UserButtonsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.userMenuService.GetUserButtons(c.Request.Context(), req.MenuID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
