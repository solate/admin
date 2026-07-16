package xredis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config Redis 连接参数。无 mapstructure tag，由 main 从 internal/config 逐字段映射。
type Config struct {
	Addr         string // 地址 host:port
	Password     string // 密码（建议 APP_REDIS_PASSWORD 注入）
	DB           int    // 库编号
	PoolSize     int    // 连接池大小，0 用 go-redis 默认
	MinIdleConns int    // 最小空闲连接，0 用默认
	MaxRetries   int    // 命令重试次数，0 用默认
	DialTimeout  int    // 建连超时(秒)，0 用默认
	ReadTimeout  int    // 读超时(秒)，0 用默认
	WriteTimeout int    // 写超时(秒)，0 用默认
}

// New 创建单节点 Redis 客户端并探活。
// 返回 go-redis 原生 *redis.Client（不套接口、不做单例），与 pkg/xgorm 返回 *gorm.DB 风格一致。
// 后期若上集群，只需把 redis.NewClient 换成 redis.NewUniversalClient 并调整返回类型，改动收敛在此一处。
func New(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(buildOptions(cfg))

	// 构造即探活：ping 失败立刻关闭并返回 error（fail-fast）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

// buildOptions 把项目 Config 映射为 go-redis 的 *redis.Options。
// 连接池/超时等可选参数仅在 >0 时覆盖，为 0 保留 go-redis 默认值——这段纯映射逻辑可零依赖单测。
func buildOptions(cfg Config) *redis.Options {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opts.MinIdleConns = cfg.MinIdleConns
	}
	if cfg.MaxRetries > 0 {
		opts.MaxRetries = cfg.MaxRetries
	}
	if cfg.DialTimeout > 0 {
		opts.DialTimeout = time.Duration(cfg.DialTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		opts.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		opts.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}
	return opts
}
