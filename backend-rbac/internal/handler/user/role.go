package user

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// AssignRoles 为用户分配角色
// @Summary 为用户分配角色
// @Description 为指定用户分配角色（覆盖式，会替换用户现有的所有角色）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.AssignRolesRequest true "分配角色请求参数"
// @Success 200 {object} response.Response "分配成功"
// @Router /api/v1/users/roles [put]
func (h *Handler) AssignRoles(c *gin.Context) {
	var req dto.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.roleSvc.AssignRoles(c.Request.Context(), req.UserID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"assigned": true})
}

// GetUserRoles 获取用户的角色列表
// @Summary 获取用户角色
// @Description 获取指定用户的角色列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id query string true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserRolesResponse} "获取成功"
// @Router /api/v1/users/roles [get]
func (h *Handler) GetUserRoles(c *gin.Context) {
	var req dto.UserDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	roles, err := h.roleSvc.GetUserRoles(c.Request.Context(), req.UserID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, roles)
}
