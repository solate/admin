package config

import (
	"fmt"
	"time"
)

// Config 全局配置结构
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Log       LogConfig       `mapstructure:"log"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type ServerConfig struct {
	Mode string    `mapstructure:"mode"`
	TLS  TLSConfig `mapstructure:"tls"`
}

// TLSConfig TLS/HTTPS 配置
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`   // 是否启用 TLS
	CertFile string `mapstructure:"cert_file"` // 证书文件路径
	KeyFile  string `mapstructure:"key_file"`  // 私钥文件路径
	// MinVersion  string `mapstructure:"min_version"`  // 最低 TLS 版本 (1.0, 1.1, 1.2, 1.3)
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	Type         string `mapstructure:"type"` // node 或 cluster
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
	DialTimeout  int    `mapstructure:"dial_timeout"`  // seconds
	ReadTimeout  int    `mapstructure:"read_timeout"`  // seconds
	WriteTimeout int    `mapstructure:"write_timeout"` // seconds
}

// JWTSettings JWT 配置（来自配置文件）
// 说明：
// - secret 为签名密钥（若不区分 access/refresh，可两者共用）
// - access_expire/refresh_expire 为过期时间（秒）
// - issuer 可选，用于在生成注册声明时设置发行者
type JWTConfig struct {
	AccessSecret  string `mapstructure:"access_secret"`
	AccessExpire  int64  `mapstructure:"access_expire"`
	RefreshSecret string `mapstructure:"refresh_secret"`
	RefreshExpire int64  `mapstructure:"refresh_expire"`
	Issuer        string `mapstructure:"issuer"`
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetConnMaxLifetime 获取连接最大生命周期
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetDialTimeout 获取Redis连接超时时间
func (c *RedisConfig) GetDialTimeout() time.Duration {
	return time.Duration(c.DialTimeout) * time.Second
}

// GetReadTimeout 获取Redis读超时时间
func (c *RedisConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout 获取Redis写超时时间
func (c *RedisConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetAddr 获取Redis地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetAccessExpire 获取访问令牌过期时间
func (c *JWTConfig) GetAccessExpire() time.Duration {
	return time.Duration(c.AccessExpire) * time.Second
}

// GetRefreshExpire 获取刷新令牌过期时间
func (c *JWTConfig) GetRefreshExpire() time.Duration {
	return time.Duration(c.RefreshExpire) * time.Second
}
