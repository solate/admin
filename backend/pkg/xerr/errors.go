// Package xerr 定义应用统一错误类型 AppError 与预定义错误码。
// AppError 携带业务码与用户可读消息，可选包装底层错误，并通过 Code()/Message()
// getter 实现 response.Coder 接口——response 只依赖该接口、不 import 本包。
// 错误码分段见 codes.go。
package xerr

import (
	"fmt"
)

// AppError 应用错误。code 为业务码(非 HTTP 码)，message 面向前端可读，
// err 为被包装的底层错误(不序列化到 JSON，仅用于日志与 errors.Is/As)。
type AppError struct {
	code    int
	message string
	err     error
}

// New 创建不含底层错误的 AppError。
func New(code int, message string) *AppError {
	return &AppError{
		code:    code,
		message: message,
	}
}

// Error 实现 error 接口。有底层错误时追加其信息，便于日志定位。
func (e *AppError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.code, e.message, e.err)
	}
	return fmt.Sprintf("[%d] %s", e.code, e.message)
}

// Code 返回业务码，实现 response.Coder 接口。
func (e *AppError) Code() int { return e.code }

// Message 返回面向前端的可读消息，实现 response.Coder 接口。
func (e *AppError) Message() string { return e.message }

// Unwrap 支持 errors.Is / errors.As 向下解包底层错误。
func (e *AppError) Unwrap() error {
	return e.err
}

// Wrap 在既有错误上包装业务码与消息，保留底层错误供解包与日志。
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		code:    code,
		message: message,
		err:     err,
	}
}
