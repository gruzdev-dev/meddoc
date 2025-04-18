package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config represents logger configuration
type Config struct {
	Level  string
	Format string
}

// Setup configures the global logger
func Setup(cfg Config) {
	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer = os.Stdout
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Configure global logger
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
}

// Debug logs a debug message
func Debug(msg string, fields ...interface{}) {
	log.Debug().Fields(fields).Msg(msg)
}

// Info logs an info message
func Info(msg string, fields ...interface{}) {
	log.Info().Fields(fields).Msg(msg)
}

// Warn logs a warning message
func Warn(msg string, fields ...interface{}) {
	log.Warn().Fields(fields).Msg(msg)
}

// Error logs an error message
func Error(msg string, err error, fields ...interface{}) {
	if err != nil {
		fields = append(fields, "error", err.Error())
	}
	log.Error().Fields(fields).Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error, fields ...interface{}) {
	if err != nil {
		fields = append(fields, "error", err.Error())
	}
	log.Fatal().Fields(fields).Msg(msg)
}

// WithContext creates a new logger with context
func WithContext(fields ...interface{}) zerolog.Logger {
	return log.With().Fields(fields).Logger()
}
