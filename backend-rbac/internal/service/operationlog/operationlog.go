package operationlog

import (
	"admin/internal/repository"

	"gorm.io/gorm"
)

// Service 操作日志服务
type Service struct {
	operationLogRepo *repository.OperationLogRepo
}

// NewService 创建操作日志服务
func NewService(db *gorm.DB) *Service {
	return &Service{
		operationLogRepo: repository.NewOperationLogRepo(db),
	}
}
