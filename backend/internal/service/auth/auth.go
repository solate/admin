package auth

import (
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/config"
	"admin/pkg/utils/jwt"
	"admin/pkg/utils/rsapwd"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Service 认证服务
type Service struct {
	userRepo     *repository.UserRepo
	userRoleRepo *repository.UserRoleRepo
	roleRepo     *repository.RoleRepo
	tenantRepo   *repository.TenantRepo
	jwt          *jwt.Manager
	rdb          redis.UniversalClient
	recorder     *audit.Recorder
	config       *config.Config
	rsaCipher    *rsapwd.RSACipher
}

// NewService 创建认证服务
func NewService(db *gorm.DB, jwtMgr *jwt.Manager, rdb redis.UniversalClient, recorder *audit.Recorder, rsaCipher *rsapwd.RSACipher, cfg *config.Config) *Service {
	return &Service{
		userRepo:     repository.NewUserRepo(db),
		userRoleRepo: repository.NewUserRoleRepo(db),
		roleRepo:     repository.NewRoleRepo(db),
		tenantRepo:   repository.NewTenantRepo(db),
		jwt:          jwtMgr,
		rdb:          rdb,
		recorder:     recorder,
		config:       cfg,
		rsaCipher:    rsaCipher,
	}
}
