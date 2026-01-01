package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

type LoginLogRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewLoginLogRepo(db *gorm.DB) *LoginLogRepo {
	return &LoginLogRepo{
		db: db,
		q:  query.Use(db),
	}
}

// GetByID 根据ID获取登录日志
func (r *LoginLogRepo) GetByID(ctx context.Context, logID string) (*model.LoginLog, error) {
	return r.q.LoginLog.WithContext(ctx).Where(r.q.LoginLog.LogID.Eq(logID)).First()
}

// ListWithFilters 根据筛选条件分页获取登录日志列表
func (r *LoginLogRepo) ListWithFilters(ctx context.Context, tenantID string, offset, limit int, userID, userName, loginType, ipAddress string, status *int16, startDate, endDate *int64) ([]*model.LoginLog, int64, error) {
	q := r.q.LoginLog.WithContext(ctx).Where(r.q.LoginLog.TenantID.Eq(tenantID))

	// 应用筛选条件
	if userID != "" {
		q = q.Where(r.q.LoginLog.UserID.Eq(userID))
	}
	if userName != "" {
		q = q.Where(r.q.LoginLog.UserName.Like("%" + userName + "%"))
	}
	if loginType != "" {
		q = q.Where(r.q.LoginLog.LoginType.Eq(loginType))
	}
	if ipAddress != "" {
		q = q.Where(r.q.LoginLog.LoginIP.Like("%" + ipAddress + "%"))
	}
	if status != nil {
		q = q.Where(r.q.LoginLog.Status.Eq(*status))
	}
	if startDate != nil && *startDate > 0 {
		q = q.Where(r.q.LoginLog.CreatedAt.Gte(*startDate))
	}
	if endDate != nil && *endDate > 0 {
		q = q.Where(r.q.LoginLog.CreatedAt.Lte(*endDate))
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	logs, err := q.Order(r.q.LoginLog.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
	return logs, total, err
}

// StatsByDate 按日期统计登录数据
// TODO: 需要时实现此方法，需要先在 model 中定义 LoginLogStat
// func (r *LoginLogRepo) StatsByDate(ctx context.Context, tenantID string, days int) ([]*model.LoginLogStat, error) {
// 	return nil, nil
// }
