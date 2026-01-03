package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/database"
	"sort"
	"context"

	"gorm.io/gorm"
)

// DictItemRepo 字典项仓库
type DictItemRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewDictItemRepo(db *gorm.DB) *DictItemRepo {
	return &DictItemRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建字典项
func (r *DictItemRepo) Create(ctx context.Context, dictItem *model.DictItem) error {
	return r.q.DictItem.WithContext(ctx).Create(dictItem)
}

// CreateBatch 批量创建字典项
func (r *DictItemRepo) CreateBatch(ctx context.Context, items []*model.DictItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.q.DictItem.WithContext(ctx).Create(items...)
}

// GetByID 根据ID获取字典项
func (r *DictItemRepo) GetByID(ctx context.Context, itemID string) (*model.DictItem, error) {
	return r.q.DictItem.WithContext(ctx).Where(r.q.DictItem.ItemID.Eq(itemID)).First()
}

// GetByTypeAndTenant 获取指定类型和租户的字典项
func (r *DictItemRepo) GetByTypeAndTenant(ctx context.Context, typeID, tenantID string) ([]*model.DictItem, error) {
	ctx = database.ManualTenantMode(ctx)
	return r.q.DictItem.WithContext(ctx).
		Where(r.q.DictItem.TypeID.Eq(typeID)).
		Where(r.q.DictItem.TenantID.Eq(tenantID)).
		Order(r.q.DictItem.Sort).
		Find()
}

// GetByTypeAndValue 获取指定类型、租户、值的字典项（用于检查是否已覆盖）
func (r *DictItemRepo) GetByTypeAndValue(ctx context.Context, typeID, tenantID, value string) (*model.DictItem, error) {
	ctx = database.ManualTenantMode(ctx)
	return r.q.DictItem.WithContext(ctx).
		Where(r.q.DictItem.TypeID.Eq(typeID)).
		Where(r.q.DictItem.TenantID.Eq(tenantID)).
		Where(r.q.DictItem.Value.Eq(value)).
		First()
}

// DeleteByTypeAndValue 删除指定类型、租户、值的字典项（恢复系统默认）
func (r *DictItemRepo) DeleteByTypeAndValue(ctx context.Context, typeID, tenantID, value string) error {
	ctx = database.ManualTenantMode(ctx)
	_, err := r.q.DictItem.WithContext(ctx).
		Where(r.q.DictItem.TypeID.Eq(typeID)).
		Where(r.q.DictItem.TenantID.Eq(tenantID)).
		Where(r.q.DictItem.Value.Eq(value)).
		Delete()
	return err
}

// DeleteByTypeID 删除指定类型的所有字典项
func (r *DictItemRepo) DeleteByTypeID(ctx context.Context, typeID string) error {
	_, err := r.q.DictItem.WithContext(ctx).
		Where(r.q.DictItem.TypeID.Eq(typeID)).
		Delete()
	return err
}

// Update 更新字典项
func (r *DictItemRepo) Update(ctx context.Context, itemID string, updates map[string]interface{}) error {
	_, err := r.q.DictItem.WithContext(ctx).Where(r.q.DictItem.ItemID.Eq(itemID)).Updates(updates)
	return err
}

// Delete 删除字典项(软删除)
func (r *DictItemRepo) Delete(ctx context.Context, itemID string) error {
	_, err := r.q.DictItem.WithContext(ctx).Where(r.q.DictItem.ItemID.Eq(itemID)).Delete()
	return err
}

// List 分页获取字典项列表
func (r *DictItemRepo) List(ctx context.Context, offset, limit int) ([]*model.DictItem, int64, error) {
	return r.q.DictItem.WithContext(ctx).FindByPage(offset, limit)
}

// ListByTypeID 根据类型ID获取字典项列表
func (r *DictItemRepo) ListByTypeID(ctx context.Context, typeID string) ([]*model.DictItem, error) {
	return r.q.DictItem.WithContext(ctx).
		Where(r.q.DictItem.TypeID.Eq(typeID)).
		Order(r.q.DictItem.Sort).
		Find()
}

// GetMergedByTypeCode 根据字典编码获取合并后的字典项（系统+租户覆盖）
// 这是字典系统的核心方法，实现了租户覆盖系统默认值的逻辑
func (r *DictItemRepo) GetMergedByTypeCode(ctx context.Context, typeCode string, defaultTenantID, currentTenantID string) ([]*model.DictItem, error) {
	ctx = database.ManualTenantMode(ctx)

	// 1. 获取字典类型
	dictType, err := query.Use(r.db).DictType.WithContext(ctx).
		Where(query.Use(r.db).DictType.TypeCode.Eq(typeCode)).
		Where(query.Use(r.db).DictType.TenantID.Eq(defaultTenantID)).
		First()
	if err != nil {
		return nil, err
	}

	// 2. 获取系统字典项
	systemItems, err := r.GetByTypeAndTenant(ctx, dictType.TypeID, defaultTenantID)
	if err != nil {
		return nil, err
	}

	// 3. 获取租户覆盖的字典项
	overrideItems, err := r.GetByTypeAndTenant(ctx, dictType.TypeID, currentTenantID)
	if err != nil {
		return nil, err
	}

	// 4. 构建覆盖映射
	overrideMap := make(map[string]*model.DictItem) // value -> DictItem
	for _, item := range overrideItems {
		overrideMap[item.Value] = item
	}

	// 5. 合并（租户覆盖的优先）
	mergedItems := make([]*model.DictItem, 0)
	for _, item := range systemItems {
		if override, exists := overrideMap[item.Value]; exists {
			// 使用租户覆盖的值
			mergedItems = append(mergedItems, override)
		} else {
			// 使用系统默认值
			mergedItems = append(mergedItems, item)
		}
	}

	// 6. 添加租户独有的字典项（系统没有的）
	valueSet := make(map[string]bool)
	for _, item := range systemItems {
		valueSet[item.Value] = true
	}
	for _, item := range overrideItems {
		if !valueSet[item.Value] {
			mergedItems = append(mergedItems, item)
		}
	}

	// 7. 按 sort 排序
	sort.Slice(mergedItems, func(i, j int) bool {
		return mergedItems[i].Sort < mergedItems[j].Sort
	})

	return mergedItems, nil
}

// GetDictTypeWithItems 获取字典类型及其合并后的字典项
func (r *DictItemRepo) GetDictTypeWithItems(ctx context.Context, typeCode string, defaultTenantID, currentTenantID string) (*model.DictType, []*model.DictItem, error) {
	ctx = database.ManualTenantMode(ctx)

	// 获取字典类型
	dictType, err := query.Use(r.db).DictType.WithContext(ctx).
		Where(query.Use(r.db).DictType.TypeCode.Eq(typeCode)).
		Where(query.Use(r.db).DictType.TenantID.Eq(defaultTenantID)).
		First()
	if err != nil {
		return nil, nil, err
	}

	// 获取合并后的字典项
	items, err := r.GetMergedByTypeCode(ctx, typeCode, defaultTenantID, currentTenantID)
	if err != nil {
		return nil, nil, err
	}

	return dictType, items, nil
}
