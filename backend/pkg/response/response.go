package response

import (
	"admin/pkg/xerr"
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
func Error(c *gin.Context, err error) {
	c.Error(err) // 供 OperationLogMiddleware 使用
	c.JSON(http.StatusOK, getResponse(c, err))
}

// ErrorWithMessage 错误响应（自定义消息）
func ErrorWithMessage(c *gin.Context, code int, message string) {
	c.Error(xerr.New(code, message)) // 供 OperationLogMiddleware 使用
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

// ErrorWithHttpCode 错误响应（自定义HTTP状态码）
func ErrorWithHttpCode(c *gin.Context, httpCode int, err error) {
	c.Error(err) // 供 OperationLogMiddleware 使用
	c.JSON(httpCode, getResponse(c, err))
}

// getResponse 获取响应
func getResponse(c *gin.Context, err error) Response {
	var resp Response
	appErr, ok := err.(*xerr.AppError) // 检查是否是 AppError 类型
	if ok {
		resp = Response{
			Code:      appErr.Code,
			Message:   appErr.Message,
			RequestID: getRequestID(c),
		}
	} else {
		resp = Response{
			Code:      xerr.ErrInternal.Code,
			Message:   err.Error(),
			RequestID: getRequestID(c),
		}
	}
	return resp
}
