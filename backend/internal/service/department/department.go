package department

import (
	"admin/internal/repository"
	"admin/pkg/audit"

	"gorm.io/gorm"
)

// Service 部门服务
type Service struct {
	deptRepo *repository.DepartmentRepo
	userRepo *repository.UserRepo
	recorder *audit.Recorder
}

// NewService 创建部门服务
func NewService(db *gorm.DB, recorder *audit.Recorder) *Service {
	return &Service{
		deptRepo: repository.NewDepartmentRepo(db),
		userRepo: repository.NewUserRepo(db),
		recorder: recorder,
	}
}
