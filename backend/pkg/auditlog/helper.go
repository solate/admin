package auditlog

import (
	"admin/pkg/constants"
	"admin/pkg/database"
	"admin/pkg/useragent"
	"admin/pkg/xcontext"
	"context"
	"net/http"
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

// RecordError 记录操作失败
func RecordError(ctx context.Context, err error) {
	if lc, ok := GetLogContext(ctx); ok {
		lc.Status = 2
		if err != nil {
			lc.ErrorMessage = err.Error()
		}
	}
}

// RecordLogin 记录登录日志（无需走中间件，异步写入）
// err 为 nil 表示登录成功，否则表示登录失败
func RecordLogin(writer *Writer, r *http.Request, tenantID, userID, userName string, err error) {
	clientInfo := useragent.GetClientInfo(r)

	// 使用 Background 创建独立的 context，设置 SkipTenantCheck
	ctx := database.SkipTenantCheck(context.Background())
	lc := Record(ctx, constants.ModuleAuth, constants.OperationLogin)

	if err != nil {
		lc.Status = 2
		lc.ErrorMessage = err.Error()
	}

	entry := &LogEntry{
		TenantID:   tenantID,
		UserID:     userID,
		UserName:   userName,
		IPAddress:  clientInfo.IP,
		UserAgent:  clientInfo.UserAgent,
		LogContext: lc,
	}

	// 异步写入，不阻塞请求
	go writer.Write(ctx, entry)
}

// RecordLogout 记录登出日志（无需走中间件，异步写入）
func RecordLogout(writer *Writer, r *http.Request, tenantID, userID, userName string) {
	clientInfo := useragent.GetClientInfo(r)

	// 使用 Background 创建独立的 context，设置 SkipTenantCheck
	ctx := database.SkipTenantCheck(context.Background())
	lc := Record(ctx, constants.ModuleAuth, constants.OperationLogout)

	entry := &LogEntry{
		TenantID:   tenantID,
		UserID:     userID,
		UserName:   userName,
		IPAddress:  clientInfo.IP,
		UserAgent:  clientInfo.UserAgent,
		LogContext: lc,
	}

	// 异步写入，不阻塞请求
	go writer.Write(ctx, entry)
}
