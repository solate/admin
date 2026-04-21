package middleware

import (
	"context"
	"net/http"
	"strings"

	"admin/pkg/utils/jwt"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.ErrorWithHttpCode(c, http.StatusUnauthorized, xerr.ErrUnauthorized)
			c.Abort()
			return
		}

		// 解析 "Bearer <token>" 格式
		parts := strings.SplitN(token, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.ErrorWithHttpCode(c, http.StatusUnauthorized, xerr.ErrTokenInvalid)
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 验证 token（签名、过期、黑名单）
		claims, err := jwtManager.VerifyAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			if err == xerr.ErrTokenExpired {
				response.ErrorWithHttpCode(c, http.StatusUnauthorized, xerr.ErrTokenExpired)
				c.Abort()
				return
			}
			response.ErrorWithHttpCode(c, http.StatusUnauthorized, xerr.ErrTokenInvalid)
			c.Abort()
			return
		}

		log.Debug().
			Str("tenant_id", claims.TenantID).
			Str("tenant_code", claims.TenantCode).
			Str("user_id", claims.UserID).
			Str("user_name", claims.UserName).
			Strs("roles", claims.Roles).
			Msg("[AuthMiddleware] JWT claims resolved")

		requestCtx := SetAuthContext(c.Request.Context(), claims)
		c.Request = c.Request.WithContext(requestCtx)

		log.Debug().
			Str("ctx_tenant_id", xcontext.GetTenantID(requestCtx)).
			Str("ctx_tenant_code", xcontext.GetTenantCode(requestCtx)).
			Msg("[AuthMiddleware] context set")

		c.Next()
	}
}

// SetAuthContext 一次性设置所有认证信息到context
func SetAuthContext(ctx context.Context, claims *jwt.Claims) context.Context {
	if claims == nil {
		return ctx
	}

	ctx = xcontext.SetTenantID(ctx, claims.TenantID)
	ctx = xcontext.SetTenantCode(ctx, claims.TenantCode)
	ctx = xcontext.SetUserID(ctx, claims.UserID)
	ctx = xcontext.SetUserName(ctx, claims.UserName)
	ctx = xcontext.SetRoles(ctx, claims.Roles)
	ctx = xcontext.SetRoleIDs(ctx, claims.RoleIDs)
	ctx = xcontext.SetTokenID(ctx, claims.TokenID)

	return ctx
}
