// Package config 定义项目配置结构与校验规则。加载机制(overlay/环境变量)见 pkg/xviper。
package config

import "admin/pkg/xviper"

// Config 是应用总配置结构。使用 mapstructure tag(xviper 约定,兼容 viper)。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`   // HTTP 服务
	Database DatabaseConfig `mapstructure:"database"` // PostgreSQL 连接
	Redis    RedisConfig    `mapstructure:"redis"`    // Redis 连接
	JWT      JWTConfig      `mapstructure:"jwt"`      // JWT 签发
	Log      LogConfig      `mapstructure:"log"`      // 日志
}

// ServerConfig 是 HTTP 服务配置。
type ServerConfig struct {
	Port            int        `mapstructure:"port"`             // 监听端口
	Mode            string     `mapstructure:"mode"`             // debug/release/test → Step 03 喂 gin.SetMode
	ReadTimeout     int        `mapstructure:"read_timeout"`     // 读超时(秒)
	WriteTimeout    int        `mapstructure:"write_timeout"`    // 写超时(秒)
	GracefulTimeout int        `mapstructure:"graceful_timeout"` // 优雅关闭超时(秒)
	Cors            CorsConfig `mapstructure:"cors"`             // 跨域
}

// CorsConfig 是跨域配置(Step 03 中间件使用)。
type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`   // 允许的来源
	AllowedMethods   []string `mapstructure:"allowed_methods"`   // 允许的方法
	AllowedHeaders   []string `mapstructure:"allowed_headers"`   // 允许的请求头
	AllowCredentials bool     `mapstructure:"allow_credentials"` // 是否允许携带凭证
}

// DatabaseConfig 是 PostgreSQL 连接配置。
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`              // 主机
	Port            int    `mapstructure:"port"`              // 端口
	User            string `mapstructure:"user"`              // 用户名
	Password        string `mapstructure:"password"`          // 密码(建议用 APP_DATABASE_PASSWORD 注入)
	DBName          string `mapstructure:"dbname"`            // 库名
	SSLMode         string `mapstructure:"sslmode"`           // SSL 模式
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int    `mapstructure:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 连接最大存活时间(秒)
}

// RedisConfig 是 Redis 连接配置。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`     // 地址(host:port)
	Password string `mapstructure:"password"` // 密码(建议用 APP_REDIS_PASSWORD 注入)
	DB       int    `mapstructure:"db"`       // 库编号
}

// JWTConfig 是 JWT 签发配置。
type JWTConfig struct {
	AccessSecret  string `mapstructure:"access_secret"`  // access token 密钥(建议用 APP_JWT_ACCESS_SECRET 注入)
	RefreshSecret string `mapstructure:"refresh_secret"` // refresh token 密钥(建议用 APP_JWT_REFRESH_SECRET 注入)
	AccessTTL     int    `mapstructure:"access_ttl"`     // access token 有效期(分钟)
	RefreshTTL    int    `mapstructure:"refresh_ttl"`    // refresh token 有效期(分钟)
	Issuer        string `mapstructure:"issuer"`         // 签发者
}

// LogConfig 是日志配置。
type LogConfig struct {
	Level     string `mapstructure:"level"`      // 级别 debug/info/warn/error
	Format    string `mapstructure:"format"`     // json(生产) / text / console
	AddSource bool   `mapstructure:"add_source"` // 记录调用位置 file:line
}

// InitConfig 按 xviper 约定加载配置并返回校验过的 *Config。
// 路径固定为 config/config.yaml(xviper 内建约定):docker/k3s 靠挂载覆盖该路径的内容,
// 环境切换靠 APP_ENV 走 overlay,均无需改路径或传参。
func InitConfig() (*Config, error) {
	cfg, err := xviper.Load[Config](xviper.WithValidate(validate))
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
