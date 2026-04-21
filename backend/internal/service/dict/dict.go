package dict

import (
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/cache"

	"gorm.io/gorm"
)

// Service 字典服务
type Service struct {
	dictTypeRepo *repository.DictTypeRepo
	dictItemRepo *repository.DictItemRepo
	tenantCache  *cache.TenantCache
	recorder     *audit.Recorder
}

// NewService 创建字典服务
func NewService(db *gorm.DB, recorder *audit.Recorder) *Service {
	return &Service{
		dictTypeRepo: repository.NewDictTypeRepo(db),
		dictItemRepo: repository.NewDictItemRepo(db),
		tenantCache:  cache.Get().Tenant,
		recorder:     recorder,
	}
}
