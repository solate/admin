package auditlog

import (
	"admin/internal/dal/model"
	"admin/pkg/idgen"
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
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
		return w.writeLoginLog(ctx, entry, lc)
	default:
		return w.writeOperationLog(ctx, entry, lc)
	}
}

// writeLoginLog 写入登录日志
func (w *Writer) writeLoginLog(ctx context.Context, entry *LogEntry, lc *LogContext) error {
	log.Info().
		Str("tenant_id", entry.TenantID).
		Str("user_id", entry.UserID).
		Str("user_name", entry.UserName).
		Str("ip", entry.IPAddress).
		Int16("status", lc.Status).
		Msg("【审计日志】准备写入登录日志...")

	id, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("【审计日志】生成UUID失败")
		return err
	}

	loginLog := &model.LoginLog{
		LogID:      id,
		TenantID:   entry.TenantID,
		UserID:     entry.UserID,
		UserName:   entry.UserName,
		LoginType:  lc.Module,
		LoginIP:    entry.IPAddress,
		UserAgent:  entry.UserAgent,
		Status:     lc.Status,
		FailReason: lc.ErrorMessage,
		CreatedAt:  lc.CreatedAt,
	}

	// 使用传入的 ctx（已通过 Detach 处理，包含租户信息）
	if err := w.db.WithContext(ctx).Create(loginLog).Error; err != nil {
		log.Error().Err(err).Msg("【审计日志】写入登录日志失败")
		return err
	}

	log.Info().Str("log_id", id).Msg("【审计日志】登录日志写入成功")
	return nil
}

// writeOperationLog 写入操作日志
func (w *Writer) writeOperationLog(ctx context.Context, entry *LogEntry, lc *LogContext) error {
	log.Info().
		Str("tenant_id", entry.TenantID).
		Str("user_id", entry.UserID).
		Str("operation", lc.OperationType).
		Msg("【审计日志】准备写入操作日志...")

	logID, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("【审计日志】生成UUID失败")
		return err
	}

	operationLog := &model.OperationLog{
		LogID:         logID,
		TenantID:      entry.TenantID,
		UserID:        entry.UserID,
		UserName:      entry.UserName,
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

	// 使用传入的 ctx（已通过 Detach 处理，包含租户信息）
	if err := w.db.WithContext(ctx).Create(operationLog).Error; err != nil {
		log.Error().Err(err).Msg("【审计日志】写入操作日志失败")
		return err
	}

	log.Info().Str("log_id", logID).Msg("【审计日志】操作日志写入成功")
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
