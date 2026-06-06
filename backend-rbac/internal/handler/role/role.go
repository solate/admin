package role

import (
	"admin/internal/rbac"
	rolesvc "admin/internal/service/role"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Handler 角色处理器
type Handler struct {
	svc *rolesvc.Service
}

// NewHandler 创建角色处理器
func NewHandler(db *gorm.DB, recorder *audit.Recorder, cache *rbac.PermissionCache) *Handler {
	return &Handler{
		svc: rolesvc.NewService(db, recorder, cache),
	}
}
