package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/database"
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
	return r.q.Position.WithContext(ctx).Create(position)
}

// GetByID 根据ID获取岗位
func (r *PositionRepo) GetByID(ctx context.Context, positionID string) (*model.Position, error) {
	return r.q.Position.WithContext(ctx).Where(r.q.Position.PositionID.Eq(positionID)).First()
}

// GetByCode 根据岗位编码获取当前租户的岗位（依赖自动模式添加 tenant_id 过滤）
func (r *PositionRepo) GetByCode(ctx context.Context, positionCode string) (*model.Position, error) {
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.PositionCode.Eq(positionCode)).
		First()
}

// GetByCodeWithTenant 根据租户ID和岗位编码获取岗位（手动模式，用于跨租户查询）
func (r *PositionRepo) GetByCodeWithTenant(ctx context.Context, tenantID, positionCode string) (*model.Position, error) {
	// 跨租户查询：使用 ManualTenantMode 禁止自动添加当前租户过滤
	ctx = database.ManualTenantMode(ctx)
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.TenantID.Eq(tenantID)).
		Where(r.q.Position.PositionCode.Eq(positionCode)).
		First()
}

// Update 更新岗位
func (r *PositionRepo) Update(ctx context.Context, positionID string, updates map[string]interface{}) error {
	_, err := r.q.Position.WithContext(ctx).Where(r.q.Position.PositionID.Eq(positionID)).Updates(updates)
	return err
}

// Delete 删除岗位(软删除)
func (r *PositionRepo) Delete(ctx context.Context, positionID string) error {
	_, err := r.q.Position.WithContext(ctx).Where(r.q.Position.PositionID.Eq(positionID)).Delete()
	return err
}

// List 分页获取岗位列表
func (r *PositionRepo) List(ctx context.Context, offset, limit int) ([]*model.Position, int64, error) {
	return r.q.Position.WithContext(ctx).FindByPage(offset, limit)
}

// ListWithFilters 根据筛选条件分页获取岗位列表（支持自动租户过滤）
func (r *PositionRepo) ListWithFilters(ctx context.Context, offset, limit int, keywordFilter string, statusFilter int) ([]*model.Position, int64, error) {
	query := r.q.Position.WithContext(ctx)

	// 应用筛选条件
	if keywordFilter != "" {
		query = query.Where(r.q.Position.PositionName.Like("%"+keywordFilter+"%")).
			Or(r.q.Position.PositionCode.Like("%"+keywordFilter+"%"))
	}
	if statusFilter != 0 {
		query = query.Where(r.q.Position.Status.Eq(int16(statusFilter)))
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	positions, err := query.Order(r.q.Position.Sort.Asc()).Offset(offset).Limit(limit).Find()
	return positions, total, err
}

// UpdateStatus 更新岗位状态
func (r *PositionRepo) UpdateStatus(ctx context.Context, positionID string, status int) error {
	_, err := r.q.Position.WithContext(ctx).Where(r.q.Position.PositionID.Eq(positionID)).Update(r.q.Position.Status, int16(status))
	return err
}

// CheckExists 检查岗位是否存在（租户过滤由 scope callback 自动处理）
func (r *PositionRepo) CheckExists(ctx context.Context, tenantID, positionCode string) (bool, error) {
	count, err := r.q.Position.WithContext(ctx).
		Where(r.q.Position.PositionCode.Eq(positionCode)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckExistsByID 检查岗位编码是否存在（排除指定ID）（租户过滤由 scope callback 自动处理）
func (r *PositionRepo) CheckExistsByID(ctx context.Context, tenantID, positionCode string, excludePositionID string) (bool, error) {
	count, err := r.q.Position.WithContext(ctx).
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
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.PositionID.In(positionIDs...)).
		Find()
}

// ListByCodes 根据岗位编码列表获取岗位列表（跳过租户过滤，支持岗位继承）
func (r *PositionRepo) ListByCodes(ctx context.Context, positionCodes []string) ([]*model.Position, error) {
	ctx = database.ManualTenantMode(ctx)
	return r.q.Position.WithContext(ctx).
		Where(r.q.Position.PositionCode.In(positionCodes...)).
		Find()
}

// ListAll 获取租户的所有岗位（按排序权重升序）
func (r *PositionRepo) ListAll(ctx context.Context) ([]*model.Position, error) {
	return r.q.Position.WithContext(ctx).
		Order(r.q.Position.Sort.Asc()).
		Find()
}
