package dict

import (
	dictsvc "admin/internal/service/dict"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Handler 字典处理器
type Handler struct {
	svc *dictsvc.Service
}

// NewHandler 创建字典处理器
func NewHandler(db *gorm.DB, recorder *audit.Recorder) *Handler {
	return &Handler{
		svc: dictsvc.NewService(db, recorder),
	}
}
