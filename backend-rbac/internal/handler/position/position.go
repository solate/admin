package position

import (
	positionsvc "admin/internal/service/position"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Handler 岗位处理器
type Handler struct {
	svc *positionsvc.Service
}

// NewHandler 创建岗位处理器
func NewHandler(db *gorm.DB, recorder *audit.Recorder) *Handler {
	return &Handler{svc: positionsvc.NewService(db, recorder)}
}
