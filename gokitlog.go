package logger

import (
	"context"
	"log/slog"
	"strings"

	gokitlog "github.com/go-kit/log"
)

type logFunc func(ctx context.Context, msg string, keysAndValues ...interface{})

func (l logFunc) Log(keyvals ...interface{}) error {
	ctx := SourceContext(context.Background(), CallerSource(2))
	l(ctx, "", keyvals...)

	return nil
}

// logger.NewLogger(os.Stdout, logger.WithJSON(true))
// logger := logger.NewGoKitLogger("info")
func NewGoKitLogger(level string) gokitlog.Logger {
	var logFunc logFunc
	switch {
	case strings.EqualFold(level, LevelDebug):
		logFunc = slog.Default().DebugContext
	case strings.EqualFold(level, LevelInfo):
		logFunc = slog.Default().InfoContext
	case strings.EqualFold(level, LevelWarn):
		logFunc = slog.Default().WarnContext
	case strings.EqualFold(level, LevelError):
		logFunc = slog.Default().ErrorContext
	default:
		logFunc = slog.Default().InfoContext
	}

	return logFunc
}
