package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

// RolePermissionRepo 角色权限关联仓储
type RolePermissionRepo struct {
	db *gorm.DB
	q  *query.Query
}

// NewRolePermissionRepo 创建角色权限仓储
func NewRolePermissionRepo(db *gorm.DB) *RolePermissionRepo {
	return &RolePermissionRepo{
		db: db,
		q:  query.Use(db),
	}
}

// AddPermission 为角色添加权限
func (r *RolePermissionRepo) AddPermission(ctx context.Context, roleID, permissionID, tenantID string) error {
	rp := &model.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		TenantID:     tenantID,
	}
	return r.q.RolePermission.WithContext(ctx).Create(rp)
}

// AddPermissions 批量添加权限
func (r *RolePermissionRepo) AddPermissions(ctx context.Context, items []*model.RolePermission) error {
	if len(items) == 0 {
		return nil
	}
	return r.q.RolePermission.WithContext(ctx).Create(items...)
}

// DeleteByRole 删除角色的所有权限
func (r *RolePermissionRepo) DeleteByRole(ctx context.Context, roleID, tenantID string) error {
	_, err := r.q.RolePermission.WithContext(ctx).
		Where(r.q.RolePermission.RoleID.Eq(roleID)).
		Where(r.q.RolePermission.TenantID.Eq(tenantID)).
		Delete()
	return err
}

// DeleteByRoles 批量删除多个角色的所有权限
func (r *RolePermissionRepo) DeleteByRoles(ctx context.Context, roleIDs []string, tenantID string) error {
	_, err := r.q.RolePermission.WithContext(ctx).
		Where(r.q.RolePermission.RoleID.In(roleIDs...)).
		Where(r.q.RolePermission.TenantID.Eq(tenantID)).
		Delete()
	return err
}

// GetPermissionIDsByRole 获取角色的权限ID列表
func (r *RolePermissionRepo) GetPermissionIDsByRole(ctx context.Context, roleID, tenantID string) ([]string, error) {
	rps, err := r.q.RolePermission.WithContext(ctx).
		Where(r.q.RolePermission.RoleID.Eq(roleID)).
		Where(r.q.RolePermission.TenantID.Eq(tenantID)).
		Find()
	if err != nil {
		return nil, err
	}

	permIDs := make([]string, len(rps))
	for i, rp := range rps {
		permIDs[i] = rp.PermissionID
	}
	return permIDs, nil
}

// GetPermissionIDsByRoles 批量获取角色的权限ID列表
func (r *RolePermissionRepo) GetPermissionIDsByRoles(ctx context.Context, roleIDs []string, tenantID string) ([]string, error) {
	rps, err := r.q.RolePermission.WithContext(ctx).
		Where(r.q.RolePermission.RoleID.In(roleIDs...)).
		Where(r.q.RolePermission.TenantID.Eq(tenantID)).
		Find()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var permIDs []string
	for _, rp := range rps {
		if !seen[rp.PermissionID] {
			seen[rp.PermissionID] = true
			permIDs = append(permIDs, rp.PermissionID)
		}
	}
	return permIDs, nil
}
