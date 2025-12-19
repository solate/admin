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

// CreateUser 创建用户
func (r *UserRepo) CreateUser(ctx context.Context, user *model.User) error {
	return r.q.User.WithContext(ctx).Create(user)
}

// GetUserByID 根据ID获取用户
func (r *UserRepo) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	return r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).First()
}

// GetUserByName 根据用户名获取用户
func (r *UserRepo) GetUserByName(ctx context.Context, userName string) (*model.User, error) {
	return r.q.User.WithContext(ctx).Where(r.q.User.UserName.Eq(userName)).First()
}

// UpdateUser 更新用户
func (r *UserRepo) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Updates(updates)
	return err
}

// DeleteUser 删除用户(软删除)
func (r *UserRepo) DeleteUser(ctx context.Context, userID string) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Delete()
	return err
}

// ListUsers 分页获取用户列表
func (r *UserRepo) ListUsers(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	return r.q.User.WithContext(ctx).FindByPage(offset, limit)
}

// UpdateUserStatus 更新用户状态
func (r *UserRepo) UpdateUserStatus(ctx context.Context, userID string, status int32) error {
	_, err := r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).Update(r.q.User.Status, status)
	return err
}

// CheckUserExists 检查用户是否存在
func (r *UserRepo) CheckUserExists(ctx context.Context, userName, tenantID string) (bool, error) {
	count, err := r.q.User.WithContext(ctx).Where(r.q.User.UserName.Eq(userName)).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
