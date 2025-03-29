package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/isauran/logger/core/handler"
)

// New creates a new logger with default configuration
func New() *slog.Logger {
	return NewWithWriter(os.Stdout)
}

// NewWithWriter creates a new logger with the specified writer
func NewWithWriter(w io.Writer) *slog.Logger {
	h := handler.New(w, nil)
	return slog.New(h)
}

// NewFromEnv creates a new logger from environment variables
func NewFromEnv() *slog.Logger {
	return New() // Simplified for now as LoadEnv is not implemented
}

// ParseLevel converts a level string to slog.Level
func ParseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
