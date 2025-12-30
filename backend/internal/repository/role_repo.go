package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/database"

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

// GetByCode 根据角色编码获取当前租户的角色（依赖自动模式添加 tenant_id 过滤）
func (r *RoleRepo) GetByCode(ctx context.Context, roleCode string) (*model.Role, error) {
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.RoleCode.Eq(roleCode)).
		First()
}

// GetByCodeWithTenant 根据租户ID和角色编码获取角色（手动模式，用于跨租户查询）
func (r *RoleRepo) GetByCodeWithTenant(ctx context.Context, tenantID, roleCode string) (*model.Role, error) {
	// 跨租户查询：使用 ManualTenantMode 禁止自动添加当前租户过滤
	ctx = database.ManualTenantMode(ctx)
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
func (r *RoleRepo) UpdateStatus(ctx context.Context, roleID string, status int) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.RoleID.Eq(roleID)).Update(r.q.Role.Status, int16(status))
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

// ListByCodes 根据角色编码列表获取角色列表（跳过租户过滤，支持角色继承）
func (r *RoleRepo) ListByCodes(ctx context.Context, roleCodes []string) ([]*model.Role, error) {
	ctx = database.ManualTenantMode(ctx)
	return r.q.Role.WithContext(ctx).
		Where(r.q.Role.RoleCode.In(roleCodes...)).
		Find()
}

// ListWithFilters 根据筛选条件分页获取角色列表
func (r *RoleRepo) ListWithFilters(ctx context.Context, tenantID string, offset, limit int, keywordFilter string, statusFilter int) ([]*model.Role, int64, error) {
	query := r.q.Role.WithContext(ctx).Where(r.q.Role.TenantID.Eq(tenantID))

	// 应用筛选条件
	if keywordFilter != "" {
		query = query.Where(r.q.Role.Name.Like("%"+keywordFilter+"%")).
			Or(r.q.Role.RoleCode.Like("%"+keywordFilter+"%"))
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
