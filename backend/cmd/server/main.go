package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"admin/internal/config"
	"admin/internal/server"
	"admin/pkg/database"
	"admin/pkg/logger"
	"admin/pkg/rdb"
)

func main() {
	// 1. 加载配置(路径约定 config/config.yaml,靠挂载覆盖内容,APP_ENV 切环境)
	cfg, err := config.InitConfig()
	if err != nil {
		panic("load config: " + err.Error())
	}

	// 2. 初始化日志
	log := logger.New(logger.Config{
		Level:     cfg.Log.Level,
		Format:    cfg.Log.Format,
		AddSource: cfg.Log.AddSource,
	})

	// 3. 连接数据库
	db, err := database.New(database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}, log)
	if err != nil {
		logger.Fatal(log, "connect database failed", slog.Any("err", err))
	}
	defer database.Close(db)

	// 4. 连接 Redis
	rdbClient, err := rdb.New(rdb.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		logger.Fatal(log, "connect redis failed", slog.Any("err", err))
	}
	defer rdbClient.Close()

	log.Info("all infrastructure initialized", slog.Int("port", cfg.Server.Port))

	// 5. 启动 HTTP 服务器
	srv, err := server.New(server.Options{
		Config: cfg,
		DB:     db,
		RDB:    rdbClient,
		Log:    log,
	})
	if err != nil {
		logger.Fatal(log, "init server failed", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal(log, "start server failed", slog.Any("err", err))
		}
	}()

	<-ctx.Done()
	stop()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.GracefulTimeout)*time.Second)
	defer cancel()
	srv.Stop(shutdownCtx)
	log.Info("server exited")
}
