package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// New 根据配置构建生产可用的 *slog.Logger，输出到 os.Stdout。
func New(cfg Config) *slog.Logger {
	return NewWith(cfg, os.Stdout)
}

// NewWith 构建写入 out 的 *slog.Logger。out 可注入（测试用 bytes.Buffer）。
//
// 设计要点：
//   - 级别用 *slog.LevelVar：并发安全，调用方持有该指针即可运行时 Set 热改级别；
//   - AddSource + ReplaceAttr(shortenSource)：记录调用位置并裁成 file:line；
//   - 包一层 ContextHandler：自动从 ctx 注入 request_id/tenant_id（需用 *Context 变体）。
func NewWith(cfg Config, out io.Writer) *slog.Logger {
	level := parseLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   cfg.AddSource,
		ReplaceAttr: shortenSource,
	}

	var h slog.Handler
	if cfg.Format == "json" {
		h = slog.NewJSONHandler(out, opts)
	} else {
		h = slog.NewTextHandler(out, opts) // "text"/"console"/"" 都走 TextHandler
	}
	return slog.New(NewContextHandler(h))
}

// parseLevel 把配置字符串映射到 *slog.LevelVar（实现 slog.Leveler）。
func parseLevel(s string) *slog.LevelVar {
	l := new(slog.LevelVar)
	switch s {
	case "debug":
		l.Set(slog.LevelDebug)
	case "warn":
		l.Set(slog.LevelWarn)
	case "error":
		l.Set(slog.LevelError)
	default:
		l.Set(slog.LevelInfo)
	}
	return l
}

// Fatal 记一条 Error 后 os.Exit(1)。slog 没有 Fatal 级别，启动失败用它。
// 手动构造 Record 并用调用者 PC，保证 source 指向 Fatal 的调用方（而非本函数）。
func Fatal(log *slog.Logger, msg string, args ...any) {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // 0=Callers, 1=Fatal, 2=调用方
	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])
	r.Add(args...)
	_ = log.Handler().Handle(context.Background(), r)
	os.Exit(1)
}
