package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

type UserTenantRoleRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewUserTenantRoleRepo(db *gorm.DB) *UserTenantRoleRepo {
	return &UserTenantRoleRepo{
		db: db,
		q:  query.Use(db),
	}
}

// CreateUserTenantRole 创建用户租户角色关系
func (r *UserTenantRoleRepo) CreateUserTenantRole(ctx context.Context, userTenantRole *model.UserTenantRole) error {
	return r.q.UserTenantRole.WithContext(ctx).Create(userTenantRole)
}

// CreateUserTenantRoles 批量创建用户租户角色关系
func (r *UserTenantRoleRepo) CreateUserTenantRoles(ctx context.Context, userTenantRoles []*model.UserTenantRole) error {
	return r.q.UserTenantRole.WithContext(ctx).CreateInBatches(userTenantRoles, 100)
}

// GetUserTenantRoleByID 根据ID获取用户租户角色关系
func (r *UserTenantRoleRepo) GetUserTenantRoleByID(ctx context.Context, userTenantRoleID string) (*model.UserTenantRole, error) {
	return r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.UserTenantRoleID.Eq(userTenantRoleID)).First()
}

// GetUserTenantRoles 根据用户ID获取租户角色关系
func (r *UserTenantRoleRepo) GetUserTenantRoles(ctx context.Context, userID string) ([]*model.UserTenantRole, error) {
	return r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.UserID.Eq(userID)).Find()
}

// GetTenantUserRoles 根据租户ID获取用户角色关系
func (r *UserTenantRoleRepo) GetTenantUserRoles(ctx context.Context, tenantID string) ([]*model.UserTenantRole, error) {
	return r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).Find()
}

// GetRoleUsers 根据角色ID获取用户关系
func (r *UserTenantRoleRepo) GetRoleUsers(ctx context.Context, roleID string) ([]*model.UserTenantRole, error) {
	return r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.RoleID.Eq(roleID)).Find()
}

// GetUserTenantRoleByUserTenant 根据用户ID和租户ID获取角色关系
func (r *UserTenantRoleRepo) GetUserTenantRoleByUserTenant(ctx context.Context, userID, tenantID string) ([]*model.UserTenantRole, error) {
	return r.q.UserTenantRole.WithContext(ctx).
		Where(r.q.UserTenantRole.UserID.Eq(userID)).
		Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).
		Find()
}

// GetUserTenantRoleByUserTenantRole 根据用户ID、租户ID和角色ID获取关系
func (r *UserTenantRoleRepo) GetUserTenantRoleByUserTenantRole(ctx context.Context, userID, tenantID, roleID string) (*model.UserTenantRole, error) {
	return r.q.UserTenantRole.WithContext(ctx).
		Where(r.q.UserTenantRole.UserID.Eq(userID)).
		Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).
		Where(r.q.UserTenantRole.RoleID.Eq(roleID)).
		First()
}

// UpdateUserTenantRole 更新用户租户角色关系
func (r *UserTenantRoleRepo) UpdateUserTenantRole(ctx context.Context, userTenantRoleID string, updates map[string]interface{}) error {
	_, err := r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.UserTenantRoleID.Eq(userTenantRoleID)).Updates(updates)
	return err
}

// DeleteUserTenantRole 删除用户租户角色关系
func (r *UserTenantRoleRepo) DeleteUserTenantRole(ctx context.Context, userTenantRoleID string) error {
	_, err := r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.UserTenantRoleID.Eq(userTenantRoleID)).Delete()
	return err
}

// DeleteUserTenantRolesByUserID 根据用户ID删除所有租户角色关系
func (r *UserTenantRoleRepo) DeleteUserTenantRolesByUserID(ctx context.Context, userID string) error {
	_, err := r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.UserID.Eq(userID)).Delete()
	return err
}

// DeleteUserTenantRolesByTenantID 根据租户ID删除所有用户角色关系
func (r *UserTenantRoleRepo) DeleteUserTenantRolesByTenantID(ctx context.Context, tenantID string) error {
	_, err := r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).Delete()
	return err
}

// DeleteUserTenantRoleByUserTenant 删除指定用户和租户的角色关系
func (r *UserTenantRoleRepo) DeleteUserTenantRoleByUserTenant(ctx context.Context, userID, tenantID string) error {
	_, err := r.q.UserTenantRole.WithContext(ctx).
		Where(r.q.UserTenantRole.UserID.Eq(userID)).
		Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).
		Delete()
	return err
}

// DeleteUserTenantRoleByUserTenantRole 删除指定的用户租户角色关系
func (r *UserTenantRoleRepo) DeleteUserTenantRoleByUserTenantRole(ctx context.Context, userID, tenantID, roleID string) error {
	_, err := r.q.UserTenantRole.WithContext(ctx).
		Where(r.q.UserTenantRole.UserID.Eq(userID)).
		Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).
		Where(r.q.UserTenantRole.RoleID.Eq(roleID)).
		Delete()
	return err
}

// ListUserTenantRoles 分页获取用户租户角色关系列表
func (r *UserTenantRoleRepo) ListUserTenantRoles(ctx context.Context, offset, limit int) ([]*model.UserTenantRole, int64, error) {
	return r.q.UserTenantRole.WithContext(ctx).FindByPage(offset, limit)
}

// CountUserTenantRolesByUser 统计用户的租户角色关系数量
func (r *UserTenantRoleRepo) CountUserTenantRolesByUser(ctx context.Context, userID string) (int64, error) {
	return r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.UserID.Eq(userID)).Count()
}

// CountUserTenantRolesByTenant 统计租户的用户角色关系数量
func (r *UserTenantRoleRepo) CountUserTenantRolesByTenant(ctx context.Context, tenantID string) (int64, error) {
	return r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).Count()
}

// CountUserTenantRolesByRole 统计角色的用户关系数量
func (r *UserTenantRoleRepo) CountUserTenantRolesByRole(ctx context.Context, roleID string) (int64, error) {
	return r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.RoleID.Eq(roleID)).Count()
}

// CheckUserTenantRoleExists 检查用户租户角色关系是否存在
func (r *UserTenantRoleRepo) CheckUserTenantRoleExists(ctx context.Context, userID, tenantID, roleID string) (bool, error) {
	count, err := r.q.UserTenantRole.WithContext(ctx).
		Where(r.q.UserTenantRole.UserID.Eq(userID)).
		Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).
		Where(r.q.UserTenantRole.RoleID.Eq(roleID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckUserHasTenantRole 检查用户是否有指定租户的角色
func (r *UserTenantRoleRepo) CheckUserHasTenantRole(ctx context.Context, userID, tenantID string) (bool, error) {
	count, err := r.q.UserTenantRole.WithContext(ctx).
		Where(r.q.UserTenantRole.UserID.Eq(userID)).
		Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserTenants 获取用户所属的所有租户ID列表
func (r *UserTenantRoleRepo) GetUserTenants(ctx context.Context, userID string) ([]string, error) {
	roles, err := r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.UserID.Eq(userID)).Find()
	if err != nil {
		return nil, err
	}

	tenantIDs := make([]string, 0, len(roles))
	seen := make(map[string]bool)
	for _, role := range roles {
		if !seen[role.TenantID] {
			tenantIDs = append(tenantIDs, role.TenantID)
			seen[role.TenantID] = true
		}
	}

	return tenantIDs, nil
}

// GetTenantUsers 获取租户下的所有用户ID列表
func (r *UserTenantRoleRepo) GetTenantUsers(ctx context.Context, tenantID string) ([]string, error) {
	roles, err := r.q.UserTenantRole.WithContext(ctx).Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).Find()
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(roles))
	seen := make(map[string]bool)
	for _, role := range roles {
		if !seen[role.UserID] {
			userIDs = append(userIDs, role.UserID)
			seen[role.UserID] = true
		}
	}

	return userIDs, nil
}

// GetUserRolesInTenant 获取用户在指定租户下的角色ID列表
func (r *UserTenantRoleRepo) GetUserRolesInTenant(ctx context.Context, userID, tenantID string) ([]string, error) {
	roles, err := r.q.UserTenantRole.WithContext(ctx).
		Where(r.q.UserTenantRole.UserID.Eq(userID)).
		Where(r.q.UserTenantRole.TenantID.Eq(tenantID)).
		Find()
	if err != nil {
		return nil, err
	}

	roleIDs := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.RoleID)
	}

	return roleIDs, nil
}
