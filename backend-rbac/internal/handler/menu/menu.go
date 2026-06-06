package menu

import (
	"admin/internal/rbac"
	menusvc "admin/internal/service/menu"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Handler 菜单处理器
type Handler struct {
	svc *menusvc.Service
}

// NewHandler 创建菜单处理器
func NewHandler(db *gorm.DB, recorder *audit.Recorder, cache *rbac.PermissionCache) *Handler {
	return &Handler{
		svc: menusvc.NewService(db, recorder, cache),
	}
}
