package tenant

import (
	"admin/internal/repository"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Service 租户服务
type Service struct {
	tenantRepo *repository.TenantRepo
	userRepo   *repository.UserRepo
	recorder   *audit.Recorder
}

// NewService 创建租户服务
func NewService(db *gorm.DB, recorder *audit.Recorder) *Service {
	return &Service{
		tenantRepo: repository.NewTenantRepo(db),
		userRepo:   repository.NewUserRepo(db),
		recorder:   recorder,
	}
}
