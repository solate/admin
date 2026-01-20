package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	menuService *service.MenuService
}

// NewMenuHandler 创建菜单处理器
func NewMenuHandler(menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
	}
}

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
func (h *MenuHandler) CreateMenu(c *gin.Context) {
	var req dto.CreateMenuRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.menuService.CreateMenu(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

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
func (h *MenuHandler) GetMenu(c *gin.Context) {
	var req dto.MenuDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.menuService.GetMenuByID(c.Request.Context(), req.MenuID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

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
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	var req dto.UpdateMenuRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.menuService.UpdateMenu(c.Request.Context(), req.MenuID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

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
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	var req dto.MenuDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.menuService.DeleteMenu(c.Request.Context(), req.MenuID); err != nil {
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
func (h *MenuHandler) BatchDeleteMenus(c *gin.Context) {
	var req dto.MenuBatchDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.menuService.BatchDeleteMenus(c.Request.Context(), req.MenuIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.MenuIDs)})
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
func (h *MenuHandler) ListMenus(c *gin.Context) {
	var req dto.ListMenusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.menuService.ListMenus(c.Request.Context(), &req)
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
func (h *MenuHandler) GetAllMenus(c *gin.Context) {
	resp, err := h.menuService.GetAllMenus(c.Request.Context())
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
func (h *MenuHandler) GetMenuTree(c *gin.Context) {
	resp, err := h.menuService.GetMenuTree(c.Request.Context())
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
func (h *MenuHandler) UpdateMenuStatus(c *gin.Context) {
	var req dto.MenuStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.menuService.UpdateMenuStatus(c.Request.Context(), req.MenuID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

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
// @Summary 获取用户菜单
// @Description 获取当前登录用户的菜单树
// @Tags 用户菜单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.UserMenuResponse} "获取成功"
// @Router /api/v1/user/menus [get]
func (h *UserMenuHandler) GetUserMenu(c *gin.Context) {
	// 从上下文获取用户名和租户编码
	// 这些信息由 Auth 中间件从 JWT token 中提取并设置到 context
	userName := xcontext.GetUserName(c.Request.Context())
	tenantCode := xcontext.GetTenantCode(c.Request.Context())

	if userName == "" || tenantCode == "" {
		response.Error(c, xerr.ErrUnauthorized)
		return
	}

	resp, err := h.userMenuService.GetUserMenu(c.Request.Context(), userName, tenantCode)
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
func (h *UserMenuHandler) GetUserButtons(c *gin.Context) {
	var req dto.MenuDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	// 从上下文获取用户名和租户编码
	// 这些信息由 Auth 中间件从 JWT token 中提取并设置到 context
	userName := xcontext.GetUserName(c.Request.Context())
	tenantCode := xcontext.GetTenantCode(c.Request.Context())

	if userName == "" || tenantCode == "" {
		response.Error(c, xerr.ErrUnauthorized)
		return
	}

	resp, err := h.userMenuService.GetUserButtons(c.Request.Context(), userName, tenantCode, req.MenuID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
