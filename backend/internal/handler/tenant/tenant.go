package tenant

import (
	tenantsvc "admin/internal/service/tenant"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Handler 租户处理器
type Handler struct {
	svc *tenantsvc.Service
}

// NewHandler 创建租户处理器
func NewHandler(db *gorm.DB, recorder *audit.Recorder) *Handler {
	return &Handler{
		svc: tenantsvc.NewService(db, recorder),
	}
}
