package xslog

import (
	"context"
	"io"
	"log/slog"
)

// Format 指定日志输出格式。
type Format = string

const (
	FormatJSON Format = "json" // 生产环境默认，结构化便于解析
	FormatText Format = "text" // 本地开发，可读性好
)

// ContextExtractor 从 context.Context 中提取要注入日志的字段。
// 返回 nil 表示无字段注入。这是 OpenTelemetry 等扩展的接入点：
// 调用方引入 otel 后实现该接口并注册到 Config.ContextExtractors，
// xlog 本身无需引入 otel 依赖。
type ContextExtractor func(ctx context.Context) []slog.Attr

// Config 是构造 logger 的配置。字段值来自 internal/config.LogConfig。
// 零值可用（Level 空串按 info、Format 空串按 JSON、Output nil 按 os.Stdout）。
type Config struct {
	Level  string    // 级别 debug/info/warn/error，空或未知值 → info
	Format Format    // 输出格式 json(默认) / text
	Output io.Writer // 输出目标，默认 os.Stdout（测试可注入 bytes.Buffer）
	// AddSource 为 true 时在日志中记录调用位置（裁成 dir/file:line）。
	AddSource bool
	// ContextExtractors 是自定义的 context 字段提取器。
	// 内置的 contextFieldsExtractor（处理 WithField/WithFields）
	// 始终在用户提取器之前运行，无需手动添加。
	ContextExtractors []ContextExtractor
	// ReplaceAttr 是用户自定义的字段改写函数（如 RedactReplaceAttr 脱敏）。
	// 它与内置的 shortenSource 串联执行（source 裁剪先跑，用户函数后跑），
	// 详见 chainReplaceAttr。
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}
