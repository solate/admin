package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/xcontext"
	"context"

	"gorm.io/gorm"
)

type PositionRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewPositionRepo(db *gorm.DB) *PositionRepo {
	return &PositionRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建岗位
func (r *PositionRepo) Create(ctx context.Context, position *model.Position) error {
	position.TenantID = xcontext.GetTenantID(ctx)
	return r.q.Position.WithContext(ctx).Create(position)
}

// GetByID 根据ID获取岗位
func (r *PositionRepo) GetByID(ctx context.Context, positionID string) (*model.Position, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionID.Eq(positionID)).
		First()
}

// GetByCode 根据岗位编码获取当前租户的岗位
func (r *PositionRepo) GetByCode(ctx context.Context, positionCode string) (*model.Position, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionCode.Eq(positionCode)).
		First()
}

// GetByCodeWithTenant 根据租户ID和岗位编码获取岗位（跨租户查询）
func (r *PositionRepo) GetByCodeWithTenant(ctx context.Context, tenantID, positionCode string) (*model.Position, error) {
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionCode.Eq(positionCode)).
		First()
}

// Update 更新岗位
func (r *PositionRepo) Update(ctx context.Context, positionID string, updates map[string]interface{}) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionID.Eq(positionID)).
		Updates(updates)
	return err
}

// Delete 删除岗位(软删除)
func (r *PositionRepo) Delete(ctx context.Context, positionID string) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionID.Eq(positionID)).
		Delete()
	return err
}

// BatchDelete 批量删除岗位
func (r *PositionRepo) BatchDelete(ctx context.Context, positionIDs []string) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionID.In(positionIDs...)).
		Delete()
	return err
}

// GetByIDs 根据岗位ID列表获取岗位信息
func (r *PositionRepo) GetByIDs(ctx context.Context, positionIDs []string) ([]*model.Position, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionID.In(positionIDs...)).
		Find()
}

// List 分页获取岗位列表
func (r *PositionRepo) List(ctx context.Context, offset, limit int) ([]*model.Position, int64, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		FindByPage(offset, limit)
}

// ListWithFilters 根据筛选条件分页获取岗位列表
func (r *PositionRepo) ListWithFilters(ctx context.Context, offset, limit int, positionName, positionCode string, statusFilter int) ([]*model.Position, int64, error) {
	tenantID := xcontext.GetTenantID(ctx)
	query := r.q.Position.WithContext(ctx).Where(r.q.Position.TenantID.Eq(tenantID))

	if positionName != "" {
		query = query.Where(r.q.Position.PositionName.Like("%" + positionName + "%"))
	}
	if positionCode != "" {
		query = query.Where(r.q.Position.PositionCode.Like("%" + positionCode + "%"))
	}
	if statusFilter != 0 {
		query = query.Where(r.q.Position.Status.Eq(int16(statusFilter)))
	}

	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	positions, err := query.Order(r.q.Position.Sort.Asc()).Offset(offset).Limit(limit).Find()
	return positions, total, err
}

// UpdateStatus 更新岗位状态
func (r *PositionRepo) UpdateStatus(ctx context.Context, positionID string, status int) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionID.Eq(positionID)).
		Update(r.q.Position.Status, int16(status))
	return err
}

// CheckExists 检查岗位是否存在
func (r *PositionRepo) CheckExists(ctx context.Context, tenantID, positionCode string) (bool, error) {
	count, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionCode.Eq(positionCode)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckExistsByID 检查岗位编码是否存在（排除指定ID）
func (r *PositionRepo) CheckExistsByID(ctx context.Context, tenantID, positionCode string, excludePositionID string) (bool, error) {
	count, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionCode.Eq(positionCode)).
		Where(r.q.Position.PositionID.Neq(excludePositionID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByIDs 根据岗位ID列表获取岗位列表
func (r *PositionRepo) ListByIDs(ctx context.Context, positionIDs []string) ([]*model.Position, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionID.In(positionIDs...)).
		Find()
}

// ListByCodes 根据岗位编码列表获取岗位列表（跨租户查询）
func (r *PositionRepo) ListByCodes(ctx context.Context, positionCodes []string) ([]*model.Position, error) {
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.PositionCode.In(positionCodes...)).
		Find()
}

// ListAll 获取租户的所有岗位（按排序权重升序）
func (r *PositionRepo) ListAll(ctx context.Context) ([]*model.Position, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Order(r.q.Position.Sort.Asc()).
		Find()
}
