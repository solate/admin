package middleware

import (
	"errors"
	"strings"

	"admin/pkg/constants"
	"admin/pkg/database"
	"admin/pkg/jwt"
	"admin/pkg/response"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// Auth JWT认证中间件
// 说明：
// - 从请求头 Authorization 提取 Bearer token
// - 调用 JWTManager 验证签名/过期，并检查是否命中黑名单
// - 验证通过后将用户信息注入到 Gin 上下文，供后续处理使用
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.Error(c, xerr.ErrUnauthorized)
			c.Abort()
			return
		}

		// 解析 "Bearer <token>" 格式
		parts := strings.SplitN(token, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, xerr.ErrTokenInvalid)
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 验证 token（签名、过期、黑名单）
		claims, err := jwtManager.VerifyAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			// 错误映射：过期 / 黑名单 / 其他无效
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.Error(c, xerr.ErrTokenExpired)
			} else if errors.Is(err, jwt.ErrTokenBlacklisted) {
				response.Error(c, xerr.ErrTokenInvalid)
			} else {
				response.Error(c, xerr.ErrTokenInvalid)
			}
			c.Abort()
			return
		}

		// 注入上下文，便于业务层读取用户信息
		c.Set(constants.CtxTenantID, claims.TenantID)
		c.Set(constants.CtxTenantCode, claims.TenantCode)
		c.Set(constants.CtxUserID, claims.UserID)
		c.Set(constants.CtxUserName, claims.UserName)
		c.Set(constants.CtxRoleType, claims.RoleType)
		c.Set(constants.CtxRoles, claims.Roles)
		c.Set(constants.CtxClaims, claims)
		c.Set(constants.CtxTokenID, claims.TokenID)

		// 注入 GORM Scope 所需的 TenantCode 到 request.Context
		if claims.TenantID != "" {
			ctx := database.WithTenantID(c.Request.Context(), claims.TenantID)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}
