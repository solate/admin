package service

import (
	"admin/internal/converter"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"gorm.io/gorm"
)

// OperationLogService 操作日志服务
type OperationLogService struct {
	operationLogRepo *repository.OperationLogRepo
}

// NewOperationLogService 创建操作日志服务
func NewOperationLogService(operationLogRepo *repository.OperationLogRepo) *OperationLogService {
	return &OperationLogService{
		operationLogRepo: operationLogRepo,
	}
}

// GetOperationLogByID 根据ID获取操作日志
func (s *OperationLogService) GetOperationLogByID(ctx context.Context, logID string) (*dto.OperationLogInfo, error) {
	log, err := s.operationLogRepo.GetByID(ctx, logID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询操作日志失败", err)
	}

	return converter.ModelToOperationLogInfo(log), nil
}

// ListOperationLogs 获取操作日志列表
func (s *OperationLogService) ListOperationLogs(ctx context.Context, req *dto.ListOperationLogsRequest) (*dto.ListOperationLogsResponse, error) {
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	logs, total, err := s.operationLogRepo.ListWithFilters(
		ctx,
		tenantID,
		req.GetOffset(),
		req.GetLimit(),
		req.Module,
		req.OperationType,
		req.ResourceType,
		req.UserName,
		req.Status,
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询操作日志列表失败", err)
	}

	return &dto.ListOperationLogsResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     converter.ModelListToOperationLogInfoList(logs),
	}, nil
}
