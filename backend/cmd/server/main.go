package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"admin/internal/config"
	"admin/internal/jobs"
	"admin/internal/server"
	"admin/pkg/xcron"
	"admin/pkg/xgorm"
	"admin/pkg/xredis"
	"admin/pkg/xslog"
)

func main() {
	// main 只负责退出码，逻辑与资源清理都在 run 里（defer 保证执行）
	if err := run(); err != nil {
		slog.Error("server exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}

// run 承载全部启动逻辑，分两层：
//   - 第一层 基础设施（被动资源）：config/log/db/redis，defer 逆序 Close
//   - 第二层 长驻组件（主动 goroutine）：HTTP server + cron scheduler，main 层 errgroup 编排
func run() error {
	// ===== 第一层：基础设施（被动资源）—— defer 逆序 Close =====

	// 1. 加载配置
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
	slog.SetDefault(log) // 设为全局默认，后续直接用 slog.Info/Error

	// 3. 连接数据库（xgorm.Config 与 config.DatabaseConfig 同构，直接类型转换）
	db, err := xgorm.New(xgorm.Config(cfg.Database), log)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer xgorm.Close(db)

	// 4. 连接 Redis
	rdbClient, err := xredis.New(xredis.Config(cfg.Redis))
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer rdbClient.Close()

	log.Info("all infrastructure initialized", slog.Int("port", cfg.Server.Port))

	// ===== 第二层：长驻组件（主动 goroutine）—— main 层 errgroup 编排 =====

	// 5. 创建 cron 调度器并注册静态任务（在 Start 之前注册）
	c := xcron.New()
	jobs.Register(c, jobs.Deps{
		DB:  db,
		RDB: rdbClient,
	})

	// 6. 组装 HTTP server
	srv, err := server.New(server.Options{
		Config: cfg,
		DB:     db,
		RDB:    rdbClient,
		Log:    log,
	})
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}

	// 7. signal.NotifyContext 把 SIGINT/SIGTERM 变成可传播的 ctx
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 8. errgroup 编排：任一组件出错 → ctx 取消 → 其余优雅关闭
	g, ctx := errgroup.WithContext(ctx)

	// 组件 1：HTTP server（阻塞）
	g.Go(func() error {
		return srv.Start()
	})
	// 组件 1 优雅关闭：ctx 取消后带超时排空在途请求
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulTimeout)
		defer cancel()
		return srv.Stop(shutdownCtx)
	})

	// 组件 2：Cron scheduler（阻塞，ctx 取消时自动等待运行中任务完成）
	g.Go(func() error {
		return xcron.Run(ctx, c)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	log.Info("all components exited")
	return nil
}
