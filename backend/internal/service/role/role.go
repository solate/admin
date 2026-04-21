package role

import (
	"admin/internal/rbac"
	"admin/internal/repository"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Service 角色服务
type Service struct {
	roleRepo       *repository.RoleRepo
	permissionRepo *repository.PermissionRepo
	menuRepo       *repository.MenuRepo
	rolePermRepo   *repository.RolePermissionRepo
	userRoleRepo   *repository.UserRoleRepo
	cache          *rbac.PermissionCache
	tenantRepo     *repository.TenantRepo
	recorder       *audit.Recorder
}

// NewService 创建角色服务
func NewService(db *gorm.DB, recorder *audit.Recorder, cache *rbac.PermissionCache) *Service {
	return &Service{
		roleRepo:       repository.NewRoleRepo(db),
		permissionRepo: repository.NewPermissionRepo(db),
		menuRepo:       repository.NewMenuRepo(db),
		rolePermRepo:   repository.NewRolePermissionRepo(db),
		userRoleRepo:   repository.NewUserRoleRepo(db),
		cache:          cache,
		tenantRepo:     repository.NewTenantRepo(db),
		recorder:       recorder,
	}
}
