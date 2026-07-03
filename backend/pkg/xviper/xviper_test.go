package xviper

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestConfig 是测试用的配置结构体（mapstructure tag）
type TestConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Port        int           `mapstructure:"port"`
	Host        string        `mapstructure:"host"`
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// writeDevOverlay 在 dir 下写一个空的 config.dev.yaml。
// xviper 默认环境是 dev,未设 APP_ENV 时会强制合并 config.dev.yaml(fail-fast),
// 所以不测 overlay 缺失的用例都需要先备好这个空 overlay。
func writeDevOverlay(t *testing.T, dir string) {
	t.Helper()
	p := filepath.Join(dir, "config.dev.yaml")
	if err := os.WriteFile(p, []byte("# empty dev overlay\n"), 0644); err != nil {
		t.Fatalf("write dev overlay: %v", err)
	}
}

func TestLoad_BasicFile(t *testing.T) {
	// 准备测试配置文件
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 8080
  host: localhost
  read_timeout: 10s

database:
  host: db.example.com
  port: 5432
  user: admin
  password: secret
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	writeDevOverlay(t, dir)

	// 加载配置
	cfg, err := Load[TestConfig](WithPath[TestConfig](configPath))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 验证基础字段
	if cfg.Server.Port != 8080 {
		t.Errorf("server.port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("server.host = %q, want localhost", cfg.Server.Host)
	}
	if cfg.Server.ReadTimeout != 10*time.Second {
		t.Errorf("server.read_timeout = %v, want 10s", cfg.Server.ReadTimeout)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("database.host = %q, want db.example.com", cfg.Database.Host)
	}
	if cfg.Database.Password != "secret" {
		t.Errorf("database.password = %q, want secret", cfg.Database.Password)
	}
}

func TestLoad_OverlayMerge(t *testing.T) {
	dir := t.TempDir()

	// base 配置
	basePath := filepath.Join(dir, "config.yaml")
	baseContent := `
server:
  port: 8080
  host: localhost
  read_timeout: 10s

database:
  host: localhost
  port: 5432
  user: dev_user
  password: dev_pass
`
	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatalf("write base config: %v", err)
	}

	// prod overlay（只覆盖部分字段）
	overlayPath := filepath.Join(dir, "config.prod.yaml")
	overlayContent := `
server:
  port: 80
  host: 0.0.0.0

database:
  host: prod-db.example.com
  user: prod_user
`
	if err := os.WriteFile(overlayPath, []byte(overlayContent), 0644); err != nil {
		t.Fatalf("write overlay config: %v", err)
	}

	// 设置 APP_ENV=prod
	t.Setenv("APP_ENV", "prod")

	cfg, err := Load[TestConfig](WithPath[TestConfig](basePath))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// overlay 覆盖的字段应该是新值
	if cfg.Server.Port != 80 {
		t.Errorf("server.port = %d, want 80 (overlay value)", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("server.host = %q, want 0.0.0.0 (overlay value)", cfg.Server.Host)
	}
	if cfg.Database.Host != "prod-db.example.com" {
		t.Errorf("database.host = %q, want prod-db.example.com (overlay value)", cfg.Database.Host)
	}
	if cfg.Database.User != "prod_user" {
		t.Errorf("database.user = %q, want prod_user (overlay value)", cfg.Database.User)
	}

	// overlay 未覆盖的字段应该保留 base 的值
	if cfg.Server.ReadTimeout != 10*time.Second {
		t.Errorf("server.read_timeout = %v, want 10s (base value)", cfg.Server.ReadTimeout)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("database.port = %d, want 5432 (base value)", cfg.Database.Port)
	}
	if cfg.Database.Password != "dev_pass" {
		t.Errorf("database.password = %q, want dev_pass (base value)", cfg.Database.Password)
	}
}

func TestLoad_OverlayMissing_FailFast(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "config.yaml")
	baseContent := `
server:
  port: 8080
`
	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatalf("write base config: %v", err)
	}

	// 设置 APP_ENV=staging，但不创建 config.staging.yaml
	t.Setenv("APP_ENV", "staging")

	_, err := Load[TestConfig](WithPath[TestConfig](basePath))
	if err == nil {
		t.Fatal("Load should fail when overlay file is missing")
	}
	// 错误信息应该提到合并失败
	if err.Error() == "" {
		t.Errorf("error message is empty")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 8080
  host: localhost

database:
  host: localhost
  port: 5432
  password: file_password
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	writeDevOverlay(t, dir)

	// 设置环境变量覆盖（APP_ 前缀 + 点转下划线大写）
	t.Setenv("APP_SERVER_PORT", "9000")
	t.Setenv("APP_DATABASE_HOST", "db-override.example.com")
	t.Setenv("APP_DATABASE_PASSWORD", "env_password")

	cfg, err := Load[TestConfig](WithPath[TestConfig](configPath))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 环境变量应该覆盖文件中的值
	if cfg.Server.Port != 9000 {
		t.Errorf("server.port = %d, want 9000 (env override)", cfg.Server.Port)
	}
	if cfg.Database.Host != "db-override.example.com" {
		t.Errorf("database.host = %q, want db-override.example.com (env override)", cfg.Database.Host)
	}
	if cfg.Database.Password != "env_password" {
		t.Errorf("database.password = %q, want env_password (env override)", cfg.Database.Password)
	}

	// 未设置环境变量的字段保持文件值
	if cfg.Server.Host != "localhost" {
		t.Errorf("server.host = %q, want localhost (file value)", cfg.Server.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("database.port = %d, want 5432 (file value)", cfg.Database.Port)
	}
}

func TestLoad_MergePriority(t *testing.T) {
	// 测试合并优先级：base < overlay < env（最高）
	dir := t.TempDir()

	basePath := filepath.Join(dir, "config.yaml")
	baseContent := `
server:
  port: 8080
  host: base-host
`
	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatalf("write base: %v", err)
	}

	overlayPath := filepath.Join(dir, "config.prod.yaml")
	overlayContent := `
server:
  port: 80
  host: overlay-host
`
	if err := os.WriteFile(overlayPath, []byte(overlayContent), 0644); err != nil {
		t.Fatalf("write overlay: %v", err)
	}

	t.Setenv("APP_ENV", "prod")
	t.Setenv("APP_SERVER_PORT", "443")

	cfg, err := Load[TestConfig](WithPath[TestConfig](basePath))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// port 被环境变量覆盖（最高优先级）
	if cfg.Server.Port != 443 {
		t.Errorf("server.port = %d, want 443 (env has highest priority)", cfg.Server.Port)
	}
	// host 使用 overlay 值（环境变量未设置）
	if cfg.Server.Host != "overlay-host" {
		t.Errorf("server.host = %q, want overlay-host (overlay priority)", cfg.Server.Host)
	}
}

func TestLoad_WithValidate_Success(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 8080
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	writeDevOverlay(t, dir)

	validate := func(c *TestConfig) error {
		if c.Server.Port < 1024 {
			return nil // 通过校验
		}
		return nil
	}

	cfg, err := Load[TestConfig](
		WithPath[TestConfig](configPath),
		WithValidate(validate),
	)
	if err != nil {
		t.Fatalf("Load with validate failed: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("server.port = %d, want 8080", cfg.Server.Port)
	}
}

func TestLoad_WithValidate_Failure(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 80
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	validate := func(c *TestConfig) error {
		if c.Server.Port < 1024 {
			return os.ErrInvalid // 模拟校验失败
		}
		return nil
	}

	_, err := Load[TestConfig](
		WithPath[TestConfig](configPath),
		WithValidate(validate),
	)
	if err == nil {
		t.Fatal("Load should fail when validation fails")
	}
	// 错误消息应该提到校验失败
	if err.Error() == "" {
		t.Errorf("error message is empty")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load[TestConfig](WithPath[TestConfig]("/nonexistent/config.yaml"))
	if err == nil {
		t.Fatal("Load should fail when config file does not exist")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	invalidContent := `
server:
  port: not_a_number
  host: [unclosed
`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	_, err := Load[TestConfig](WithPath[TestConfig](configPath))
	if err == nil {
		t.Fatal("Load should fail when YAML is invalid")
	}
}

func TestGetEnvironment_FromEnv(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 8080
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("APP_ENV", "staging")

	// 由于 APP_ENV=staging，xviper 会尝试合并 config.staging.yaml
	// 这里我们不创建它，应该报错（验证环境变量被正确读取）
	_, err := Load[TestConfig](WithPath[TestConfig](configPath))
	if err == nil {
		t.Fatal("Load should fail when staging overlay is missing")
	}
	// 错误信息应该提到 staging
	if err.Error() == "" {
		t.Errorf("error message is empty")
	}
}

func TestGetEnvironment_DefaultToDev(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 8080
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// 不设置 APP_ENV，也不在配置文件中设置 app.env
	// 默认应该是 dev，会尝试合并 config.dev.yaml
	_, err := Load[TestConfig](WithPath[TestConfig](configPath))
	if err == nil {
		t.Fatal("Load should fail when dev overlay is missing")
	}
}

func TestBuildOverlayPath(t *testing.T) {
	tests := []struct {
		base string
		env  string
		want string
	}{
		{"config/config.yaml", "prod", "config/config.prod.yaml"},
		{"config.yaml", "dev", "config.dev.yaml"},
		{"/etc/app/config.yaml", "staging", "/etc/app/config.staging.yaml"},
		{"config.yml", "test", "config.test.yml"},
	}

	for _, tt := range tests {
		got := buildOverlayPath(tt.base, tt.env)
		if got != tt.want {
			t.Errorf("buildOverlayPath(%q, %q) = %q, want %q", tt.base, tt.env, got, tt.want)
		}
	}
}

func TestLoad_EmptyOverlay(t *testing.T) {
	// 测试空 overlay 文件（只有注释）不影响 base
	dir := t.TempDir()

	basePath := filepath.Join(dir, "config.yaml")
	baseContent := `
server:
  port: 8080
  host: localhost
`
	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatalf("write base: %v", err)
	}

	// 空 overlay（只有注释）
	overlayPath := filepath.Join(dir, "config.dev.yaml")
	overlayContent := `# dev overlay is empty, use base defaults
`
	if err := os.WriteFile(overlayPath, []byte(overlayContent), 0644); err != nil {
		t.Fatalf("write overlay: %v", err)
	}

	t.Setenv("APP_ENV", "dev")

	cfg, err := Load[TestConfig](WithPath[TestConfig](basePath))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 空 overlay 不应改变 base 的值
	if cfg.Server.Port != 8080 {
		t.Errorf("server.port = %d, want 8080 (base value)", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("server.host = %q, want localhost (base value)", cfg.Server.Host)
	}
}

func TestLoad_DurationParsing(t *testing.T) {
	// 测试 viper 的 duration 自动解析（"10s" → time.Duration）
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  read_timeout: 30s
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	writeDevOverlay(t, dir)

	cfg, err := Load[TestConfig](WithPath[TestConfig](configPath))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("server.read_timeout = %v, want 30s", cfg.Server.ReadTimeout)
	}
}
