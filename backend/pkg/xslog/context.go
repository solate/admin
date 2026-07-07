package xslog

import (
	"context"
	"log/slog"
)

// ctxKey 是非导出类型，避免与其他包的 context key 冲突。
type ctxKey struct{}

// fieldsKey 用作存放 ctx 注入字段的唯一 key。
var fieldsKey = ctxKey{}

// WithField 把单个字段存入 ctx，后续在该 ctx 下打的日志会自动带上该字段。
func WithField(ctx context.Context, key string, val any) context.Context {
	return WithFields(ctx, slog.Any(key, val))
}

// WithFields 把一组字段存入 ctx。多次调用会累积（后写入的在后），
// 返回的新 ctx 持有合并后的字段切片副本，不影响父 ctx。
func WithFields(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	existing := fieldsFromContext(ctx)
	// 复制一份，避免多个子 ctx 共享同一底层数组产生数据竞争。
	merged := make([]slog.Attr, 0, len(existing)+len(attrs))
	merged = append(merged, existing...)
	merged = append(merged, attrs...)
	return context.WithValue(ctx, fieldsKey, merged)
}

// fieldsFromContext 取出通过 WithField(s) 存入的字段，无则返回 nil。
func fieldsFromContext(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	attrs, _ := ctx.Value(fieldsKey).([]slog.Attr)
	return attrs
}

// contextFieldsExtractor 是内置提取器，读取 WithField(s) 写入的字段。
// 它始终排在用户自定义 ContextExtractors 之前执行。
func contextFieldsExtractor(ctx context.Context) []slog.Attr {
	return fieldsFromContext(ctx)
}
