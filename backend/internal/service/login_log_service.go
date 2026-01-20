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

// LoginLogService 登录日志服务
type LoginLogService struct {
	loginLogRepo *repository.LoginLogRepo
}

// NewLoginLogService 创建登录日志服务
func NewLoginLogService(loginLogRepo *repository.LoginLogRepo) *LoginLogService {
	return &LoginLogService{
		loginLogRepo: loginLogRepo,
	}
}

// GetLoginLogByID 根据ID获取登录日志
func (s *LoginLogService) GetLoginLogByID(ctx context.Context, logID string) (*dto.LoginLogInfo, error) {
	log, err := s.loginLogRepo.GetByID(ctx, logID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询登录日志失败", err)
	}

	return converter.ModelToLoginLogInfo(log), nil
}

// ListLoginLogs 获取登录日志列表
func (s *LoginLogService) ListLoginLogs(ctx context.Context, req *dto.ListLoginLogsRequest) (*dto.ListLoginLogsResponse, error) {
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	logs, total, err := s.loginLogRepo.ListWithFilters(
		ctx,
		tenantID,
		req.GetOffset(),
		req.GetLimit(),
		req.UserID,
		req.UserName,
		req.OperationType,
		req.LoginType,
		req.IPAddress,
		req.Status,
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询登录日志列表失败", err)
	}

	return &dto.ListLoginLogsResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     converter.ModelListToLoginLogInfoList(logs),
	}, nil
}
