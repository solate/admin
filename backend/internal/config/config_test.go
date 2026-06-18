package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeProjectConfig(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

const fullProjectYAML = `
server:
  port: 8080
  mode: debug
  read_timeout: 10s
  write_timeout: 10s
  graceful_timeout: 30s
  cors:
    allowed_origins: ["http://localhost:5173"]
    allowed_methods: [GET, POST]
    allowed_headers: [Content-Type]
    allow_credentials: true
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: admin_dev
  sslmode: disable
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600
redis:
  addr: localhost:6379
  password: ""
  db: 0
jwt:
  access_secret: "access"
  refresh_secret: "refresh"
  access_ttl: 30
  refresh_ttl: 10080
  issuer: "admin"
log:
  level: debug
  format: console
`

func TestLoad_HappyPath(t *testing.T) {
	cfg, err := Load(writeProjectConfig(t, fullProjectYAML))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("port = %d", cfg.Server.Port)
	}
	if cfg.Server.GracefulTimeout.Seconds() != 30 {
		t.Errorf("graceful_timeout = %v", cfg.Server.GracefulTimeout)
	}
	if len(cfg.Server.Cors.AllowedOrigins) != 1 {
		t.Errorf("cors origins = %v", cfg.Server.Cors.AllowedOrigins)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("DB_PASSWORD", "env-secret")
	cfg, err := Load(writeProjectConfig(t, fullProjectYAML))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Database.Password != "env-secret" {
		t.Errorf("password = %q, want env-secret", cfg.Database.Password)
	}
}

func TestLoad_ValidateFailure(t *testing.T) {
	// 缺 database.host 与 dbname → validate 失败
	bad := `
server:
  port: 8080
  mode: debug
  read_timeout: 10s
  write_timeout: 10s
  graceful_timeout: 30s
  cors:
    allowed_origins: ["http://localhost:5173"]
database:
  port: 5432
  password: x
redis:
  addr: localhost:6379
jwt:
  access_secret: "a"
  refresh_secret: "b"
  access_ttl: 30
  refresh_ttl: 10080
log:
  level: debug
  format: console
`
	if _, err := Load(writeProjectConfig(t, bad)); err == nil {
		t.Fatal("expected validate error, got nil")
	}
}
