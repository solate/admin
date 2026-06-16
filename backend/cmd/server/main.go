package main

import (
	"context"
	"flag"
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
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		panic("load config: " + err.Error())
	}

	// 2. 初始化日志
	log := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
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
		log.Fatal().Err(err).Msg("connect database failed")
	}
	defer database.Close(db)

	// 4. 连接 Redis
	rdbClient, err := rdb.New(rdb.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("connect redis failed")
	}
	defer rdbClient.Close()

	log.Info().Int("port", cfg.Server.Port).Msg("all infrastructure initialized")

	// 5. 启动 HTTP 服务器
	srv, err := server.New(server.Options{
		Config: cfg,
		DB:     db,
		RDB:    rdbClient,
		Log:    log,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("init server failed")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("start server failed")
		}
	}()

	<-ctx.Done()
	stop()
	log.Info().Msg("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Stop(shutdownCtx)
	log.Info().Msg("server exited")
}
