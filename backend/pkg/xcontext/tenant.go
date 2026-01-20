package xcontext

import "context"

// 定义context key类型，避免冲突
type contextKey string

const (
	// 租户相关
	TenantIDKey        contextKey = "tenant_id"
	TenantCodeKey      contextKey = "tenant_code"
	SkipTenantCheckKey contextKey = "skip_tenant_check"
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

// GetTenantCode 从context获取租户代码，如果不存在返回空字符串
func GetTenantCode(ctx context.Context) string {
	value := ctx.Value(TenantCodeKey)
	if value == nil {
		return ""
	}
	tenantCode, ok := value.(string)
	if !ok {
		return ""
	}
	return tenantCode
}

// SkipTenantCheck 设置跳过租户检查标记到context
func SkipTenantCheck(ctx context.Context) context.Context {
	return context.WithValue(ctx, SkipTenantCheckKey, true)
}

// ShouldSkipTenantCheck 从context获取是否跳过租户检查
func ShouldSkipTenantCheck(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	value := ctx.Value(SkipTenantCheckKey)
	skip, ok := value.(bool)
	return ok && skip
}

// CopyContext 将认证相关上下文信息拷贝到 background context
// 用于异步场景：避免请求取消影响后台任务，同时保留租户、用户和角色信息
func CopyContext(ctx context.Context) context.Context {
	bg := context.Background()
	if tenantID := GetTenantID(ctx); tenantID != "" {
		bg = SetTenantID(bg, tenantID)
	}
	if tenantCode := GetTenantCode(ctx); tenantCode != "" {
		bg = SetTenantCode(bg, tenantCode)
	}
	if userID := GetUserID(ctx); userID != "" {
		bg = SetUserID(bg, userID)
	}
	if userName := GetUserName(ctx); userName != "" {
		bg = SetUserName(bg, userName)
	}
	if tokenID := GetTokenID(ctx); tokenID != "" {
		bg = SetTokenID(bg, tokenID)
	}
	if roles := GetRoles(ctx); roles != nil {
		bg = SetRoles(bg, roles)
	}
	return bg
}
