package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// TestNew_SetGet 用 miniredis 起内存 Redis（纯 Go、test-only、不进生产二进制），
// 验证 New 成功连接并可 Set/Get 往返。miniredis 是 Go 社区测 go-redis 代码的事实标准，
// 且满足项目 reusable-package.md「测试自包含、不依赖外部服务」的要求。
func TestNew_SetGet(t *testing.T) {
	mr := miniredis.RunT(t)

	client, err := New(Config{Addr: mr.Addr()})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Set(ctx, "k", "v", time.Minute).Err(); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	got, err := client.Get(ctx, "k").Result()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != "v" {
		t.Errorf("Get() = %q, want %q", got, "v")
	}
}

// TestNew_ConnRefused 验证连不上时 New 返回 error（fail-fast），无需任何外部库。
func TestNew_ConnRefused(t *testing.T) {
	// 127.0.0.1:1 无监听，ping 必失败
	if _, err := New(Config{Addr: "127.0.0.1:1", DialTimeout: 1}); err == nil {
		t.Fatal("New() error = nil, want ping failure")
	}
}

// TestBuildOptions 纯映射逻辑单测（零依赖、零网络）：验证 >0 时覆盖、为 0 时保留 go-redis 默认。
// 这是本包唯一自有的逻辑（Config → *redis.Options 映射），单独抽出来测最有价值。
func TestBuildOptions(t *testing.T) {
	// 全部 >0：应逐字段覆盖
	opts := buildOptions(Config{
		Addr:         "h:6379",
		Password:     "p",
		DB:           2,
		PoolSize:     5,
		MinIdleConns: 1,
		MaxRetries:   2,
		DialTimeout:  3,
		ReadTimeout:  4,
		WriteTimeout: 5,
	})
	if opts.Addr != "h:6379" || opts.Password != "p" || opts.DB != 2 {
		t.Errorf("基础字段映射错误: %+v", opts)
	}
	if opts.PoolSize != 5 || opts.MinIdleConns != 1 || opts.MaxRetries != 2 {
		t.Errorf("pool/retry 覆盖错误: %+v", opts)
	}
	if opts.DialTimeout != 3*time.Second || opts.ReadTimeout != 4*time.Second || opts.WriteTimeout != 5*time.Second {
		t.Errorf("timeout 覆盖错误: dial=%v read=%v write=%v", opts.DialTimeout, opts.ReadTimeout, opts.WriteTimeout)
	}

	// 全部为 0：buildOptions 不显式赋值，字段留零值（交由 redis.NewClient 内部填默认）
	def := buildOptions(Config{Addr: "h:6379"})
	if def.PoolSize != 0 {
		t.Errorf("PoolSize = %d, want 0(交由 go-redis 取默认)", def.PoolSize)
	}
	if def.DialTimeout != 0 {
		t.Errorf("DialTimeout = %v, want 0(交由 go-redis 取默认)", def.DialTimeout)
	}
}
