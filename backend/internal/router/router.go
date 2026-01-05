package router

import (
	"admin/internal/middleware"
	"admin/pkg/audit"
	filesystem "admin/static"
	"fmt"
	"net/http"
	"strings"

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
			// 登录/注册相关接口 - 需要审计日志
			auth := public.Group("/auth")

			{
				auth.GET("/captcha", app.Handlers.CaptchaHandler.Get)   // 获取验证码 （无需租户）
				auth.POST("/refresh", app.Handlers.AuthHandler.Refresh) // 刷新令牌 （无需租户）

				tenantAuth := auth.Group("/:tenant_code")
				tenantAuth.Use(middleware.TenantFromCode(app.DB)) //中间件
				tenantAuth.Use(audit.AuditMiddleware())           // 登录接口也需要审计日志
				{
					tenantAuth.POST("/login", app.Handlers.AuthHandler.Login) // 用户登录
				}
				// auth.POST("/register", func(c *gin.Context) {
				// 	c.JSON(200, gin.H{"registered": true})
				// }) // 用户注册
			}

		}

		// 需要认证 + Casbin 权限检查的路由
		authorized := v1.Group("")
		authorized.Use(middleware.AuthMiddleware(app.JWT))
		authorized.Use(middleware.CasbinMiddleware(app.Enforcer)) // 权限校验（超管跳过，其他走 Casbin）
		authorized.Use(audit.AuditMiddleware())                   // 审计日志中间件（提取请求信息）
		{
			// 用户个人资料相关（当前登录用户操作自己的数据）
			userSelf := authorized.Group("")
			{
				userSelf.GET("/profile", app.Handlers.UserHandler.GetProfile)              // 获取当前用户信息（含角色）
				userSelf.POST("/password/change", app.Handlers.UserHandler.ChangePassword) // 修改自己的密码
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
				tenant.POST("", app.Handlers.TenantHandler.CreateTenant)                                // 创建租户
				tenant.GET("", app.Handlers.TenantHandler.ListTenants)                                  // 获取租户列表（管理视角）
				tenant.GET("/:tenant_id", app.Handlers.TenantHandler.GetTenant)                         // 获取租户详情
				tenant.PUT("/:tenant_id", app.Handlers.TenantHandler.UpdateTenant)                      // 更新租户
				tenant.DELETE("/:tenant_id", app.Handlers.TenantHandler.DeleteTenant)                   // 删除租户
				tenant.PUT("/:tenant_id/status/:status", app.Handlers.TenantHandler.UpdateTenantStatus) // 更新租户状态
			}

			// 用户管理（租户管理员+超管）
			user := authorized.Group("/users")
			{
				user.POST("", app.Handlers.UserHandler.CreateUser)                              // 创建用户
				user.GET("", app.Handlers.UserHandler.ListUsers)                                // 获取用户列表
				user.GET("/:user_id", app.Handlers.UserHandler.GetUser)                         // 获取用户详情
				user.PUT("/:user_id", app.Handlers.UserHandler.UpdateUser)                      // 更新用户
				user.DELETE("/:user_id", app.Handlers.UserHandler.DeleteUser)                   // 删除用户
				user.PUT("/:user_id/status/:status", app.Handlers.UserHandler.UpdateUserStatus) // 更新用户状态
				user.GET("/:user_id/roles", app.Handlers.UserHandler.GetUserRoles)              // 获取用户角色
				user.PUT("/:user_id/roles", app.Handlers.UserHandler.AssignRoles)               // 为用户分配角色
				user.POST("/:user_id/password/reset", app.Handlers.UserHandler.ResetPassword)   // 超管重置用户密码
			}

			// 角色管理（租户管理员+超管）
			role := authorized.Group("/roles")
			{
				role.POST("", app.Handlers.RoleHandler.CreateRole)                              // 创建角色
				role.GET("", app.Handlers.RoleHandler.ListRoles)                                // 获取角色列表
				role.GET("/:role_id", app.Handlers.RoleHandler.GetRole)                         // 获取角色详情
				role.PUT("/:role_id", app.Handlers.RoleHandler.UpdateRole)                      // 更新角色
				role.DELETE("/:role_id", app.Handlers.RoleHandler.DeleteRole)                   // 删除角色
				role.PUT("/:role_id/status/:status", app.Handlers.RoleHandler.UpdateRoleStatus) // 更新角色状态
				role.PUT("/:role_id/permissions", app.Handlers.RoleHandler.AssignPermissions)   // 为角色分配权限（菜单+按钮）
				role.GET("/:role_id/permissions", app.Handlers.RoleHandler.GetRolePermissions)  // 获取角色的权限
				// 保留旧路由向后兼容
				role.PUT("/:role_id/menus", app.Handlers.RoleHandler.AssignMenus)  // 为角色分配菜单（已弃用）
				role.GET("/:role_id/menus", app.Handlers.RoleHandler.GetRoleMenus) // 获取角色的菜单（已弃用）
			}

			// 菜单接口
			menu := authorized.Group("/menus")
			{
				menu.POST("", app.Handlers.MenuHandler.CreateMenu)                              // 创建菜单
				menu.GET("", app.Handlers.MenuHandler.ListMenus)                                // 获取菜单列表（分页）
				menu.GET("/all", app.Handlers.MenuHandler.GetAllMenus)                          // 获取所有菜单（平铺）
				menu.GET("/tree", app.Handlers.MenuHandler.GetMenuTree)                         // 获取菜单树
				menu.GET("/:menu_id", app.Handlers.MenuHandler.GetMenu)                         // 获取菜单详情
				menu.PUT("/:menu_id", app.Handlers.MenuHandler.UpdateMenu)                      // 更新菜单
				menu.DELETE("/:menu_id", app.Handlers.MenuHandler.DeleteMenu)                   // 删除菜单
				menu.PUT("/:menu_id/status/:status", app.Handlers.MenuHandler.UpdateMenuStatus) // 更新菜单状态
			}

			// 部门管理（租户管理员+超管）
			dept := authorized.Group("/departments")
			{
				dept.POST("", app.Handlers.DepartmentHandler.CreateDepartment)                                    // 创建部门
				dept.GET("", app.Handlers.DepartmentHandler.ListDepartments)                                      // 获取部门列表
				dept.GET("/tree", app.Handlers.DepartmentHandler.GetDepartmentTree)                               // 获取部门树
				dept.GET("/:department_id", app.Handlers.DepartmentHandler.GetDepartment)                         // 获取部门详情
				dept.PUT("/:department_id", app.Handlers.DepartmentHandler.UpdateDepartment)                      // 更新部门
				dept.DELETE("/:department_id", app.Handlers.DepartmentHandler.DeleteDepartment)                   // 删除部门
				dept.PUT("/:department_id/status/:status", app.Handlers.DepartmentHandler.UpdateDepartmentStatus) // 更新部门状态
				dept.GET("/:department_id/children", app.Handlers.DepartmentHandler.GetChildren)                  // 获取子部门
			}

			// 岗位管理（租户管理员+超管）
			position := authorized.Group("/positions")
			{
				position.POST("", app.Handlers.PositionHandler.CreatePosition)                                  // 创建岗位
				position.GET("", app.Handlers.PositionHandler.ListPositions)                                    // 获取岗位列表
				position.GET("/all", app.Handlers.PositionHandler.ListAllPositions)                             // 获取所有岗位（不分页）
				position.GET("/:position_id", app.Handlers.PositionHandler.GetPosition)                         // 获取岗位详情
				position.PUT("/:position_id", app.Handlers.PositionHandler.UpdatePosition)                      // 更新岗位
				position.DELETE("/:position_id", app.Handlers.PositionHandler.DeletePosition)                   // 删除岗位
				position.PUT("/:position_id/status/:status", app.Handlers.PositionHandler.UpdatePositionStatus) // 更新岗位状态
			}

			// 用户菜单接口（基于权限动态加载）
			userMenu := authorized.Group("/user")
			{
				userMenu.GET("/menus", app.Handlers.UserMenuHandler.GetUserMenu)               // 获取用户菜单树
				userMenu.GET("/buttons/:menu_id", app.Handlers.UserMenuHandler.GetUserButtons) // 获取菜单按钮权限
			}

			// 字典管理接口
			dict := authorized.Group("/dict")
			{
				dict.GET("/:type_code", app.Handlers.DictHandler.GetDict)                       // 获取字典（合并系统+覆盖）
				dict.PUT("/items", app.Handlers.DictHandler.BatchUpdateDictItems)               // 批量更新字典项
				dict.DELETE("/:type_code/items/:value", app.Handlers.DictHandler.ResetDictItem) // 恢复系统默认值
			}
			// 系统字典管理接口（超管专用）
			systemDict := authorized.Group("/system/dict")
			{
				systemDict.POST("", app.Handlers.DictHandler.CreateSystemDict)                               // 创建系统字典
				systemDict.GET("", app.Handlers.DictHandler.ListSystemDictTypes)                             // 获取系统字典列表
				systemDict.PUT("/:type_code", app.Handlers.DictHandler.UpdateSystemDict)                     // 更新系统字典
				systemDict.DELETE("/:type_code", app.Handlers.DictHandler.DeleteSystemDict)                  // 删除系统字典
				systemDict.DELETE("/:type_code/items/:value", app.Handlers.DictHandler.DeleteSystemDictItem) // 删除系统字典项
			}
			// 字典类型列表接口
			authorized.GET("/dict-types", app.Handlers.DictHandler.ListDictTypes) // 获取字典类型列表

			// 审计日志接口
			// 登录日志
			loginLog := authorized.Group("/login-logs")
			{
				loginLog.GET("", app.Handlers.LoginLogHandler.ListLoginLogs)       // 获取登录日志列表
				loginLog.GET("/:log_id", app.Handlers.LoginLogHandler.GetLoginLog) // 获取登录日志详情
			}
			// 操作日志
			operationLog := authorized.Group("/operation-logs")
			{
				operationLog.GET("", app.Handlers.OperationLogHandler.ListOperationLogs)       // 获取操作日志列表
				operationLog.GET("/:log_id", app.Handlers.OperationLogHandler.GetOperationLog) // 获取操作日志详情
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
		path := c.Request.URL.Path
		// API 路由返回 404 JSON
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found", "path": path})
			return
		}
		// 其他路由返回前端 index.html（SPA 路由）
		fmt.Printf("[NoRoute] %s doesn't exist, serving index.html\n", path)
		c.FileFromFS("/index.html", frontendFS)
	})

}
