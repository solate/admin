package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func New(cfg Config, log *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormLogger(log),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// gormSlogger 将 GORM SQL 日志桥接到 slog。
// 全程用 *Context 变体，把 ctx 透传给 slog；请求级字段（request_id 等）
// 由后续 xcontext/中间件阶段注入，本适配器不关心具体字段。
type gormSlogger struct {
	log      *slog.Logger
	logLevel gormlogger.LogLevel
}

func newGormLogger(log *slog.Logger) gormlogger.Interface {
	return &gormSlogger{log: log, logLevel: gormlogger.Info}
}

func (l *gormSlogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &gormSlogger{log: l.log, logLevel: level}
}

func (l *gormSlogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.log.InfoContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *gormSlogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.log.WarnContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *gormSlogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.log.ErrorContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *gormSlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	attrs := []slog.Attr{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}
	if err != nil {
		l.log.LogAttrs(ctx, slog.LevelError, "sql error", append(attrs, slog.Any("err", err))...)
		return
	}
	l.log.LogAttrs(ctx, slog.LevelDebug, "sql", attrs...)
}
