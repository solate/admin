package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"strconv"

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
// @Router /menus [post]
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
// @Param menu_id path string true "菜单ID"
// @Success 200 {object} response.Response{data=dto.MenuInfo} "获取成功"
// @Router /menus/:menu_id [get]
func (h *MenuHandler) GetMenu(c *gin.Context) {
	menuID := c.Param("menu_id")
	if menuID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.menuService.GetMenuByID(c.Request.Context(), menuID)
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
// @Param menu_id path string true "菜单ID"
// @Param request body dto.UpdateMenuRequest true "更新菜单请求参数"
// @Success 200 {object} response.Response{data=dto.MenuInfo} "更新成功"
// @Router /menus/:menu_id [put]
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	menuID := c.Param("menu_id")
	if menuID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	var req dto.UpdateMenuRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.menuService.UpdateMenu(c.Request.Context(), menuID, &req)
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
// @Param menu_id path string true "菜单ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /menus/:menu_id [delete]
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	menuID := c.Param("menu_id")
	if menuID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.menuService.DeleteMenu(c.Request.Context(), menuID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
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
// @Param type query string false "类型筛选(MENU:菜单, BUTTON:按钮)"
// @Param status query int false "状态筛选(1:显示, 2:隐藏)"
// @Success 200 {object} response.Response{data=dto.ListMenusResponse} "获取成功"
// @Router /menus [get]
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
// @Router /menus/all [get]
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
// @Router /menus/tree [get]
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
// @Param menu_id path string true "菜单ID"
// @Param status path int true "状态(1:显示, 2:隐藏)"
// @Success 200 {object} response.Response "更新成功"
// @Router /menus/:menu_id/status/:status [put]
func (h *MenuHandler) UpdateMenuStatus(c *gin.Context) {
	menuID := c.Param("menu_id")
	if menuID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	statusStr := c.Param("status")
	if statusStr == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.menuService.UpdateMenuStatus(c.Request.Context(), menuID, status); err != nil {
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
// @Router /user/menus [get]
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
// @Param menu_id path string true "菜单ID"
// @Success 200 {object} response.Response{data=dto.UserButtonsResponse} "获取成功"
// @Router /user/buttons/{menu_id} [get]
func (h *UserMenuHandler) GetUserButtons(c *gin.Context) {
	menuID := c.Param("menu_id")
	if menuID == "" {
		response.Error(c, xerr.ErrInvalidParams)
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

	resp, err := h.userMenuService.GetUserButtons(c.Request.Context(), userName, tenantCode, menuID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

