package database

import (
	"admin/internal/dal/query"
	"context"

	"gorm.io/gorm"
)

// Tx 事务上下文，持有事务中的 DB 实例
type Tx struct {
	DB *gorm.DB
	Q  *query.Query // GORM Gen 查询实例
}

// InTransaction 在事务中执行函数
//
// 用法示例：
//
//	database.InTransaction(db, func(tx *database.Tx) error {
//	    // 使用 tx.DB 或 tx.Q 进行业务操作
//	    playlistRepo := repository.NewDevicePlaylistRepo(tx.DB)
//	    deviceRepo := repository.NewDeviceRepo(tx.DB)
//	    // ...
//	    return nil
//	})
func InTransaction(db *gorm.DB, fc func(tx *Tx) error) error {
	return db.Transaction(func(txDB *gorm.DB) error {
		return fc(&Tx{
			DB: txDB,
			Q:  query.Use(txDB),
		})
	})
}

// InTransactionWithCtx 在事务中执行函数，携带 context
func InTransactionWithCtx(ctx context.Context, db *gorm.DB, fc func(ctx context.Context, tx *Tx) error) error {
	return db.WithContext(ctx).Transaction(func(txDB *gorm.DB) error {
		return fc(ctx, &Tx{
			DB: txDB,
			Q:  query.Use(txDB),
		})
	})
}
