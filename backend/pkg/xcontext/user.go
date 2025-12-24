package xcontext

import "context"

const (
	// 用户相关
	UserIDKey   contextKey = "user_id"
	UserNameKey contextKey = "user_name"
)

// UserContext 用户上下文信息
type UserContext struct {
	UserID   string
	UserName string
}

// SetUserID 设置用户ID到context
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID 从context获取用户ID，如果不存在返回空字符串
func GetUserID(ctx context.Context) string {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return ""
	}
	userID, _ := value.(string)
	return userID
}

// SetUserName 设置用户名到context
func SetUserName(ctx context.Context, userName string) context.Context {
	return context.WithValue(ctx, UserNameKey, userName)
}

// GetUserName 从context获取用户名，如果不存在返回空字符串
func GetUserName(ctx context.Context) string {
	value := ctx.Value(UserNameKey)
	if value == nil {
		return ""
	}
	userName, _ := value.(string)
	return userName
}
