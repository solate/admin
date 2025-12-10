package response

import (
	"admin/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "X-Request-ID"

// Response 统一响应结构
type Response struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id"`
}

// getRequestID 从context中获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		return requestID.(string)
	}
	return ""
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Error 错误响应
func Error(c *gin.Context, httpCode int, err *errors.AppError) {
	c.JSON(httpCode, Response{
		Code:      err.Code,
		Message:   err.Message,
		RequestID: getRequestID(c),
	})
}

// ErrorWithMessage 错误响应（自定义消息）
func ErrorWithMessage(c *gin.Context, httpCode int, code int, message string) {
	c.JSON(httpCode, Response{
		Code:      code,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

// SuccessResponse 成功响应（别名，保持兼容性）
func SuccessResponse(c *gin.Context, data any) {
	Success(c, data)
}

// ErrorResponse 错误响应（简化版，直接使用code和message）
func ErrorResponse(c *gin.Context, httpCode int, code int, message string) {
	ErrorWithMessage(c, httpCode, code, message)
}
