package logger

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

// shortenSource 把 source 字段裁成 "file:line" 字符串（Lshortfile 风格），
// 避免泄露部署路径、输出紧凑（默认是 {"function":...,"file":...,"line":...} 对象）。
//
// 类型断言 *slog.Source 是 slog 官方写法 —— slog.Value 为零分配设计成不透明类型，
// 取结构化内容只能 .Any() 后断言，无类型安全访问器（见 golang/go#59280）。
// 成熟写法是把 a.Value 重写成 slog.StringValue，而非原地改 src.File。
func shortenSource(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.SourceKey {
		return a
	}
	src, ok := a.Value.Any().(*slog.Source)
	if !ok || src == nil {
		return a
	}
	a.Value = slog.StringValue(fmt.Sprintf("%s:%d", filepath.Base(src.File), src.Line))
	return a
}
