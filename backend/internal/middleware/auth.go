package middleware

import (
	"errors"
	"net/http"
	"strings"

	"admin/internal/constants"
	apierrors "admin/pkg/errors"
	"admin/pkg/jwt"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// Auth JWT认证中间件
// 说明：
// - 从请求头 Authorization 提取 Bearer token
// - 调用 JWTManager 验证签名/过期，并检查是否命中黑名单
// - 验证通过后将用户信息注入到 Gin 上下文，供后续处理使用
func Auth(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.Error(c, http.StatusUnauthorized, apierrors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 解析 "Bearer <token>" 格式
		parts := strings.SplitN(token, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, apierrors.ErrTokenInvalid)
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 验证 token（签名、过期、黑名单）
		claims, err := jwtManager.VerifyAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			// 错误映射：过期 / 黑名单 / 其他无效
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.Error(c, http.StatusUnauthorized, apierrors.ErrTokenExpired)
			} else if errors.Is(err, jwt.ErrTokenBlacklisted) {
				response.Error(c, http.StatusUnauthorized, apierrors.ErrTokenInvalid)
			} else {
				response.Error(c, http.StatusUnauthorized, apierrors.ErrTokenInvalid)
			}
			c.Abort()
			return
		}

		// 注入上下文，便于业务层读取用户信息
		c.Set(constants.CtxUserID, claims.UserID)
		c.Set(constants.CtxTenantID, claims.TenantID)
		c.Set(constants.CtxRoleID, claims.RoleID)
		c.Set(constants.CtxClaims, claims)
		c.Set(constants.CtxTokenID, claims.TokenID)

		c.Next()
	}
}
