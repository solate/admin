package seeds

import (
	"admin/internal/dal/model"
	"admin/pkg/constants"
	"fmt"

	"gorm.io/gorm"
)

// SeedTenant 初始化默认租户
func SeedTenant(db *gorm.DB, tenantID string) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := db.Where("tenant_code = ?", constants.DefaultTenantCode).First(&tenant).Error; err != nil {
		// 租户不存在，创建新租户
		tenant = model.Tenant{
			TenantID:   tenantID,
			TenantCode: constants.DefaultTenantCode,
			Name:       "默认租户",
			Status:     1,
		}
		if err := db.Create(&tenant).Error; err != nil {
			return nil, fmt.Errorf("创建默认租户失败: %w", err)
		}
		fmt.Printf("✅ 默认租户创建成功 tenant_id=%s code=%s name=%s\n", tenant.TenantID, tenant.TenantCode, tenant.Name)
	} else {
		fmt.Printf("ℹ️  默认租户已存在 tenant_id=%s code=%s name=%s\n", tenant.TenantID, tenant.TenantCode, tenant.Name)
	}

	return &tenant, nil
}
