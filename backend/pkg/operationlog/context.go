package operationlog

import (
	"context"
	"time"
)

// LogContext 操作日志上下文，用于在请求生命周期中传递日志信息
type LogContext struct {
	// 必填字段
	TenantID      string       // 租户ID
	Module        string       // 模块名称 (如: user, role, tenant)
	OperationType string       // 操作类型 (CREATE, UPDATE, DELETE, QUERY, etc.)

	// 资源信息 (可选)
	ResourceType string   // 资源类型 (如: user, role)
	ResourceID   string   // 资源ID
	ResourceName string   // 资源名称

	// 数据变更 (可选)
	OldValue any // 操作前数据 (会被序列化为 JSON)
	NewValue any // 操作后数据 (会被序列化为 JSON)

	// 操作结果 (在响应时设置)
	Status       int16   // 操作状态 (1:成功, 2:失败)
	ErrorMessage string  // 错误信息 (失败时)

	// 时间戳
	CreatedAt int64 // 创建时间 (毫秒时间戳)
}

// contextKey 是 context.Context 的 key 类型，防止冲突
type contextKey int

const (
	// CtxLogContext 操作日志上下文的 key
	CtxLogContext contextKey = iota
)

// WithLogContext 将 LogContext 存入 context
func WithLogContext(ctx context.Context, lc *LogContext) context.Context {
	return context.WithValue(ctx, CtxLogContext, lc)
}

// GetLogContext 从 context 获取 LogContext
func GetLogContext(ctx context.Context) (*LogContext, bool) {
	lc, ok := ctx.Value(CtxLogContext).(*LogContext)
	return lc, ok
}

// MustGetLogContext 从 context 获取 LogContext，如果不存在则返回新的
func MustGetLogContext(ctx context.Context) *LogContext {
	if lc, ok := GetLogContext(ctx); ok {
		return lc
	}
	return &LogContext{
		CreatedAt: time.Now().UnixMilli(),
	}
}

// SetResource 设置资源信息 (链式调用)
func (lc *LogContext) SetResource(resourceType, resourceID, resourceName string) *LogContext {
	lc.ResourceType = resourceType
	lc.ResourceID = resourceID
	lc.ResourceName = resourceName
	return lc
}

// SetDataChange 设置数据变更 (链式调用)
func (lc *LogContext) SetDataChange(oldValue, newValue any) *LogContext {
	lc.OldValue = oldValue
	lc.NewValue = newValue
	return lc
}

// SetSuccess 设置操作成功 (链式调用)
func (lc *LogContext) SetSuccess() *LogContext {
	lc.Status = 1
	return lc
}

// SetError 设置操作失败 (链式调用)
func (lc *LogContext) SetError(err error) *LogContext {
	lc.Status = 2
	if err != nil {
		lc.ErrorMessage = err.Error()
	}
	return lc
}

// OperationBuilder 操作日志构建器
type OperationBuilder struct {
	lc *LogContext
}

// NewOperation 创建操作日志构建器
func NewOperation(module, operationType string) *OperationBuilder {
	return &OperationBuilder{
		lc: &LogContext{
			Module:        module,
			OperationType: operationType,
			Status:        1, // 默认成功
			CreatedAt:     time.Now().UnixMilli(),
		},
	}
}

// Resource 设置资源信息
func (b *OperationBuilder) Resource(resourceType, resourceID, resourceName string) *OperationBuilder {
	b.lc.ResourceType = resourceType
	b.lc.ResourceID = resourceID
	b.lc.ResourceName = resourceName
	return b
}

// Data 设置数据变更
func (b *OperationBuilder) Data(oldValue, newValue any) *OperationBuilder {
	b.lc.OldValue = oldValue
	b.lc.NewValue = newValue
	return b
}

// Error 设置操作失败
func (b *OperationBuilder) Error(err error) *OperationBuilder {
	b.lc.Status = 2
	if err != nil {
		b.lc.ErrorMessage = err.Error()
	}
	return b
}

// Build 构建并返回 LogContext
func (b *OperationBuilder) Build() *LogContext {
	return b.lc
}
