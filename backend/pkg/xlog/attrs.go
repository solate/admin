package xlog

import (
	"log/slog"
	"slices"
	"strings"
)

// Err 是记录 error 的快捷构造，统一用 "error" 作为 key，便于日志检索与告警。
// err 为 nil 时返回一个空 Attr（slog 会将其忽略）。
func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.String("error", err.Error())
}

// RedactReplaceAttr 返回一个 slog.HandlerOptions.ReplaceAttr 函数，
// 对 key 命中 sensitiveKeys（大小写不敏感）的字段，把值替换为 "[REDACTED]"。
// 用于避免密码、token 等敏感字段被明文写入日志。
//
// 用法：
//
//	cfg := xlog.Config{ReplaceAttr: xlog.RedactReplaceAttr("password", "token")}
//
// 注意：分组（WithGroup）下的字段，groups 参数会带上组名前缀，这里只按叶子 key 匹配。
func RedactReplaceAttr(sensitiveKeys ...string) func(groups []string, a slog.Attr) slog.Attr {
	lowered := make([]string, len(sensitiveKeys))
	for i, k := range sensitiveKeys {
		lowered[i] = strings.ToLower(k)
	}
	return func(_ []string, a slog.Attr) slog.Attr {
		// 跳过 group 属性：若组名恰好命中敏感词，用 slog.String 覆盖会摧毁整组结构，
		// 而组容器本身不该是敏感值。只对叶子字段做脱敏。
		if a.Value.Kind() == slog.KindGroup {
			return a
		}
		if slices.Contains(lowered, strings.ToLower(a.Key)) {
			return slog.String(a.Key, "[REDACTED]")
		}
		return a
	}
}
