package loginlog

import (
	"admin/internal/repository"

	"gorm.io/gorm"
)

// Service 登录日志服务
type Service struct {
	loginLogRepo *repository.LoginLogRepo
}

// NewService 创建登录日志服务
func NewService(db *gorm.DB) *Service {
	return &Service{
		loginLogRepo: repository.NewLoginLogRepo(db),
	}
}
