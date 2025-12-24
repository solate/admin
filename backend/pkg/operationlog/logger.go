package operationlog

import (
	"admin/internal/dal/model"
	"admin/pkg/idgen"
	"context"
	"encoding/json"

	"gorm.io/gorm"
)

// Logger 操作日志写入器
type Logger struct {
	db *gorm.DB
}

// NewLogger 创建操作日志写入器
func NewLogger(db *gorm.DB) *Logger {
	return &Logger{db: db}
}

// LogEntry 日志条目 (由中间件收集后传递给 Write)
type LogEntry struct {
	// 从 request.Context 收集的信息 (由 AuthMiddleware 注入)
	TenantID string
	UserID   string
	UserName string

	// 从 request 收集的信息
	RequestMethod string
	RequestPath   string
	RequestParams string
	IPAddress     string
	UserAgent     string

	// 业务代码设置的日志上下文
	LogContext *LogContext
}

// Write 写入操作日志 (异步，不阻塞主流程)
func (l *Logger) Write(ctx context.Context, entry *LogEntry) error {
	if entry == nil || entry.LogContext == nil {
		return nil
	}

	lc := entry.LogContext

	// 序列化数据变更 (old_value, new_value)
	oldValueJSON := serializeJSON(lc.OldValue)
	newValueJSON := serializeJSON(lc.NewValue)

	// 构建日志记录
	log := &model.OperationLog{
		LogID:         generateLogID(),
		TenantID:      entry.TenantID,
		Module:        lc.Module,
		OperationType: lc.OperationType,
		ResourceType:  toStringPtr(lc.ResourceType),
		ResourceID:    toStringPtr(lc.ResourceID),
		ResourceName:  toStringPtr(lc.ResourceName),
		UserID:        entry.UserID,
		UserName:      entry.UserName,
		RequestMethod: toStringPtr(entry.RequestMethod),
		RequestPath:   toStringPtr(entry.RequestPath),
		RequestParams: toStringPtr(entry.RequestParams),
		OldValue:      oldValueJSON,
		NewValue:      newValueJSON,
		Status:        lc.Status,
		ErrorMessage:  toStringPtr(lc.ErrorMessage),
		IPAddress:     toStringPtr(entry.IPAddress),
		UserAgent:     toStringPtr(entry.UserAgent),
		CreatedAt:     lc.CreatedAt,
	}

	// 异步写入，不阻塞主流程
	go func() {
		_ = l.db.Create(log).Error
	}()

	return nil
}

// serializeJSON 序列化为 JSON 指针
func serializeJSON(v any) *string {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(data)
	return &s
}

// toStringPtr 字符串转指针 (空字符串返回 nil)
func toStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// generateLogID 生成日志 ID
func generateLogID() string {
	if id, err := idgen.GenerateUUID(); err == nil {
		return id
	}
	// fallback (实际不太可能走到这里)
	return "log_" + randomString(16)
}

// randomString 生成随机字符串 (fallback 用)
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[int64(i)*9%36]
	}
	return string(b)
}
