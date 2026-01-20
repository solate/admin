package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// TenantHandler 租户处理器
type TenantHandler struct {
	tenantService *service.TenantService
}

// NewTenantHandler 创建租户处理器
func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

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
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req dto.TenantCreateRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.tenantService.CreateTenant(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

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
func (h *TenantHandler) GetTenant(c *gin.Context) {
	var req dto.TenantDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.tenantService.GetTenantByID(c.Request.Context(), req.TenantID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

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
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	var req dto.TenantUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.tenantService.UpdateTenant(c.Request.Context(), req.TenantID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

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
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	var req dto.TenantDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.tenantService.DeleteTenant(c.Request.Context(), req.TenantID); err != nil {
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
func (h *TenantHandler) BatchDeleteTenants(c *gin.Context) {
	var req dto.TenantBatchDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.tenantService.BatchDeleteTenants(c.Request.Context(), req.TenantIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.TenantIDs)})
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
func (h *TenantHandler) ListTenants(c *gin.Context) {
	var req dto.TenantListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.tenantService.ListTenants(c.Request.Context(), &req)
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
func (h *TenantHandler) ListAllTenants(c *gin.Context) {
	resp, err := h.tenantService.GetAllTenants(c.Request.Context())
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
func (h *TenantHandler) UpdateTenantStatus(c *gin.Context) {
	var req dto.TenantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.tenantService.UpdateTenantStatus(c.Request.Context(), req.TenantID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
