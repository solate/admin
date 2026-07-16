package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"admin/pkg/response"
	"admin/pkg/xerr"
)

// Recovery 是 panic 恢复中间件，捕获 panic 后记录栈并返回统一错误响应。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := c.Request.Context()
				stack := string(debug.Stack())
				slog.ErrorContext(ctx, "panic recovered",
					slog.Any("error", err),
					slog.String("stack", stack),
				)
				// 通过 response.Error 返回统一 envelope，request_id 自动带上
				response.Error(c, xerr.ErrInternal)
			}
		}()
		c.Next()
	}
}
