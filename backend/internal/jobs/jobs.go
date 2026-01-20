package jobs

import (
	"admin/pkg/xcron"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Init 初始化并注册所有定时任务
func Init(cronMgr *xcron.Manager, db *gorm.DB) error {
	// 测试任务 - 每5秒执行一次
	if err := cronMgr.Add("test_job", "*/5 * * * * ?", testJob); err != nil {
		return err
	}

	log.Info().Msg("定时任务注册完成")
	return nil
}

// ============================================
// 定时任务实现
// ============================================

// testJob 测试任务 - 每5秒执行
func testJob() {
	log.Info().Msg("🕐 定时任务测试：每5秒执行一次")
}

// cleanupLogs 清理过期日志
func cleanupLogs() {
	log.Info().Msg("开始清理过期日志...")
	// TODO: 实现你的业务逻辑
	// 例如：删除 30 天前的日志
	log.Info().Msg("清理过期日志完成")
}

// backupDatabase 数据库备份
func backupDatabase() {
	log.Info().Msg("开始数据库备份...")
	// TODO: 实现你的业务逻辑
	// 例如：导出数据库到文件
	log.Info().Msg("数据库备份完成")
}

// syncData 同步数据
func syncData() {
	log.Info().Msg("开始同步数据...")
	// TODO: 实现你的业务逻辑
	// 例如：从外部API同步数据
	log.Info().Msg("数据同步完成")
}

// ============================================
// 你可以在这里添加更多任务
// ============================================

// exampleTask 示例任务 - 每5分钟执行
func exampleTask() {
	log.Info().Msg("执行示例任务...")
	// TODO: 你的业务逻辑
}
