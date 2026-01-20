package converter

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// ModelToOperationLogInfo 将数据库模型转换为操作日志信息 DTO
func ModelToOperationLogInfo(log *model.OperationLog) *dto.OperationLogInfo {
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

	// 处理可选字段
	resp.RequestParams = log.RequestParams
	resp.OldValue = log.OldValue
	resp.NewValue = log.NewValue

	return resp
}

// ModelListToOperationLogInfoList 批量将数据库模型转换为操作日志信息 DTO
func ModelListToOperationLogInfoList(logs []*model.OperationLog) []*dto.OperationLogInfo {
	if len(logs) == 0 {
		return nil
	}

	result := make([]*dto.OperationLogInfo, len(logs))
	for i, log := range logs {
		result[i] = ModelToOperationLogInfo(log)
	}
	return result
}
