package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"admin/pkg/xcontext"
	"context"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建用户
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	return r.q.User.WithContext(ctx).Create(user)
}

// GetByID 根据ID获取用户
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
	return r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).First()
}

// GetByTenantAndUserName 根据租户ID和用户名获取用户（用于登录，手动模式）
func (r *UserRepo) GetByTenantAndUserName(ctx context.Context, tenantID, userName string) (*model.User, error) {
	ctx = xcontext.SkipTenantCheck(ctx)
	return r.q.User.WithContext(ctx).
		Where(r.q.User.TenantID.Eq(tenantID)).
		Where(r.q.User.UserName.Eq(userName)).
		First()
}

// GetByEmail 根据邮箱获取用户（用于登录，邮箱全局唯一，手动模式）
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx = xcontext.SkipTenantCheck(ctx)
	return r.q.User.WithContext(ctx).
		Where(r.q.User.Email.Eq(email)).
		First()
}

// GetByPhone 根据手机号获取用户（用于手机号登录，手机号全局唯一，手动模式）
func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	ctx = xcontext.SkipTenantCheck(ctx)
	return r.q.User.WithContext(ctx).
		Where(r.q.User.Phone.Eq(phone)).
		First()
}

// Update 更新用户
func (r *UserRepo) Update(ctx context.Context, userID string, updates map[string]interface{}) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Updates(updates)
	return err
}

// UpdateManual 手动模式更新用户（不自动添加租户过滤）
// 适用于：登录等场景，此时 context 中没有租户信息
func (r *UserRepo) UpdateManual(ctx context.Context, userID string, updates map[string]interface{}) error {
	ctx = xcontext.SkipTenantCheck(ctx)
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Updates(updates)
	return err
}

// Delete 删除用户(软删除)
func (r *UserRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Delete()
	return err
}

// BatchDelete 批量删除用户
func (r *UserRepo) BatchDelete(ctx context.Context, userIDs []string) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.In(userIDs...)).Delete()
	return err
}

// List 分页获取用户列表
func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	return r.q.User.WithContext(ctx).FindByPage(offset, limit)
}

// ListWithFilters 根据筛选条件分页获取用户列表
func (r *UserRepo) ListWithFilters(ctx context.Context, offset, limit int, nicknameFilter string, statusFilter int) ([]*model.User, int64, error) {
	return r.ListWithFiltersAndTenant(ctx, offset, limit, nicknameFilter, statusFilter, "")
}

// ListWithFiltersAndTenant 根据筛选条件和租户ID分页获取用户列表
// tenantID 为空时不过滤租户，非空时使用手动模式跨租户查询
func (r *UserRepo) ListWithFiltersAndTenant(ctx context.Context, offset, limit int, nicknameFilter string, statusFilter int, tenantID string) ([]*model.User, int64, error) {
	query := r.q.User.WithContext(ctx)

	// 如果指定了租户ID，使用手动模式跨租户查询
	if tenantID != "" {
		ctx = xcontext.SkipTenantCheck(ctx)
		query = r.q.User.WithContext(ctx).Where(r.q.User.TenantID.Eq(tenantID))
	}

	// 应用筛选条件
	if nicknameFilter != "" {
		query = query.Where(r.q.User.Nickname.Like("%" + nicknameFilter + "%"))
	}
	if statusFilter != 0 {
		query = query.Where(r.q.User.Status.Eq(int16(statusFilter)))
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	users, err := query.Order(r.q.User.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
	return users, total, err
}

// UpdateStatus 更新用户状态
func (r *UserRepo) UpdateStatus(ctx context.Context, userID string, status int) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Update(r.q.User.Status, status)
	return err
}

// CheckExists 检查用户是否存在（租户过滤由 scope callback 自动处理）
func (r *UserRepo) CheckExists(ctx context.Context, tenantID, userName string) (bool, error) {
	count, err := r.q.User.WithContext(ctx).
		Where(r.q.User.UserName.Eq(userName)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByIDsAndFilters 根据用户ID列表和筛选条件分页获取用户列表
func (r *UserRepo) ListByIDsAndFilters(ctx context.Context, userIDs []string, offset, limit int, keywordFilter string, statusFilter int) ([]*model.User, int64, error) {
	query := r.q.User.WithContext(ctx).Where(r.q.User.UserID.In(userIDs...))

	// 应用筛选条件
	if keywordFilter != "" {
		query = query.Where(r.q.User.UserName.Like("%" + keywordFilter + "%")).
			Or(r.q.User.Nickname.Like("%" + keywordFilter + "%"))
	}
	if statusFilter != 0 {
		query = query.Where(r.q.User.Status.Eq(int16(statusFilter)))
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	users, err := query.Order(r.q.User.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
	return users, total, err
}

// CountByDept 统计部门下的用户数
func (r *UserRepo) CountByDept(ctx context.Context, departmentID string) (int64, error) {
	return r.q.User.WithContext(ctx).
		Where(r.q.User.DepartmentID.Eq(departmentID)).
		Count()
}

// CountByTenantID 统计租户下的用户数（使用手动模式跨租户查询）
func (r *UserRepo) CountByTenantID(ctx context.Context, tenantID string) (int64, error) {
	ctx = xcontext.SkipTenantCheck(ctx)
	return r.q.User.WithContext(ctx).
		Where(r.q.User.TenantID.Eq(tenantID)).
		Count()
}

// CountByTenantIDs 批量统计多个租户下的用户数（使用手动模式跨租户查询）
// 返回 map[tenantID]userCount
func (r *UserRepo) CountByTenantIDs(ctx context.Context, tenantIDs []string) (map[string]int64, error) {
	ctx = xcontext.SkipTenantCheck(ctx)

	// 查询所有租户的用户统计
	type TenantUserCount struct {
		TenantID  string
		UserCount int64
	}

	var results []TenantUserCount
	err := r.db.WithContext(ctx).
		Table(r.q.User.TableName()).
		Select("tenant_id, count(*) as user_count").
		Where("tenant_id IN ?", tenantIDs).
		Group("tenant_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为 map
	countMap := make(map[string]int64, len(results))
	for _, r := range results {
		countMap[r.TenantID] = r.UserCount
	}

	// 为没有用户的租户补充 0
	for _, tenantID := range tenantIDs {
		if _, exists := countMap[tenantID]; !exists {
			countMap[tenantID] = 0
		}
	}

	return countMap, nil
}

// ListByDeptWithChildren 按部门及子部门查询用户
func (r *UserRepo) ListByDeptWithChildren(ctx context.Context, departmentIDs []string) ([]*model.User, error) {
	return r.q.User.WithContext(ctx).
		Where(r.q.User.DepartmentID.In(departmentIDs...)).
		Find()
}

// UpdatePassword 更新用户密码
func (r *UserRepo) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	_, err := r.q.User.WithContext(ctx).
		Where(r.q.User.UserID.Eq(userID)).
		Update(r.q.User.Password, hashedPassword)
	return err
}

// GetByIDs 根据用户ID列表获取用户信息
func (r *UserRepo) GetByIDs(ctx context.Context, userIDs []string) ([]*model.User, error) {
	return r.q.User.WithContext(ctx).Where(r.q.User.UserID.In(userIDs...)).Find()
}

// FindUserIDsByName 根据昵称模糊匹配获取用户ID列表（支持同名用户）
func (r *UserRepo) FindUserIDsByName(ctx context.Context, nickname string) ([]string, error) {
	users, err := r.q.User.WithContext(ctx).
		Where(r.q.User.Nickname.Like("%" + nickname + "%")).
		Select(r.q.User.UserID).
		Find()
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.UserID
	}
	return userIDs, nil
}
