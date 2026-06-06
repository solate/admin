package xredis

import (
	"context"
	"testing"
	"time"
)

func localConfig() Config {
	return Config{
		Type:         "node",
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "123456",
		DB:           0,
		PoolSize:     5,
		MinIdleConns: 1,
		MaxRetries:   1,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

func connectOrSkip(t *testing.T) {
	t.Helper()
	_, err := Connect(localConfig())
	if err != nil {
		t.Skipf("local redis not available: %v", err)
	}
}

func TestConnectAndSetGet(t *testing.T) {
	connectOrSkip(t)

	ctx := context.Background()
	key := "redis:test:key"
	val := "ok"

	c := GetRedis()
	if c == nil {
		t.Fatalf("GetRedis returned nil")
	}

	if err := c.Set(ctx, key, val, 5*time.Second).Err(); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	got, err := c.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got != val {
		t.Fatalf("Get returned %q, want %q", got, val)
	}
}

func TestHealthCheck(t *testing.T) {
	connectOrSkip(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck error: %v", err)
	}
}
