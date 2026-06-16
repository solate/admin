package database

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func New(cfg Config, log zerolog.Logger) (*gorm.DB, error) {
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

// gormZerologger 将 GORM SQL 日志桥接到 zerolog
type gormZerologger struct {
	log      zerolog.Logger
	logLevel gormlogger.LogLevel
}

func newGormLogger(log zerolog.Logger) gormlogger.Interface {
	return &gormZerologger{
		log:      log,
		logLevel: gormlogger.Info,
	}
}

func (l *gormZerologger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &gormZerologger{log: l.log, logLevel: level}
}

func (l *gormZerologger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.log.Info().Msgf(msg, args...)
}

func (l *gormZerologger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.log.Warn().Msgf(msg, args...)
}

func (l *gormZerologger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.log.Error().Msgf(msg, args...)
}

func (l *gormZerologger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	ev := l.log.Debug().Dur("elapsed", elapsed).Int64("rows", rows).Str("sql", sql)
	if err != nil {
		ev = l.log.Error().Err(err).Dur("elapsed", elapsed).Int64("rows", rows).Str("sql", sql)
	}
	ev.Msg("sql")
}
