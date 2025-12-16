package handler

import (
	"admin/internal/constants"
	"admin/pkg/config"
	"admin/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// AuthHandler JWT 认证处理器
// 提供登录、刷新、登出等接口的处理函数
type AuthHandler struct {
	config  *config.Config
	manager *jwt.JWTManager
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(cfg *config.Config, manager *jwt.JWTManager) *AuthHandler {
	return &AuthHandler{
		config:  cfg,
		manager: manager,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Login 处理登录请求
// 实际使用中应该先验证用户名和密码，这里仅作示例
// 生成并返回 token pair
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// TODO: 验证用户名和密码
	// 这里应该调用你的用户服务进行身份验证
	// 演示代码假设验证通过
	// 假设用户 admin, 角色 admin, 租户 tenant-1

	tenantID := "tenant-1"
	userID := "user-2"
	roleID := "user"
	if req.Username == "admin" {
		userID = "user-1"
		roleID = "admin"
	}

	// 生成 token 对（demo 数据）
	tokenPair, err := h.manager.GenerateTokenPair(
		c.Request.Context(),
		tenantID, // 租户 ID
		userID,   // 用户 ID
		roleID,   // 角色 ID
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(200, LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    h.config.JWT.AccessExpire,
		TokenType:    "Bearer",
	})
}

// RefreshRequest 刷新请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 处理刷新 token 请求
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tokenPair, err := h.manager.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(200, LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    h.config.JWT.AccessExpire,
		TokenType:    "Bearer",
	})
}

// Logout 处理登出请求
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取 Claims
	claims, exists := c.Get(constants.CtxClaims)
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	jwtClaims, ok := claims.(*jwt.Claims)
	if !ok {
		c.JSON(500, gin.H{"error": "invalid claims type"})
		return
	}

	// 撤销当前 token (加入黑名单)
	err := h.manager.RevokeToken(c.Request.Context(), jwtClaims.TokenID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to logout"})
		return
	}

	c.JSON(200, gin.H{"message": "logged out successfully"})
}
