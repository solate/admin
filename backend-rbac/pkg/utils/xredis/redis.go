package xredis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// client Redis通用客户端（全局） 通过接口来实现单节点和集群 Client/ClusterClient
	client redis.UniversalClient
	once   sync.Once
)

// Config Redis配置
type Config struct {
	Type         string // node 或 cluster
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Connect 连接Redis
func Connect(cfg Config) (redis.UniversalClient, error) {

	once.Do(func() {
		client = newClient(cfg)
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return client, fmt.Errorf("failed to ping redis: %w", err)
	}
	return client, nil
}

func newClient(cfg Config) redis.UniversalClient {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// If type is explicitly cluster, use ClusterClient
	if cfg.Type == "cluster" {
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        []string{addr},
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		})
	}

	// Default to single node client
	// We can use NewClient which implements UniversalClient (via *Client)
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})
}

func GetRedis() redis.UniversalClient {
	return client
}

// Close 关闭Redis连接
func Close() error {
	if client != nil {
		if err := client.Close(); err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
	}
	return nil
}

// HealthCheck 检查Redis连接是否正常
func HealthCheck(ctx context.Context) error {
	c := GetRedis()
	if c == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := c.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}
