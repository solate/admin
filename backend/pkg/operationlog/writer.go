package operationlog

import (
	"admin/internal/dal/model"
	"admin/pkg/idgen"
	"context"
	"encoding/json"

	"gorm.io/gorm"
)

// Writer 日志写入器
type Writer struct {
	db *gorm.DB
}

// NewWriter 创建日志写入器
func NewWriter(db *gorm.DB) *Writer {
	return &Writer{db: db}
}

// Write 写入日志
func (w *Writer) Write(ctx context.Context, entry *LogEntry) error {
	if entry == nil || entry.LogContext == nil {
		return nil
	}

	lc := entry.LogContext

	// 根据操作类型写入不同的表
	switch lc.OperationType {
	case "LOGIN", "LOGOUT":
		return w.writeLoginLog(entry, lc)
	default:
		return w.writeOperationLog(entry, lc)
	}
}

// writeLoginLog 写入登录日志
func (w *Writer) writeLoginLog(entry *LogEntry, lc *LogContext) error {
	id, err := idgen.GenerateUUID()
	if err != nil {
		return err
	}

	log := &model.LoginLog{
		LogID:      id,
		TenantID:   entry.TenantID,
		UserID:     entry.UserID,
		UserName:   entry.UserName,
		Nickname:   entry.Nickname,
		LoginType:  lc.Module,
		LoginIP:    entry.IPAddress,
		UserAgent:  entry.UserAgent,
		Status:     lc.Status,
		FailReason: lc.ErrorMessage,
		CreatedAt:  lc.CreatedAt,
	}

	go w.db.Create(log)
	return nil
}

// writeOperationLog 写入操作日志
func (w *Writer) writeOperationLog(entry *LogEntry, lc *LogContext) error {
	id, err := idgen.GenerateUUID()
	if err != nil {
		return err
	}

	log := &model.OperationLog{
		LogID:         id,
		TenantID:      entry.TenantID,
		UserID:        entry.UserID,
		UserName:      entry.UserName,
		Nickname:      entry.Nickname,
		Module:        lc.Module,
		OperationType: lc.OperationType,
		ResourceType:  lc.ResourceType,
		ResourceID:    lc.ResourceID,
		ResourceName:  lc.ResourceName,
		RequestMethod: entry.RequestMethod,
		RequestPath:   entry.RequestPath,
		RequestParams: entry.RequestParams,
		OldValue:      serializeJSON(lc.OldValue),
		NewValue:      serializeJSON(lc.NewValue),
		Status:        lc.Status,
		ErrorMessage:  lc.ErrorMessage,
		IPAddress:     entry.IPAddress,
		UserAgent:     entry.UserAgent,
		CreatedAt:     lc.CreatedAt,
	}

	go w.db.Create(log)
	return nil
}

func serializeJSON(v any) string {
	if v == nil {
		return ""
	}
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}
