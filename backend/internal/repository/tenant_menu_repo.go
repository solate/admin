package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

// TenantMenuRepo 租户菜单边界数据访问层
type TenantMenuRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewTenantMenuRepo(db *gorm.DB) *TenantMenuRepo {
	return &TenantMenuRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建租户菜单关联
func (r *TenantMenuRepo) Create(ctx context.Context, tm *model.TenantMenu) error {
	return r.q.TenantMenu.WithContext(ctx).Create(tm)
}

// CreateBatch 批量创建租户菜单关联
func (r *TenantMenuRepo) CreateBatch(ctx context.Context, tms []*model.TenantMenu) error {
	if len(tms) == 0 {
		return nil
	}
	return r.q.TenantMenu.WithContext(ctx).Create(tms...)
}

// DeleteByTenant 删除租户的所有菜单关联
func (r *TenantMenuRepo) DeleteByTenant(ctx context.Context, tenantID string) error {
	_, err := r.q.TenantMenu.WithContext(ctx).
		Where(r.q.TenantMenu.TenantID.Eq(tenantID)).
		Delete()
	return err
}

// GetMenuIDsByTenant 获取租户分配的菜单ID列表
func (r *TenantMenuRepo) GetMenuIDsByTenant(ctx context.Context, tenantID string) ([]string, error) {
	tms, err := r.q.TenantMenu.WithContext(ctx).
		Where(r.q.TenantMenu.TenantID.Eq(tenantID)).
		Find()
	if err != nil {
		return nil, err
	}

	menuIDs := make([]string, 0, len(tms))
	for _, tm := range tms {
		menuIDs = append(menuIDs, tm.MenuID)
	}
	return menuIDs, nil
}

// Exists 检查租户是否分配了指定菜单
func (r *TenantMenuRepo) Exists(ctx context.Context, tenantID, menuID string) (bool, error) {
	count, err := r.q.TenantMenu.WithContext(ctx).
		Where(r.q.TenantMenu.TenantID.Eq(tenantID)).
		Where(r.q.TenantMenu.MenuID.Eq(menuID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByTenant 获取租户的所有菜单关联
func (r *TenantMenuRepo) GetByTenant(ctx context.Context, tenantID string) ([]*model.TenantMenu, error) {
	return r.q.TenantMenu.WithContext(ctx).
		Where(r.q.TenantMenu.TenantID.Eq(tenantID)).
		Find()
}
