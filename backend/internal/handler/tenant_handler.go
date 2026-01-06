package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xerr"
	"strconv"

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
// @Success 200 {object} response.Response{data=dto.TenantResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
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
// @Param tenant_id path string true "租户ID"
// @Success 200 {object} response.Response{data=dto.TenantResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tenants/{tenant_id} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.tenantService.GetTenantByID(c.Request.Context(), tenantID)
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
// @Param tenant_id path string true "租户ID"
// @Param request body dto.TenantUpdateRequest true "更新租户请求参数"
// @Success 200 {object} response.Response{data=dto.TenantResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tenants/{tenant_id} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	var req dto.TenantUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.tenantService.UpdateTenant(c.Request.Context(), tenantID, &req)
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
// @Param tenant_id path string true "租户ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tenants/{tenant_id} [delete]
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.tenantService.DeleteTenant(c.Request.Context(), tenantID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
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
// @Param code query string false "租户编码（模糊查询）"
// @Param name query string false "租户名称（模糊查询）"
// @Param status query int false "状态筛选" Enums(1,2)
// @Success 200 {object} response.Response{data=dto.TenantListResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 500 {object} response.Response "服务器内部错误"
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

// UpdateTenantStatus 更新租户状态
// @Summary 更新租户状态
// @Description 启用或禁用租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id path string true "租户ID"
// @Param status path int true "状态(1:启用,2:禁用)"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tenants/{tenant_id}/status/{status} [put]
func (h *TenantHandler) UpdateTenantStatus(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
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

	if err := h.tenantService.UpdateTenantStatus(c.Request.Context(), tenantID, status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
