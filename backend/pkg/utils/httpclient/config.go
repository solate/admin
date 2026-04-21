package httpclient

import (
	"time"
)

// Config HTTP 客户端配置
type Config struct {
	// 基础配置
	BaseURL string // 基础 URL

	// 超时配置
	Timeout         time.Duration // 总超时时间
	ConnectionTimeout time.Duration // 连接超时时间

	// 重试配置
	MaxRetries       int           // 最大重试次数（0 表示不重试）
	RetryWaitTime    time.Duration // 重试等待时间
	RetryMaxWaitTime time.Duration // 重试最大等待时间

	// 连接池配置
	MaxIdleConns        int           // 最大空闲连接数
	MaxIdleConnsPerHost int           // 每个主机的最大空闲连接数
	IdleConnTimeout     time.Duration // 空闲连接超时时间

	// 其他配置
	Debug              bool // 是否开启调试模式
	DisableKeepAlives  bool // 是否禁用长连接
	InsecureSkipVerify bool // 是否跳过 SSL 验证（不推荐生产环境使用）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:             30 * time.Second,
		ConnectionTimeout:   10 * time.Second,
		MaxRetries:          3,
		RetryWaitTime:       100 * time.Millisecond,
		RetryMaxWaitTime:    1 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		Debug:               false,
		DisableKeepAlives:   false,
		InsecureSkipVerify:  false,
	}
}

// Option 配置选项函数
type Option func(*Config)

// WithBaseURL 设置基础 URL
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithConnectionTimeout 设置连接超时时间
func WithConnectionTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ConnectionTimeout = timeout
	}
}

// WithRetry 设置重试配置
func WithRetry(maxRetries int, waitTime, maxWaitTime time.Duration) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
		c.RetryWaitTime = waitTime
		c.RetryMaxWaitTime = maxWaitTime
	}
}

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *Config) {
		c.MaxIdleConns = maxIdleConns
	}
}

// WithMaxIdleConnsPerHost 设置每个主机的最大空闲连接数
func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return func(c *Config) {
		c.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

// WithDebug 开启/关闭调试模式
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithInsecureSkipVerify 设置是否跳过 SSL 验证
func WithInsecureSkipVerify(skip bool) Option {
	return func(c *Config) {
		c.InsecureSkipVerify = skip
	}
}
