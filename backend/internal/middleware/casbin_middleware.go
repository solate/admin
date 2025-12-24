package middleware

import (
	"admin/pkg/casbin"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// CasbinMiddleware creates a middleware that enforces Casbin RBAC policies
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
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
