package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"admin/internal/config"
	"admin/internal/server"
	"admin/pkg/database"
	"admin/pkg/xredis"
	"admin/pkg/xslog"
)

func main() {
	// main 只负责退出码：run 返回 error 说明启动/运行失败，打日志后以非零码退出，
	// 让容器/systemd 感知崩溃并重启。真正的逻辑与资源清理都在 run 里（defer 保证执行）。
	if err := run(); err != nil {
		slog.Error("server exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}

// run 承载全部启动逻辑，所有资源用 defer 逆序清理。
// 任何一步失败都 return error，由 main 统一打日志 + os.Exit(1)，
// 从而避免在日志调用里隐藏 os.Exit（slog 刻意不提供 Fatal，见 pkg/xslog）。
func run() error {
	// 1. 加载配置(路径约定 config/config.yaml,靠挂载覆盖内容,APP_ENV 切环境)
	cfg, err := config.InitConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 2. 初始化日志
	log := xslog.New(xslog.Config{
		Level:     cfg.Log.Level,
		Format:    cfg.Log.Format,
		AddSource: cfg.Log.AddSource,
	}).With("service", cfg.App.Name, "env", cfg.App.Env)

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
		return fmt.Errorf("connect database: %w", err)
	}
	defer database.Close(db)

	// 4. 连接 Redis
	rdbClient, err := xredis.New(xredis.Config{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
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
		return fmt.Errorf("init server: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// srvErr 缓冲 1，避免 Start 失败时 goroutine 因无人接收而泄漏。
	// 启动失败不再直接 os.Exit（那会跳过上面的 defer），而是把 error 送回主流程走统一 shutdown。
	srvErr := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			srvErr <- err
		}
	}()

	// 等待：要么收到信号优雅退出，要么 server 启动/运行出错。
	select {
	case err := <-srvErr:
		return fmt.Errorf("start server: %w", err)
	case <-ctx.Done():
		stop()
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.GracefulTimeout)*time.Second)
	defer cancel()
	srv.Stop(shutdownCtx)
	log.Info("server exited")
	return nil
}
