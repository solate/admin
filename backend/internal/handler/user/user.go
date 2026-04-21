package user

import (
	"admin/internal/rbac"
	usersvc "admin/internal/service/user"
	"admin/pkg/audit"
	"admin/pkg/utils/rsapwd"

	"gorm.io/gorm"
)

// Handler 用户处理器
type Handler struct {
	svc     *usersvc.Service
	roleSvc *usersvc.RoleService
	menuSvc *usersvc.MenuService
}

// NewHandler 创建用户处理器
func NewHandler(db *gorm.DB, recorder *audit.Recorder, rsaCipher *rsapwd.RSACipher, cache *rbac.PermissionCache) *Handler {
	return &Handler{
		svc:     usersvc.NewService(db, recorder, rsaCipher),
		roleSvc: usersvc.NewRoleService(db, recorder),
		menuSvc: usersvc.NewMenuService(db, cache),
	}
}
