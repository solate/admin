// Package xlog 封装标准库 log/slog，为 SaaS 后端提供：
//   - 基于 context 的字段自动注入（request_id / tenant_id 等），由自定义
//     contextHandler 实现（slog 官方推荐的扩展方式）；
//   - 面向 OpenTelemetry 的零依赖扩展点（ContextExtractor）；
//   - 级别字符串解析（配置文件友好，默认 info）、source 裁成 dir/file:line、
//     敏感字段脱敏。
//
// 设计上不包装 slog 类型：对外直接暴露 *slog.Logger 供 DI，扩展全部落在自定义
// slog.Handler 上。输出固定交给 Config.Output（默认 os.Stdout），轮转交给容器/systemd。
package xslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
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

// New 根据 Config 构造一个 *slog.Logger，底层 handler 被 contextHandler 包裹，
// 支持从 context 自动注入字段。返回的 logger 可直接用于 DI，也可通过
// 标准库的 slog.SetDefault 设为全局默认供 slog 包级函数使用。
func New(cfg Config) *slog.Logger {
	// 填充默认值
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	// 构造底层 handler（JSON 或 Text）
	opts := &slog.HandlerOptions{
		Level:       parseLevel(cfg.Level),
		AddSource:   cfg.AddSource,
		ReplaceAttr: chainReplaceAttr(cfg.ReplaceAttr),
	}
	var base slog.Handler
	switch cfg.Format {
	case FormatText:
		base = slog.NewTextHandler(cfg.Output, opts)
	default: // "json"/"" 都走 JSONHandler，兑现"零值默认 JSON"契约
		base = slog.NewJSONHandler(cfg.Output, opts)
	}
	// 用 contextHandler 包裹，注入 context 字段
	handler := newContextHandler(base, cfg.ContextExtractors)
	return slog.New(handler)
}

// parseLevel 把配置字符串映射到 slog.Level（实现 slog.Leveler）。
// 直接复用 slog.Level.UnmarshalText：大小写不敏感（DEBUG/Debug 均可），
// 还支持 "info+2" 这类偏移语法，语义与 slog 自身完全一致，无需自维护映射表。
// 先 TrimSpace 兜住配置里的首尾空白（UnmarshalText 本身不容忍空白）。
// 空值、纯空白或非法值一律回落到 Info（UnmarshalText 失败时不改动目标变量）。
func parseLevel(s string) slog.Level {
	var lv slog.Level
	if err := lv.UnmarshalText([]byte(strings.TrimSpace(s))); err != nil {
		return slog.LevelInfo
	}
	return lv
}

// chainReplaceAttr 把内置的 shortenSource 与用户自定义的 ReplaceAttr 串联：
// source 裁剪永远生效（只动 SourceKey），用户函数（如 RedactReplaceAttr 脱敏）
// 随后叠加。二者作用的 key 不重叠，互不干扰。user 为 nil 时只跑 shortenSource。
func chainReplaceAttr(user func(groups []string, a slog.Attr) slog.Attr) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		a = shortenSource(groups, a)
		if user != nil {
			a = user(groups, a)
		}
		return a
	}
}

// shortenSource 把 source 字段裁成 "dir/file:line" 字符串（对齐 zap 的
// ShortCallerEncoder）：保留末尾一级目录做包消歧（避免多个 service.go/handler.go
// 只剩文件名无法区分），又不泄露完整部署路径、输出紧凑（默认是
// {"function":...,"file":...,"line":...} 对象）。
//
// 类型断言 *slog.Source 是 slog 官方写法 —— slog.Value 为零分配设计成不透明类型，
// 取结构化内容只能 .Any() 后断言，无类型安全访问器（见 golang/go#59280）。
// 成熟写法是把 a.Value 重写成 slog.StringValue，而非原地改 src.File。
//
// src.File 由 runtime 填充，分隔符恒为 "/"（与运行平台无关），故用 strings 而非
// filepath 处理，保证跨平台一致。
func shortenSource(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.SourceKey {
		return a
	}
	src, ok := a.Value.Any().(*slog.Source)
	if !ok || src == nil {
		return a
	}
	a.Value = slog.StringValue(fmt.Sprintf("%s:%d", shortFile(src.File), src.Line))
	return a
}

// shortFile 取路径末尾两段 "dir/file"。不足两段（无分隔符）时原样返回。
func shortFile(file string) string {
	if i := strings.LastIndexByte(file, '/'); i >= 0 {
		if j := strings.LastIndexByte(file[:i], '/'); j >= 0 {
			return file[j+1:]
		}
	}
	return file
}
