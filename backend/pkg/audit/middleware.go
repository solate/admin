package audit

import (
	"admin/pkg/bodyreader"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计日志中间件
// 功能：提取请求信息并存入 context，供日志记录使用
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 提取并存储客户端信息
		clientInfo := &ClientInfo{
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}
		ctx = WithClientInfo(ctx, clientInfo)

		// 提取并存储请求信息
		requestInfo := &RequestInfo{
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			Params: extractParams(c),
		}
		ctx = WithRequestInfo(ctx, requestInfo)

		// 更新请求的 context
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// extractParams 提取请求参数（带脱敏）
func extractParams(c *gin.Context) string {
	if c.Request.Method == "GET" {
		return c.Request.URL.RawQuery
	}
	bodyStr, restoredBody := bodyreader.ReadBodyString(c.Request.Body)
	if restoredBody != nil {
		c.Request.Body = restoredBody
	}
	return bodyreader.SanitizeParams(bodyStr)
}
