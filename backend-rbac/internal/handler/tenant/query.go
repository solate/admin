package tenant

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetTenant 获取租户详情
// @Summary 获取租户详情
// @Description 根据租户ID获取租户详细信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id query string true "租户ID"
// @Success 200 {object} response.Response{data=dto.TenantInfo} "获取成功"
// @Router /api/v1/tenants/detail [get]
func (h *Handler) GetTenant(c *gin.Context) {
	var req dto.TenantDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetTenantByID(c.Request.Context(), req.TenantID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListTenants 获取租户列表
// @Summary 获取租户列表
// @Description 分页获取租户列表，支持筛选
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param tenant_code query string false "租户编码（模糊查询）"
// @Param name query string false "租户名称（模糊查询）"
// @Param status query int false "状态筛选(1:启用,2:禁用)" Enums(1,2)
// @Success 200 {object} response.Response{data=dto.TenantListResponse} "获取成功"
// @Router /api/v1/tenants [get]
func (h *Handler) ListTenants(c *gin.Context) {
	var req dto.TenantListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ListTenants(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListAllTenants 获取所有启用的租户列表（不分页）
// @Summary 获取所有启用的租户列表
// @Description 获取所有启用状态的租户列表，不分页，用于下拉选择
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]dto.TenantInfo} "获取成功"
// @Router /api/v1/tenants/all [get]
func (h *Handler) ListAllTenants(c *gin.Context) {
	resp, err := h.svc.GetAllTenants(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
