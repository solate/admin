package menu

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetMenu 获取菜单详情
// @Summary 获取菜单详情
// @Description 根据ID获取菜单详情
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param menu_id query string true "菜单ID"
// @Success 200 {object} response.Response{data=dto.MenuInfo} "获取成功"
// @Router /api/v1/menus/detail [get]
func (h *Handler) GetMenu(c *gin.Context) {
	var req dto.MenuDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetMenuByID(c.Request.Context(), req.MenuID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListMenus 获取菜单列表
// @Summary 获取菜单列表
// @Description 分页获取菜单列表，支持按关键词和状态筛选
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param name query string false "菜单名称搜索"
// @Param status query int false "状态筛选(1:显示,2:隐藏)" Enums(1,2)
// @Success 200 {object} response.Response{data=dto.ListMenusResponse} "获取成功"
// @Router /api/v1/menus [get]
func (h *Handler) ListMenus(c *gin.Context) {
	var req dto.ListMenusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ListMenus(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetAllMenus 获取所有菜单（平铺）
// @Summary 获取所有菜单
// @Description 获取当前租户的所有菜单（平铺列表）
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.AllMenusResponse} "获取成功"
// @Router /api/v1/menus/all [get]
func (h *Handler) GetAllMenus(c *gin.Context) {
	resp, err := h.svc.GetAllMenus(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetMenuTree 获取菜单树
// @Summary 获取菜单树
// @Description 获取当前租户的菜单树结构
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.MenuTreeResponse} "获取成功"
// @Router /api/v1/menus/tree [get]
func (h *Handler) GetMenuTree(c *gin.Context) {
	resp, err := h.svc.GetMenuTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
