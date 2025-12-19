package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/constants"
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

// CreateRole 创建角色
func (r *RoleRepo) CreateRole(ctx context.Context, role *model.Role) error {
	return r.q.Role.WithContext(ctx).Create(role)
}

// GetRoleByID 根据ID获取角色
func (r *RoleRepo) GetRoleByID(ctx context.Context, roleID string) (*model.Role, error) {
	return r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).First()
}

// GetRoleByCode 根据角色编码获取角色
func (r *RoleRepo) GetRoleByCode(ctx context.Context, tenantID, roleCode string) (*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleCode.Eq(roleCode)).
		First()
}

// UpdateRole 更新角色
func (r *RoleRepo) UpdateRole(ctx context.Context, roleID string, updates map[string]interface{}) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).Updates(updates)
	return err
}

// DeleteRole 删除角色(软删除)
func (r *RoleRepo) DeleteRole(ctx context.Context, roleID string) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).Delete()
	return err
}

// ListRoles 分页获取角色列表
func (r *RoleRepo) ListRoles(ctx context.Context, offset, limit int) ([]*model.Role, int64, error) {
	return r.q.Role.WithContext(ctx).FindByPage(offset, limit)
}

// // ListRolesByTenant 根据租户ID获取角色列表
// func (r *RoleRepo) ListRolesByTenant(ctx context.Context, tenantID string, offset, limit int) ([]*model.Role, int64, error) {
// 	return r.q.Role.WithContext(ctx).
// 		Where(r.q.Role.TenantID.Eq(tenantID)).
// 		FindByPage(offset, limit)
// }

// UpdateRoleStatus 更新角色状态
func (r *RoleRepo) UpdateRoleStatus(ctx context.Context, roleID string, status int32) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).Update(r.q.Role.Status, status)
	return err
}

// CheckRoleExists 检查角色是否存在
func (r *RoleRepo) CheckRoleExists(ctx context.Context, tenantID, roleCode string) (bool, error) {
	count, err := r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleCode.Eq(roleCode)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetActiveRolesByTenant 获取租户下的活跃角色
func (r *RoleRepo) GetActiveRolesByTenant(ctx context.Context, tenantID string) ([]*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.Status.Eq(constants.StatusEnabled)).
		Find()
}

// ListRolesByIDs 根据角色ID列表获取角色列表
func (r *RoleRepo) ListRolesByIDs(ctx context.Context, roleIDs []string) ([]*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.RoleID.In(roleIDs...)).
		Find()
}
