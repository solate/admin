package router

import (
	"admin/internal/handler"
	"admin/internal/jobs"
	"admin/internal/repository"
	"admin/internal/service"
	"admin/pkg/audit"
	"admin/pkg/cache"
	"admin/pkg/casbin"
	"admin/pkg/config"
	"admin/pkg/constants"
	"admin/pkg/database"
	"admin/pkg/jwt"
	"admin/pkg/logger"
	"admin/pkg/rsapwd"
	"admin/pkg/xcron"
	"admin/pkg/xredis"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type App struct {
	Config    *config.Config        // 配置
	Router    *gin.Engine           // 路由
	DB        *gorm.DB              // 数据库连接
	Redis     redis.UniversalClient // Redis连接
	JWT       *jwt.Manager          // JWT管理器
	Enforcer  *casbin.Enforcer      // Casbin enforce
	RSACipher *rsapwd.RSACipher     // RSA 密码解密器（全局单例）
	Cron      *xcron.Manager        // 定时任务管理器
	Handlers  *Handlers             // 处理器层容器
}

// Handlers 处理器层容器
type Handlers struct {
	HealthHandler       *handler.HealthHandler
	CaptchaHandler      *handler.CaptchaHandler
	AuthHandler         *handler.AuthHandler
	UserHandler         *handler.UserHandler
	TenantHandler       *handler.TenantHandler
	RoleHandler         *handler.RoleHandler
	MenuHandler         *handler.MenuHandler
	UserMenuHandler     *handler.UserMenuHandler
	LoginLogHandler     *handler.LoginLogHandler
	OperationLogHandler *handler.OperationLogHandler
	DepartmentHandler   *handler.DepartmentHandler
	PositionHandler     *handler.PositionHandler
	DictHandler         *handler.DictHandler
}

func NewApp() (*App, error) {
	app := &App{}
	// 1. 加载配置 (必须最先加载，其他组件依赖配置)
	if err := app.initConfig(); err != nil {
		return nil, fmt.Errorf("failed to init config: %w", err)
	}

	// 2. 初始化日志
	if err := app.initLogger(app.Config); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	// 3. 初始化数据库
	if err := app.initDatabase(app.Config); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// 3.5 初始化缓存（租户缓存等）
	if err := cache.Init(app.DB); err != nil {
		return nil, fmt.Errorf("failed to init cache: %w", err)
	}

	// 4. 初始化Redis
	if err := app.initRedis(app.Config); err != nil {
		return nil, fmt.Errorf("failed to init redis: %w", err)
	}

	// 5. 初始化JWT
	if err := app.initJWT(app.Config); err != nil {
		return nil, fmt.Errorf("failed to init jwt: %w", err)
	}

	// 6. 初始化Casbin
	if err := app.initCasbin(); err != nil {
		return nil, fmt.Errorf("failed to init casbin: %w", err)
	}

	// 6.5 初始化RSA密码解密器
	if err := app.initRSACipher(); err != nil {
		return nil, fmt.Errorf("failed to init rsa cipher: %w", err)
	}

	// 6.6 初始化定时任务
	if err := app.initCron(); err != nil {
		return nil, fmt.Errorf("failed to init cron: %w", err)
	}

	// 7. 初始化新的审计日志 Recorder
	auditDB := audit.NewDB(app.DB)
	auditRecorder := audit.NewRecorder(auditDB)

	// 8. 初始化处理器层
	if err := app.initHandlers(auditRecorder); err != nil {
		return nil, fmt.Errorf("failed to init handlers: %w", err)
	}

	// 9. 初始化路由
	if err := app.initRouter(); err != nil {
		return nil, fmt.Errorf("failed to init router: %w", err)
	}

	log.Info().Msg("Application initialized successfully")
	return app, nil
}

func (s *App) initConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	s.Config = cfg

	log.Info().
		Str("app", s.Config.App.Name).
		Str("env", s.Config.App.Env).
		Int("port", s.Config.App.Port).
		Msg("Starting Admin Backend")
	return nil
}

// initLogger 初始化日志系统（使用配置文件中的设置）
func (s *App) initLogger(cfg *config.Config) error {
	logger.Init(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	log.Info().
		Str("level", cfg.Log.Level).
		Str("format", cfg.Log.Format).
		Msg("Logger initialized")
	return nil
}

// initDatabase 初始化数据库连接
func (s *App) initDatabase(cfg *config.Config) error {
	db, err := database.Connect(database.Config{
		DSN:             cfg.Database.GetDSN(),
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.GetConnMaxLifetime(),
		LogLevel:        cfg.Log.Level,
	})
	if err != nil {
		return err
	}
	s.DB = db

	log.Info().
		Str("host", cfg.Database.Host).
		Int("port", cfg.Database.Port).
		Str("dbname", cfg.Database.DBName).
		Msg("Database connected")
	return nil
}

// initRedis 初始化Redis连接
func (a *App) initRedis(cfg *config.Config) error {
	config := xredis.Config{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		Type:         cfg.Redis.Type,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.GetDialTimeout(),
		ReadTimeout:  cfg.Redis.GetReadTimeout(),
		WriteTimeout: cfg.Redis.GetWriteTimeout(),
	}

	redisClient, err := xredis.Connect(config)
	if err != nil {
		return err
	}

	a.Redis = redisClient
	log.Info().
		Str("addr", cfg.Redis.GetAddr()).
		Int("db", cfg.Redis.DB).
		Msg("Redis connected")
	return nil
}

// initJWT 初始化JWT管理器
func (a *App) initJWT(cfg *config.Config) error {
	config := &jwt.JWTConfig{
		AccessSecret:  []byte(cfg.JWT.AccessSecret),
		AccessExpire:  cfg.JWT.AccessExpire,
		RefreshSecret: []byte(cfg.JWT.RefreshSecret),
		RefreshExpire: cfg.JWT.RefreshExpire,
		Issuer:        cfg.JWT.Issuer,
	}

	a.JWT = jwt.NewManager(config, jwt.NewRedisStore(a.Redis))
	return nil
}

// initCasbin 初始化Casbin
func (a *App) initCasbin() error {
	enforcer, err := casbin.NewEnforcerManager(a.DB, casbin.DefaultModel())
	if err != nil {
		return fmt.Errorf("failed to init casbin: %w", err)
	}

	a.Enforcer = enforcer
	return nil
}

// initRSACipher 初始化RSA密码解密器
func (a *App) initRSACipher() error {
	// 使用测试密钥创建 RSA 加密器
	// 注意：生产环境应该从配置文件或环境变量中读取私钥
	cipher := rsapwd.MustNew(constants.RSAKey)
	a.RSACipher = cipher

	log.Info().Msg("RSA cipher initialized")
	return nil
}

// initCron 初始化定时任务
func (a *App) initCron() error {
	// 初始化 cron 管理器
	cronMgr, err := xcron.Init(xcron.Config{
		WithSeconds: true,
	})
	if err != nil {
		return err
	}
	a.Cron = cronMgr

	// 注册所有定时任务
	if err := jobs.Init(cronMgr, a.DB); err != nil {
		return fmt.Errorf("failed to register jobs: %w", err)
	}

	// 启动定时任务
	a.Cron.Start()

	log.Info().Msg("Cron jobs initialized and started")
	return nil
}

// initRouter 初始化路由
func (s *App) initRouter() error {
	gin.SetMode(s.Config.Server.Mode)

	r := gin.New()

	Setup(r, s)
	s.Router = r

	log.Info().
		Str("mode", s.Config.Server.Mode).
		Msg("Router initialized")
	return nil
}

// Run 启动应用
func (s *App) Run() error {
	addr := fmt.Sprintf(":%d", s.Config.App.Port)

	// 如果启用了 TLS，使用 HTTPS
	if s.Config.Server.TLS.Enabled {
		certFile := s.Config.Server.TLS.CertFile
		keyFile := s.Config.Server.TLS.KeyFile

		// 验证证书文件存在
		if certFile == "" || keyFile == "" {
			return fmt.Errorf("TLS enabled but cert_file or key_file not specified")
		}

		log.Info().
			Str("addr", addr).
			Str("cert", certFile).
			Str("key", keyFile).
			Msg("Starting HTTPS server")
		return s.Router.RunTLS(addr, certFile, keyFile)
	}

	log.Info().Str("addr", addr).Msg("Starting HTTP server")
	return s.Router.Run(addr)
}

// Close 关闭资源
func (s *App) Close() error {
	// 停止定时任务
	if s.Cron != nil {
		s.Cron.Stop()
	}

	if err := database.Close(); err != nil {
		return err
	}
	if err := xredis.Close(); err != nil {
		return err
	}
	return nil
}

// initHandlers 初始化处理器层
func (s *App) initHandlers(auditRecorder *audit.Recorder) error {
	// 初始化仓库层
	userRepo := repository.NewUserRepo(s.DB)
	userRoleRepo := repository.NewUserRoleRepo(s.Enforcer)
	roleRepo := repository.NewRoleRepo(s.DB)
	tenantRepo := repository.NewTenantRepo(s.DB)
	menuRepo := repository.NewMenuRepo(s.DB)
	permissionRepo := repository.NewPermissionRepo(s.DB)
	loginLogRepo := repository.NewLoginLogRepo(s.DB)
	operationLogRepo := repository.NewOperationLogRepo(s.DB)
	departmentRepo := repository.NewDepartmentRepo(s.DB)
	positionRepo := repository.NewPositionRepo(s.DB)
	dictTypeRepo := repository.NewDictTypeRepo(s.DB)
	dictItemRepo := repository.NewDictItemRepo(s.DB)

	// 初始化服务层
	authService := service.NewAuthService(userRepo, userRoleRepo, roleRepo, tenantRepo, s.JWT, s.Redis, s.Enforcer, auditRecorder, s.Config, s.RSACipher) // 初始化认证服务
	userRoleService := service.NewUserRoleService(userRepo, userRoleRepo, roleRepo, tenantRepo, s.Enforcer, auditRecorder)                                // 初始化用户角色服务
	userService := service.NewUserService(userRepo, userRoleService, roleRepo, tenantRepo, s.Enforcer, auditRecorder, s.RSACipher)                        // 初始化用户服务
	tenantService := service.NewTenantService(tenantRepo, userRepo, auditRecorder)                                                                        // 初始化租户服务
	roleService := service.NewRoleService(roleRepo, permissionRepo, menuRepo, s.Enforcer, cache.Get().Tenant, auditRecorder)                              // 初始化角色服务（需要 enforcer、tenantCache 和 menuRepo 用于角色继承、菜单权限管理和 API 权限关联）
	menuService := service.NewMenuService(menuRepo, s.Enforcer, auditRecorder)                                                                            // 初始化菜单服务
	userMenuService := service.NewUserMenuService(menuRepo, permissionRepo, s.Enforcer)                                                                   // 初始化用户菜单服务（根据设计文档：无需取交集）
	loginLogService := service.NewLoginLogService(loginLogRepo)                                                                                           // 初始化登录日志服务
	operationLogService := service.NewOperationLogService(operationLogRepo)                                                                               // 初始化操作日志服务
	departmentService := service.NewDepartmentService(departmentRepo, userRepo, auditRecorder)                                                            // 初始化部门服务
	positionService := service.NewPositionService(positionRepo, auditRecorder)                                                                            // 初始化岗位服务
	dictService := service.NewDictService(dictTypeRepo, dictItemRepo, cache.Get().Tenant, auditRecorder)                                                  // 初始化字典服务
	// 初始化接入设备服务

	s.Handlers = &Handlers{
		HealthHandler:       handler.NewHealthHandler(),
		CaptchaHandler:      handler.NewCaptchaHandler(s.Redis),
		AuthHandler:         handler.NewAuthHandler(authService),
		UserHandler:         handler.NewUserHandler(userService, userRoleService),
		TenantHandler:       handler.NewTenantHandler(tenantService),
		RoleHandler:         handler.NewRoleHandler(roleService),
		MenuHandler:         handler.NewMenuHandler(menuService),
		UserMenuHandler:     handler.NewUserMenuHandler(userMenuService),
		LoginLogHandler:     handler.NewLoginLogHandler(loginLogService),
		OperationLogHandler: handler.NewOperationLogHandler(operationLogService),
		DepartmentHandler:   handler.NewDepartmentHandler(departmentService),
		PositionHandler:     handler.NewPositionHandler(positionService),
		DictHandler:         handler.NewDictHandler(dictService),
	}
	return nil
}
