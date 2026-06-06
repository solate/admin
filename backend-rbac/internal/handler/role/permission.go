package role

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// AssignPermissions 为角色分配权限（菜单+按钮）
// @Summary 为角色分配权限
// @Description 为指定角色分配菜单和按钮权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.AssignPermissionsRequest true "分配权限请求参数"
// @Success 200 {object} response.Response "分配成功"
// @Router /api/v1/roles/permissions [put]
func (h *Handler) AssignPermissions(c *gin.Context) {
	var req dto.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.AssignPermissions(c.Request.Context(), req.RoleID, &req); err != nil {
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
// @Param role_id query string true "角色ID"
// @Success 200 {object} response.Response{data=dto.RolePermissionsResponse} "获取成功"
// @Router /api/v1/roles/permissions [get]
func (h *Handler) GetRolePermissions(c *gin.Context) {
	var req dto.RoleDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	permissions, err := h.svc.GetRolePermissions(c.Request.Context(), req.RoleID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, permissions)
}

// AssignMenus 为角色分配菜单（已弃用，保留向后兼容）
// @Summary 为角色分配菜单（已弃用）
// @Deprecated
// @Description 为指定角色分配菜单权限。已弃用，请使用 AssignPermissions 接口
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path string true "角色ID"
// @Param request body dto.AssignPermissionsRequest true "菜单ID列表"
// @Success 200 {object} response.Response "分配成功"
// @Router /api/v1/roles/{role_id}/menus [put]
// Deprecated: 使用 AssignPermissions 代替
func (h *Handler) AssignMenus(c *gin.Context) {
	h.AssignPermissions(c)
}

// GetRoleMenus 获取角色的菜单权限（已弃用，保留向后兼容）
// @Summary 获取角色菜单权限（已弃用）
// @Deprecated
// @Description 获取指定角色的菜单权限。已弃用，请使用 GetRolePermissions 接口
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id query string true "角色ID"
// @Success 200 {object} response.Response{data=dto.RolePermissionsResponse} "获取成功"
// @Router /api/v1/roles/menus [get]
// Deprecated: 使用 GetRolePermissions 代替
func (h *Handler) GetRoleMenus(c *gin.Context) {
	var req dto.RoleDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	permissions, err := h.svc.GetRolePermissions(c.Request.Context(), req.RoleID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 旧接口只返回菜单ID
	response.Success(c, &dto.RolePermissionsResponse{MenuPermIDs: permissions.MenuPermIDs})
}
