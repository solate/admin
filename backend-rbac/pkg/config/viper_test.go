package config

import (
	"testing"
)

func TestLoadConfigFromRoot(t *testing.T) {
	t.Setenv("CONFIG_DIR", "../../config")
	t.Setenv("APP_ENV", "prod")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	t.Logf("App: name=%s port=%d env=%s", cfg.App.Name, cfg.App.Port, cfg.App.Env)
	t.Logf("DB: host=%s port=%d user=%s dbname=%s sslmode=%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.DBName, cfg.Database.SSLMode)
	t.Logf("Redis: addr=%s db=%d", cfg.Redis.GetAddr(), cfg.Redis.DB)
}
