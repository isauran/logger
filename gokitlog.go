package logger

import (
	"log/slog"
	"strings"

	gokitlog "github.com/go-kit/log"
)

type logFunc func(msg string, keysAndValues ...interface{})

func (l logFunc) Log(keyvals ...interface{}) error {
	defer ResetCallerSource()
	DefaultCallerSource()
	l("", keyvals...)

	return nil
}

// logger.NewLogger(os.Stdout, logger.WithJSON(true))
// logger := logger.NewGoKitLogger("info")
func NewGoKitLogger(level string) gokitlog.Logger {
	var logFunc logFunc
	switch {
	case strings.EqualFold(level, LevelDebug):
		logFunc = slog.Default().Debug
	case strings.EqualFold(level, LevelInfo):
		logFunc = slog.Default().Info
	case strings.EqualFold(level, LevelWarn):
		logFunc = slog.Default().Warn
	case strings.EqualFold(level, LevelError):
		logFunc = slog.Default().Error
	default:
		logFunc = slog.Default().Info
	}

	return logFunc
}
