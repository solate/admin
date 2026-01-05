package audit

import (
	"admin/internal/dal/model"
	"admin/pkg/idgen"
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DB 数据库写入接口（简化命名）
type DB struct {
	db *gorm.DB
}

// NewDB 创建写入器
func NewDB(db *gorm.DB) *DB {
	return &DB{db: db}
}

// Write 写入日志（统一入口）
func (w *DB) Write(ctx context.Context, entry *LogEntry) error {
	if entry == nil {
		return nil
	}

	switch entry.OperationType {
	case OperationLogin, OperationLogout:
		return w.writeLoginLog(ctx, entry)
	default:
		return w.writeOperationLog(ctx, entry)
	}
}

func (w *DB) writeLoginLog(ctx context.Context, entry *LogEntry) error {
	log.Info().
		Str("tenant_id", entry.TenantID).
		Str("user_id", entry.UserID).
		Str("user_name", entry.UserName).
		Str("ip", entry.IPAddress).
		Int16("status", entry.Status).
		Msg("[审计日志] 准备写入登录日志...")

	logID, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("[审计日志] 生成UUID失败")
		return err
	}

	loginLog := &model.LoginLog{
		LogID:         logID,
		TenantID:      entry.TenantID,
		UserID:        entry.UserID,
		UserName:      entry.UserName,
		LoginType:     entry.Module,
		LoginIP:       entry.IPAddress,
		LoginLocation: "", // 暂不实现 IP 地址解析
		UserAgent:     entry.UserAgent,
		Status:        entry.Status,
		FailReason:    entry.ErrorMessage,
		CreatedAt:     entry.CreatedAt,
	}

	if err := w.db.WithContext(ctx).Create(loginLog).Error; err != nil {
		log.Error().Err(err).Msg("[审计日志] 写入登录日志失败")
		return err
	}

	log.Info().Str("log_id", logID).Msg("[审计日志] 登录日志写入成功")
	return nil
}

func (w *DB) writeOperationLog(ctx context.Context, entry *LogEntry) error {
	log.Info().
		Str("tenant_id", entry.TenantID).
		Str("user_id", entry.UserID).
		Str("operation", entry.OperationType).
		Msg("[审计日志] 准备写入操作日志...")

	logID, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("[审计日志] 生成UUID失败")
		return err
	}

	operationLog := &model.OperationLog{
		LogID:         logID,
		TenantID:      entry.TenantID,
		UserID:        entry.UserID,
		UserName:      entry.UserName,
		Module:        entry.Module,
		OperationType: entry.OperationType,
		ResourceType:  entry.ResourceType,
		ResourceID:    entry.ResourceID,
		ResourceName:  entry.ResourceName,
		RequestMethod: entry.RequestMethod,
		RequestPath:   entry.RequestPath,
		RequestParams: entry.RequestParams,
		OldValue:      toJSON(entry.OldValue),
		NewValue:      toJSON(entry.NewValue),
		Status:        entry.Status,
		ErrorMessage:  entry.ErrorMessage,
		IPAddress:     entry.IPAddress,
		UserAgent:     entry.UserAgent,
		CreatedAt:     entry.CreatedAt,
	}

	if err := w.db.WithContext(ctx).Create(operationLog).Error; err != nil {
		log.Error().Err(err).Msg("[审计日志] 写入操作日志失败")
		return err
	}

	log.Info().Str("log_id", logID).Msg("[审计日志] 操作日志写入成功")
	return nil
}

func toJSON(v any) string {
	if v == nil {
		return ""
	}
	data, _ := json.Marshal(v)
	return string(data)
}
