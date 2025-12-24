package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

type RoleRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewRoleRepo(db *gorm.DB) *RoleRepo {
	return &RoleRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建角色
func (r *RoleRepo) Create(ctx context.Context, role *model.Role) error {
	return r.q.Role.WithContext(ctx).Create(role)
}

// GetByID 根据ID获取角色
func (r *RoleRepo) GetByID(ctx context.Context, roleID string) (*model.Role, error) {
	return r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).First()
}

// GetByRoleCode 根据角色编码获取角色
func (r *RoleRepo) GetByRoleCode(ctx context.Context, tenantID, roleCode string) (*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleCode.Eq(roleCode)).
		First()
}

// Update 更新角色
func (r *RoleRepo) Update(ctx context.Context, roleID string, updates map[string]interface{}) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).Updates(updates)
	return err
}

// Delete 删除角色(软删除)
func (r *RoleRepo) Delete(ctx context.Context, roleID string) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).Delete()
	return err
}

// List 分页获取角色列表
func (r *RoleRepo) List(ctx context.Context, offset, limit int) ([]*model.Role, int64, error) {
	return r.q.Role.WithContext(ctx).FindByPage(offset, limit)
}

// UpdateStatus 更新角色状态
func (r *RoleRepo) UpdateStatus(ctx context.Context, roleID string, status int32) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).Update(r.q.Role.Status, status)
	return err
}

// CheckExists 检查角色是否存在
func (r *RoleRepo) CheckExists(ctx context.Context, tenantID, roleCode string) (bool, error) {
	count, err := r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleCode.Eq(roleCode)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByIDs 根据角色ID列表获取角色列表
func (r *RoleRepo) ListByIDs(ctx context.Context, roleIDs []string) ([]*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.RoleID.In(roleIDs...)).
		Find()
}
