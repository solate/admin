package tenant

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeleteTenant 删除租户
// @Summary 删除租户
// @Description 软删除租户（只有租户下无用户时才能删除）
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.TenantDeleteRequest true "删除租户请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/tenants [delete]
func (h *Handler) DeleteTenant(c *gin.Context) {
	var req dto.TenantDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeleteTenant(c.Request.Context(), req.TenantID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// BatchDeleteTenants 批量删除租户
// @Summary 批量删除租户
// @Description 批量软删除租户（只有租户下无用户时才能删除）
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.TenantBatchDeleteRequest true "批量删除请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/tenants/batch-delete [delete]
func (h *Handler) BatchDeleteTenants(c *gin.Context) {
	var req dto.TenantBatchDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.BatchDeleteTenants(c.Request.Context(), req.TenantIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.TenantIDs)})
}
