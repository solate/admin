package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 是日志中间件，记录每个请求的访问日志与错误日志。
// request_id 由 xslog 的 ContextExtractor 自动注入，此处无需手动添加。
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		ctx := c.Request.Context()

		// 访问日志（request_id 靠 xslog extractor 自动带）
		log.InfoContext(ctx, "request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", duration),
			slog.String("client_ip", c.ClientIP()),
		)

		// 错误日志（如果 handler 通过 c.Error 压入了错误）
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.ErrorContext(ctx, "request error", slog.Any("error", e.Err))
			}
		}
	}
}
