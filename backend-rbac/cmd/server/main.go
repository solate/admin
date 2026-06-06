package main

import (
	"admin/internal/router"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
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

func main() {
	// 初始化应用
	app, err := router.NewApp()
	if err != nil {
		fmt.Printf("Failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	// 启动服务器
	go func() {
		if err := app.Run(); err != nil {
			fmt.Printf("Failed to run app: %v\n", err)
			os.Exit(1)
		}
	}()

	// 等待关闭信号
	waitForShutdown()

	// 关闭应用
	if err := app.Close(); err != nil {
		fmt.Printf("Failed to close app: %v\n", err)
		os.Exit(1)
	}

	log.Info().Msg("Server stopped gracefully")
}

// waitForShutdown 等待关闭信号
func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")
}
