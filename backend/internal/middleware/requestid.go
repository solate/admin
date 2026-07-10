// Package middleware 提供 gin 中间件：request-id、logger、recovery、cors。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"admin/pkg/xslog"
)

const RequestIDKey = "X-Request-ID"

// RequestID 是 request-id 中间件，优先从 W3C traceparent 提取 trace-id，
// 兜底取 X-Request-ID header，最后生成 uuid。
//
// id 写入两处，各由不同消费方读取：
//   - gin.Context（键 RequestIDKey）：供下游处理器按需读取；
//   - request context（xslog.WithField）：供 xslog 内置 extractor 自动把 request_id
//     注入该请求后续每行日志，无需在日志调用处手写。
//
// 同时写入 X-Request-ID 响应头供前端排障——注意 request_id 不进响应 body，
// response envelope 只保留 {code, message, data}。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var id string

		// 1. 优先解析 traceparent（W3C Trace Context，格式：00-<trace-id>-<span-id>-<flags>）
		if traceparent := c.GetHeader("traceparent"); traceparent != "" {
			parts := strings.Split(traceparent, "-")
			if len(parts) == 4 {
				id = parts[1] // trace-id 是 32 字符 hex
			}
		}

		// 2. 兜底取 X-Request-ID（兼容上游网关/旧客户端传入的 id）
		if id == "" {
			id = c.GetHeader(RequestIDKey)
		}

		// 3. 最终生成 uuid
		if id == "" {
			id = uuid.New().String()
		}

		// 写入 gin.Context 供下游处理器读取
		c.Set(RequestIDKey, id)
		// 写入 request context 供 xslog extractor 自动注入日志
		c.Request = c.Request.WithContext(xslog.WithField(c.Request.Context(), "request_id", id))
		// 响应 header 回传，供客户端排障
		c.Writer.Header().Set(RequestIDKey, id)

		c.Next()
	}
}
