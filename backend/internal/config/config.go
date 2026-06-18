// Package config 是项目专属的配置层:定义本项目的类型化 Config 结构、
// 显式环境变量覆盖、显式校验。加载机制复用 admin/pkg/xconfig。
//
// 新项目复用框架时,只需改本包:Config 字段、Override 的环境变量名、validate 规则。
package config

import (
	"fmt"
	"time"

	"admin/pkg/xconfig"
)

// Config 是应用总配置结构。仅 mapstructure tag(给 viper 用)。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 是 HTTP 服务配置。
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"` // debug/release/test → Step 03 喂 gin.SetMode
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	GracefulTimeout time.Duration `mapstructure:"graceful_timeout"` // 优雅关闭超时
	Cors            CorsConfig    `mapstructure:"cors"`
}

// CorsConfig 是跨域配置(Step 03 中间件使用)。
type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// DatabaseConfig 是 PostgreSQL 连接配置。
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
}

// RedisConfig 是 Redis 连接配置。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig 是 JWT 签发配置。
type JWTConfig struct {
	AccessSecret  string `mapstructure:"access_secret"`
	RefreshSecret string `mapstructure:"refresh_secret"`
	AccessTTL     int    `mapstructure:"access_ttl"`  // 分钟
	RefreshTTL    int    `mapstructure:"refresh_ttl"` // 分钟
	Issuer        string `mapstructure:"issuer"`
}

// LogConfig 是日志配置。
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load 从 path 读取配置并返回校验过的 *Config。
// 流程:xconfig 加载文件 → 显式环境变量覆盖敏感字段 → validate 校验。
func Load(path string) (*Config, error) {
	var cfg Config
	loader := xconfig.New(xconfig.WithFile(path))
	if err := loader.Load(&cfg); err != nil {
		return nil, err
	}
	// 显式环境变量覆盖:类型化、无反射、集中在项目层。
	xconfig.Override(&cfg.Database.Password, "DB_PASSWORD")
	xconfig.Override(&cfg.Redis.Password, "REDIS_PASSWORD")
	xconfig.Override(&cfg.JWT.AccessSecret, "JWT_ACCESS_SECRET")
	xconfig.Override(&cfg.JWT.RefreshSecret, "JWT_REFRESH_SECRET")
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}

// validate 显式校验关键字段(无反射、无第三方校验库)。
func validate(c *Config) error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be in [1,65535], got %d", c.Server.Port)
	}
	switch c.Server.Mode {
	case "", "debug", "release", "test":
	default:
		return fmt.Errorf("server.mode invalid: %q", c.Server.Mode)
	}
	if c.Server.ReadTimeout <= 0 || c.Server.WriteTimeout <= 0 || c.Server.GracefulTimeout <= 0 {
		return fmt.Errorf("server timeouts must be positive")
	}
	if len(c.Server.Cors.AllowedOrigins) == 0 {
		return fmt.Errorf("server.cors.allowed_origins required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database.dbname required")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("database.port must be in [1,65535]")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database.password required (set DB_PASSWORD or config)")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr required")
	}
	if c.JWT.AccessSecret == "" || c.JWT.RefreshSecret == "" {
		return fmt.Errorf("jwt secrets required (set JWT_ACCESS_SECRET/JWT_REFRESH_SECRET)")
	}
	if c.JWT.AccessTTL < 1 || c.JWT.RefreshTTL < 1 {
		return fmt.Errorf("jwt ttl must be positive (minutes)")
	}
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log.level invalid: %q", c.Log.Level)
	}
	switch c.Log.Format {
	case "json", "console":
	default:
		return fmt.Errorf("log.format invalid: %q", c.Log.Format)
	}
	return nil
}
