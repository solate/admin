package middleware

import (
	"context"
	"strings"

	"admin/pkg/jwt"
	"admin/pkg/response"
	"admin/pkg/xcontext"
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
			response.Error(c, err)
			c.Abort()
			return
		}

		// 将认证信息注入到 request.Context 中，供 service 层使用
		// SetAuthContext 已经包含了租户ID设置，与database包使用相同的key
		requestCtx := SetAuthContext(c.Request.Context(), claims)
		// 更新 request 的 context
		c.Request = c.Request.WithContext(requestCtx)

		c.Next()
	}
}

// SetAuthContext 一次性设置所有认证信息到context
func SetAuthContext(ctx context.Context, claims *jwt.Claims) context.Context {
	if claims == nil {
		return ctx
	}

	// 设置租户信息
	ctx = xcontext.SetTenantID(ctx, claims.TenantID)
	ctx = xcontext.SetTenantCode(ctx, claims.TenantCode)

	// 设置用户信息
	ctx = xcontext.SetUserID(ctx, claims.UserID)
	ctx = xcontext.SetUserName(ctx, claims.UserName)
	ctx = xcontext.SetRoles(ctx, claims.Roles)
	ctx = xcontext.SetTokenID(ctx, claims.TokenID)

	return ctx
}
