package operationlog

import "context"

// contextKey 防止 context key 冲突
type contextKey int

const (
	CtxLogContext contextKey = iota
)

// WithLogContext 存入 context
func WithLogContext(ctx context.Context, lc *LogContext) context.Context {
	return context.WithValue(ctx, CtxLogContext, lc)
}

// GetLogContext 获取 LogContext
func GetLogContext(ctx context.Context) (*LogContext, bool) {
	lc, ok := ctx.Value(CtxLogContext).(*LogContext)
	return lc, ok
}

// MustGetLogContext 获取或创建 LogContext
func MustGetLogContext(ctx context.Context) *LogContext {
	if lc, ok := GetLogContext(ctx); ok {
		return lc
	}
	return &LogContext{}
}
