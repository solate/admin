package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/xcontext"
	"context"

	"gorm.io/gorm"
)

// DictTypeRepo 字典类型仓库
type DictTypeRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewDictTypeRepo(db *gorm.DB) *DictTypeRepo {
	return &DictTypeRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建字典类型
func (r *DictTypeRepo) Create(ctx context.Context, dictType *model.DictType) error {
	dictType.TenantID = xcontext.GetTenantID(ctx)
	return r.q.DictType.WithContext(ctx).Create(dictType)
}

// GetByID 根据ID获取字典类型
func (r *DictTypeRepo) GetByID(ctx context.Context, typeID string) (*model.DictType, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeID.Eq(typeID)).
		First()
}

// GetByCodeAndTenant 根据租户ID和字典编码获取字典类型（跨租户查询）
func (r *DictTypeRepo) GetByCodeAndTenant(ctx context.Context, typeCode, tenantID string) (*model.DictType, error) {
	return r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeCode.Eq(typeCode)).
		First()
}

// GetByCode 根据字典编码获取当前租户的字典类型
func (r *DictTypeRepo) GetByCode(ctx context.Context, typeCode string) (*model.DictType, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeCode.Eq(typeCode)).
		First()
}

// Update 更新字典类型
func (r *DictTypeRepo) Update(ctx context.Context, typeID string, updates map[string]interface{}) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeID.Eq(typeID)).
		Updates(updates)
	return err
}

// Delete 删除字典类型(软删除)
func (r *DictTypeRepo) Delete(ctx context.Context, typeID string) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeID.Eq(typeID)).
		Delete()
	return err
}

// BatchDelete 批量删除字典类型
func (r *DictTypeRepo) BatchDelete(ctx context.Context, typeIDs []string) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeID.In(typeIDs...)).
		Delete()
	return err
}

// GetByIDs 根据字典类型ID列表获取字典类型信息
func (r *DictTypeRepo) GetByIDs(ctx context.Context, typeIDs []string) ([]*model.DictType, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeID.In(typeIDs...)).
		Find()
}

// List 分页获取字典类型列表
func (r *DictTypeRepo) List(ctx context.Context, offset, limit int) ([]*model.DictType, int64, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		FindByPage(offset, limit)
}

// ListWithFilters 根据筛选条件分页获取字典类型列表
func (r *DictTypeRepo) ListWithFilters(ctx context.Context, offset, limit int, typeName, typeCode string) ([]*model.DictType, int64, error) {
	tenantID := xcontext.GetTenantID(ctx)
	query := r.q.DictType.WithContext(ctx).Where(r.q.DictType.TenantID.Eq(tenantID))

	if typeName != "" {
		query = query.Where(r.q.DictType.TypeName.Like("%" + typeName + "%"))
	}
	if typeCode != "" {
		query = query.Where(r.q.DictType.TypeCode.Like("%" + typeCode + "%"))
	}

	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	dictTypes, err := query.Order(r.q.DictType.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
	return dictTypes, total, err
}

// ListAll 获取所有字典类型（不分页）
func (r *DictTypeRepo) ListAll(ctx context.Context) ([]*model.DictType, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Find()
}

// CheckExists 检查字典类型是否存在
func (r *DictTypeRepo) CheckExists(ctx context.Context, tenantID, typeCode string) (bool, error) {
	count, err := r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeCode.Eq(typeCode)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckExistsByID 检查字典编码是否存在（排除指定ID）
func (r *DictTypeRepo) CheckExistsByID(ctx context.Context, tenantID, typeCode string, excludeTypeID string) (bool, error) {
	count, err := r.q.DictType.WithContext(ctx).
		Where(r.q.DictType.TenantID.Eq(tenantID)).
		Where(r.q.DictType.TypeCode.Eq(typeCode)).
		Where(r.q.DictType.TypeID.Neq(excludeTypeID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
