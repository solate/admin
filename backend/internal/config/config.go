// Package config 是项目专属配置层,包含类型化 Config 结构体与校验逻辑。
//
// 配置加载委托给 pkg/xviper(通用可复用库),本包只定义业务结构与校验规则。
//
// 环境变量覆盖由 xviper 的 AutomaticEnv + ExperimentalBindStruct 自动完成:
// 配置 key server.port 对应环境变量 APP_SERVER_PORT(APP_ 前缀 + 点转下划线大写)。
// 密钥字段用 APP_DATABASE_PASSWORD / APP_REDIS_PASSWORD / APP_JWT_ACCESS_SECRET / APP_JWT_REFRESH_SECRET 注入。
//
// 多环境用 APP_ENV 环境变量:未设=dev(合并 config.dev.yaml),设为 prod=合并 config.prod.yaml(overlay 只写差异,DRY)。
package config

import (
	"fmt"
	"time"

	"admin/pkg/xviper"
)

// Config 是应用总配置结构。使用 mapstructure tag(xviper 约定,兼容 viper)。
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
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`     // json(生产) / text / console
	AddSource bool   `mapstructure:"add_source"` // 记录调用位置 file:line
}

// Load 从 basePath 读取配置并返回校验过的 *Config。
//
// 加载流程:xviper.Load 读 base(+APP_ENV overlay,环境变量自动覆盖)→ validate 校验。
//
// APP_ENV 决定 overlay:未设=dev(合并 config.dev.yaml),设为 prod=合并 config.prod.yaml(只写环境差异)。
// overlay 文件必须存在(缺失由 xviper MergeInConfig fail-fast),故已补建空的 config.dev.yaml。
//
// 环境变量覆盖:xviper 的 AutomaticEnv(APP_ 前缀)自动完成,key server.port → APP_SERVER_PORT。
// 密钥字段用 APP_DATABASE_PASSWORD / APP_REDIS_PASSWORD / APP_JWT_ACCESS_SECRET / APP_JWT_REFRESH_SECRET 注入。
func Load(basePath string) (*Config, error) {
	cfg, err := xviper.Load[Config](xviper.WithPath[Config](basePath), xviper.WithValidate(validate))
	if err != nil {
		return nil, err
	}
	return cfg, nil
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
		return fmt.Errorf("database.password required (set APP_DATABASE_PASSWORD or config)")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr required")
	}
	if c.JWT.AccessSecret == "" || c.JWT.RefreshSecret == "" {
		return fmt.Errorf("jwt secrets required (set APP_JWT_ACCESS_SECRET/APP_JWT_REFRESH_SECRET)")
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
	case "", "json", "text", "console":
	default:
		return fmt.Errorf("log.format invalid: %q", c.Log.Format)
	}
	return nil
}
