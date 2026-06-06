package user

import (
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/utils/rsapwd"

	"gorm.io/gorm"
)

// Service 用户服务
type Service struct {
	userRepo        *repository.UserRepo
	userRoleRepo    *repository.UserRoleRepo
	userRoleService *RoleService
	roleRepo        *repository.RoleRepo
	tenantRepo      *repository.TenantRepo
	recorder        *audit.Recorder
	rsaCipher       *rsapwd.RSACipher
}

// NewService 创建用户服务
func NewService(db *gorm.DB, recorder *audit.Recorder, rsaCipher *rsapwd.RSACipher) *Service {
	roleSvc := NewRoleService(db, recorder)
	return &Service{
		userRepo:        repository.NewUserRepo(db),
		userRoleRepo:    repository.NewUserRoleRepo(db),
		userRoleService: roleSvc,
		roleRepo:        repository.NewRoleRepo(db),
		tenantRepo:      repository.NewTenantRepo(db),
		recorder:        recorder,
		rsaCipher:       rsaCipher,
	}
}
