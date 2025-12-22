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
// 实际使用中应该先验证用户名和密码，这里仅作示例
// 生成并返回 token pair
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.authService.Login(c, &req)
	if err != nil {
		response.Error(c, err.(*xerr.AppError))
		return
	}

	response.Success(c, resp)
}

// RefreshRequest 刷新请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 处理刷新 token 请求
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.authService.RefreshToken(c, req.RefreshToken)
	if err != nil {
		response.Error(c, err.(*xerr.AppError))
		return
	}

	response.Success(c, resp)
}

// Logout 处理登出请求
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
