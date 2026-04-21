package auth

import (
	authsvc "admin/internal/service/auth"
	"admin/pkg/audit"
	"admin/pkg/config"
	"admin/pkg/utils/jwt"
	"admin/pkg/utils/rsapwd"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Handler JWT 认证处理器
// 提供登录、选择租户、刷新、登出等接口的处理函数
type Handler struct {
	svc *authsvc.Service
}

// NewHandler 创建认证处理器
func NewHandler(db *gorm.DB, jwtMgr *jwt.Manager, rdb redis.UniversalClient, recorder *audit.Recorder, rsaCipher *rsapwd.RSACipher, cfg *config.Config) *Handler {
	return &Handler{svc: authsvc.NewService(db, jwtMgr, rdb, recorder, rsaCipher, cfg)}
}
