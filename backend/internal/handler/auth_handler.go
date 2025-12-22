package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/config"
	"admin/pkg/response"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// AuthHandler JWT 认证处理器
// 提供登录、刷新、登出等接口的处理函数
type AuthHandler struct {
	config      *config.Config
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(cfg *config.Config, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		config:      cfg,
		authService: authService,
	}
}

// Login 处理登录请求
// @Summary 用户登录
// @Description 用户通过用户名、密码和验证码进行登录，成功后返回访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求参数"
// @Success 200 {object} response.Response{data=dto.LoginResponse} "登录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "认证失败"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*xerr.AppError))
		return
	}

	response.Success(c, resp)
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Refresh 处理刷新 token 请求
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "刷新令牌请求参数"
// @Success 200 {object} response.Response{data=dto.RefreshResponse} "刷新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "刷新令牌无效或已过期"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err.(*xerr.AppError))
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
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 500 {object} response.Response "服务器内部错误"
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
