package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler JWT 认证处理器
// 提供登录、选择租户、刷新、登出等接口的处理函数
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 处理登录请求
// @Summary 用户登录
// @Description 用户通过用户名、密码和验证码进行登录。如果用户只有一个租户，直接返回token；如果有多个租户，返回租户列表让用户选择。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求参数"
// @Success 200 {object} response.Response{data=dto.LoginResponse} "登录成功或需要选择租户"
// @Success 200 {object} response.Response "请求参数错误"
// @Success 200 {object} response.Response "认证失败"
// @Success 200 {object} response.Response "服务器内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), c.Request, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// Refresh 处理刷新 token 请求
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "刷新令牌请求参数"
// @Success 200 {object} response.Response{data=dto.RefreshResponse} "刷新成功"
// @Success 200 {object} response.Response "请求参数错误"
// @Success 200 {object} response.Response "刷新令牌无效或已过期"
// @Success 200 {object} response.Response "服务器内部错误"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// Logout 处理登出请求
// @Summary 用户登出
// @Description 用户登出，将当前访问令牌加入黑名单
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response "登出成功"
// @Success 200 {object} response.Response "未授权访问"
// @Success 200 {object} response.Response "服务器内部错误"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	if err := h.authService.Logout(c.Request.Context(), c.Request); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
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
// @Success 200 {object} response.Response "请求参数错误"
// @Success 200 {object} response.Response "租户不存在"
// @Success 200 {object} response.Response "无该租户访问权限"
// @Success 200 {object} response.Response "服务器内部错误"
// @Router /auth/switch-tenant [post]
func (h *AuthHandler) SwitchTenant(c *gin.Context) {
	var req dto.SwitchTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.authService.SwitchTenant(c.Request.Context(), &req)
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
// @Success 200 {object} response.Response "服务器内部错误"
// @Router /auth/available-tenants [get]
func (h *AuthHandler) GetAvailableTenants(c *gin.Context) {
	resp, err := h.authService.GetAvailableTenants(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
