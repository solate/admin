package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func newTestLogger(level string) (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := NewWith(Config{Level: level, Format: "json", AddSource: true}, &buf)
	return l, &buf
}

func TestNewWith_LevelFilter(t *testing.T) {
	log, buf := newTestLogger("warn")
	log.Info("should be filtered") // 低于 Warn，丢弃
	log.Warn("kept", slog.String("k", "v"))

	out := buf.String()
	if strings.Contains(out, "should be filtered") {
		t.Fatalf("info leaked through warn level: %s", out)
	}
	if !strings.Contains(out, "kept") || !strings.Contains(out, `"k":"v"`) {
		t.Fatalf("warn record missing/malformed: %s", out)
	}
}

func TestNewWith_JSONFormat(t *testing.T) {
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
	if !strings.Contains(src, "logger_test.go:") {
		t.Fatalf("source not shortened to file:line: %q", src)
	}
	if strings.Contains(src, "/") {
		t.Fatalf("source still has path separator: %q", src)
	}
}

func TestContextHandler_InjectsRequestID(t *testing.T) {
	log, buf := newTestLogger("info")
	ctx := context.WithValue(context.Background(), RequestIDKey{}, "req-123")
	log.InfoContext(ctx, "req")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if m["request_id"] != "req-123" {
		t.Fatalf("request_id not injected: %v", m["request_id"])
	}
}

func TestContextHandler_WithAttrsPreservesInjection(t *testing.T) {
	log, buf := newTestLogger("info")
	ctx := context.WithValue(context.Background(), RequestIDKey{}, "req-456")
	// logger.With 生成子 logger —— 必须仍能注入 request_id（验证 WithAttrs 正确重写）
	sub := log.With(slog.String("component", "svc"))
	sub.InfoContext(ctx, "sub call")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if m["request_id"] != "req-456" {
		t.Fatalf("sub-logger lost request_id injection: %v", m["request_id"])
	}
	if m["component"] != "svc" {
		t.Fatalf("With attr lost: %v", m["component"])
	}
}
