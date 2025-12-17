package middleware

// import (
// 	"admin/pkg/database"
// 	"context"

// 	"admin/internal/constants"

// 	"github.com/gin-gonic/gin"
// )

// // TenantContext 从请求中提取租户 ID 并注入到上下文,
// // 当没有jwt 的时候，依然可以从 X-Tenant-ID 头中提取租户 ID
// func TenantContext() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		tenantID := c.GetString(constants.CtxTenantID)
// 		if tenantID == "" {
// 			tenantID = c.GetHeader("X-Tenant-ID")
// 		}
// 		if tenantID != "" {
// 			// 使用 database.WithTenantID 注入，以便 GORM scopes 自动识别
// 			ctx := database.WithTenantID(c.Request.Context(), tenantID)
// 			// 同时保留 constants 上下文（如果需要的话，但通常 scopes 只看 database 的 key）
// 			ctx = context.WithValue(ctx, constants.CtxTenantID, tenantID)
// 			c.Request = c.Request.WithContext(ctx)
// 		}
// 		c.Next()
// 	}
// }
