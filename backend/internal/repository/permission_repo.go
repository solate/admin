package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

// PermissionRepo 权限数据访问层，处理权限的基础 CRUD 操作
type PermissionRepo struct {
	db *gorm.DB
	q  *query.Query
}

// NewPermissionRepo 创建权限仓库
func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建权限
func (r *PermissionRepo) Create(ctx context.Context, permission *model.Permission) error {
	return r.q.Permission.WithContext(ctx).Create(permission)
}

// CreateBatch 批量创建权限
func (r *PermissionRepo) CreateBatch(ctx context.Context, permissions []*model.Permission) error {
	return r.q.Permission.WithContext(ctx).Create(permissions...)
}

// GetByID 根据ID获取权限
func (r *PermissionRepo) GetByID(ctx context.Context, permissionID string) (*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.Eq(permissionID)).
		First()
}

// GetByIDs 根据ID列表获取权限列表
func (r *PermissionRepo) GetByIDs(ctx context.Context, ids []string) ([]*model.Permission, error) {
	if len(ids) == 0 {
		return []*model.Permission{}, nil
	}
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.In(ids...)).
		Order(r.q.Permission.Sort.Asc()).
		Find()
}

// Update 更新权限
func (r *PermissionRepo) Update(ctx context.Context, permissionID string, updates map[string]interface{}) error {
	_, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.Eq(permissionID)).
		Updates(updates)
	return err
}

// Delete 删除权限（软删除）
func (r *PermissionRepo) Delete(ctx context.Context, permissionID string) error {
	_, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.Eq(permissionID)).
		Delete()
	return err
}

// DeleteBatch 批量删除权限（软删除）
func (r *PermissionRepo) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.In(ids...)).
		Delete()
	return err
}

// List 根据筛选条件获取权限列表
func (r *PermissionRepo) List(ctx context.Context, tenantID string, offset, limit int) ([]*model.Permission, int64, error) {
	query := r.q.Permission.WithContext(ctx).Where(r.q.Permission.TenantID.Eq(tenantID))

	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	permissions, err := query.
		Order(r.q.Permission.Sort.Asc()).
		Offset(offset).
		Limit(limit).
		Find()
	return permissions, total, err
}

// ListWithFilters 根据筛选条件分页获取权限列表
func (r *PermissionRepo) ListWithFilters(ctx context.Context, tenantID string, offset, limit int, filters map[string]interface{}) ([]*model.Permission, int64, error) {
	query := r.q.Permission.WithContext(ctx).Where(r.q.Permission.TenantID.Eq(tenantID))

	// 应用筛选条件
	if nameFilter, ok := filters["name"].(string); ok && nameFilter != "" {
		query = query.Where(r.q.Permission.Name.Like("%" + nameFilter + "%"))
	}
	if typeFilter, ok := filters["type"].(string); ok && typeFilter != "" {
		query = query.Where(r.q.Permission.Type.Eq(typeFilter))
	}
	if statusFilter, ok := filters["status"].(int16); ok {
		query = query.Where(r.q.Permission.Status.Eq(statusFilter))
	}
	if sourceTypeFilter, ok := filters["source_type"].(string); ok && sourceTypeFilter != "" {
		query = query.Where(r.q.Permission.SourceType.Eq(sourceTypeFilter))
	}
	if parentIDFilter, ok := filters["parent_id"].(string); ok && parentIDFilter != "" {
		query = query.Where(r.q.Permission.ParentID.Eq(parentIDFilter))
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	permissions, err := query.
		Order(r.q.Permission.Sort.Asc()).
		Offset(offset).
		Limit(limit).
		Find()
	return permissions, total, err
}

// CheckExistsByName 检查权限名称是否存在（租户内）
func (r *PermissionRepo) CheckExistsByName(ctx context.Context, tenantID, name string) (bool, error) {
	count, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.TenantID.Eq(tenantID)).
		Where(r.q.Permission.Name.Eq(name)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetChildrenCount 获取子权限数量
func (r *PermissionRepo) GetChildrenCount(ctx context.Context, parentID string) (int64, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.ParentID.Eq(parentID)).
		Count()
}
