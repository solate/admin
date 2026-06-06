package auth

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// Refresh 处理刷新 token 请求
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "刷新令牌请求参数"
// @Success 200 {object} response.Response{data=dto.RefreshResponse} "刷新成功"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// SwitchTenant 处理切换租户请求
// @Summary 切换租户
// @Description 用户切换到指定租户，超管可切换任意租户，普通用户只能切换已加入的租户
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.SwitchTenantRequest true "切换租户请求参数"
// @Success 200 {object} response.Response{data=dto.LoginResponse} "切换成功"
// @Router /api/v1/auth/switch-tenant [post]
func (h *Handler) SwitchTenant(c *gin.Context) {
	var req dto.SwitchTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.SwitchTenant(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetAvailableTenants 获取可切换的租户列表
// @Summary 获取可切换租户列表
// @Description 获取当前用户可以切换的租户列表，超管获取所有租户，普通用户获取已加入的租户
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.AvailableTenantsResponse} "获取成功"
// @Router /api/v1/auth/available-tenants [get]
func (h *Handler) GetAvailableTenants(c *gin.Context) {
	resp, err := h.svc.GetAvailableTenants(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
