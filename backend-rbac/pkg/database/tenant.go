package database

import (
	"context"

	"admin/pkg/xcontext"

	"gorm.io/gorm"
)

// TenantScope 返回 GORM Scope 函数，为查询添加租户过滤条件
// 使用 GORM 原生 Scopes 模式：db.Scopes(database.TenantScope(tenantID))
func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if tenantID == "" {
			return db
		}
		return db.Where("tenant_id = ?", tenantID)
	}
}

// TenantScoped 接口：带 tenant_id 的模型可实现此接口
type TenantScoped interface {
	SetTenantID(tenantID string)
}

// SetTenantIDFromCtx 从 context 获取 tenant_id 并设置到模型
// 推荐直接在 Repository 中赋值：model.TenantID = xcontext.GetTenantID(ctx)
func SetTenantIDFromCtx(ctx context.Context, model TenantScoped) {
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID != "" {
		model.SetTenantID(tenantID)
	}
}
