package captcha

import (
	"context"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

// Manager 验证码管理器
type Manager struct {
	store *RedisStore
}

// RedisStore Redis存储实现
type RedisStore struct {
	expiration time.Duration
	keyPrefix  string
	redis      redis.UniversalClient
}

// NewManager 创建验证码管理器（使用全局Redis客户端）
func NewManager(rdb redis.UniversalClient) *Manager {
	store := &RedisStore{
		expiration: time.Minute * 5, // 验证码5分钟过期
		keyPrefix:  "captcha:",
		redis:      rdb,
	}

	return &Manager{
		store: store,
	}
}

// Generate 生成图形验证码
// 返回：验证码ID, Base64图片数据, 验证码答案(用于调试), 错误
func (m *Manager) Generate() (id, b64s, answer string, err error) {
	// 配置验证码：高度80，宽度240，4位数字，干扰度0.7，最多80个干扰点
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, m.store)
	return c.Generate()
}

// Verify 验证验证码
func (m *Manager) Verify(id, answer string) bool {
	return m.store.Verify(id, answer, true)
}

// Set 实现base64Captcha.Store接口
func (s *RedisStore) Set(id string, value string) error {
	ctx := context.Background()
	key := s.keyPrefix + id
	return s.redis.Set(ctx, key, value, s.expiration).Err()
}

// Get 实现base64Captcha.Store接口
func (s *RedisStore) Get(id string, clear bool) string {
	ctx := context.Background()
	key := s.keyPrefix + id
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return ""
	}

	if clear {
		s.redis.Del(ctx, key)
	}

	return val
}

// Verify 实现base64Captcha.Store接口
func (s *RedisStore) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	return val != "" && val == answer
}
