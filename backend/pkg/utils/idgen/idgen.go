package idgen

import "github.com/google/uuid"

// NewV7String 生成一个 UUIDv7 字符串（时间有序、做主键索引友好）。
//
// 备用封装：主键默认走数据库端 uuid 列的 DEFAULT uuidv7()（PG18 库端自动生成），
// 插入时无需传 ID；仅当需要在应用层预生成 ID（如批量种子、跨表外键预填）时才调用本方法。
//
// 无 error 返回：uuid.NewV7 仅在读系统随机源（crypto/rand）失败时返回 error，
// 正常运行的机器上几乎不可能发生，故用官方 uuid.Must 做 fail-fast。
// 相比原 sonyflake 方案，这里没有机器 ID 协调 / 时钟回拨 / init panic 这些概念。
func NewV7String() string {
	return uuid.Must(uuid.NewV7()).String()
}
