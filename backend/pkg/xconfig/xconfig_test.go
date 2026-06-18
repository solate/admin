package xconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeYAML 在 dir 下写一个临时 yaml 文件,返回其路径。
func writeYAML(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

type testServer struct {
	Port        int           `mapstructure:"port"`
	Mode        string        `mapstructure:"mode"`
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
}

type testRoot struct {
	Server testServer `mapstructure:"server"`
}

func TestLoader_HappyPath(t *testing.T) {
	dir := t.TempDir()
	p := writeYAML(t, dir, "config.yaml", `
server:
  port: 9090
  mode: release
  read_timeout: 15s
`)
	var cfg testRoot
	if err := New(WithFile(p)).Load(&cfg); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("mode = %q, want release", cfg.Server.Mode)
	}
	if cfg.Server.ReadTimeout != 15*time.Second {
		t.Errorf("read_timeout = %v, want 15s", cfg.Server.ReadTimeout)
	}
}

func TestLoader_DefaultsApplied(t *testing.T) {
	dir := t.TempDir()
	// yaml 缺省 port,靠 WithDefaults 兜底
	p := writeYAML(t, dir, "config.yaml", `
server:
  mode: debug
`)
	var cfg testRoot
	if err := New(WithFile(p), WithDefaults(map[string]any{"server.port": 8080})).Load(&cfg); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("port = %d, want default 8080", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("mode = %q, want debug", cfg.Server.Mode)
	}
}

func TestLoader_SearchPaths(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "myconf.yaml", `
server:
  port: 7000
`)
	var cfg testRoot
	if err := New(
		WithSearchPaths(dir),
		WithFileName("myconf"),
		WithFileType("yaml"),
	).Load(&cfg); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 7000 {
		t.Errorf("port = %d, want 7000", cfg.Server.Port)
	}
}

func TestLoader_FileNotFound(t *testing.T) {
	var cfg testRoot
	err := New(WithFile(filepath.Join(t.TempDir(), "nope.yaml"))).Load(&cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "read config") {
		t.Errorf("error = %q, want contain 'read config'", err.Error())
	}
}

func TestLoader_NilTarget(t *testing.T) {
	dir := t.TempDir()
	p := writeYAML(t, dir, "config.yaml", "server:\n  port: 1\n")
	if err := New(WithFile(p)).Load(nil); err == nil {
		t.Fatal("expected error for nil target, got nil")
	}
}

func TestOverride_Set(t *testing.T) {
	t.Setenv("XCONFIG_TEST_KEY", "from-env")
	var s string
	Override(&s, "XCONFIG_TEST_KEY")
	if s != "from-env" {
		t.Errorf("s = %q, want from-env", s)
	}
}

func TestOverride_EmptyOverrides(t *testing.T) {
	// env 设为空串时也要覆盖(set-to-empty 语义,用于显式清空密钥)
	t.Setenv("XCONFIG_TEST_EMPTY", "")
	s := "original"
	Override(&s, "XCONFIG_TEST_EMPTY")
	if s != "" {
		t.Errorf("s = %q, want empty (env set-to-empty overrides)", s)
	}
}

func TestOverride_Unset(t *testing.T) {
	// env 未设置时保持原值
	os.Unsetenv("XCONFIG_TEST_UNSET")
	s := "original"
	Override(&s, "XCONFIG_TEST_UNSET")
	if s != "original" {
		t.Errorf("s = %q, want original (env unset keeps value)", s)
	}
}

func TestOverride_NilDst(t *testing.T) {
	t.Setenv("XCONFIG_TEST_NIL", "x")
	Override(nil, "XCONFIG_TEST_NIL") // 不应 panic
}
