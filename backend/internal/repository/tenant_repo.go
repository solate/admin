package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

type TenantRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewTenantRepo(db *gorm.DB) *TenantRepo {
	return &TenantRepo{
		db: db,
		q:  query.Use(db),
	}
}

// GetByID 根据租户ID获取租户信息
func (r *TenantRepo) GetByID(ctx context.Context, tenantID string) (*model.Tenant, error) {
	return r.q.Tenant.WithContext(ctx).Where(r.q.Tenant.TenantID.Eq(tenantID)).First()
}

// GetByCode 根据租户编码获取租户信息
func (r *TenantRepo) GetByCode(ctx context.Context, tenantCode string) (*model.Tenant, error) {
	return r.q.Tenant.WithContext(ctx).Where(r.q.Tenant.TenantCode.Eq(tenantCode)).First()
}

// List 列出所有租户
func (r *TenantRepo) List(ctx context.Context) ([]*model.Tenant, error) {
	return r.q.Tenant.WithContext(ctx).Find()
}

// Create 创建租户
func (r *TenantRepo) Create(ctx context.Context, tenant *model.Tenant) error {
	return r.q.Tenant.WithContext(ctx).Create(tenant)
}

// Update 更新租户
func (r *TenantRepo) Update(ctx context.Context, tenantID string, updates map[string]interface{}) error {
	_, err := r.q.Tenant.WithContext(ctx).Where(r.q.Tenant.TenantID.Eq(tenantID)).Updates(updates)
	return err
}

// Delete 删除租户
func (r *TenantRepo) Delete(ctx context.Context, tenantID string) error {
	_, err := r.q.Tenant.WithContext(ctx).Where(r.q.Tenant.TenantID.Eq(tenantID)).Delete()
	return err
}

// GetByIDs 根据租户ID列表获取租户信息
func (r *TenantRepo) GetByIDs(ctx context.Context, tenantIDs []string) (map[string]*model.Tenant, error) {
	tenants, err := r.q.Tenant.WithContext(ctx).Where(r.q.Tenant.TenantID.In(tenantIDs...)).Find()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*model.Tenant, len(tenants))
	for _, tenant := range tenants {
		result[tenant.TenantID] = tenant
	}
	return result, nil
}
