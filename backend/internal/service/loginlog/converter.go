package loginlog

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// modelToLoginLogInfo 将数据库模型转换为登录日志信息 DTO
func modelToLoginLogInfo(log *model.LoginLog) *dto.LoginLogInfo {
	if log == nil {
		return nil
	}

	return &dto.LoginLogInfo{
		LogID:         log.LogID,
		TenantID:      log.TenantID,
		UserID:        log.UserID,
		UserName:      log.UserName,
		OperationType: log.OperationType,
		LoginType:     log.LoginType,
		LoginIP:       log.LoginIP,
		LoginLocation: log.LoginLocation,
		UserAgent:     log.UserAgent,
		Status:        log.Status,
		FailReason:    log.FailReason,
		CreatedAt:     log.CreatedAt,
	}
}

// modelListToLoginLogInfoList 批量将数据库模型转换为登录日志信息 DTO
func modelListToLoginLogInfoList(logs []*model.LoginLog) []*dto.LoginLogInfo {
	if len(logs) == 0 {
		return nil
	}

	result := make([]*dto.LoginLogInfo, len(logs))
	for i, log := range logs {
		result[i] = modelToLoginLogInfo(log)
	}
	return result
}
