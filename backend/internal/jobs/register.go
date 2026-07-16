// Package jobs 集中注册所有静态 cron 任务（启动时固定的）。
// 动态任务（运行时 Add/Remove）通过 HTTP API 操作 scheduler，不在这里。
package jobs

import (
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Deps 任务函数依赖的 service/cache 等组件。
type Deps struct {
	DB  *gorm.DB
	RDB *redis.Client
	// 后续按需扩充，例如：
	// OperationLogSvc *operationlog.Service
	// RBACCache       *rbac.PermissionCache
}

// Register 注册所有静态任务到 scheduler（在 Start 之前调用）。
func Register(s *cron.Cron, deps Deps) {
	// 示例（后续根据实际需求填充）：
	// s.AddFunc("0 0 2 * * *", func() {
	//     deps.OperationLogSvc.CleanOldLogs(context.Background(), 30)
	// })
}
