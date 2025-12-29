package cache

import (
	"gorm.io/gorm"
	"sync"
)

// Cache 缓存管理器
type Cache struct {
	db     *gorm.DB
	Tenant *TenantCache
}

var (
	globalCache *Cache
	once        sync.Once
)

// NewCache 创建缓存管理器
func NewCache(db *gorm.DB) *Cache {
	return &Cache{
		db:     db,
		Tenant: &TenantCache{},
	}
}

// Init 初始化所有缓存（全局单例，应用启动时调用）
func Init(db *gorm.DB) error {
	var err error
	once.Do(func() {
		globalCache = NewCache(db)
		err = globalCache.init()
	})
	return err
}

// Get 获取全局缓存实例
func Get() *Cache {
	if globalCache == nil {
		panic("cache not initialized, call cache.Init() first")
	}
	return globalCache
}

// init 初始化所有缓存
func (c *Cache) init() error {
	return c.Tenant.Init(c.db)
}

// Reset 重置所有缓存（主要用于测试）
func Reset() {
	if globalCache != nil {
		globalCache.Tenant.Reset()
	}
	once = sync.Once{}
	globalCache = nil
}
