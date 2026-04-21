package router

import (
	"admin/internal/handler/auth"
	"admin/internal/handler/captcha"
	"admin/internal/handler/department"
	"admin/internal/handler/dict"
	"admin/internal/handler/health"
	"admin/internal/handler/loginlog"
	"admin/internal/handler/menu"
	"admin/internal/handler/operationlog"
	"admin/internal/handler/position"
	"admin/internal/handler/role"
	"admin/internal/handler/tenant"
	"admin/internal/handler/user"
	"admin/internal/jobs"

	"admin/internal/rbac"
	"admin/pkg/audit"
	"admin/pkg/cache"
	"admin/pkg/config"
	"admin/pkg/constants"
	"admin/pkg/database"
	"admin/pkg/utils/jwt"
	"admin/pkg/utils/logger"
	"admin/pkg/utils/rsapwd"
	"admin/pkg/utils/xcron"
	"admin/pkg/utils/xredis"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type App struct {
	Config    *config.Config
	Router    *gin.Engine
	DB        *gorm.DB
	Redis     redis.UniversalClient
	JWT       *jwt.Manager
	RBAC      *rbac.PermissionCache
	RSACipher *rsapwd.RSACipher
	Cron      *xcron.Manager
	Handlers  *Handlers
	Audit     *audit.Recorder
}

type Handlers struct {
	HealthHandler       *health.Handler
	CaptchaHandler      *captcha.Handler
	AuthHandler         *auth.Handler
	UserHandler         *user.Handler
	TenantHandler       *tenant.Handler
	RoleHandler         *role.Handler
	MenuHandler         *menu.Handler
	LoginLogHandler     *loginlog.Handler
	OperationLogHandler *operationlog.Handler
	DepartmentHandler   *department.Handler
	PositionHandler     *position.Handler
	DictHandler         *dict.Handler
}

func NewApp() (*App, error) {
	app := &App{}
	// 1. 加载配置
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

	// 3.5 初始化缓存（租户缓存）
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

	// 6. 初始化RBAC权限缓存
	if err := app.initRBAC(); err != nil {
		return nil, fmt.Errorf("failed to init rbac: %w", err)
	}

	// 6.5 初始化RSA密码解密器
	if err := app.initRSACipher(); err != nil {
		return nil, fmt.Errorf("failed to init rsa cipher: %w", err)
	}

	// 6.6 初始化定时任务
	if err := app.initCron(); err != nil {
		return nil, fmt.Errorf("failed to init cron: %w", err)
	}

	// 7. 创建审计 Recorder
	auditDB := audit.NewDB(app.DB)
	app.Audit = audit.NewRecorder(auditDB)

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

	log.Info().Str("app", s.Config.App.Name).Str("env", s.Config.App.Env).Int("port", s.Config.App.Port).Msg("Starting Admin Backend")
	return nil
}

func (s *App) initLogger(cfg *config.Config) error {
	logger.Init(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	log.Info().Str("level", cfg.Log.Level).Str("format", cfg.Log.Format).Msg("Logger initialized")
	return nil
}

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

	log.Info().Str("host", cfg.Database.Host).Int("port", cfg.Database.Port).Str("dbname", cfg.Database.DBName).Msg("Database connected")
	return nil
}

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
	log.Info().Str("addr", cfg.Redis.GetAddr()).Int("db", cfg.Redis.DB).Msg("Redis connected")
	return nil
}

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

func (a *App) initRBAC() error {
	a.RBAC = rbac.NewPermissionCache(a.DB, 30*time.Second)
	return nil
}

func (a *App) initRSACipher() error {
	cipher := rsapwd.MustNew(constants.RSAKey)
	a.RSACipher = cipher
	log.Info().Msg("RSA cipher initialized")
	return nil
}

func (a *App) initCron() error {
	cronMgr, err := xcron.Init(xcron.Config{
		WithSeconds: true,
	})
	if err != nil {
		return err
	}
	a.Cron = cronMgr

	if err := jobs.Init(cronMgr, a.DB); err != nil {
		return fmt.Errorf("failed to register jobs: %w", err)
	}

	a.Cron.Start()
	log.Info().Msg("Cron jobs initialized and started")
	return nil
}

func (s *App) initHandlers() error {
	s.Handlers = &Handlers{
		HealthHandler:       health.NewHandler(),
		CaptchaHandler:      captcha.NewHandler(s.Redis),
		AuthHandler:         auth.NewHandler(s.DB, s.JWT, s.Redis, s.Audit, s.RSACipher, s.Config),
		UserHandler:         user.NewHandler(s.DB, s.Audit, s.RSACipher, s.RBAC),
		TenantHandler:       tenant.NewHandler(s.DB, s.Audit),
		RoleHandler:         role.NewHandler(s.DB, s.Audit, s.RBAC),
		MenuHandler:         menu.NewHandler(s.DB, s.Audit, s.RBAC),
		LoginLogHandler:     loginlog.NewHandler(s.DB),
		OperationLogHandler: operationlog.NewHandler(s.DB),
		DepartmentHandler:   department.NewHandler(s.DB, s.Audit),
		PositionHandler:     position.NewHandler(s.DB, s.Audit),
		DictHandler:         dict.NewHandler(s.DB, s.Audit),
	}
	return nil
}

func (s *App) initRouter() error {
	gin.SetMode(s.Config.Server.Mode)

	r := gin.New()

	Setup(r, s.Handlers, s.Config, s.JWT, s.RBAC)
	s.Router = r

	log.Info().Str("mode", s.Config.Server.Mode).Msg("Router initialized")
	return nil
}

func (s *App) Run() error {
	addr := fmt.Sprintf(":%d", s.Config.App.Port)

	if s.Config.Server.TLS.Enabled {
		certFile := s.Config.Server.TLS.CertFile
		keyFile := s.Config.Server.TLS.KeyFile

		if certFile == "" || keyFile == "" {
			return fmt.Errorf("TLS enabled but cert_file or key_file not specified")
		}

		log.Info().Str("addr", addr).Str("cert", certFile).Str("key", keyFile).Msg("Starting HTTPS server")
		return s.Router.RunTLS(addr, certFile, keyFile)
	}

	log.Info().Str("addr", addr).Msg("Starting HTTP server")
	return s.Router.Run(addr)
}

func (s *App) Close() error {
	if s.RBAC != nil {
		s.RBAC.Stop()
	}

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
