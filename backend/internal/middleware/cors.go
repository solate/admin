package middleware

import (
	"github.com/gin-gonic/gin"
)

// allowedMethods/allowedHeaders/allowCredentials 是固定的 CORS 策略，不对外暴露配置。
// 只有 AllowedOrigins 按环境配置（开发填 localhost，生产填真实域名）。
const (
	corsAllowMethods     = "GET, POST, PUT, DELETE, OPTIONS"
	corsAllowHeaders     = "Content-Type, Authorization"
	corsAllowCredentials = "true"
)

// CORS 是跨域中间件，只接收允许的来源列表。
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", corsAllowCredentials)
			c.Writer.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
			c.Writer.Header().Set("Access-Control-Allow-Headers", corsAllowHeaders)
		}

		// 预检请求直接返回 204
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed 检查 origin 是否在允许列表中。
func isOriginAllowed(origin string, allowed []string) bool {
	for _, o := range allowed {
		if o == origin {
			return true
		}
	}
	return false
}
