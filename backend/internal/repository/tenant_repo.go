package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/database"
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

// ListWithFilters 条件查询租户列表
func (r *TenantRepo) ListWithFilters(ctx context.Context, offset, limit int, code, name string, status int) ([]*model.Tenant, int64, error) {
	q := r.q.Tenant.WithContext(ctx)

	// 构建查询条件
	if code != "" {
		q = q.Where(r.q.Tenant.TenantCode.Like("%" + code + "%"))
	}
	if name != "" {
		q = q.Where(r.q.Tenant.Name.Like("%" + name + "%"))
	}
	if status != 0 {
		q = q.Where(r.q.Tenant.Status.Eq(int16(status)))
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	tenants, err := q.
		Order(r.q.Tenant.CreatedAt.Desc()).
		Offset(offset).
		Limit(limit).
		Find()

	return tenants, total, err
}

// CheckExists 检查租户编码是否存在
func (r *TenantRepo) CheckExists(ctx context.Context, tenantCode string, excludeTenantID ...string) (bool, error) {
	q := r.q.Tenant.WithContext(ctx).Where(r.q.Tenant.TenantCode.Eq(tenantCode))

	if len(excludeTenantID) > 0 && excludeTenantID[0] != "" {
		q = q.Where(r.q.Tenant.TenantID.Neq(excludeTenantID[0]))
	}

	count, err := q.Count()
	return count > 0, err
}

// UpdateStatus 更新租户状态
func (r *TenantRepo) UpdateStatus(ctx context.Context, tenantID string, status int) error {
	_, err := r.q.Tenant.WithContext(ctx).
		Where(r.q.Tenant.TenantID.Eq(tenantID)).
		Update(r.q.Tenant.Status, status)
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

// GetByCodeManual 根据租户编码获取租户信息（手动模式，用于跨租户查询）
// 使用场景：登录时需要通过租户code查询租户，此时还没有当前租户信息
func (r *TenantRepo) GetByCodeManual(ctx context.Context, tenantCode string) (*model.Tenant, error) {
	// 跨租户查询：使用 ManualTenantMode 禁止自动添加当前租户过滤
	ctx = database.ManualTenantMode(ctx)
	return r.q.Tenant.WithContext(ctx).
		Where(r.q.Tenant.TenantCode.Eq(tenantCode)).
		First()
}

// GetByIDManual 根据租户ID获取租户信息（手动模式，用于跨租户查询）
// 使用场景：需要查询任意租户信息，不受当前用户租户限制
func (r *TenantRepo) GetByIDManual(ctx context.Context, tenantID string) (*model.Tenant, error) {
	ctx = database.ManualTenantMode(ctx)
	return r.q.Tenant.WithContext(ctx).
		Where(r.q.Tenant.TenantID.Eq(tenantID)).
		First()
}

// GetByCodes 根据租户编码列表批量获取租户信息（手动模式，用于跨租户查询）
// 使用场景：需要查询任意租户信息，不受当前用户租户限制
func (r *TenantRepo) GetByCodes(ctx context.Context, tenantCodes []string) ([]*model.Tenant, error) {
	// 跨租户查询：使用 ManualTenantMode 禁止自动添加当前租户过滤
	ctx = database.ManualTenantMode(ctx)
	return r.q.Tenant.WithContext(ctx).Where(r.q.Tenant.TenantCode.In(tenantCodes...)).Find()
}

// ListAll 获取所有启用的租户列表（手动模式，用于跨租户查询）
// 使用场景：超管需要查看所有可切换的租户
func (r *TenantRepo) ListAll(ctx context.Context) ([]*model.Tenant, error) {
	// 跨租户查询：使用 ManualTenantMode 禁止自动添加当前租户过滤
	ctx = database.ManualTenantMode(ctx)
	return r.q.Tenant.WithContext(ctx).
		Where(r.q.Tenant.Status.Eq(int16(1))). // 1 = 启用状态
		Find()
}
