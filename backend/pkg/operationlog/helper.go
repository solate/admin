package operationlog

import (
	"admin/pkg/constants"
	"admin/pkg/xcontext"
	"context"
	"time"
)

// Record 记录操作 (便捷方法，直接设置到上下文)
// 用法示例:
//
//	ctx = operationlog.Record(ctx, "user", constants.OperationCreate).
//	    Resource("user", "123", "张三").
//	    Data(nil, user).
//	    BuildToContext(ctx)
func Record(ctx context.Context, module, operationType string) *OperationBuilder {
	lc := &LogContext{
		TenantID:      xcontext.GetTenantID(ctx),
		Module:        module,
		OperationType: operationType,
		CreatedAt:     time.Now().UnixMilli(),
	}

	// 返回构建器，支持链式调用
	return &OperationBuilder{lc: lc}
}

// RecordCreate 记录创建操作
func RecordCreate(ctx context.Context, module string, resourceType, resourceID, resourceName string, newValue any) context.Context {
	return Record(ctx, module, constants.OperationCreate).
		Resource(resourceType, resourceID, resourceName).
		Data(nil, newValue).
		BuildToContext(ctx)
}

// RecordUpdate 记录更新操作
func RecordUpdate(ctx context.Context, module string, resourceType, resourceID, resourceName string, oldValue, newValue any) context.Context {
	return Record(ctx, module, constants.OperationUpdate).
		Resource(resourceType, resourceID, resourceName).
		Data(oldValue, newValue).
		BuildToContext(ctx)
}

// RecordDelete 记录删除操作
func RecordDelete(ctx context.Context, module string, resourceType, resourceID, resourceName string, oldValue any) context.Context {
	return Record(ctx, module, constants.OperationDelete).
		Resource(resourceType, resourceID, resourceName).
		Data(oldValue, nil).
		BuildToContext(ctx)
}

// RecordQuery 记录查询操作
func RecordQuery(ctx context.Context, module, resourceType string) context.Context {
	return Record(ctx, module, constants.OperationQuery).
		Resource(resourceType, "", "").
		BuildToContext(ctx)
}

// RecordLogin 记录登录操作
func RecordLogin(ctx context.Context, userID, userName string) context.Context {
	return Record(ctx, constants.ModuleAuth, constants.OperationLogin).
		Resource("user", userID, userName).
		BuildToContext(ctx)
}

// RecordLogout 记录登出操作
func RecordLogout(ctx context.Context, userID, userName string) context.Context {
	return Record(ctx, constants.ModuleAuth, constants.OperationLogout).
		Resource("user", userID, userName).
		BuildToContext(ctx)
}

// RecordError 记录操作失败
func RecordError(ctx context.Context, err error) {
	if lc, ok := GetLogContext(ctx); ok {
		lc.SetError(err)
	}
}

// BuildToContext 将构建的 LogContext 存入 context 并返回
func (b *OperationBuilder) BuildToContext(ctx context.Context) context.Context {
	return WithLogContext(ctx, b.lc)
}
