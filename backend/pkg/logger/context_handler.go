package logger

import (
	"context"
	"log/slog"
)

// RequestIDKey / TenantIDKey 是存入 context.Context 的请求级字段 key。
// 中间件用 context.WithValue(ctx, logger.RequestIDKey{}, id) 写入；
// ContextHandler.Handle 读取并自动注入到每条日志。
type RequestIDKey struct{}
type TenantIDKey struct{}

// ContextHandler 包裹任意 slog.Handler：Handle 时从 ctx 自动追加
// request_id / tenant_id。调用方全程用 *Context 变体（InfoContext 等）即自动生效，
// 无需在每个 log 调用里手写字段。
//
// Gin 接入前为 inert（ctx 里暂无这些 key），接入 RequestID 中间件后自然生效。
type ContextHandler struct {
	slog.Handler
}

// NewContextHandler 包裹一个 Handler。
func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{Handler: h}
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if v, ok := ctx.Value(RequestIDKey{}).(string); ok && v != "" {
		r.AddAttrs(slog.String("request_id", v))
	}
	if v, ok := ctx.Value(TenantIDKey{}).(string); ok && v != "" {
		r.AddAttrs(slog.String("tenant_id", v))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs / WithGroup 必须重新包裹返回 *ContextHandler，
// 否则 logger.With(...) 生成的子 logger 会退化为内层 Handler，丢失 ctx 注入。
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithGroup(name)}
}
