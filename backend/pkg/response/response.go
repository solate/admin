// Package response 提供 gin 统一响应封装：把业务结果渲染成 {code, message, data} envelope。
//
// 设计约束：
//   - request_id 不进 body：它已由 requestid 中间件写入 X-Request-ID 响应头与每行日志，
//     前端从 header 读取即可，envelope 无需重复携带。
//   - 错误渲染直接认 xerr.AppError（同为项目 pkg 业务包，无跨项目复用诉求，故不做接口反转）；
//     用 errors.As 断言而非 type switch，以穿透 fmt.Errorf %w 包装。
//   - 参数校验中文提示由本包的 validator.go 处理：getResponse 识别 validator.ValidationErrors
//     并用官方 zh 翻译器翻成中文，handler 直接把 c.ShouldBind 的 err 交给 response.Error 即可。
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"admin/pkg/xerr"
)

// Response 统一响应结构。request_id 走响应头，不在 body 内。
type Response struct {
	Code    int    `json:"code" example:"200"`        // 业务状态码
	Message string `json:"message" example:"success"` // 响应消息
	Data    any    `json:"data,omitempty"`            // 响应数据
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// SuccessWithCode 成功响应（自定义业务码）
func SuccessWithCode(c *gin.Context, code int, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	c.Error(err) // 供 OperationLogMiddleware 使用
	c.JSON(http.StatusOK, getResponse(err))
}

// ErrorWithMessage 错误响应（自定义消息）
func ErrorWithMessage(c *gin.Context, code int, message string) {
	c.Error(errors.New(message)) // 供 OperationLogMiddleware 使用
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithHttpCode 错误响应（自定义HTTP状态码）
func ErrorWithHttpCode(c *gin.Context, httpCode int, err error) {
	c.Error(err) // 供 OperationLogMiddleware 使用
	c.JSON(httpCode, getResponse(err))
}

// getResponse 把错误渲染成 envelope。
func getResponse(err error) Response {
	// 优先识别参数校验错误（validator.ValidationErrors），翻译成中文
	if msg, ok := translateValidationError(err); ok {
		return Response{
			Code:    xerr.ErrInvalidParams.Code(),
			Message: msg,
		}
	}

	// 其次识别业务错误 *xerr.AppError（errors.As 穿透 %w 包装）
	var appErr *xerr.AppError
	if errors.As(err, &appErr) {
		return Response{
			Code:    appErr.Code(),
			Message: appErr.Message(),
		}
	}

	// 非 AppError：统一兜底，不泄露内部细节，兜底码复用 xerr.ErrInternal
	return Response{
		Code:    xerr.ErrInternal.Code(),
		Message: "服务器内部错误",
	}
}
