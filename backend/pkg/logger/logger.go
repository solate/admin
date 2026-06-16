package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(cfg Config) zerolog.Logger {
	var output io.Writer = os.Stdout

	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	return zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}
