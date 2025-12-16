package middleware

import (
	"context"

	"admin/internal/constants"

	"github.com/gin-gonic/gin"
)

func TenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString(constants.CtxTenantID)
		if tenantID == "" {
			tenantID = c.GetHeader("X-Tenant-ID")
		}
		if tenantID != "" {
			ctx := context.WithValue(c.Request.Context(), constants.CtxTenantID, tenantID)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
