package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var globalLogger zerolog.Logger

// Config 日志配置
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, console
}

// Init 初始化日志
func Init(cfg Config) error {
	// 设置日志级别
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// 设置时间格式
	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer = os.Stdout

	// 设置格式
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	globalLogger = zerolog.New(output).With().Timestamp().Caller().Logger()
	log.Logger = globalLogger

	return nil
}

// Get 获取全局logger
func Get() *zerolog.Logger {
	return &globalLogger
}
