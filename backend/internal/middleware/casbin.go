package middleware

import (
	"net/http"

	"admin/internal/constants"
	"admin/pkg/casbin"
	"admin/pkg/response"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// CasbinMiddleware creates a middleware that enforces Casbin RBAC policies
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString(constants.CtxUserID)
		tenantID := c.GetString(constants.CtxTenantID)

		if userID == "" {
			response.Error(c, http.StatusUnauthorized, xerr.ErrUnauthorized)
			c.Abort()
			return
		}

		// Fallback to header if tenantID is not in context (e.g. optional in token but required for access)
		if tenantID == "" {
			tenantID = c.GetHeader("X-Tenant-ID")
		}

		if tenantID == "" {
			// In strict multi-tenant mode, tenantID is required.
			// You might want to allow empty tenantID for some global admin routes,
			// but usually those are handled by a separate logic or a "system" tenant.
			response.Error(c, http.StatusForbidden, xerr.New(http.StatusForbidden, "Tenant context missing"))
			c.Abort()
			return
		}

		// Object: Request Path
		obj := c.Request.URL.Path
		// Action: Request Method
		act := c.Request.Method

		// Enforce policy: sub, dom, obj, act
		ok, err := enforcer.Enforce(userID, tenantID, obj, act)
		if err != nil {
			// Log error
			response.Error(c, http.StatusInternalServerError, xerr.ErrInternal)
			c.Abort()
			return
		}

		if !ok {
			response.Error(c, http.StatusForbidden, xerr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
