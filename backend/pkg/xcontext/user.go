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

// GetUserID 从context获取用户ID
func GetUserID(ctx context.Context) (string, bool) {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return "", false
	}
	userID, ok := value.(string)
	return userID, ok && userID != ""
}

// SetUserName 设置用户名到context
func SetUserName(ctx context.Context, userName string) context.Context {
	return context.WithValue(ctx, UserNameKey, userName)
}

// GetUserName 从context获取用户名
func GetUserName(ctx context.Context) (string, bool) {
	value := ctx.Value(UserNameKey)
	if value == nil {
		return "", false
	}
	userName, ok := value.(string)
	return userName, ok
}
