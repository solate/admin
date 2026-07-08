package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// defaultSlowThreshold 慢查询阈值，对齐 GORM 原生默认（200ms）。
// 超过此耗时的 SQL 记为 Warn，便于在日志里区分慢查询。
const defaultSlowThreshold = 200 * time.Millisecond

// gormSlogger 将 GORM 的 SQL 日志桥接到项目的 slog。
// 行为对齐 GORM 原生 logger：正常 SQL→Info、慢查询→Warn、出错→Error；
// 忽略 ErrRecordNotFound（正常业务未命中，不当错误刷屏）。
//
// 级别门禁不在本适配器做——交给 slog handler（已按 cfg.Log.Level 过滤）。
// 全程用 *Context 变体，把 ctx 透传给 slog；请求级字段（request_id 等）
// 由后续 xcontext/中间件阶段注入，本适配器不关心具体字段。
type gormSlogger struct {
	log           *slog.Logger
	slowThreshold time.Duration
}

func newGormLogger(log *slog.Logger) gormlogger.Interface {
	return &gormSlogger{log: log, slowThreshold: defaultSlowThreshold}
}

// LogMode 接口要求实现；级别门禁由 slog 统一管，这里直接返回自身。
func (l *gormSlogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
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

// Trace 每条 SQL 执行后由 GORM 回调，按原生三分支语义分级输出到 slog。
func (l *gormSlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	attrs := []slog.Attr{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}
	switch {
	// 出错 → Error；忽略 RecordNotFound（正常业务未命中，不当错误刷屏）
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.log.LogAttrs(ctx, slog.LevelError, "sql error", append(attrs, slog.Any("err", err))...)
	// 慢查询 → Warn
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		l.log.LogAttrs(ctx, slog.LevelWarn, "slow sql", append(attrs, slog.Duration("threshold", l.slowThreshold))...)
	// 正常 → Info（不再是 Debug，生产 info level 即可见）
	default:
		l.log.LogAttrs(ctx, slog.LevelInfo, "sql", attrs...)
	}
}
