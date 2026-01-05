package xcontext

import (
	"admin/pkg/database"
	"context"
)

// 定义context key类型，避免冲突
type contextKey string

const (
	// 租户相关
	TenantIDKey   contextKey = "tenant_id"
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

// GetTenantCode 从context获取租户代码，如果不存在返回空字符串
func GetTenantCode(ctx context.Context) string {
	value := ctx.Value(TenantCodeKey)
	if value == nil {
		return ""
	}
	tenantCode, _ := value.(string)
	return tenantCode
}

// CopyContext 将认证相关上下文信息拷贝到 background context
// 用于异步场景：避免请求取消影响后台任务，同时保留租户和用户信息
func CopyContext(ctx context.Context) context.Context {
	bg := context.Background()
	if tenantID := GetTenantID(ctx); tenantID != "" {
		bg = SetTenantID(bg, tenantID)
		// 同时设置 database 包需要的租户ID（GORM 租户插件使用 string 类型的 key）
		bg = database.WithTenantID(bg, tenantID)
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
	return bg
}
