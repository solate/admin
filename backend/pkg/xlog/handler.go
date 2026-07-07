package xlog

import (
	"context"
	"log/slog"
)

// contextHandler 包装底层 handler（JSON/Text），在每条日志输出前从 ctx 中
// 提取字段并追加到 Record。它本身不做格式化，只负责“注入”，格式化仍交给底层。
//
// 实现 slog.Handler 的四个方法时，WithAttrs / WithGroup 必须把调用透传给
// 底层 handler 并保持 contextHandler 包装，否则会丢失 group 语义与预置字段。
type contextHandler struct {
	inner      slog.Handler
	extractors []ContextExtractor
}

// newContextHandler 用给定的提取器包裹 inner。内置的 contextFieldsExtractor
// 始终排在最前，保证 WithField(s) 写入的字段优先被处理。
func newContextHandler(inner slog.Handler, extractors []ContextExtractor) *contextHandler {
	all := make([]ContextExtractor, 0, len(extractors)+1)
	all = append(all, contextFieldsExtractor)
	all = append(all, extractors...)
	return &contextHandler{inner: inner, extractors: all}
}

// Enabled 直接委托底层 handler。
func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle 依次运行所有提取器，把得到的 Attr 追加到 Record，再交给底层输出。
//
// 已知限制：注入的字段是通过 rec.AddAttrs 追加的，会落在“当前 group 作用域”里。
// 若在调用链上先 logger.WithGroup("g") 再打日志，注入的 request_id/tenant_id 会被
// 嵌进 "g" 对象而非顶层。本项目身份字段依赖顶层可检索，故约定：需要顶层身份字段的
// logger 不要套 WithGroup（slog 官方 handler guide 亦提及此类 wrapping 陷阱）。
func (h *contextHandler) Handle(ctx context.Context, rec slog.Record) error {
	for _, extract := range h.extractors {
		if attrs := extract(ctx); len(attrs) > 0 {
			rec.AddAttrs(attrs...)
		}
	}
	return h.inner.Handle(ctx, rec)
}

// WithAttrs 透传给底层并保持 contextHandler 包装。
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{inner: h.inner.WithAttrs(attrs), extractors: h.extractors}
}

// WithGroup 透传给底层并保持 contextHandler 包装。
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{inner: h.inner.WithGroup(name), extractors: h.extractors}
}
