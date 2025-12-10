package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

var globalConfig *Config

// Load 加载配置文件
// 支持通过 APP_ENV 环境变量指定环境：dev, prod
// 加载顺序：config.yaml (基础配置) -> config.{env}.yaml (环境配置) -> 环境变量
func Load() (*Config, error) {

	// 1. 配置文件查找路径
	viper.SetConfigType("yaml")
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		viper.AddConfigPath(dir)
	}
	viper.AddConfigPath("./config")                   // 当前目录下的 config 文件夹
	viper.AddConfigPath("/root/admin/admin-backend/") // 系统配置目录

	// 2. 读取基础配置文件 config.yaml
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read base config file: %w", err)
	}
	fmt.Printf("Loaded base config: %s\n", viper.ConfigFileUsed())

	// 3. 获取环境变量，决定加载哪个配置文件
	env := getEnvironment()
	fmt.Printf("Loading configuration for environment: %s\n", env)

	// 4. 合并环境特定配置文件 config.{env}.yaml
	envConfigName := fmt.Sprintf("config.%s", env)
	viper.SetConfigName(envConfigName)
	if err := viper.MergeInConfig(); err != nil {
		// 环境配置文件不存在是可以接受的，使用默认配置即可
		fmt.Printf("No environment-specific config file found for %s (this is OK)\n", env)
	} else {
		fmt.Printf("Merged environment config: %s\n", viper.ConfigFileUsed())
	}

	// 5. 环境变量覆盖（支持下划线命名，如 APP_NAME, DATABASE_HOST）
	// viper.SetEnvPrefix("AB")                               // 环境变量前缀，AB_
	// viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 环境变量. 替换为 _，如 app.name -> AB_APP_NAME
	viper.AutomaticEnv() // 自动加载环境变量，优先级高

	// 6. 反序列化配置到结构体
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 7. 验证配置
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 8. 保存全局配置
	globalConfig = &cfg
	fmt.Printf("Configuration loaded successfully. App: %s, Port: %d, Env: %s\n",
		cfg.App.Name, cfg.App.Port, cfg.App.Env)

	return &cfg, nil
}

// getEnvironment 获取当前运行环境
func getEnvironment() string {
	// 优先级：APP_ENV > GIN_MODE > 默认值
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}

	// 兼容 Gin 的环境变量
	if ginMode := os.Getenv("GIN_MODE"); ginMode != "" {
		switch ginMode {
		case "release":
			return "prod"
		case "test":
			return "test"
		default:
			return "dev"
		}
	}

	if configEnv := viper.GetString("app.env"); configEnv != "" {
		return configEnv
	}

	// 默认开发环境
	return "dev"
}

// validateConfig 验证配置的有效性
func validateConfig(cfg *Config) error {
	// 验证必填项
	if cfg.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if cfg.App.Port <= 0 || cfg.App.Port > 65535 {
		return fmt.Errorf("invalid app.port: %d", cfg.App.Port)
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("database.dbname is required")
	}

	// 生产环境额外检查
	if cfg.App.Env == "prod" {
		if cfg.Server.Mode != "release" {
			fmt.Printf("WARNING: server.mode should be 'release' in production\n")
		}
	}

	return nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
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
