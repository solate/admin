package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"admin/internal/config"
)

type Server struct {
	httpSrv *http.Server
	db      *gorm.DB
	rdb     *redis.Client
	log     zerolog.Logger
	// Step 05 起追加：cronRunner *cron.Runner
}

type Options struct {
	Config *config.Config
	DB     *gorm.DB
	RDB    *redis.Client
	Log    zerolog.Logger
}

func New(opts Options) (*Server, error) {
	// Step 03 起替换为：engine := gin.New(); router.Setup(engine, ...)
	// gin.SetMode(opts.Config.Server.Mode) // Step 03 接入:把 Server.Mode 喂给 gin
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", opts.Config.Server.Port),
		Handler:      mux,
		ReadTimeout:  opts.Config.Server.ReadTimeout,
		WriteTimeout: opts.Config.Server.WriteTimeout,
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
	s.log.Info().Str("addr", s.httpSrv.Addr).Msg("server starting")
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) {
	// Step 05 起逆序追加：s.cronRunner.Stop()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		s.log.Error().Err(err).Msg("server shutdown error")
	}
}
