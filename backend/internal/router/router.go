package router

import (
	"admin/internal/middleware"
	filesystem "admin/static"
	"fmt"
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 设置路由
func Setup(r *gin.Engine, app *App) {

	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())

	if app.Config.RateLimit.Enabled {
		r.Use(middleware.RateLimitMiddleware(
			app.Config.RateLimit.RequestsPerSecond,
			app.Config.RateLimit.Burst,
		))
	}

	r.GET("/health", app.Handlers.HealthHandler.Check)
	r.GET("/ping", app.Handlers.HealthHandler.Ping)

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{

		// 公开接口（无需认证）
		public := v1.Group("")
		{
			// 登录/注册相关接口 - 需要跳过租户检查
			// 说明：登录时需要跨租户查询用户关联的租户信息，因此使用 SkipTenantCheck 中间件
			auth := public.Group("/auth")
			auth.Use(middleware.SkipTenantCheck())
			auth.Use(middleware.OperationLogMiddleware(app.OperationLogLogger))
			{
				auth.GET("/captcha", app.Handlers.CaptchaHandler.Get)   // 获取验证码
				auth.POST("/login", app.Handlers.AuthHandler.Login)     // 用户登录
				auth.POST("/refresh", app.Handlers.AuthHandler.Refresh) // 刷新令牌
				// auth.POST("/register", func(c *gin.Context) {
				// 	c.JSON(200, gin.H{"registered": true})
				// }) // 用户注册
			}

		}

		// 需要认证的路由
		authenticated := v1.Group("")
		authenticated.Use(middleware.AuthMiddleware(app.JWT))
		authenticated.Use(middleware.CasbinMiddleware(app.Enforcer))
		authenticated.Use(middleware.OperationLogMiddleware(app.OperationLogLogger))
		{

			// 用户接口
			user := authenticated.Group("/users")
			{
				user.POST("", app.Handlers.UserHandler.CreateUser)                              // 创建用户
				user.GET("", app.Handlers.UserHandler.ListUsers)                                // 获取用户列表
				user.GET("/:user_id", app.Handlers.UserHandler.GetUser)                         // 获取用户详情
				user.PUT("/:user_id", app.Handlers.UserHandler.UpdateUser)                      // 更新用户
				user.DELETE("/:user_id", app.Handlers.UserHandler.DeleteUser)                   // 删除用户
				user.PUT("/:user_id/status/:status", app.Handlers.UserHandler.UpdateUserStatus) // 更新用户状态

			}

			// 租户接口
			tenant := authenticated.Group("/tenants")
			{
				tenant.POST("", app.Handlers.TenantHandler.CreateTenant)                              // 创建租户
				tenant.GET("", app.Handlers.TenantHandler.ListTenants)                                // 获取租户列表
				tenant.GET("/:tenant_id", app.Handlers.TenantHandler.GetTenant)                       // 获取租户详情
				tenant.PUT("/:tenant_id", app.Handlers.TenantHandler.UpdateTenant)                    // 更新租户
				tenant.DELETE("/:tenant_id", app.Handlers.TenantHandler.DeleteTenant)                // 删除租户
				tenant.PUT("/:tenant_id/status/:status", app.Handlers.TenantHandler.UpdateTenantStatus) // 更新租户状态
			}

		}

	}

	// 设置嵌入的前端静态文件
	setupEmbedFrontend(r)

}

// setupEmbedFrontend 设置嵌入的前端静态文件
func setupEmbedFrontend(r *gin.Engine) {
	// 静态文件服务 - 前端资源（使用 embed.FS）
	// 使用 gin-contrib/static 提供嵌入的前端静态文件服务
	frontendFS, err := static.EmbedFolder(filesystem.FrontEnd, "frontend")
	if err != nil {
		panic(err)
	}
	r.Use(static.Serve("/", frontendFS))

	// 处理 SPA 前端路由（NoRoute 必须在最后注册）
	r.NoRoute(func(c *gin.Context) {
		// path := c.Request.URL.Path
		// API 路由返回 404 JSON
		// if strings.HasPrefix(path, "/api") {
		// 	c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
		// 	return
		// }
		// 其他路由返回前端 index.html（SPA 路由）
		// c.FileFromFS("/index.html", frontendFS)
		fmt.Printf("%s doesn't exists, redirect on /\n", c.Request.URL.Path)
		c.Redirect(http.StatusMovedPermanently, "/")
	})

}
