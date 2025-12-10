package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		duration := time.Since(start)

		// 日志记录
		log.Info().
			Str("request_id", GetRequestID(c)).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Str("ip", c.ClientIP()).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP Request")

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
