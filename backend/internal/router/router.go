package router

import (
	"admin/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 设置路由
func Setup(r *gin.Engine, app *App) {

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

	r.GET("/health", app.Handlers.HealthHandler.Check)
	r.GET("/ping", app.Handlers.HealthHandler.Ping)

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
			auth.POST("/login", app.Handlers.AuthHandler.Login)
		}

		// 需要认证的路由
		authenticated := v1.Group("")
		authenticated.Use(middleware.Auth(app.JWT))
		authenticated.Use(middleware.CasbinMiddleware(app.Enforcer))
		{
			// 用户接口
			user := authenticated.Group("/users")
			{
				user.GET("/info", func(c *gin.Context) {
					c.JSON(200, gin.H{"info": "user"})
				})
			}

			// policy := authenticated.Group("/policy")
			// {
			// 	policy.POST("/add", app.Handlers.PolicyHandler.AddPolicy)
			// 	policy.POST("/role/add", app.Handlers.PolicyHandler.AddRole)
			// }

			// 超级管理员专属接口
			super := authenticated.Group("/super")
			super.Use(middleware.SuperAdmin())
			{
				// 租户管理接口
				tenant := super.Group("/tenants")
				{
					tenant.POST("", app.Handlers.TenantHandler.CreateTenant)                        // 创建租户
					tenant.GET("", app.Handlers.TenantHandler.ListTenants)                          // 获取租户列表
					tenant.GET("/:tenant_id", app.Handlers.TenantHandler.GetTenant)                 // 获取租户详情
					tenant.PUT("/:tenant_id", app.Handlers.TenantHandler.UpdateTenant)              // 更新租户
					tenant.DELETE("/:tenant_id", app.Handlers.TenantHandler.DeleteTenant)           // 删除租户
					tenant.PUT("/:tenant_id/status", app.Handlers.TenantHandler.UpdateTenantStatus) // 更新租户状态
				}
			}
		}

	}
}
