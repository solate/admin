package router

import (
	"admin/internal/handler"
	"admin/internal/middleware"
	"admin/internal/service"
	"admin/pkg/config"
	"admin/pkg/jwt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Config 路由配置
type Config struct {
	HealthHandler *handler.HealthHandler
	AuthHandler   *handler.AuthHandler
	PolicyHandler *handler.PolicyHandler
	AppConfig     *config.Config
	JWTManager    *jwt.JWTManager
	CasbinService *service.CasbinService
}

// Setup 设置路由
func Setup(r *gin.Engine, cfg *Config) {

	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.TenantContext())

	// if cfg.AppConfig.RateLimit.Enabled {
	//     r.Use(middleware.RateLimit(
	//         cfg.AppConfig.RateLimit.RequestsPerSecond,
	//         cfg.AppConfig.RateLimit.Burst,
	//     ))
	// }

	r.GET("/health", cfg.HealthHandler.Check)
	r.GET("/ping", cfg.HealthHandler.Ping)

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{

		// 公开接口（无需认证）
		public := v1.Group("/public")
		{

			// 设备查询证书公钥
			public.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})
		}

		auth := v1.Group("/auth")
		{
			// 验证码（无需认证）
			v1.GET("/captcha", func(c *gin.Context) {
				c.JSON(200, gin.H{"captcha": "stub"})
			})
			v1.POST("/captcha/verify", func(c *gin.Context) {
				c.JSON(200, gin.H{"verified": true})
			})

			auth.POST("/register", func(c *gin.Context) {
				c.JSON(200, gin.H{"registered": true})
			})
			auth.POST("/login", cfg.AuthHandler.Login)
		}

		// 需要认证的路由
		authenticated := v1.Group("")
		authenticated.Use(middleware.Auth(cfg.JWTManager))
		authenticated.Use(middleware.CasbinMiddleware(cfg.CasbinService.Enforcer))
		{
			// 用户接口
			user := authenticated.Group("/users")
			{
				user.GET("/info", func(c *gin.Context) {
					c.JSON(200, gin.H{"info": "user"})
				})
			}

			policy := authenticated.Group("/policy")
			{
				policy.POST("/add", cfg.PolicyHandler.AddPolicy)
				policy.POST("/role/add", cfg.PolicyHandler.AddRole)
			}
		}

	}
}
