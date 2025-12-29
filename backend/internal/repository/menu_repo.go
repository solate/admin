package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

// MenuRepo 菜单数据访问层
type MenuRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewMenuRepo(db *gorm.DB) *MenuRepo {
	return &MenuRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建菜单
func (r *MenuRepo) Create(ctx context.Context, menu *model.Menu) error {
	return r.q.Menu.WithContext(ctx).Create(menu)
}

// CreateBatch 批量创建菜单
func (r *MenuRepo) CreateBatch(ctx context.Context, menus []*model.Menu) error {
	if len(menus) == 0 {
		return nil
	}
	return r.q.Menu.WithContext(ctx).Create(menus...)
}

// GetByID 根据ID获取菜单
func (r *MenuRepo) GetByID(ctx context.Context, menuID string) (*model.Menu, error) {
	return r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.MenuID.Eq(menuID)).
		First()
}

// GetByIDs 根据ID列表获取菜单列表
func (r *MenuRepo) GetByIDs(ctx context.Context, ids []string) ([]*model.Menu, error) {
	if len(ids) == 0 {
		return []*model.Menu{}, nil
	}
	return r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.MenuID.In(ids...)).
		Order(r.q.Menu.Sort.Asc()).
		Find()
}

// Update 更新菜单
func (r *MenuRepo) Update(ctx context.Context, menuID string, updates map[string]interface{}) error {
	_, err := r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.MenuID.Eq(menuID)).
		Updates(updates)
	return err
}

// Delete 删除菜单(软删除)
func (r *MenuRepo) Delete(ctx context.Context, menuID string) error {
	_, err := r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.MenuID.Eq(menuID)).
		Delete()
	return err
}

// DeleteBatch 批量删除菜单(软删除)
func (r *MenuRepo) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.MenuID.In(ids...)).
		Delete()
	return err
}

// List 获取所有菜单列表
func (r *MenuRepo) List(ctx context.Context) ([]*model.Menu, error) {
	return r.q.Menu.WithContext(ctx).
		Order(r.q.Menu.Sort.Asc()).
		Find()
}

// ListWithFilters 根据筛选条件获取菜单列表
func (r *MenuRepo) ListWithFilters(ctx context.Context, offset, limit int, nameFilter string, statusFilter *int16) ([]*model.Menu, int64, error) {
	query := r.q.Menu.WithContext(ctx)

	// 应用筛选条件
	if nameFilter != "" {
		query = query.Where(r.q.Menu.Name.Like("%" + nameFilter + "%"))
	}
	if statusFilter != nil {
		query = query.Where(r.q.Menu.Status.Eq(*statusFilter))
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	menus, err := query.
		Order(r.q.Menu.Sort.Asc()).
		Offset(offset).
		Limit(limit).
		Find()
	return menus, total, err
}

// ListByParentID 根据父菜单ID获取子菜单列表
func (r *MenuRepo) ListByParentID(ctx context.Context, parentID string) ([]*model.Menu, error) {
	return r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.ParentID.Eq(parentID)).
		Order(r.q.Menu.Sort.Asc()).
		Find()
}

// ListRootMenus 获取顶级菜单（parent_id 为空字符串）
func (r *MenuRepo) ListRootMenus(ctx context.Context) ([]*model.Menu, error) {
	return r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.ParentID.Eq("")).
		Order(r.q.Menu.Sort.Asc()).
		Find()
}

// GetChildrenCount 获取子菜单数量
func (r *MenuRepo) GetChildrenCount(ctx context.Context, parentID string) (int64, error) {
	return r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.ParentID.Eq(parentID)).
		Count()
}

// UpdateStatus 更新菜单状态
func (r *MenuRepo) UpdateStatus(ctx context.Context, menuID string, status int16) error {
	return r.Update(ctx, menuID, map[string]interface{}{
		"status": status,
	})
}

// CheckExistsByName 检查菜单名称是否存在
func (r *MenuRepo) CheckExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.Name.Eq(name)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Exists 检查菜单是否存在
func (r *MenuRepo) Exists(ctx context.Context, menuID string) (bool, error) {
	count, err := r.q.Menu.WithContext(ctx).
		Where(r.q.Menu.MenuID.Eq(menuID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
