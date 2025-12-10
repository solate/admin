package router

import (
	"admin/internal/handler"
	"admin/internal/middleware"
	"admin/pkg/config"

	"github.com/gin-gonic/gin"
)

// Config 路由配置
type Config struct {
	HealthHandler *handler.HealthHandler
	AppConfig     *config.Config
}

// Setup 设置路由
func Setup(r *gin.Engine, cfg *Config) {

	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// if cfg.AppConfig.RateLimit.Enabled {
	//     r.Use(middleware.RateLimit(
	//         cfg.AppConfig.RateLimit.RequestsPerSecond,
	//         cfg.AppConfig.RateLimit.Burst,
	//     ))
	// }

	r.GET("/health", cfg.HealthHandler.Check)
	r.GET("/ping", cfg.HealthHandler.Ping)

}
