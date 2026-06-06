package role

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Description 根据ID获取角色详情
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id query string true "角色ID"
// @Success 200 {object} response.Response{data=dto.RoleInfo} "获取成功"
// @Router /api/v1/roles/detail [get]
func (h *Handler) GetRole(c *gin.Context) {
	var req dto.RoleDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetRoleByID(c.Request.Context(), req.RoleID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
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
// @Param role_name query string false "角色名称(模糊匹配)"
// @Param role_code query string false "角色编码(模糊匹配)"
// @Param status query int false "状态筛选(1:启用,2:禁用)" Enums(1, 2)
// @Success 200 {object} response.Response{data=dto.ListRolesResponse} "获取成功"
// @Router /api/v1/roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	var req dto.ListRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ListRoles(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetAllRoles 获取所有角色（不分页）
// @Summary 获取所有角色
// @Description 获取所有角色列表（不分页），支持按关键词和状态筛选
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_name query string false "角色名称(模糊匹配)"
// @Param role_code query string false "角色编码(模糊匹配)"
// @Param status query int false "状态筛选(1:启用,2:禁用)" Enums(1, 2)
// @Success 200 {object} response.Response{data=dto.GetAllRolesResponse} "获取成功"
// @Router /api/v1/roles/all [get]
func (h *Handler) GetAllRoles(c *gin.Context) {
	var req dto.GetAllRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetAllRoles(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
