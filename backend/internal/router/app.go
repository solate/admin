package router

import (
	"admin/internal/handler"
	"admin/internal/repository"
	"admin/internal/service"
	"admin/pkg/casbin"
	"admin/pkg/config"
	"admin/pkg/database"
	"admin/pkg/jwt"
	"admin/pkg/operationlog"
	"admin/pkg/logger"
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
	Handlers  *Handlers             // 处理器层容器
	OperationLogLogger *operationlog.Logger // 操作日志写入器
}

// Handlers 处理器层容器
type Handlers struct {
	HealthHandler  *handler.HealthHandler
	CaptchaHandler *handler.CaptchaHandler
	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
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

	// 7. 初始化操作日志写入器
	if err := app.initOperationLogLogger(); err != nil {
		return nil, fmt.Errorf("failed to init operation log logger: %w", err)
	}

	// 8. 初始化处理器层
	if err := app.initHandlers(); err != nil {
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
	log.Info().Str("addr", addr).Msg("Starting server")
	return s.Router.Run(addr)
}

// Close 关闭资源
func (s *App) Close() error {
	if err := database.Close(); err != nil {
		return err
	}
	if err := xredis.Close(); err != nil {
		return err
	}
	return nil
}

// initOperationLogLogger 初始化操作日志写入器
func (s *App) initOperationLogLogger() error {
	s.OperationLogLogger = operationlog.NewLogger(s.DB)
	return nil
}

// initHandlers 初始化处理器层
func (s *App) initHandlers() error {
	// 初始化仓库层
	userRepo := repository.NewUserRepo(s.DB)
	userTenantRoleRepo := repository.NewUserTenantRoleRepo(s.DB)
	roleRepo := repository.NewRoleRepo(s.DB)
	tenantRepo := repository.NewTenantRepo(s.DB)

	// 初始化认证服务
	authService := service.NewAuthService(userRepo, userTenantRoleRepo, roleRepo, tenantRepo, s.JWT, s.Redis, s.Config)

	// 初始化用户服务
	userService := service.NewUserService(userRepo, userTenantRoleRepo, tenantRepo)

	s.Handlers = &Handlers{
		HealthHandler:  handler.NewHealthHandler(),
		CaptchaHandler: handler.NewCaptchaHandler(s.Redis),
		AuthHandler:    handler.NewAuthHandler(authService),
		UserHandler:    handler.NewUserHandler(userService),
		// PolicyHandler: handler.NewPolicyHandler(s.Services.CasbinService), // Will be nil for now
		// TenantHandler: handler.NewTenantHandler(s.Services.TenantService),
	}
	return nil
}
