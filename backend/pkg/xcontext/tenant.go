package xcontext

import "context"

// 定义context key类型，避免冲突
type contextKey string

const (
	// 租户相关 - 与database包使用相同的key
	TenantIDKey   contextKey = "tenant_id" // 与database包保持一致
	TenantCodeKey contextKey = "tenant_code"
)

// TenantContext 租户上下文信息
type TenantContext struct {
	TenantID   string
	TenantCode string
}

// SetTenantID 设置租户ID到context
func SetTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID 从context获取租户ID，如果不存在返回空字符串
func GetTenantID(ctx context.Context) string {
	value := ctx.Value(TenantIDKey)
	if value == nil {
		return ""
	}
	tenantID, ok := value.(string)
	if !ok {
		return ""
	}
	return tenantID
}

// SetTenantCode 设置租户代码到context
func SetTenantCode(ctx context.Context, tenantCode string) context.Context {
	return context.WithValue(ctx, TenantCodeKey, tenantCode)
}

// GetTenantCode 从context获取租户代码
func GetTenantCode(ctx context.Context) (string, bool) {
	value := ctx.Value(TenantCodeKey)
	if value == nil {
		return "", false
	}
	tenantCode, ok := value.(string)
	return tenantCode, ok
}
