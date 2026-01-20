package httpclient

import (
	"fmt"
)

// HTTPClientError HTTP 客户端错误
type HTTPClientError struct {
	StatusCode int    // HTTP 状态码
	Message    string // 错误消息
	Err        error  // 原始错误
}

// Error 实现 error 接口
func (e *HTTPClientError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[HTTP %d] %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("[HTTP %d] %s", e.StatusCode, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *HTTPClientError) Unwrap() error {
	return e.Err
}

// 预定义错误
var (
	ErrRequestFailed      = &HTTPClientError{StatusCode: 0, Message: "请求失败"}
	ErrTimeout            = &HTTPClientError{StatusCode: 0, Message: "请求超时"}
	ErrInvalidResponse    = &HTTPClientError{StatusCode: 0, Message: "无效的响应"}
	ErrMaxRetriesExceeded = &HTTPClientError{StatusCode: 0, Message: "超过最大重试次数"}
)

// IsStatusCodeError 判断是否为特定状态码错误
func IsStatusCodeError(err error, statusCode int) bool {
	if httpErr, ok := err.(*HTTPClientError); ok {
		return httpErr.StatusCode == statusCode
	}
	return false
}

// IsTimeout 判断是否为超时错误
func IsTimeout(err error) bool {
	if httpErr, ok := err.(*HTTPClientError); ok {
		return httpErr.Message == ErrTimeout.Message
	}
	return false
}

// IsNetworkError 判断是否为网络错误（状态码为 0）
func IsNetworkError(err error) bool {
	if httpErr, ok := err.(*HTTPClientError); ok {
		return httpErr.StatusCode == 0
	}
	return false
}

// NewError 创建 HTTP 错误
func NewError(statusCode int, message string, err error) *HTTPClientError {
	return &HTTPClientError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}
