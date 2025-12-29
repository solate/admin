package service

import (
	"admin/internal/dal/model"
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
func (s *OperationLogService) GetOperationLogByID(ctx context.Context, logID string) (*dto.OperationLogResponse, error) {
	log, err := s.operationLogRepo.GetByID(ctx, logID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询操作日志失败", err)
	}

	return s.toOperationLogResponse(log), nil
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

	responses := make([]*dto.OperationLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = s.toOperationLogResponse(log)
	}

	return &dto.ListOperationLogsResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     responses,
	}, nil
}

// toOperationLogResponse 转换为操作日志响应格式
func (s *OperationLogService) toOperationLogResponse(log *model.OperationLog) *dto.OperationLogResponse {
	resp := &dto.OperationLogResponse{
		LogID:     log.LogID,
		TenantID:  log.TenantID,
		UserID:    log.UserID,
		UserName:  log.UserName,
		Nickname:  log.Nickname,
		Module:    log.Module,
		OperationType:   log.OperationType,
		ResourceType:    log.ResourceType,
		ResourceID:      log.ResourceID,
		ResourceName:    log.ResourceName,
		RequestMethod:   log.RequestMethod,
		RequestPath:     log.RequestPath,
		Status:          int(log.Status),
		ErrorMessage:    log.ErrorMessage,
		IPAddress:       log.IPAddress,
		UserAgent:       log.UserAgent,
		CreatedAt:       log.CreatedAt,
	}

	// 处理可选字段 - 现在字段是 string 类型，直接赋值
	resp.RequestParams = log.RequestParams
	resp.OldValue = log.OldValue
	resp.NewValue = log.NewValue

	return resp
}
