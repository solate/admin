package main

import (
	"admin/internal/handler"
	"admin/internal/router"
	"admin/internal/service"
	"admin/pkg/casbin"
	"admin/pkg/config"
	"admin/pkg/database"
	"admin/pkg/jwt"
	"admin/pkg/logger"
	"admin/pkg/xredis"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// @title 管理后台 API
// @version 1.0
// @description 管理后台 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// App 应用实例，包含所有需要的组件
type App struct {
	cfg           *config.Config
	db            *gorm.DB
	router        *gin.Engine
	redis         redis.UniversalClient
	jwt           *jwt.JWTManager
	casbinService *service.CasbinService
}

func main() {
	// 初始化应用
	app, err := initApp()
	if err != nil {
		fmt.Printf("Failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	// 启动服务器
	go startServer(app)

	// 等待关闭信号
	waitForShutdown()

	// 清理资源
	cleanup(app)
}

// initApp 初始化应用的所有组件
func initApp() (*App, error) {
	app := &App{}

	// 1. 加载配置 (必须最先加载，其他组件依赖配置)
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	app.cfg = cfg

	log.Info().
		Str("app", cfg.App.Name).
		Str("env", cfg.App.Env).
		Int("port", cfg.App.Port).
		Msg("Starting Admin Backend")

	// 2. 初始化日志 (使用配置中的日志设置)
	if err := initLogger(cfg); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	// 3. 连接数据库
	db, err := initDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}
	app.db = db

	// Run migrations placeholder

	// 4. 连接Redis（全局）
	redisClient, err := initRedis(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init redis: %w", err)
	}
	app.redis = redisClient

	// 5. jwt
	app.jwt = initJWTManager(cfg, app.redis)

	// 6. 初始化 Casbin Service
	casbinService, err := service.InitCasbinService(app.db, casbin.DefaultModel())
	if err != nil {
		return nil, fmt.Errorf("failed to init casbin: %w", err)
	}
	app.casbinService = casbinService

	// Seed initial policies for admin (user-1)
	app.casbinService.AddPolicy("user-1", "tenant-1", "/api/v1/policy/add", "POST")
	app.casbinService.AddPolicy("user-1", "tenant-1", "/api/v1/policy/role/add", "POST")

	// 7. 初始化路由
	router := initRouter(cfg, app)
	app.router = router

	log.Info().Msg("Application initialized successfully")
	return app, nil
}

// loadConfig 加载配置文件
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// initLogger 初始化日志系统（使用配置文件中的设置）
func initLogger(cfg *config.Config) error {
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
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.Connect(database.Config{
		DSN:             cfg.Database.GetDSN(),
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.GetConnMaxLifetime(),
		LogLevel:        cfg.Log.Level,
	})
	if err != nil {
		return nil, err
	}
	log.Info().
		Str("host", cfg.Database.Host).
		Int("port", cfg.Database.Port).
		Str("dbname", cfg.Database.DBName).
		Msg("Database connected")
	return db, nil
}

// initRedis 初始化全局Redis连接
func initRedis(cfg *config.Config) (redis.UniversalClient, error) {
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
		return nil, err
	}

	log.Info().
		Str("addr", cfg.Redis.GetAddr()).
		Int("db", cfg.Redis.DB).
		Msg("Redis connected")
	return redisClient, nil
}

func initJWTManager(cfg *config.Config, rdb redis.UniversalClient) *jwt.JWTManager {
	config := &jwt.JWTConfig{
		AccessSecret:  []byte(cfg.JWT.AccessSecret),
		AccessExpire:  cfg.JWT.AccessExpire,
		RefreshSecret: []byte(cfg.JWT.RefreshSecret),
		RefreshExpire: cfg.JWT.RefreshExpire,
		Issuer:        cfg.JWT.Issuer,
	}

	return jwt.NewJWTManager(config, jwt.NewRedisStore(rdb))
}

// Handlers 处理器层容器
// initRouter 初始化路由
func initRouter(cfg *config.Config, app *App) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()

	routerCfg := &router.Config{
		HealthHandler: handler.NewHealthHandler(),
		AuthHandler:   handler.NewAuthHandler(cfg, app.jwt),
		PolicyHandler: handler.NewPolicyHandler(app.casbinService),
		AppConfig:     cfg,
		JWTManager:    app.jwt,
		CasbinService: app.casbinService,
	}

	router.Setup(r, routerCfg)

	log.Info().
		Str("mode", cfg.Server.Mode).
		Msg("Router initialized")
	return r
}

// initRouter 初始化路由
// 移除业务路由版本，保留简化版本

// startServer 启动HTTP服务器
func startServer(app *App) {
	addr := fmt.Sprintf(":%d", app.cfg.App.Port)
	log.Info().
		Str("addr", addr).
		Str("env", app.cfg.App.Env).
		Msg("Server starting")

	if err := app.router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

// waitForShutdown 等待关闭信号
func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")
}

// cleanup 清理资源
func cleanup(app *App) {
	// 关闭数据库连接
	if app.db != nil {
		if err := database.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database")
		} else {
			log.Info().Msg("Database connection closed")
		}
	}

	// 关闭Redis连接
	if app.redis != nil {
		if err := xredis.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close redis")
		} else {
			log.Info().Msg("Redis connection closed")
		}
	}

	log.Info().Msg("Server stopped gracefully")
}
