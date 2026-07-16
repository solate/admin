package xgorm

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config 数据库连接参数（纯连接配置，不含日志字段）。
// 由 internal/config.DatabaseConfig 逐字段映射而来（见 cmd/server/main.go）。
// SQL 日志的可见性由 slog 的 cfg.Log.Level 统一控制，不在此单独配置。
type Config struct {
	Host            string // 数据库主机地址
	Port            int    // 端口
	User            string // 用户名
	Password        string // 密码
	DBName          string // 库名
	SSLMode         string // SSL 模式：disable / require / verify-full 等
	MaxIdleConns    int    // 连接池最大空闲连接数
	MaxOpenConns    int    // 连接池最大打开连接数
	ConnMaxLifetime int    // 连接最大存活时长（秒）
}

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
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
