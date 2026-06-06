package tenant

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreateTenant 创建租户
// @Summary 创建租户
// @Description 超级管理员创建新租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.TenantCreateRequest true "创建租户请求参数"
// @Success 200 {object} response.Response{data=dto.TenantInfo} "创建成功"
// @Router /api/v1/tenants [post]
func (h *Handler) CreateTenant(c *gin.Context) {
	var req dto.TenantCreateRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.CreateTenant(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
