package router

import (
	"admin/internal/middleware"
	"admin/pkg/audit"
	filesystem "admin/static"
	"fmt"

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

	// 静态文件服务 - uploads 目录（视频文件等）
	r.Static("/uploads", "./uploads")
	// 设置嵌入的前端静态文件
	setupEmbedFrontend(r)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{

		// 登录/注册相关接口
		auth := v1.Group("/auth")
		{
			auth.GET("/captcha", app.Handlers.CaptchaHandler.Get)                        // 获取验证码
			auth.POST("/login", audit.AuditMiddleware(), app.Handlers.AuthHandler.Login) // 用户登录（通过邮箱登录，需要审计日志）
			auth.POST("/refresh", app.Handlers.AuthHandler.Refresh)                      // 刷新令牌
		}

		// 需要认证 + Casbin 权限检查的路由
		authorized := v1.Group("")
		authorized.Use(middleware.AuthMiddleware(app.JWT))
		authorized.Use(middleware.CasbinMiddleware(app.Enforcer)) // 权限校验（超管跳过，其他走 Casbin）
		authorized.Use(middleware.TenantSkipMiddleware())         // 租户检查跳过（超管和审核员可查看所有租户数据）
		authorized.Use(audit.AuditMiddleware())                   // 审计日志中间件（提取请求信息）
		{
			// 用户个人资料相关（当前登录用户操作自己的数据）
			userSelf := authorized.Group("/user")
			{
				userSelf.GET("/profile", app.Handlers.UserHandler.GetProfile)              // 获取当前用户信息（含角色）
				userSelf.POST("/password/change", app.Handlers.UserHandler.ChangePassword) // 修改自己的密码
				// 菜单权限
				userSelf.GET("/menus", app.Handlers.UserMenuHandler.GetUserMenu)      // 获取用户菜单树
				userSelf.GET("/buttons", app.Handlers.UserMenuHandler.GetUserButtons) // 获取菜单按钮权限
			}

			// 认证接口（无需 Casbin 权限检查）
			auth := authorized.Group("/auth")
			{
				auth.POST("/logout", app.Handlers.AuthHandler.Logout)                        // 用户登出
				auth.POST("/switch-tenant", app.Handlers.AuthHandler.SwitchTenant)           // 切换租户
				auth.GET("/available-tenants", app.Handlers.AuthHandler.GetAvailableTenants) // 获取可切换租户（当前用户视角）

			}

			// 租户管理（仅超管）
			tenant := authorized.Group("/tenants")
			{
				tenant.POST("", app.Handlers.TenantHandler.CreateTenant)                      // 创建租户
				tenant.GET("", app.Handlers.TenantHandler.ListTenants)                        // 获取租户列表（管理视角）
				tenant.GET("/all", app.Handlers.TenantHandler.ListAllTenants)                 // 获取所有启用的租户列表（不分页）
				tenant.GET("/detail", app.Handlers.TenantHandler.GetTenant)                   // 获取租户详情
				tenant.PUT("", app.Handlers.TenantHandler.UpdateTenant)                       // 更新租户
				tenant.DELETE("", app.Handlers.TenantHandler.DeleteTenant)                    // 删除租户
				tenant.DELETE("/batch-delete", app.Handlers.TenantHandler.BatchDeleteTenants) // 批量删除租户
				tenant.PUT("/status", app.Handlers.TenantHandler.UpdateTenantStatus)          // 更新租户状态
			}

			// 用户管理（租户管理员+超管）
			user := authorized.Group("/users")
			{
				user.POST("", app.Handlers.UserHandler.CreateUser)                      // 创建用户
				user.GET("", app.Handlers.UserHandler.ListUsers)                        // 获取用户列表
				user.GET("/detail", app.Handlers.UserHandler.GetUser)                   // 获取用户详情
				user.PUT("", app.Handlers.UserHandler.UpdateUser)                       // 更新用户
				user.DELETE("", app.Handlers.UserHandler.DeleteUser)                    // 删除用户
				user.DELETE("/batch-delete", app.Handlers.UserHandler.BatchDeleteUsers) // 批量删除用户
				user.PUT("/status", app.Handlers.UserHandler.UpdateUserStatus)          // 更新用户状态
				user.GET("/roles", app.Handlers.UserHandler.GetUserRoles)               // 获取用户角色
				user.PUT("/roles", app.Handlers.UserHandler.AssignRoles)                // 为用户分配角色
				user.POST("/password/reset", app.Handlers.UserHandler.ResetPassword)    // 超管重置用户密码
			}

			// 角色管理（租户管理员+超管）
			role := authorized.Group("/roles")
			{
				role.POST("", app.Handlers.RoleHandler.CreateRole)                    // 创建角色
				role.GET("", app.Handlers.RoleHandler.ListRoles)                      // 获取角色列表（分页）
				role.GET("/all", app.Handlers.RoleHandler.GetAllRoles)                // 获取所有角色（不分页）
				role.GET("/detail", app.Handlers.RoleHandler.GetRole)                 // 获取角色详情
				role.PUT("", app.Handlers.RoleHandler.UpdateRole)                     // 更新角色
				role.DELETE("", app.Handlers.RoleHandler.DeleteRole)                  // 删除角色
				role.PUT("/status", app.Handlers.RoleHandler.UpdateRoleStatus)        // 更新角色状态
				role.PUT("/permissions", app.Handlers.RoleHandler.AssignPermissions)  // 为角色分配权限（菜单+按钮）
				role.GET("/permissions", app.Handlers.RoleHandler.GetRolePermissions) // 获取角色的权限
			}

			// 菜单接口
			menu := authorized.Group("/menus")
			{
				menu.POST("", app.Handlers.MenuHandler.CreateMenu)             // 创建菜单
				menu.GET("", app.Handlers.MenuHandler.ListMenus)               // 获取菜单列表（分页）
				menu.GET("/all", app.Handlers.MenuHandler.GetAllMenus)         // 获取所有菜单（平铺）
				menu.GET("/tree", app.Handlers.MenuHandler.GetMenuTree)        // 获取菜单树
				menu.GET("/detail", app.Handlers.MenuHandler.GetMenu)          // 获取菜单详情
				menu.PUT("", app.Handlers.MenuHandler.UpdateMenu)              // 更新菜单
				menu.DELETE("", app.Handlers.MenuHandler.DeleteMenu)           // 删除菜单
				menu.PUT("/status", app.Handlers.MenuHandler.UpdateMenuStatus) // 更新菜单状态
			}

			// 部门管理（租户管理员+超管）
			dept := authorized.Group("/departments")
			{
				dept.POST("", app.Handlers.DepartmentHandler.CreateDepartment)             // 创建部门
				dept.GET("", app.Handlers.DepartmentHandler.ListDepartments)               // 获取部门列表
				dept.GET("/tree", app.Handlers.DepartmentHandler.GetDepartmentTree)        // 获取部门树
				dept.GET("/detail", app.Handlers.DepartmentHandler.GetDepartment)          // 获取部门详情
				dept.PUT("", app.Handlers.DepartmentHandler.UpdateDepartment)              // 更新部门
				dept.DELETE("", app.Handlers.DepartmentHandler.DeleteDepartment)           // 删除部门
				dept.PUT("/status", app.Handlers.DepartmentHandler.UpdateDepartmentStatus) // 更新部门状态
				dept.GET("/children", app.Handlers.DepartmentHandler.GetChildren)          // 获取子部门
			}

			// 岗位管理（租户管理员+超管）
			position := authorized.Group("/positions")
			{
				position.POST("", app.Handlers.PositionHandler.CreatePosition)             // 创建岗位
				position.GET("", app.Handlers.PositionHandler.ListPositions)               // 获取岗位列表
				position.GET("/all", app.Handlers.PositionHandler.ListAllPositions)        // 获取所有岗位（不分页）
				position.GET("/detail", app.Handlers.PositionHandler.GetPosition)          // 获取岗位详情
				position.PUT("", app.Handlers.PositionHandler.UpdatePosition)              // 更新岗位
				position.DELETE("", app.Handlers.PositionHandler.DeletePosition)           // 删除岗位
				position.PUT("/status", app.Handlers.PositionHandler.UpdatePositionStatus) // 更新岗位状态
			}

			// 字典管理接口
			dicts := authorized.Group("/dicts")
			{
				dicts.GET("", app.Handlers.DictHandler.GetDict)                    // 获取字典（合并系统+覆盖）
				dicts.GET("/types", app.Handlers.DictHandler.ListDictTypes)        // 获取字典类型列表
				dicts.PUT("/items", app.Handlers.DictHandler.BatchUpdateDictItems) // 批量更新字典项
				dicts.DELETE("/items", app.Handlers.DictHandler.ResetDictItem)     // 恢复系统默认值
			}
			// 系统字典管理接口（超管专用）
			systemDict := authorized.Group("/system/dicts")
			{
				systemDict.POST("", app.Handlers.DictHandler.CreateSystemDict)             // 创建系统字典
				systemDict.GET("", app.Handlers.DictHandler.ListSystemDictTypes)           // 获取系统字典列表
				systemDict.PUT("", app.Handlers.DictHandler.UpdateSystemDict)              // 更新系统字典
				systemDict.DELETE("", app.Handlers.DictHandler.DeleteSystemDict)           // 删除系统字典
				systemDict.DELETE("/items", app.Handlers.DictHandler.DeleteSystemDictItem) // 删除系统字典项
			}

			// 审计日志接口
			logs := authorized.Group("/logs")
			{
				logs.GET("/login", app.Handlers.LoginLogHandler.ListLoginLogs)                  // 获取登录日志列表
				logs.GET("/login/detail", app.Handlers.LoginLogHandler.GetLoginLog)             // 获取登录日志详情
				logs.GET("/operation", app.Handlers.OperationLogHandler.ListOperationLogs)      // 获取操作日志列表
				logs.GET("/operation/detail", app.Handlers.OperationLogHandler.GetOperationLog) // 获取操作日志详情
			}

		}
	}
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
		path := c.Request.URL.Path
		// // API 路由返回 404 JSON
		// if strings.HasPrefix(path, "/api") {
		// 	c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found", "path": path})
		// 	return
		// }
		// 其他路由返回前端 index.html（SPA 路由）
		fmt.Printf("[NoRoute] %s doesn't exist, serving index.html\n", path)
		c.FileFromFS("/index.html", frontendFS)
	})

}
