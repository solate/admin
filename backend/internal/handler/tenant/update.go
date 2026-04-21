package tenant

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// UpdateTenant 更新租户
// @Summary 更新租户
// @Description 更新租户信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.TenantUpdateRequest true "更新租户请求参数"
// @Success 200 {object} response.Response{data=dto.TenantInfo} "更新成功"
// @Router /api/v1/tenants [put]
func (h *Handler) UpdateTenant(c *gin.Context) {
	var req dto.TenantUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.UpdateTenant(c.Request.Context(), req.TenantID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateTenantStatus 更新租户状态
// @Summary 更新租户状态
// @Description 启用或禁用租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.TenantStatusRequest true "更新状态请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/tenants/status [put]
func (h *Handler) UpdateTenantStatus(c *gin.Context) {
	var req dto.TenantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.UpdateTenantStatus(c.Request.Context(), req.TenantID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
