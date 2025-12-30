package cache

import (
	"admin/internal/dal/model"
	"admin/pkg/constants"
	"admin/pkg/database"
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// TenantCache 租户缓存
type TenantCache struct {
	once   sync.Once
	tenant *model.Tenant
}

// Init 初始化租户缓存
func (c *TenantCache) Init(db *gorm.DB) error {
	var err error
	c.once.Do(func() {
		var tenant model.Tenant
		// 跳过租户检查，查询默认租户
		ctx := database.SkipTenantCheck(context.Background())
		if dbErr := db.WithContext(ctx).Where("tenant_code = ?", constants.DefaultTenantCode).First(&tenant).Error; dbErr != nil {
			err = fmt.Errorf("failed to load default tenant: %w", dbErr)
			return
		}
		c.tenant = &tenant
	})
	return err
}

// GetDefaultTenantID 获取默认租户ID
func (c *TenantCache) GetDefaultTenantID() string {
	if c.tenant == nil {
		panic("default tenant not initialized, call cache.Init() first")
	}
	return c.tenant.TenantID
}

// IsDefaultTenant 判断是否为默认租户
func (c *TenantCache) IsDefaultTenant(tenantID string) bool {
	return c.GetDefaultTenantID() == tenantID
}

// Reload 重新加载缓存
func (c *TenantCache) Reload(db *gorm.DB) error {
	var tenant model.Tenant
	// 跳过租户检查，查询默认租户
	ctx := database.SkipTenantCheck(context.Background())
	if err := db.WithContext(ctx).Where("tenant_code = ?", constants.DefaultTenantCode).First(&tenant).Error; err != nil {
		return fmt.Errorf("failed to reload default tenant: %w", err)
	}
	c.tenant = &tenant
	return nil
}

// Reset 重置缓存（主要用于测试）
func (c *TenantCache) Reset() {
	c.tenant = nil
	c.once = sync.Once{}
}
