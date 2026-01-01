package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
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

// GetByTenantAndUserName 根据租户ID和用户名获取用户（用于登录）
func (r *UserRepo) GetByTenantAndUserName(ctx context.Context, tenantID, userName string) (*model.User, error) {
	return r.q.User.WithContext(ctx).
		Where(r.q.User.TenantID.Eq(tenantID)).
		Where(r.q.User.UserName.Eq(userName)).
		First()
}

// Update 更新用户
func (r *UserRepo) Update(ctx context.Context, userID string, updates map[string]interface{}) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Updates(updates)
	return err
}

// Delete 删除用户(软删除)
func (r *UserRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Delete()
	return err
}

// List 分页获取用户列表
func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	return r.q.User.WithContext(ctx).FindByPage(offset, limit)
}

// ListWithFilters 根据筛选条件分页获取用户列表
func (r *UserRepo) ListWithFilters(ctx context.Context, offset, limit int, userNameFilter string, statusFilter int) ([]*model.User, int64, error) {
	query := r.q.User.WithContext(ctx)

	// 应用筛选条件
	if userNameFilter != "" {
		query = query.Where(r.q.User.UserName.Like("%" + userNameFilter + "%"))
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

// CheckExists 检查用户是否存在（租户内唯一）
func (r *UserRepo) CheckExists(ctx context.Context, tenantID, userName string) (bool, error) {
	count, err := r.q.User.WithContext(ctx).
		Where(r.q.User.TenantID.Eq(tenantID)).
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
