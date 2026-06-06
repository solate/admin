package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

// PermissionRepo 权限点数据访问层
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

// Create 创建权限点
func (r *PermissionRepo) Create(ctx context.Context, permission *model.Permission) error {
	return r.q.Permission.WithContext(ctx).Create(permission)
}

// CreateBatch 批量创建权限点
func (r *PermissionRepo) CreateBatch(ctx context.Context, permissions []*model.Permission) error {
	if len(permissions) == 0 {
		return nil
	}
	return r.q.Permission.WithContext(ctx).Create(permissions...)
}

// GetByID 根据ID获取权限点
func (r *PermissionRepo) GetByID(ctx context.Context, permissionID string) (*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.Eq(permissionID)).
		First()
}

// GetByIDs 根据ID列表获取权限点列表
func (r *PermissionRepo) GetByIDs(ctx context.Context, ids []string) ([]*model.Permission, error) {
	if len(ids) == 0 {
		return []*model.Permission{}, nil
	}
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.In(ids...)).
		Find()
}

// GetByResource 根据资源标识获取权限点
func (r *PermissionRepo) GetByResource(ctx context.Context, resource string) (*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.Resource.Eq(resource)).
		First()
}

// Update 更新权限点
func (r *PermissionRepo) Update(ctx context.Context, permissionID string, updates map[string]interface{}) error {
	_, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.Eq(permissionID)).
		Updates(updates)
	return err
}

// Delete 删除权限点（软删除）
func (r *PermissionRepo) Delete(ctx context.Context, permissionID string) error {
	_, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.Eq(permissionID)).
		Delete()
	return err
}

// DeleteBatch 批量删除权限点（软删除）
func (r *PermissionRepo) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.In(ids...)).
		Delete()
	return err
}

// List 获取所有权限点列表
func (r *PermissionRepo) List(ctx context.Context) ([]*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).Find()
}

// ListByType 根据类型获取权限点列表
func (r *PermissionRepo) ListByType(ctx context.Context, permissionType string) ([]*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.Type.Eq(permissionType)).
		Find()
}

// ListWithFilters 根据筛选条件分页获取权限点列表
func (r *PermissionRepo) ListWithFilters(ctx context.Context, offset, limit int, nameFilter, typeFilter string) ([]*model.Permission, int64, error) {
	query := r.q.Permission.WithContext(ctx)

	// 应用筛选条件
	if nameFilter != "" {
		query = query.Where(r.q.Permission.Name.Like("%" + nameFilter + "%"))
	}
	if typeFilter != "" {
		query = query.Where(r.q.Permission.Type.Eq(typeFilter))
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	permissions, err := query.
		Order(r.q.Permission.PermissionID.Desc()).
		Offset(offset).
		Limit(limit).
		Find()
	return permissions, total, err
}

// CheckExistsByName 检查权限名称是否存在
func (r *PermissionRepo) CheckExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.Name.Eq(name)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckExistsByResource 检查资源标识是否存在
func (r *PermissionRepo) CheckExistsByResource(ctx context.Context, resource string) (bool, error) {
	count, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.Resource.Eq(resource)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Exists 检查权限点是否存在
func (r *PermissionRepo) Exists(ctx context.Context, permissionID string) (bool, error) {
	count, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.Eq(permissionID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
