package middleware

import (
	"admin/pkg/casbin"
	"admin/pkg/constants"
	"admin/pkg/database"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// CasbinMiddleware Casbin权限中间件
// 职责：
//  1. 超管（super_admin 角色）：跳过权限检查和数据层租户检查
//  2. 普通用户：通过 Casbin 验证 (username, tenantCode, path, method)
//
// 执行顺序：RoleMiddleware（租户校验） → CasbinMiddleware（权限校验）
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// 超管跳过权限检查和租户检查
		if xcontext.HasRole(ctx, constants.SuperAdmin) {
			ctx := database.SkipTenantCheck(ctx)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		userName := xcontext.GetUserName(ctx)
		tenantCode := xcontext.GetTenantCode(ctx)

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
