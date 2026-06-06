package router

import (
	"admin/internal/middleware"
	"admin/internal/rbac"
	"admin/pkg/audit"
	"admin/pkg/config"
	"admin/pkg/utils/jwt"
	filesystem "admin/static"
	"fmt"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 设置路由
func Setup(r *gin.Engine, handlers *Handlers, cfg *config.Config, jwtMgr *jwt.Manager, rbacCache *rbac.PermissionCache) {

	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())

	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimitMiddleware(
			cfg.RateLimit.RequestsPerSecond,
			cfg.RateLimit.Burst,
		))
	}

	r.GET("/health", handlers.HealthHandler.Check)
	r.GET("/ping", handlers.HealthHandler.Ping)

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 静态文件服务 - uploads 目录（视频文件等）
	r.Static("/uploads", "./uploads")
	// 设置嵌入的前端静态文件
	setupEmbedFrontend(r)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{

		// 登录/注册相关接口
		authGroup := v1.Group("/auth")
		{
			authGroup.GET("/captcha", handlers.CaptchaHandler.Get)
			authGroup.POST("/login", audit.AuditMiddleware(), handlers.AuthHandler.Login)
			authGroup.POST("/refresh", handlers.AuthHandler.Refresh)
		}

		// 需要认证 + RBAC 权限检查的路由
		authorized := v1.Group("")
		authorized.Use(middleware.AuthMiddleware(jwtMgr))
		authorized.Use(middleware.RBACMiddleware(rbacCache))
		authorized.Use(audit.AuditMiddleware())
		{
			// 用户个人资料相关
			userSelf := authorized.Group("/user")
			{
				userSelf.GET("/profile", handlers.UserHandler.GetProfile)
				userSelf.POST("/password/change", handlers.UserHandler.ChangePassword)
				userSelf.GET("/menus", handlers.UserHandler.GetUserMenu)
				userSelf.GET("/buttons", handlers.UserHandler.GetUserButtons)
			}

			// 认证接口
			authGroup := authorized.Group("/auth")
			{
				authGroup.POST("/logout", handlers.AuthHandler.Logout)
				authGroup.POST("/switch-tenant", handlers.AuthHandler.SwitchTenant)
				authGroup.GET("/available-tenants", handlers.AuthHandler.GetAvailableTenants)
			}

			// 租户管理
			tenant := authorized.Group("/tenants")
			{
				tenant.POST("", handlers.TenantHandler.CreateTenant)
				tenant.GET("", handlers.TenantHandler.ListTenants)
				tenant.GET("/all", handlers.TenantHandler.ListAllTenants)
				tenant.GET("/detail", handlers.TenantHandler.GetTenant)
				tenant.PUT("", handlers.TenantHandler.UpdateTenant)
				tenant.DELETE("", handlers.TenantHandler.DeleteTenant)
				tenant.DELETE("/batch-delete", handlers.TenantHandler.BatchDeleteTenants)
				tenant.PUT("/status", handlers.TenantHandler.UpdateTenantStatus)
			}

			// 用户管理
			userGroup := authorized.Group("/users")
			{
				userGroup.POST("", handlers.UserHandler.CreateUser)
				userGroup.GET("", handlers.UserHandler.ListUsers)
				userGroup.GET("/detail", handlers.UserHandler.GetUser)
				userGroup.PUT("", handlers.UserHandler.UpdateUser)
				userGroup.DELETE("", handlers.UserHandler.DeleteUser)
				userGroup.DELETE("/batch-delete", handlers.UserHandler.BatchDeleteUsers)
				userGroup.PUT("/status", handlers.UserHandler.UpdateUserStatus)
				userGroup.GET("/roles", handlers.UserHandler.GetUserRoles)
				userGroup.PUT("/roles", handlers.UserHandler.AssignRoles)
				userGroup.POST("/password/reset", handlers.UserHandler.ResetPassword)
			}

			// 角色管理
			roleGroup := authorized.Group("/roles")
			{
				roleGroup.POST("", handlers.RoleHandler.CreateRole)
				roleGroup.GET("", handlers.RoleHandler.ListRoles)
				roleGroup.GET("/all", handlers.RoleHandler.GetAllRoles)
				roleGroup.GET("/detail", handlers.RoleHandler.GetRole)
				roleGroup.PUT("", handlers.RoleHandler.UpdateRole)
				roleGroup.DELETE("", handlers.RoleHandler.DeleteRole)
				roleGroup.PUT("/status", handlers.RoleHandler.UpdateRoleStatus)
				roleGroup.PUT("/permissions", handlers.RoleHandler.AssignPermissions)
				roleGroup.GET("/permissions", handlers.RoleHandler.GetRolePermissions)
			}

			// 菜单接口
			menuGroup := authorized.Group("/menus")
			{
				menuGroup.POST("", handlers.MenuHandler.CreateMenu)
				menuGroup.GET("", handlers.MenuHandler.ListMenus)
				menuGroup.GET("/all", handlers.MenuHandler.GetAllMenus)
				menuGroup.GET("/tree", handlers.MenuHandler.GetMenuTree)
				menuGroup.GET("/detail", handlers.MenuHandler.GetMenu)
				menuGroup.PUT("", handlers.MenuHandler.UpdateMenu)
				menuGroup.DELETE("", handlers.MenuHandler.DeleteMenu)
				menuGroup.PUT("/status", handlers.MenuHandler.UpdateMenuStatus)
			}

			// 部门管理
			dept := authorized.Group("/departments")
			{
				dept.POST("", handlers.DepartmentHandler.CreateDepartment)
				dept.GET("", handlers.DepartmentHandler.ListDepartments)
				dept.GET("/tree", handlers.DepartmentHandler.GetDepartmentTree)
				dept.GET("/detail", handlers.DepartmentHandler.GetDepartment)
				dept.PUT("", handlers.DepartmentHandler.UpdateDepartment)
				dept.DELETE("", handlers.DepartmentHandler.DeleteDepartment)
				dept.PUT("/status", handlers.DepartmentHandler.UpdateDepartmentStatus)
				dept.GET("/children", handlers.DepartmentHandler.GetChildren)
			}

			// 岗位管理
			position := authorized.Group("/positions")
			{
				position.POST("", handlers.PositionHandler.CreatePosition)
				position.GET("", handlers.PositionHandler.ListPositions)
				position.GET("/all", handlers.PositionHandler.ListAllPositions)
				position.GET("/detail", handlers.PositionHandler.GetPosition)
				position.PUT("", handlers.PositionHandler.UpdatePosition)
				position.DELETE("", handlers.PositionHandler.DeletePosition)
				position.PUT("/status", handlers.PositionHandler.UpdatePositionStatus)
			}

			// 字典管理接口
			dicts := authorized.Group("/dicts")
			{
				dicts.GET("", handlers.DictHandler.GetDict)
				dicts.GET("/types", handlers.DictHandler.ListDictTypes)
				dicts.PUT("/items", handlers.DictHandler.BatchUpdateDictItems)
				dicts.DELETE("/items", handlers.DictHandler.ResetDictItem)
			}
			systemDict := authorized.Group("/system/dicts")
			{
				systemDict.POST("", handlers.DictHandler.CreateSystemDict)
				systemDict.GET("", handlers.DictHandler.ListSystemDictTypes)
				systemDict.PUT("", handlers.DictHandler.UpdateSystemDict)
				systemDict.DELETE("", handlers.DictHandler.DeleteSystemDict)
				systemDict.DELETE("/items", handlers.DictHandler.DeleteSystemDictItem)
			}

			// 审计日志接口
			logs := authorized.Group("/logs")
			{
				logs.GET("/login", handlers.LoginLogHandler.ListLoginLogs)
				logs.GET("/login/detail", handlers.LoginLogHandler.GetLoginLog)
				logs.GET("/operation", handlers.OperationLogHandler.ListOperationLogs)
				logs.GET("/operation/detail", handlers.OperationLogHandler.GetOperationLog)
			}

		}
	}
}

// setupEmbedFrontend 设置嵌入的前端静态文件
func setupEmbedFrontend(r *gin.Engine) {
	frontendFS, err := static.EmbedFolder(filesystem.FrontEnd, "frontend")
	if err != nil {
		panic(err)
	}
	r.Use(static.Serve("/", frontendFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		fmt.Printf("[NoRoute] %s doesn't exist, serving index.html\n", path)
		c.FileFromFS("/index.html", frontendFS)
	})
}
