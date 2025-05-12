package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Level  string
	Format string
}

func Setup(cfg Config) {
	zerolog.TimeFieldFormat = time.RFC3339

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
}

func Debug(msg string, fields ...any) {
	log.Debug().Fields(fields).Msg(msg)
}

func Info(msg string, fields ...any) {
	log.Info().Fields(fields).Msg(msg)
}

func Warn(msg string, fields ...any) {
	log.Warn().Fields(fields).Msg(msg)
}

func Error(msg string, err error, fields ...any) {
	if err != nil {
		fields = append(fields, "error", err.Error())
	}
	log.Error().Fields(fields).Msg(msg)
}

func Fatal(msg string, err error, fields ...any) {
	if err != nil {
		fields = append(fields, "error", err.Error())
	}
	log.Fatal().Fields(fields).Msg(msg)
}

func WithContext(fields ...any) zerolog.Logger {
	return log.With().Fields(fields).Logger()
}
