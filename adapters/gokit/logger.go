package gokit

import (
	"context"
	"log/slog"
	"strings"

	gokitlog "github.com/go-kit/log"
)

type logFunc func(ctx context.Context, msg string, keysAndValues ...interface{})

func (l logFunc) Log(keyvals ...interface{}) error {
	// Extract message if present
	var msg string
	for i := 0; i < len(keyvals)-1; i += 2 {
		if key, ok := keyvals[i].(string); ok && key == "msg" {
			if msgVal, ok := keyvals[i+1].(string); ok {
				msg = msgVal
				// Remove message from keyvals
				keyvals = append(keyvals[:i], keyvals[i+2:]...)
				break
			}
		}
	}
	
	ctx := context.Background()
	l(ctx, msg, keyvals...)
	return nil
}

// NewLogger creates a new Go-kit logger adapter
func NewLogger(lvl string) gokitlog.Logger {
	var logFunc logFunc

	switch strings.ToLower(lvl) {
	case "debug":
		logFunc = slog.Default().DebugContext
	case "info":
		logFunc = slog.Default().InfoContext
	case "warn":
		logFunc = slog.Default().WarnContext
	case "error":
		logFunc = slog.Default().ErrorContext
	default:
		logFunc = slog.Default().InfoContext
	}

	return logFunc
}
