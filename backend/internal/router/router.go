// Package router 配置 gin 路由与中间件。
package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"admin/internal/config"
	"admin/internal/middleware"
	"admin/pkg/response"
)

// Setup 配置 gin engine：注册中间件、健康检查路由。
// 中间件顺序：requestid（最先，下游全依赖）→ logger → recovery → cors。
func Setup(r *gin.Engine, cfg *config.Config, log *slog.Logger) {
	// 中间件（顺序关键：requestid 最先，其余可读取 request_id）
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS(cfg.Server.Cors))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// 业务路由组（Step 04+ 追加）
	// api := r.Group("/api/v1")
	// api.Use(middleware.JWT(...))
	// api.POST("/login", handler.Login)
	// ...
}
