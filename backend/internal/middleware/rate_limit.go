package middleware

import (
    "admin/internal/dto"
    "admin/pkg/errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	limiterMap map[string]*rate.Limiter
	mu         sync.RWMutex
	r          rate.Limit
	b          int
}

// NewRateLimiter 创建限流器
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiterMap: make(map[string]*rate.Limiter),
		r:          r,
		b:          b,
	}
}

// getLimiter 获取指定IP的限流器
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiterMap[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.limiterMap[key] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

// RateLimit 限流中间件
func RateLimit(requestsPerSecond int, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate.Limit(requestsPerSecond), burst)

	return func(c *gin.Context) {
		// 使用IP作为限流key
		key := c.ClientIP()

		// 获取该IP的限流器
		l := limiter.getLimiter(key)

		// 检查是否允许请求
		if !l.Allow() {
			dto.Error(c, http.StatusTooManyRequests, errors.ErrTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}
