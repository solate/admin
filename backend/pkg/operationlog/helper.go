package operationlog

import (
	"admin/pkg/constants"
	"admin/pkg/xcontext"
	"context"
	"time"
)

// Record 记录操作（返回 LogContext）
func Record(ctx context.Context, module, operationType string) *LogContext {
	return &LogContext{
		TenantID:      xcontext.GetTenantID(ctx),
		Module:        module,
		OperationType: operationType,
		Status:        1, // 默认成功
		CreatedAt:     time.Now().UnixMilli(),
	}
}

// RecordCreate 记录创建操作
func RecordCreate(ctx context.Context, module string, resourceType, resourceID, resourceName string, newValue any) context.Context {
	lc := Record(ctx, module, constants.OperationCreate)
	lc.ResourceType = resourceType
	lc.ResourceID = resourceID
	lc.ResourceName = resourceName
	lc.NewValue = newValue
	return WithLogContext(ctx, lc)
}

// RecordUpdate 记录更新操作
func RecordUpdate(ctx context.Context, module string, resourceType, resourceID, resourceName string, oldValue, newValue any) context.Context {
	lc := Record(ctx, module, constants.OperationUpdate)
	lc.ResourceType = resourceType
	lc.ResourceID = resourceID
	lc.ResourceName = resourceName
	lc.OldValue = oldValue
	lc.NewValue = newValue
	return WithLogContext(ctx, lc)
}

// RecordDelete 记录删除操作
func RecordDelete(ctx context.Context, module string, resourceType, resourceID, resourceName string, oldValue any) context.Context {
	lc := Record(ctx, module, constants.OperationDelete)
	lc.ResourceType = resourceType
	lc.ResourceID = resourceID
	lc.ResourceName = resourceName
	lc.OldValue = oldValue
	return WithLogContext(ctx, lc)
}

// RecordQuery 记录查询操作
func RecordQuery(ctx context.Context, module, resourceType string) context.Context {
	lc := Record(ctx, module, constants.OperationQuery)
	lc.ResourceType = resourceType
	return WithLogContext(ctx, lc)
}

// RecordLogin 记录登录操作
func RecordLogin(ctx context.Context, userID, userName string) context.Context {
	lc := Record(ctx, constants.ModuleAuth, constants.OperationLogin)
	lc.ResourceType = "user"
	lc.ResourceID = userID
	lc.ResourceName = userName
	return WithLogContext(ctx, lc)
}

// RecordLogout 记录登出操作
func RecordLogout(ctx context.Context, userID, userName string) context.Context {
	lc := Record(ctx, constants.ModuleAuth, constants.OperationLogout)
	lc.ResourceType = "user"
	lc.ResourceID = userID
	lc.ResourceName = userName
	return WithLogContext(ctx, lc)
}

// RecordError 记录操作失败
func RecordError(ctx context.Context, err error) {
	if lc, ok := GetLogContext(ctx); ok {
		lc.Status = 2
		if err != nil {
			lc.ErrorMessage = err.Error()
		}
	}
}
