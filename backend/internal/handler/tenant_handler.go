package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"admin/pkg/xerr"
)

type TenantHandler struct {
	tenantService *service.TenantService
}

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
// @Param request body dto.TenantCreateRequest true "创建租户请求"
// @Success 200 {object} dto.TenantResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/super/tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req dto.TenantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, http.StatusBadRequest, xerr.ErrBadRequest.Code, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.tenantService.CreateTenant(c.Request.Context(), &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// GetTenant 获取租户详情
// @Summary 获取租户详情
// @Description 根据租户ID获取租户详细信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID"
// @Success 200 {object} dto.TenantResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/super/tenants/{tenant_id} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.ErrorWithMessage(c, http.StatusBadRequest, xerr.ErrBadRequest.Code, "租户ID不能为空")
		return
	}

	result, err := h.tenantService.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// UpdateTenant 更新租户
// @Summary 更新租户
// @Description 更新租户信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID"
// @Param request body dto.TenantUpdateRequest true "更新租户请求"
// @Success 200 {object} dto.TenantResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/super/tenants/{tenant_id} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.ErrorWithMessage(c, http.StatusBadRequest, xerr.ErrBadRequest.Code, "租户ID不能为空")
		return
	}

	var req dto.TenantUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, http.StatusBadRequest, xerr.ErrBadRequest.Code, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.tenantService.UpdateTenant(c.Request.Context(), tenantID, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// DeleteTenant 删除租户
// @Summary 删除租户
// @Description 软删除租户（只有租户下无用户时才能删除）
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/super/tenants/{tenant_id} [delete]
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.ErrorWithMessage(c, http.StatusBadRequest, xerr.ErrBadRequest.Code, "租户ID不能为空")
		return
	}

	err := h.tenantService.DeleteTenant(c.Request.Context(), tenantID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListTenants 获取租户列表
// @Summary 获取租户列表
// @Description 分页获取租户列表，支持筛选
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param code query string false "租户编码（模糊查询）"
// @Param name query string false "租户名称（模糊查询）"
// @Param status query int false "状态筛选" Enums(1,2)
// @Success 200 {object} dto.TenantListResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/super/tenants [get]
func (h *TenantHandler) ListTenants(c *gin.Context) {
	var req dto.TenantListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorWithMessage(c, http.StatusBadRequest, xerr.ErrBadRequest.Code, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.tenantService.ListTenants(c.Request.Context(), &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// UpdateTenantStatus 更新租户状态
// @Summary 更新租户状态
// @Description 启用或禁用租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID"
// @Param request body dto.TenantStatusRequest true "状态更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/super/tenants/{tenant_id}/status [put]
func (h *TenantHandler) UpdateTenantStatus(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.ErrorWithMessage(c, http.StatusBadRequest, xerr.ErrBadRequest.Code, "租户ID不能为空")
		return
	}

	var req dto.TenantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, http.StatusBadRequest, xerr.ErrBadRequest.Code, "请求参数错误: "+err.Error())
		return
	}

	err := h.tenantService.UpdateTenantStatus(c.Request.Context(), tenantID, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

func writeServiceError(c *gin.Context, err error) {
	appErr, ok := err.(*xerr.AppError)
	if !ok {
		response.Error(c, http.StatusInternalServerError, xerr.ErrInternal)
		return
	}
	response.Error(c, httpStatusFromAppError(appErr), appErr)
}

func httpStatusFromAppError(err *xerr.AppError) int {
	switch err.Code {
	case xerr.ErrBadRequest.Code, xerr.ErrInvalidateParam.Code:
		return http.StatusBadRequest
	case xerr.ErrUnauthorized.Code:
		return http.StatusUnauthorized
	case xerr.ErrForbidden.Code:
		return http.StatusForbidden
	case xerr.ErrNotFound.Code:
		return http.StatusNotFound
	case xerr.ErrConflict.Code:
		return http.StatusConflict
	case xerr.ErrTooManyRequests.Code:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
