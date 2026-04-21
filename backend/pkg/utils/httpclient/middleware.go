package httpclient

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

// RetryCondition 重试条件函数
type RetryCondition func(resp *Response, err error) bool

// Response 响应包装
type Response struct {
	StatusCode int
	Header     map[string][]string
	Body       []byte
}

// RetryIf 根据状态码判断是否重试
func RetryIf(statusCodes ...int) RetryCondition {
	return func(resp *Response, err error) bool {
		if err != nil {
			return true // 网络错误时重试
		}
		for _, code := range statusCodes {
			if resp.StatusCode == code {
				return true
			}
		}
		return false
	}
}

// DefaultRetryCondition 默认重试条件
func DefaultRetryCondition() RetryCondition {
	return RetryIf(
		429, // Too Many Requests
		500, // Internal Server Error
		502, // Bad Gateway
		503, // Service Unavailable
		504, // Gateway Timeout
	)
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	token     string
	tokenType string
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(token string) *AuthMiddleware {
	return &AuthMiddleware{
		token:     token,
		tokenType: "Bearer",
	}
}

// NewAuthMiddlewareWithType 创建指定类型的认证中间件
func NewAuthMiddlewareWithType(token, tokenType string) *AuthMiddleware {
	return &AuthMiddleware{
		token:     token,
		tokenType: tokenType,
	}
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	enableRequestBodyLog  bool
	enableResponseBodyLog bool
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		enableRequestBodyLog:  false,
		enableResponseBodyLog: false,
	}
}

// WithRequestBodyLog 启用请求体日志
func (l *LoggingMiddleware) WithRequestBodyLog() *LoggingMiddleware {
	l.enableRequestBodyLog = true
	return l
}

// WithResponseBodyLog 启用响应体日志
func (l *LoggingMiddleware) WithResponseBodyLog() *LoggingMiddleware {
	l.enableResponseBodyLog = true
	return l
}

// LogRequest 记录请求日志
func (l *LoggingMiddleware) LogRequest(ctx context.Context, method, url string, headers map[string]string, body []byte) {
	event := log.Ctx(ctx).Info().
		Str("method", method).
		Str("url", url)

	if len(headers) > 0 {
		safeHeaders := make(map[string]string)
		for k, v := range headers {
			if k != "Authorization" && k != "Cookie" && k != "Set-Cookie" {
				safeHeaders[k] = v
			}
		}
		event = event.Interface("headers", safeHeaders)
	}

	if l.enableRequestBodyLog && len(body) > 0 {
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			event = event.Interface("body", jsonBody)
		} else {
			event = event.Str("body", string(body))
		}
	}

	event.Msg("发送 HTTP 请求")
}

// LogResponse 记录响应日志
func (l *LoggingMiddleware) LogResponse(ctx context.Context, statusCode int, duration time.Duration, body []byte) {
	event := log.Ctx(ctx).Info().
		Int("status_code", statusCode).
		Dur("duration", duration)

	if l.enableResponseBodyLog && len(body) > 0 {
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			event = event.Interface("body", jsonBody)
		} else {
			event = event.Str("body", string(body))
		}
	}

	event.Msg("收到 HTTP 响应")
}
