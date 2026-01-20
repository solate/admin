package middleware

import (
	"admin/pkg/bodyreader"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// shouldLogBody 判断是否应该记录请求体
// 跳过文件上传、二进制数据等请求
func shouldLogBody(c *gin.Context, bodyStr string) bool {
	// 检查 Content-Type
	contentType := c.GetHeader("Content-Type")

	// 跳过 multipart/form-data（文件上传）
	if strings.Contains(contentType, "multipart/form-data") {
		return false
	}

	// 跳过其他二进制类型
	binaryTypes := []string{
		"application/octet-stream",
		"image/",
		"video/",
		"audio/",
		"application/pdf",
		"application/zip",
		"application/gzip",
	}

	for _, btype := range binaryTypes {
		if strings.Contains(contentType, btype) {
			return false
		}
	}

	// 检查 body 内容，如果包含大量非 ASCII 字符，可能是二进制数据
	if len(bodyStr) > 100 {
		nonASCII := 0
		for _, b := range bodyStr {
			if b > 127 {
				nonASCII++
			}
		}
		// 如果超过 20% 是非 ASCII 字符，认为是二进制数据
		if float64(nonASCII)/float64(len(bodyStr)) > 0.2 {
			return false
		}
	}

	return true
}

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
			// 只有在应该记录的情况下才进行脱敏处理
			if shouldLogBody(c, bodyStr) {
				bodyParams = bodyreader.SanitizeParams(bodyStr)
			}
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
