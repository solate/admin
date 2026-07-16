package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"admin/internal/config"
	"admin/internal/router"
)

type Server struct {
	httpSrv *http.Server
	db      *gorm.DB
	rdb     *redis.Client
	log     *slog.Logger
	// Step 05 起追加：cronRunner *cron.Runner
}

type Options struct {
	Config *config.Config
	DB     *gorm.DB
	RDB    *redis.Client
	Log    *slog.Logger
}

func New(opts Options) (*Server, error) {
	// 按配置设置 gin 运行模式（debug/release/test）
	gin.SetMode(opts.Config.Server.Mode)

	// gin.New() 不带默认中间件，中间件全部由 router.Setup 显式注册
	engine := gin.New()
	router.Setup(engine, opts.Config)

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", opts.Config.Server.Port),
		Handler:      engine,
		ReadTimeout:  time.Duration(opts.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(opts.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return &Server{
		httpSrv: httpSrv,
		db:      opts.DB,
		rdb:     opts.RDB,
		log:     opts.Log,
	}, nil
}

func (s *Server) Start() error {
	// Step 05 起追加：go s.cronRunner.Start(ctx)
	s.log.Info("server starting", slog.String("addr", s.httpSrv.Addr))
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	// Step 05 起逆序追加：s.cronRunner.Stop()
	err := s.httpSrv.Shutdown(ctx)
	if err != nil {
		s.log.Error("server shutdown error", slog.Any("err", err))
	}
	return err
}
