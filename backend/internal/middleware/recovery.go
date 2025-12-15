package middleware

import (
	"admin/pkg/response"
	"admin/pkg/xerr"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Recovery Panic恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := string(debug.Stack())

				// 记录panic日志（包含堆栈）
				log.Error().
					Str("request_id", GetRequestID(c)).
					Interface("error", err).
					Str("path", c.Request.URL.Path).
					Str("stack", stack).
					Msg("Panic Recovered")

				// // 打印到标准输出便于调试
				// fmt.Printf("\n=== PANIC ===\n")
				// fmt.Printf("Error: %v\n", err)
				// fmt.Printf("Path: %s\n", c.Request.URL.Path)
				// fmt.Printf("Stack:\n%s\n", stack)
				// fmt.Printf("=============\n\n")

				// 返回错误响应
				response.Error(c, http.StatusInternalServerError, xerr.ErrInternal)
				c.Abort()
			}
		}()

		c.Next()
	}
}
