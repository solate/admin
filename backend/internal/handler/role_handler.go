package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xerr"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建新的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateRoleRequest true "创建角色请求参数"
// @Success 200 {object} response.Response{data=dto.RoleResponse} "创建成功"
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.roleService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Description 根据ID获取角色详情
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path string true "角色ID"
// @Success 200 {object} response.Response{data=dto.RoleResponse} "获取成功"
// @Router /roles/:role_id [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	roleID := c.Param("role_id")
	if roleID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.roleService.GetRoleByID(c.Request.Context(), roleID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path string true "角色ID"
// @Param request body dto.UpdateRoleRequest true "更新角色请求参数"
// @Success 200 {object} response.Response{data=dto.RoleResponse} "更新成功"
// @Router /roles/:role_id [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	roleID := c.Param("role_id")
	if roleID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.roleService.UpdateRole(c.Request.Context(), roleID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除角色（软删除）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path string true "角色ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /roles/:role_id [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleID := c.Param("role_id")
	if roleID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.roleService.DeleteRole(c.Request.Context(), roleID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// ListRoles 获取角色列表
// @Summary 获取角色列表
// @Description 分页获取角色列表，支持按关键词和状态筛选
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "关键词搜索（角色名称/编码）"
// @Param status query int false "状态筛选(1:启用,2:禁用)"
// @Success 200 {object} response.Response{data=dto.ListRolesResponse} "获取成功"
// @Router /roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	var req dto.ListRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.roleService.ListRoles(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateRoleStatus 更新角色状态
// @Summary 更新角色状态
// @Description 更新角色状态
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path string true "角色ID"
// @Param status path int true "状态(1:启用,2:禁用)"
// @Success 200 {object} response.Response "更新成功"
// @Router /roles/:role_id/status/:status [put]
func (h *RoleHandler) UpdateRoleStatus(c *gin.Context) {
	roleID := c.Param("role_id")
	if roleID == "" {
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

	if err := h.roleService.UpdateRoleStatus(c.Request.Context(), roleID, status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

// AssignPermissions 为角色分配权限（菜单+按钮）
// @Summary 为角色分配权限
// @Description 为指定角色分配菜单和按钮权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path string true "角色ID"
// @Param request body dto.AssignPermissionsRequest true "权限列表"
// @Success 200 {object} response.Response "分配成功"
// @Router /roles/:role_id/permissions [put]
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	roleID := c.Param("role_id")
	if roleID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	var req dto.AssignPermissionsRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.roleService.AssignPermissions(c.Request.Context(), roleID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"assigned": true})
}

// GetRolePermissions 获取角色的权限（菜单+按钮）
// @Summary 获取角色权限
// @Description 获取指定角色的菜单和按钮权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path string true "角色ID"
// @Success 200 {object} response.Response{data=dto.RolePermissionsResponse} "获取成功"
// @Router /roles/:role_id/permissions [get]
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("role_id")
	if roleID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	permissions, err := h.roleService.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, permissions)
}

// AssignMenus 为角色分配菜单（已弃用，保留向后兼容）
// Deprecated: 使用 AssignPermissions 代替
func (h *RoleHandler) AssignMenus(c *gin.Context) {
	h.AssignPermissions(c)
}

// GetRoleMenus 获取角色的菜单权限（已弃用，保留向后兼容）
// Deprecated: 使用 GetRolePermissions 代替
func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	roleID := c.Param("role_id")
	if roleID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	permissions, err := h.roleService.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 旧接口只返回菜单ID
	response.Success(c, &dto.RoleMenusResponse{MenuIDs: permissions.MenuIDs})
}
