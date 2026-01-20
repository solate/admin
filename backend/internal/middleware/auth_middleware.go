package middleware

import (
	"context"
	"net/http"

	"admin/pkg/jwt"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
			response.ErrorWithHttpCode(c, http.StatusUnauthorized, xerr.ErrUnauthorized)
			c.Abort()
			return
		}

		// // 解析 "Bearer <token>" 格式
		// parts := strings.SplitN(token, " ", 2)
		// if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		// 	response.ErrorWithHttpCode(c, http.StatusUnauthorized, xerr.ErrTokenInvalid)
		// 	c.Abort()
		// 	return
		// }
		// tokenString := parts[1]

		// 验证 token（签名、过期、黑名单）
		claims, err := jwtManager.VerifyAccessToken(c.Request.Context(), token)
		if err != nil {
			// VerifyAccessToken 已经将 JWT 错误转换为 xerr 错误
			// 这里只需要判断是否是 xerr.ErrTokenExpired 即可
			if err == xerr.ErrTokenExpired {
				response.ErrorWithHttpCode(c, http.StatusUnauthorized, xerr.ErrTokenExpired)
				c.Abort()
				return
			}
			// 其他未知错误也返回 token 无效
			response.ErrorWithHttpCode(c, http.StatusUnauthorized, xerr.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 调试日志：打印 JWT claims 中的租户信息
		log.Debug().
			Str("tenant_id", claims.TenantID).
			Str("tenant_code", claims.TenantCode).
			Str("user_id", claims.UserID).
			Str("user_name", claims.UserName).
			Strs("roles", claims.Roles).
			Msg("[AuthMiddleware] JWT claims 解析成功")

		// 将认证信息注入到 request.Context 中，供 service 层使用
		// SetAuthContext 已经包含了租户ID设置，与database包使用相同的key
		requestCtx := SetAuthContext(c.Request.Context(), claims)
		// 更新 request 的 context
		c.Request = c.Request.WithContext(requestCtx)

		// 调试日志：验证上下文是否设置成功
		log.Debug().
			Str("ctx_tenant_id", xcontext.GetTenantID(requestCtx)).
			Str("ctx_tenant_code", xcontext.GetTenantCode(requestCtx)).
			Msg("[AuthMiddleware] 上下文设置完成")

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
