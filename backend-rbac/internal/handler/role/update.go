package role

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdateRoleRequest true "更新角色请求参数"
// @Success 200 {object} response.Response{data=dto.RoleInfo} "更新成功"
// @Router /api/v1/roles [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.UpdateRole(c.Request.Context(), req.RoleID, &req)
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
// @Param request body dto.RoleStatusRequest true "更新状态请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/roles/status [put]
func (h *Handler) UpdateRoleStatus(c *gin.Context) {
	var req dto.RoleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.UpdateRoleStatus(c.Request.Context(), req.RoleID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
