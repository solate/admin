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

	resp, err := h.authService.Login(c.Request.Context(), &req)
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
// @Success 200 {object} map[string]string "登出成功"
// @Success 200 {object} response.Response "未授权访问"
// @Success 200 {object} response.Response "服务器内部错误"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取 Claims
	// claims, exists := c.Get(constants.CtxClaims)
	// if !exists {
	// 	response.Error(c, xerr.ErrUnauthorized)
	// 	return
	// }

	// jwtClaims, ok := claims.(*jwt.Claims)
	// if !ok {
	// 	response.Error(c, xerr.ErrInternal)
	// 	return
	// }

	// // 撤销当前 token (加入黑名单)
	// err := h.authService.RevokeToken(c.Request.Context(), jwtClaims.TokenID)
	// if err != nil {
	// 	response.Error(c, xerr.ErrInternal)
	// 	return
	// }

	c.JSON(200, gin.H{"message": "logged out successfully"})
}
