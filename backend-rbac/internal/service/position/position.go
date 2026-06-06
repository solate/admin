package position

import (
	"admin/internal/repository"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Service 岗位服务
type Service struct {
	positionRepo *repository.PositionRepo
	recorder     *audit.Recorder
}

// NewService 创建岗位服务
func NewService(db *gorm.DB, recorder *audit.Recorder) *Service {
	return &Service{
		positionRepo: repository.NewPositionRepo(db),
		recorder:     recorder,
	}
}
