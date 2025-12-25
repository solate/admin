package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/constants"
	"context"

	"gorm.io/gorm"
)

// MenuRepo 菜单数据访问层，处理菜单相关的业务逻辑
type MenuRepo struct {
	db             *gorm.DB
	q              *query.Query
	permissionRepo *PermissionRepo
}

func NewMenuRepo(db *gorm.DB) *MenuRepo {
	return &MenuRepo{
		db:             db,
		q:              query.Use(db),
		permissionRepo: NewPermissionRepo(db),
	}
}

// Create 创建菜单
func (r *MenuRepo) Create(ctx context.Context, menu *model.Permission) error {
	return r.permissionRepo.Create(ctx, menu)
}

// GetByID 根据ID获取菜单
func (r *MenuRepo) GetByID(ctx context.Context, menuID string) (*model.Permission, error) {
	return r.permissionRepo.GetByID(ctx, menuID)
}

// Update 更新菜单
func (r *MenuRepo) Update(ctx context.Context, menuID string, updates map[string]interface{}) error {
	return r.permissionRepo.Update(ctx, menuID, updates)
}

// Delete 删除菜单(软删除)
func (r *MenuRepo) Delete(ctx context.Context, menuID string) error {
	return r.permissionRepo.Delete(ctx, menuID)
}

// UpdateStatus 更新菜单状态
func (r *MenuRepo) UpdateStatus(ctx context.Context, menuID string, status int16) error {
	return r.permissionRepo.Update(ctx, menuID, map[string]interface{}{
		"status": status,
	})
}

// ListByTenant 根据租户ID获取菜单列表（包含菜单和按钮）
func (r *MenuRepo) ListByTenant(ctx context.Context, tenantID string) ([]*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.TenantID.Eq(tenantID)).
		Where(r.q.Permission.Type.In(constants.TypeMenu, constants.TypeButton)).
		Order(r.q.Permission.Sort.Asc()).
		Find()
}

// ListByTenantAndType 根据租户ID和类型获取菜单列表
func (r *MenuRepo) ListByTenantAndType(ctx context.Context, tenantID, permissionType string) ([]*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.TenantID.Eq(tenantID)).
		Where(r.q.Permission.Type.Eq(permissionType)).
		Order(r.q.Permission.Sort.Asc()).
		Find()
}

// ListByParentID 根据父菜单ID获取子菜单列表
func (r *MenuRepo) ListByParentID(ctx context.Context, tenantID, parentID string) ([]*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.TenantID.Eq(tenantID)).
		Where(r.q.Permission.ParentID.Eq(parentID)).
		Order(r.q.Permission.Sort.Asc()).
		Find()
}

// ListRootMenus 获取顶级菜单（父菜单ID为空）
func (r *MenuRepo) ListRootMenus(ctx context.Context, tenantID string) ([]*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.TenantID.Eq(tenantID)).
		Where(r.q.Permission.Type.Eq(constants.TypeMenu)).
		Where(r.q.Permission.ParentID.IsNull()).
		Order(r.q.Permission.Sort.Asc()).
		Find()
}

// GetChildrenCount 获取子菜单数量
func (r *MenuRepo) GetChildrenCount(ctx context.Context, parentID string) (int64, error) {
	return r.permissionRepo.GetChildrenCount(ctx, parentID)
}

// CheckExistsByName 检查菜单名称是否存在（租户内）
func (r *MenuRepo) CheckExistsByName(ctx context.Context, tenantID, name string) (bool, error) {
	return r.permissionRepo.CheckExistsByName(ctx, tenantID, name)
}

// ListWithFilters 根据筛选条件分页获取菜单列表
func (r *MenuRepo) ListWithFilters(ctx context.Context, tenantID string, offset, limit int, nameFilter, typeFilter string, statusFilter *int16) ([]*model.Permission, int64, error) {
	filters := make(map[string]interface{})
	if nameFilter != "" {
		filters["name"] = nameFilter
	}
	if typeFilter != "" {
		filters["type"] = typeFilter
	}
	if statusFilter != nil {
		filters["status"] = *statusFilter
	}
	return r.permissionRepo.ListWithFilters(ctx, tenantID, offset, limit, filters)
}

// GetByIDs 根据ID列表获取菜单列表
func (r *MenuRepo) GetByIDs(ctx context.Context, ids []string) ([]*model.Permission, error) {
	return r.permissionRepo.GetByIDs(ctx, ids)
}
