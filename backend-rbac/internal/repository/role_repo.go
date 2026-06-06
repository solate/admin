package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/xcontext"
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
	role.TenantID = xcontext.GetTenantID(ctx)
	return r.q.Role.WithContext(ctx).Create(role)
}

// GetByID 根据ID获取角色
func (r *RoleRepo) GetByID(ctx context.Context, roleID string) (*model.Role, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleID.Eq(roleID)).
		First()
}

// GetByCode 根据角色编码获取当前租户的角色
func (r *RoleRepo) GetByCode(ctx context.Context, roleCode string) (*model.Role, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleCode.Eq(roleCode)).
		First()
}

// GetByCodeWithTenant 根据租户ID和角色编码获取角色（跨租户查询）
func (r *RoleRepo) GetByCodeWithTenant(ctx context.Context, tenantID, roleCode string) (*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleCode.Eq(roleCode)).
		First()
}

// Update 更新角色
func (r *RoleRepo) Update(ctx context.Context, roleID string, updates map[string]interface{}) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleID.Eq(roleID)).
		Updates(updates)
	return err
}

// Delete 删除角色(软删除)
func (r *RoleRepo) Delete(ctx context.Context, roleID string) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleID.Eq(roleID)).
		Delete()
	return err
}

// BatchDelete 批量删除角色
func (r *RoleRepo) BatchDelete(ctx context.Context, roleIDs []string) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleID.In(roleIDs...)).
		Delete()
	return err
}

// List 分页获取角色列表
func (r *RoleRepo) List(ctx context.Context, offset, limit int) ([]*model.Role, int64, error) {
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		FindByPage(offset, limit)
}

// ListWithFilters 根据筛选条件分页获取角色列表
func (r *RoleRepo) ListWithFilters(ctx context.Context, offset, limit int, roleName, roleCode string, statusFilter int) ([]*model.Role, int64, error) {
	tenantID := xcontext.GetTenantID(ctx)
	query := r.q.Role.WithContext(ctx).Where(r.q.Role.TenantID.Eq(tenantID))

	// 应用筛选条件
	if roleName != "" {
		query = query.Where(r.q.Role.Name.Like("%" + roleName + "%"))
	}
	if roleCode != "" {
		query = query.Where(r.q.Role.RoleCode.Like("%" + roleCode + "%"))
	}
	if statusFilter != 0 {
		query = query.Where(r.q.Role.Status.Eq(int16(statusFilter)))
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	roles, err := query.Order(r.q.Role.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
	return roles, total, err
}

// UpdateStatus 更新角色状态
func (r *RoleRepo) UpdateStatus(ctx context.Context, roleID string, status int) error {
	tenantID := xcontext.GetTenantID(ctx)
	_, err := r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleID.Eq(roleID)).
		Update(r.q.Role.Status, int16(status))
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
	tenantID := xcontext.GetTenantID(ctx)
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleID.In(roleIDs...)).
		Find()
}

// ListByCodes 根据角色编码列表获取角色列表（跨租户查询）
func (r *RoleRepo) ListByCodes(ctx context.Context, roleCodes []string) ([]*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.RoleCode.In(roleCodes...)).
		Find()
}

// ListByCodesWithTenant 根据租户ID和角色编码列表获取角色列表（跨租户查询）
func (r *RoleRepo) ListByCodesWithTenant(ctx context.Context, tenantID string, roleCodes []string) ([]*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleCode.In(roleCodes...)).
		Find()
}

// ListAll 根据筛选条件获取所有角色（不分页）
func (r *RoleRepo) ListAll(ctx context.Context, roleName, roleCode string, statusFilter int) ([]*model.Role, error) {
	tenantID := xcontext.GetTenantID(ctx)
	query := r.q.Role.WithContext(ctx).Where(r.q.Role.TenantID.Eq(tenantID))

	// 应用筛选条件
	if roleName != "" {
		query = query.Where(r.q.Role.Name.Like("%" + roleName + "%"))
	}
	if roleCode != "" {
		query = query.Where(r.q.Role.RoleCode.Like("%" + roleCode + "%"))
	}
	if statusFilter != 0 {
		query = query.Where(r.q.Role.Status.Eq(int16(statusFilter)))
	}

	return query.Order(r.q.Role.CreatedAt.Desc()).Find()
}

// ListByTenant 根据租户ID获取所有角色（跨租户查询）
func (r *RoleRepo) ListByTenant(ctx context.Context, tenantID, roleName, roleCode string, statusFilter int) ([]*model.Role, error) {
	query := r.q.Role.WithContext(ctx).Where(r.q.Role.TenantID.Eq(tenantID))

	if roleName != "" {
		query = query.Where(r.q.Role.Name.Like("%" + roleName + "%"))
	}
	if roleCode != "" {
		query = query.Where(r.q.Role.RoleCode.Like("%" + roleCode + "%"))
	}
	if statusFilter != 0 {
		query = query.Where(r.q.Role.Status.Eq(int16(statusFilter)))
	}

	return query.Order(r.q.Role.CreatedAt.Desc()).Find()
}

// ListByTenantWithFilters 根据租户ID分页获取角色列表（跨租户查询）
func (r *RoleRepo) ListByTenantWithFilters(ctx context.Context, tenantID string, offset, limit int, roleName, roleCode string, statusFilter int) ([]*model.Role, int64, error) {
	query := r.q.Role.WithContext(ctx).Where(r.q.Role.TenantID.Eq(tenantID))

	if roleName != "" {
		query = query.Where(r.q.Role.Name.Like("%" + roleName + "%"))
	}
	if roleCode != "" {
		query = query.Where(r.q.Role.RoleCode.Like("%" + roleCode + "%"))
	}
	if statusFilter != 0 {
		query = query.Where(r.q.Role.Status.Eq(int16(statusFilter)))
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	roles, err := query.Order(r.q.Role.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
	return roles, total, err
}

// CheckExistsByID 检查角色编码是否存在（排除指定ID）
func (r *RoleRepo) CheckExistsByID(ctx context.Context, tenantID, roleCode string, excludeRoleID string) (bool, error) {
	count, err := r.q.Role.WithContext(ctx).
		Where(r.q.Role.TenantID.Eq(tenantID)).
		Where(r.q.Role.RoleCode.Eq(roleCode)).
		Where(r.q.Role.RoleID.Neq(excludeRoleID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByIDs 根据角色ID列表获取角色信息（跨租户查询，用于权限缓存）
func (r *RoleRepo) GetByIDs(ctx context.Context, roleIDs []string) ([]*model.Role, error) {
	return r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.In(roleIDs...)).Find()
}
