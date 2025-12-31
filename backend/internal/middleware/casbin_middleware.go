package middleware

import (
	"admin/pkg/casbin"
	"admin/pkg/database"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// CasbinMiddleware creates a middleware that enforces Casbin RBAC policies
// 说明：
// - 超管（super_admin 角色）跳过权限检查和租户检查
// - 普通用户通过 Casbin 验证 (username, tenantCode, path, method)，并受租户 scope 限制
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 超管跳过权限检查和租户检查
		if xcontext.IsSuperAdmin(c.Request.Context()) {
			ctx := database.SkipTenantCheck(c.Request.Context())
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		userName := xcontext.GetUserName(c.Request.Context())
		tenantCode := xcontext.GetTenantCode(c.Request.Context())

		if userName == "" || tenantCode == "" {
			response.Error(c, xerr.ErrUnauthorized)
			c.Abort()
			return
		}

		// Object: Request Path
		obj := c.Request.URL.Path
		// Action: Request Method
		act := c.Request.Method

		// Enforce policy: sub, dom, obj, act
		ok, err := enforcer.Enforce(userName, tenantCode, obj, act)
		if err != nil {
			// Log error
			response.Error(c, xerr.ErrInternal)
			c.Abort()
			return
		}

		if !ok {
			response.Error(c, xerr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperAdminMiddleware 超管中间件
// 只有超级管理员才能访问通过此中间件的路由
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !xcontext.IsSuperAdmin(c.Request.Context()) {
			response.Error(c, xerr.ErrForbidden)
			c.Abort()
			return
		}

		ctx := database.SkipTenantCheck(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
