package operationlog

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// modelToOperationLogInfo 将数据库模型转换为操作日志信息 DTO
func modelToOperationLogInfo(log *model.OperationLog) *dto.OperationLogInfo {
	if log == nil {
		return nil
	}

	resp := &dto.OperationLogInfo{
		LogID:         log.LogID,
		TenantID:      log.TenantID,
		UserID:        log.UserID,
		UserName:      log.UserName,
		Module:        log.Module,
		OperationType: log.OperationType,
		ResourceType:  log.ResourceType,
		ResourceID:    log.ResourceID,
		ResourceName:  log.ResourceName,
		RequestMethod: log.RequestMethod,
		RequestPath:   log.RequestPath,
		Status:        int(log.Status),
		ErrorMessage:  log.ErrorMessage,
		IPAddress:     log.IPAddress,
		UserAgent:     log.UserAgent,
		CreatedAt:     log.CreatedAt,
	}

	resp.RequestParams = log.RequestParams
	resp.OldValue = log.OldValue
	resp.NewValue = log.NewValue

	return resp
}

// modelListToOperationLogInfoList 批量将数据库模型转换为操作日志信息 DTO
func modelListToOperationLogInfoList(logs []*model.OperationLog) []*dto.OperationLogInfo {
	if len(logs) == 0 {
		return nil
	}

	result := make([]*dto.OperationLogInfo, len(logs))
	for i, log := range logs {
		result[i] = modelToOperationLogInfo(log)
	}
	return result
}
