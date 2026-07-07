package config

import "fmt"

// validate 是配置校验总入口，按业务块分派到各子校验函数。
// 无反射、无第三方校验库；如需跨块校验（字段间关系），在此函数内追加。
func validate(c *Config) error {
	if err := validateApp(&c.App); err != nil {
		return err
	}
	if err := validateServer(&c.Server); err != nil {
		return err
	}
	if err := validateDatabase(&c.Database); err != nil {
		return err
	}
	if err := validateRedis(&c.Redis); err != nil {
		return err
	}
	if err := validateJWT(&c.JWT); err != nil {
		return err
	}
	return validateLog(&c.Log)
}

// validateApp 校验应用元信息配置。Name 用作日志 service 字段，必填；Env 可空。
func validateApp(a *AppConfig) error {
	if a.Name == "" {
		return fmt.Errorf("app.name required")
	}
	return nil
}

// validateServer 校验 HTTP 服务配置。
func validateServer(s *ServerConfig) error {
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("server.port must be in [1,65535], got %d", s.Port)
	}
	switch s.Mode {
	case "", "debug", "release", "test":
	default:
		return fmt.Errorf("server.mode invalid: %q", s.Mode)
	}
	if s.ReadTimeout <= 0 || s.WriteTimeout <= 0 || s.GracefulTimeout <= 0 {
		return fmt.Errorf("server timeouts must be positive")
	}
	if len(s.Cors.AllowedOrigins) == 0 {
		return fmt.Errorf("server.cors.allowed_origins required")
	}
	return nil
}

// validateDatabase 校验 PostgreSQL 连接配置。
func validateDatabase(d *DatabaseConfig) error {
	if d.Host == "" {
		return fmt.Errorf("database.host required")
	}
	if d.DBName == "" {
		return fmt.Errorf("database.dbname required")
	}
	if d.Port < 1 || d.Port > 65535 {
		return fmt.Errorf("database.port must be in [1,65535]")
	}
	if d.Password == "" {
		return fmt.Errorf("database.password required (set APP_DATABASE_PASSWORD or config)")
	}
	return nil
}

// validateRedis 校验 Redis 连接配置。
func validateRedis(r *RedisConfig) error {
	if r.Addr == "" {
		return fmt.Errorf("redis.addr required")
	}
	return nil
}

// validateJWT 校验 JWT 签发配置。
func validateJWT(j *JWTConfig) error {
	if j.AccessSecret == "" || j.RefreshSecret == "" {
		return fmt.Errorf("jwt secrets required (set APP_JWT_ACCESS_SECRET/APP_JWT_REFRESH_SECRET)")
	}
	if j.AccessTTL < 1 || j.RefreshTTL < 1 {
		return fmt.Errorf("jwt ttl must be positive (minutes)")
	}
	return nil
}

// validateLog 校验日志配置。
func validateLog(l *LogConfig) error {
	switch l.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log.level invalid: %q", l.Level)
	}
	switch l.Format {
	case "", "json", "text":
	default:
		return fmt.Errorf("log.format invalid: %q", l.Format)
	}
	return nil
}
