package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/constants"
	"context"
	"gorm.io/gorm"
)

// UserMenuRepo 用户菜单数据访问层
type UserMenuRepo struct {
	db      *gorm.DB
	q       *query.Query
	menuRepo *MenuRepo
}

// NewUserMenuRepo 创建用户菜单仓库
func NewUserMenuRepo(db *gorm.DB) *UserMenuRepo {
	return &UserMenuRepo{
		db:      db,
		q:       query.Use(db),
		menuRepo: NewMenuRepo(db),
	}
}

// GetByIDs 根据ID列表获取菜单
func (r *UserMenuRepo) GetByIDs(ctx context.Context, ids []string) ([]*model.Permission, error) {
	return r.menuRepo.GetByIDs(ctx, ids)
}

// GetButtonsByMenuID 获取菜单下的按钮权限
func (r *UserMenuRepo) GetButtonsByMenuID(ctx context.Context, tenantID, menuID string) ([]*model.Permission, error) {
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.TenantID.Eq(tenantID)).
		Where(r.q.Permission.Type.Eq(constants.TypeButton)).
		Where(r.q.Permission.ParentID.Eq(menuID)).
		Order(r.q.Permission.Sort.Asc()).
		Find()
}

// GetTenantAvailableMenus 获取租户可用的所有菜单（包括系统菜单和租户自定义菜单）
func (r *UserMenuRepo) GetTenantAvailableMenus(ctx context.Context, tenantID string) ([]*model.Permission, error) {
	// 获取默认租户的系统菜单
	systemMenus, err := r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.TenantID.Eq(constants.DefaultTenant)).
		Where(r.q.Permission.Type.Eq(constants.TypeMenu)).
		Order(r.q.Permission.Sort.Asc()).
		Find()
	if err != nil {
		return nil, err
	}

	// 如果不是默认租户，再获取租户自定义菜单
	if tenantID != constants.DefaultTenant {
		customMenus, err := r.q.Permission.WithContext(ctx).
			Where(r.q.Permission.TenantID.Eq(tenantID)).
			Where(r.q.Permission.Type.Eq(constants.TypeMenu)).
			Where(r.q.Permission.SourceType.Eq(constants.SourceTypeCustom)).
			Order(r.q.Permission.Sort.Asc()).
			Find()
		if err != nil {
			return nil, err
		}
		systemMenus = append(systemMenus, customMenus...)
	}

	return systemMenus, nil
}

// GetByPermissionIDs 根据权限ID列表获取权限详情（用于获取用户的所有权限）
func (r *UserMenuRepo) GetByPermissionIDs(ctx context.Context, permissionIDs []string) ([]*model.Permission, error) {
	if len(permissionIDs) == 0 {
		return []*model.Permission{}, nil
	}
	return r.q.Permission.WithContext(ctx).
		Where(r.q.Permission.PermissionID.In(permissionIDs...)).
		Order(r.q.Permission.Sort.Asc()).
		Find()
}
