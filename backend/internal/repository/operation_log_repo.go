package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

type OperationLogRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewOperationLogRepo(db *gorm.DB) *OperationLogRepo {
	return &OperationLogRepo{
		db: db,
		q:  query.Use(db),
	}
}

// GetByID 根据ID获取操作日志
func (r *OperationLogRepo) GetByID(ctx context.Context, logID string) (*model.OperationLog, error) {
	return r.q.OperationLog.WithContext(ctx).Where(r.q.OperationLog.LogID.Eq(logID)).First()
}

// ListWithFilters 根据筛选条件分页获取操作日志列表
func (r *OperationLogRepo) ListWithFilters(ctx context.Context, tenantID string, offset, limit int, module, operationType, resourceType, userName string, status int, startDate, endDate int64) ([]*model.OperationLog, int64, error) {
	q := r.q.OperationLog.WithContext(ctx).Where(r.q.OperationLog.TenantID.Eq(tenantID))

	// 应用筛选条件
	if module != "" {
		q = q.Where(r.q.OperationLog.Module.Eq(module))
	}
	if operationType != "" {
		q = q.Where(r.q.OperationLog.OperationType.Eq(operationType))
	}
	if resourceType != "" {
		q = q.Where(r.q.OperationLog.ResourceType.Eq(resourceType))
	}
	if userName != "" {
		q = q.Where(r.q.OperationLog.UserName.Like("%" + userName + "%"))
	}
	if status != 0 {
		q = q.Where(r.q.OperationLog.Status.Eq(int16(status)))
	}
	if startDate > 0 {
		q = q.Where(r.q.OperationLog.CreatedAt.Gte(startDate))
	}
	if endDate > 0 {
		q = q.Where(r.q.OperationLog.CreatedAt.Lte(endDate))
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	logs, err := q.Order(r.q.OperationLog.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
	return logs, total, err
}
