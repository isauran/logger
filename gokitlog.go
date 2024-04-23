package logger

import (
	"log/slog"
	"runtime"
	"strings"

	gokitlog "github.com/go-kit/log"
)

type logFunc func(msg string, keysAndValues ...interface{})

func (l logFunc) Log(keyvals ...interface{}) error {
	defer ResetCaller()
	_, file, line, _ := runtime.Caller(1)
	Caller(file, line)
	l("", keyvals...)

	return nil
}

// logger.NewLogger(os.Stdout, logger.WithJSON(true))
// gokitlog := logger.NewGoKitLogger("info")
// gokitlog.Log("msg", "init", "logger", "go-kit/log", "format", "json")
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
