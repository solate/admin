package menu

import (
	"admin/internal/repository"
	"admin/internal/rbac"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Service 菜单服务
type Service struct {
	menuRepo     *repository.MenuRepo
	rolePermRepo *repository.RolePermissionRepo
	cache        *rbac.PermissionCache
	recorder     *audit.Recorder
}

// NewService 创建菜单服务
func NewService(db *gorm.DB, recorder *audit.Recorder, cache *rbac.PermissionCache) *Service {
	return &Service{
		menuRepo:     repository.NewMenuRepo(db),
		rolePermRepo: repository.NewRolePermissionRepo(db),
		cache:        cache,
		recorder:     recorder,
	}
}
