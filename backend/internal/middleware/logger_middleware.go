package middleware

import (
	"admin/pkg/bodyreader"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger 请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 提取请求参数（POST/PUT 等需要读取 Body）
		var bodyParams string
		if c.Request.Body != nil {
			bodyStr, restoredBody := bodyreader.ReadBodyString(c.Request.Body)
			if restoredBody != nil {
				c.Request.Body = restoredBody
			}
			bodyParams = bodyreader.SanitizeParams(bodyStr)
		}

		// 处理请求
		c.Next()

		// 结束时间
		duration := time.Since(start)

		// 构建日志字段
		event := log.Info().
			Str("request_id", GetRequestID(c)).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Str("ip", c.ClientIP()).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Str("user_agent", c.Request.UserAgent())

		// 如果有 body 参数，添加到日志中（已脱敏）
		if bodyParams != "" {
			event.Str("body", bodyParams)
		}

		event.Msg("HTTP Request")

		// 如果有错误，记录错误日志
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Error().
					Str("request_id", GetRequestID(c)).
					Err(e.Err).
					Msg("Request Error")
			}
		}
	}
}
