package audit

import (
	"context"
)

// contextKey 用于 context 的 key 类型
type contextKey int

const (
	ctxKeyClientInfo  contextKey = 0
	ctxKeyRequestInfo contextKey = 1
)

// ClientInfo 客户端信息
type ClientInfo struct {
	IP        string
	UserAgent string
}

// RequestInfo 请求信息
type RequestInfo struct {
	Method string
	Path   string
	Params string
}

// WithClientInfo 将客户端信息存入 context
func WithClientInfo(ctx context.Context, info *ClientInfo) context.Context {
	return context.WithValue(ctx, ctxKeyClientInfo, info)
}

// WithRequestInfo 将请求信息存入 context
func WithRequestInfo(ctx context.Context, info *RequestInfo) context.Context {
	return context.WithValue(ctx, ctxKeyRequestInfo, info)
}

// GetClientInfo 从 context 获取客户端信息
func GetClientInfo(ctx context.Context) *ClientInfo {
	if info, ok := ctx.Value(ctxKeyClientInfo).(*ClientInfo); ok {
		return info
	}
	return nil
}

// GetRequestInfo 从 context 获取请求信息
func GetRequestInfo(ctx context.Context) *RequestInfo {
	if info, ok := ctx.Value(ctxKeyRequestInfo).(*RequestInfo); ok {
		return info
	}
	return nil
}

