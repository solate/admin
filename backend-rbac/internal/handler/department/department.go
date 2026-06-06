package department

import (
	departmentsvc "admin/internal/service/department"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Handler 部门处理器
type Handler struct {
	svc *departmentsvc.Service
}

// NewHandler 创建部门处理器
func NewHandler(db *gorm.DB, recorder *audit.Recorder) *Handler {
	return &Handler{svc: departmentsvc.NewService(db, recorder)}
}
