package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

// UserRoleRepo 用户角色仓储（基于 user_roles 表）
type UserRoleRepo struct {
	db *gorm.DB
	q  *query.Query
}

// NewUserRoleRepo 创建用户角色仓储
func NewUserRoleRepo(db *gorm.DB) *UserRoleRepo {
	return &UserRoleRepo{
		db: db,
		q:  query.Use(db),
	}
}

// GetUserRoleIDs 获取用户在指定租户下的角色ID列表
func (r *UserRoleRepo) GetUserRoleIDs(ctx context.Context, userID, tenantID string) ([]string, error) {
	userRoles, err := r.q.UserRole.WithContext(ctx).
		Where(r.q.UserRole.UserID.Eq(userID)).
		Where(r.q.UserRole.TenantID.Eq(tenantID)).
		Find()
	if err != nil {
		return nil, err
	}

	roleIDs := make([]string, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}
	return roleIDs, nil
}

// AddUserRole 为用户添加角色
func (r *UserRoleRepo) AddUserRole(ctx context.Context, userID, roleID, tenantID string) error {
	userRole := &model.UserRole{
		UserID:   userID,
		RoleID:   roleID,
		TenantID: tenantID,
	}
	return r.q.UserRole.WithContext(ctx).Create(userRole)
}

// DeleteUserRole 删除用户的指定角色
func (r *UserRoleRepo) DeleteUserRole(ctx context.Context, userID, roleID, tenantID string) error {
	_, err := r.q.UserRole.WithContext(ctx).
		Where(r.q.UserRole.UserID.Eq(userID)).
		Where(r.q.UserRole.RoleID.Eq(roleID)).
		Where(r.q.UserRole.TenantID.Eq(tenantID)).
		Delete()
	return err
}

// DeleteUserRoles 删除用户在指定租户下的所有角色
func (r *UserRoleRepo) DeleteUserRoles(ctx context.Context, userID, tenantID string) error {
	_, err := r.q.UserRole.WithContext(ctx).
		Where(r.q.UserRole.UserID.Eq(userID)).
		Where(r.q.UserRole.TenantID.Eq(tenantID)).
		Delete()
	return err
}

// CheckUserRole 检查用户是否拥有指定角色
func (r *UserRoleRepo) CheckUserRole(ctx context.Context, userID, roleID, tenantID string) (bool, error) {
	count, err := r.q.UserRole.WithContext(ctx).
		Where(r.q.UserRole.UserID.Eq(userID)).
		Where(r.q.UserRole.RoleID.Eq(roleID)).
		Where(r.q.UserRole.TenantID.Eq(tenantID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRoleUsers 获取指定角色下的所有用户ID
func (r *UserRoleRepo) GetRoleUsers(ctx context.Context, roleID, tenantID string) ([]string, error) {
	userRoles, err := r.q.UserRole.WithContext(ctx).
		Where(r.q.UserRole.RoleID.Eq(roleID)).
		Where(r.q.UserRole.TenantID.Eq(tenantID)).
		Find()
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, len(userRoles))
	for i, ur := range userRoles {
		userIDs[i] = ur.UserID
	}
	return userIDs, nil
}

// AssignRoles 为用户批量分配角色（覆盖式）
func (r *UserRoleRepo) AssignRoles(ctx context.Context, userID string, roleIDs []string, tenantID string) error {
	// 1. 删除用户在该租户下的所有现有角色
	if err := r.DeleteUserRoles(ctx, userID, tenantID); err != nil {
		return err
	}

	// 2. 添加新角色
	for _, roleID := range roleIDs {
		if err := r.AddUserRole(ctx, userID, roleID, tenantID); err != nil {
			return err
		}
	}

	return nil
}

// AddRoles 为用户批量添加角色（增量式）
func (r *UserRoleRepo) AddRoles(ctx context.Context, userID string, roleIDs []string, tenantID string) error {
	existingRoles, err := r.GetUserRoleIDs(ctx, userID, tenantID)
	if err != nil {
		return err
	}

	existingMap := make(map[string]bool, len(existingRoles))
	for _, id := range existingRoles {
		existingMap[id] = true
	}

	for _, roleID := range roleIDs {
		if !existingMap[roleID] {
			if err := r.AddUserRole(ctx, userID, roleID, tenantID); err != nil {
				return err
			}
		}
	}

	return nil
}

// RemoveRoles 为用户批量删除角色
func (r *UserRoleRepo) RemoveRoles(ctx context.Context, userID string, roleIDs []string, tenantID string) error {
	for _, roleID := range roleIDs {
		if err := r.DeleteUserRole(ctx, userID, roleID, tenantID); err != nil {
			return err
		}
	}
	return nil
}

// GetUserRolesByRoleIDs 根据角色ID列表获取所有用户角色关联
func (r *UserRoleRepo) GetUserRolesByRoleIDs(ctx context.Context, roleIDs []string, tenantID string) ([]*model.UserRole, error) {
	return r.q.UserRole.WithContext(ctx).
		Where(r.q.UserRole.RoleID.In(roleIDs...)).
		Where(r.q.UserRole.TenantID.Eq(tenantID)).
		Find()
}

// DeleteRoles 批量删除指定角色的所有关联
func (r *UserRoleRepo) DeleteRoles(ctx context.Context, roleIDs []string, tenantID string) error {
	_, err := r.q.UserRole.WithContext(ctx).
		Where(r.q.UserRole.RoleID.In(roleIDs...)).
		Where(r.q.UserRole.TenantID.Eq(tenantID)).
		Delete()
	return err
}
