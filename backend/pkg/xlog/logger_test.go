package xlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"

	"admin/pkg/xlog"
)

// newTestLogger 构造一个写入 buf 的 JSON logger,默认开启 AddSource。
func newTestLogger(level string) (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := xlog.New(xlog.Config{Level: level, Format: "json", AddSource: true, Output: &buf})
	return l, &buf
}

// TestLevelFilter 验证字符串级别解析与过滤:warn 级别下 info 被丢弃。
func TestLevelFilter(t *testing.T) {
	log, buf := newTestLogger("warn")
	log.Info("should be filtered") // 低于 Warn,丢弃
	log.Warn("kept", slog.String("k", "v"))

	out := buf.String()
	if strings.Contains(out, "should be filtered") {
		t.Fatalf("info leaked through warn level: %s", out)
	}
	if !strings.Contains(out, "kept") || !strings.Contains(out, `"k":"v"`) {
		t.Fatalf("warn record missing/malformed: %s", out)
	}
}

// TestDefaultLevelInfo 验证空级别默认 info:debug 丢弃、info 保留。
func TestDefaultLevelInfo(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{Format: "json", Output: &buf}) // Level 空
	log.Debug("debug dropped")
	if buf.Len() > 0 {
		t.Fatalf("debug should be dropped at default info level: %s", buf.String())
	}
	log.Info("info kept")
	if buf.Len() == 0 {
		t.Fatalf("info should appear at default info level")
	}
}

// TestJSONFormat 验证 JSON 格式输出结构正确。
func TestJSONFormat(t *testing.T) {
	log, buf := newTestLogger("info")
	log.Info("hello", slog.String("user", "alice"))

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output not valid JSON: %v (%s)", err, buf.String())
	}
	if m["msg"] != "hello" {
		t.Fatalf("msg mismatch: %v", m["msg"])
	}
	if m["level"] != "INFO" {
		t.Fatalf("level mismatch: %v", m["level"])
	}
}

// TestTextFormat 验证 Text 格式输出 key=value。
func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{Format: "text", Output: &buf})
	log.Info("text format test", "key", "value")

	out := buf.String()
	if !strings.Contains(out, "text format test") {
		t.Errorf("output should contain message: %s", out)
	}
	if !strings.Contains(out, "key=value") {
		t.Errorf("output should contain key=value: %s", out)
	}
}

// TestShortenSource 验证 source 被裁成 file:line(无路径分隔符)。
func TestShortenSource(t *testing.T) {
	log, buf := newTestLogger("info")
	log.Info("x") // 触发 source 记录

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not JSON: %v (%s)", err, buf.String())
	}
	src, ok := m["source"].(string)
	if !ok {
		t.Fatalf("source not a string (got %T): %v", m["source"], m["source"])
	}
	if !strings.Contains(src, "xlog/logger_test.go:") {
		t.Fatalf("source not shortened to dir/file:line: %q", src)
	}
	if strings.Count(src, "/") != 1 {
		t.Fatalf("source should keep exactly one dir segment: %q", src)
	}
}

// TestShortenSourceWithRedact 是 chainReplaceAttr 的关键验证:
// 内置 source 裁剪与用户传入的 RedactReplaceAttr 必须同时生效。
func TestShortenSourceWithRedact(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{
		Level:       "info",
		Format:      "json",
		AddSource:   true,
		Output:      &buf,
		ReplaceAttr: xlog.RedactReplaceAttr("password"),
	})
	log.Info("login", "user", "alice", "password", "secret123")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not JSON: %v (%s)", err, buf.String())
	}
	// source 仍被裁剪
	src, ok := m["source"].(string)
	if !ok || strings.Count(src, "/") != 1 || !strings.Contains(src, ":") {
		t.Fatalf("source not shortened alongside redact: %v", m["source"])
	}
	// password 仍被脱敏
	if m["password"] != "[REDACTED]" {
		t.Fatalf("password should be redacted: %v", m["password"])
	}
	if m["user"] != "alice" {
		t.Fatalf("user should not be redacted: %v", m["user"])
	}
}

// TestContextInjection 验证 WithFields 写入的字段自动注入日志。
func TestContextInjection(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{Format: "json", Output: &buf, Level: "info"})

	ctx := xlog.WithFields(context.Background(),
		slog.String("request_id", "req-123"),
		slog.String("tenant_id", "tenant-abc"),
	)
	log.InfoContext(ctx, "test message", "extra", "value")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["request_id"] != "req-123" {
		t.Errorf("missing or wrong request_id: %v", m["request_id"])
	}
	if m["tenant_id"] != "tenant-abc" {
		t.Errorf("missing or wrong tenant_id: %v", m["tenant_id"])
	}
	if m["extra"] != "value" {
		t.Errorf("missing or wrong extra: %v", m["extra"])
	}
}

// TestContextInjectionEmpty 验证无 context 字段时不出现注入字段。
func TestContextInjectionEmpty(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{Format: "json", Output: &buf})
	log.InfoContext(context.Background(), "no fields")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["request_id"]; ok {
		t.Errorf("unexpected request_id in log without context fields")
	}
}

// TestWithFieldAccumulate 验证 WithField 多次调用累积,且不污染父 ctx。
func TestWithFieldAccumulate(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{Format: "json", Output: &buf})

	parent := xlog.WithField(context.Background(), "a", "1")
	child := xlog.WithField(parent, "b", "2")

	log.InfoContext(child, "both")
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["a"] != "1" || m["b"] != "2" {
		t.Errorf("child should carry both fields: %v", m)
	}

	// 父 ctx 不应含有 b
	buf.Reset()
	log.InfoContext(parent, "parent only")
	var pm map[string]any
	if err := json.Unmarshal(buf.Bytes(), &pm); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := pm["b"]; ok {
		t.Errorf("parent ctx polluted with child field b: %v", pm)
	}
}

// TestContextExtractor 验证自定义 ContextExtractor 注入字段。
func TestContextExtractor(t *testing.T) {
	type customKey struct{}
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{
		Format: "json",
		Output: &buf,
		ContextExtractors: []xlog.ContextExtractor{
			func(ctx context.Context) []slog.Attr {
				if v, ok := ctx.Value(customKey{}).(string); ok {
					return []slog.Attr{slog.String("custom_field", v)}
				}
				return nil
			},
		},
	})
	ctx := context.WithValue(context.Background(), customKey{}, "custom_value")
	log.InfoContext(ctx, "test custom extractor")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["custom_field"] != "custom_value" {
		t.Errorf("missing or wrong custom_field: %v", m["custom_field"])
	}
}

// TestRedactReplaceAttr 验证敏感字段脱敏。
func TestRedactReplaceAttr(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{
		Format:      "json",
		Output:      &buf,
		ReplaceAttr: xlog.RedactReplaceAttr("password", "token"),
	})
	log.InfoContext(context.Background(), "login",
		"user", "alice",
		"password", "secret123",
		"token", "abc-xyz",
	)
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["user"] != "alice" {
		t.Errorf("user should not be redacted: %v", m["user"])
	}
	if m["password"] != "[REDACTED]" {
		t.Errorf("password should be redacted, got: %v", m["password"])
	}
	if m["token"] != "[REDACTED]" {
		t.Errorf("token should be redacted, got: %v", m["token"])
	}
}

// TestErrAttr 验证 Err 快捷构造统一 error 字段。
func TestErrAttr(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{Format: "json", Output: &buf})
	log.InfoContext(context.Background(), "operation failed", xlog.Err(errors.New("something went wrong")))

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["error"] != "something went wrong" {
		t.Errorf("missing or wrong error field: %v", m["error"])
	}
}

// TestErrAttrNil 验证 Err(nil) 返回空 Attr,不出现在日志中。
func TestErrAttrNil(t *testing.T) {
	var buf bytes.Buffer
	log := xlog.New(xlog.Config{Format: "json", Output: &buf})
	log.InfoContext(context.Background(), "no error", xlog.Err(nil))

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["error"]; ok {
		t.Errorf("error field should not appear when Err(nil)")
	}
}

// TestHandlerCompliance 用官方 slogtest 验证 contextHandler 符合 slog.Handler 规范。
func TestHandlerCompliance(t *testing.T) {
	var buf bytes.Buffer
	logger := xlog.New(xlog.Config{Format: "json", Output: &buf})

	results := func() []map[string]any {
		var ms []map[string]any
		for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
			if line == "" {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				t.Fatalf("unmarshal log line: %v", err)
			}
			ms = append(ms, m)
		}
		return ms
	}
	if err := slogtest.TestHandler(logger.Handler(), results); err != nil {
		t.Fatal(err)
	}
}
