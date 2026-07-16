// Package xcron 是 robfig/cron v3 的极简封装，仅提供：
//   - New()：带项目预设（秒级 cron + panic 恢复 + slog 日志）的构造工厂
//   - Run()：阻塞式启动，优雅停止与 server 生命周期集成
//
// 不封装任务管理（AddJob/Remove/List）—— 直接用 robfig/cron 原生 *cron.Cron。
// 适用于单副本场景（无分布式锁，内存调度）。
//
// 用法：
//
//	c := xcron.New()
//	c.AddFunc("0 0 2 * * *", func() { /* 凌晨2点执行 */ })
//	return xcron.Run(ctx, c) // 阻塞，ctx 取消时优雅停止
package xcron

import (
	"context"
	"log/slog"

	"github.com/robfig/cron/v3"
)

// New 创建原生 cron.Cron（6 字段秒级 cron 表达式 + panic 恢复 + 日志）。
// 返回原生 *cron.Cron，调用方直接用 AddFunc/AddJob/Remove/Entries 等原生方法。
func New() *cron.Cron {
	return cron.New(
		cron.WithSeconds(),                          // 6 字段：秒 分 时 日 月 周
		cron.WithLogger(slogAdapter{}),              // 日志适配
		cron.WithChain(cron.Recover(slogAdapter{})), // panic 恢复
	)
}

// Run 阻塞式启动调度器。ctx 取消时自动优雅停止，等待运行中任务完成后返回。
// 通常在 errgroup goroutine 中调用：
//
//	g.Go(func() error { return xcron.Run(ctx, c) })
func Run(ctx context.Context, c *cron.Cron) error {
	c.Start()
	slog.Info("cron scheduler started", "jobs", len(c.Entries()))

	<-ctx.Done() // 等待 server lifecycle ctx 取消（SIGTERM 等）

	slog.Info("cron scheduler stopping")
	<-c.Stop().Done() // 等待运行中任务完成
	slog.Info("cron scheduler stopped")

	return nil
}

// slogAdapter 把 robfig/cron 的日志调用适配到全局 slog
type slogAdapter struct{}

func (slogAdapter) Info(msg string, keysAndValues ...interface{}) {
	slog.Info(msg, keysAndValues...)
}

func (slogAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	attrs := make([]any, 0, len(keysAndValues)+2)
	attrs = append(attrs, "error", err)
	attrs = append(attrs, keysAndValues...)
	slog.Error(msg, attrs...)
}
