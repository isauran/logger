package logger

import (
	"io"

	gokitlog "github.com/go-kit/log"
)

type logFunc func(msg string, keysAndValues ...interface{})

func (l logFunc) Log(keyvals ...interface{}) error {
	l("", keyvals...)
	return nil
}

// logger := logger.NewGoKitLogger(os.Stdout, gokitlogger.WithJSON(true))
func NewGoKitLogger(w io.Writer, options ...Option) gokitlog.Logger {
	opts := LoggerOptions(options...)
	logger := NewLogger(w, options...)

	var logFunc logFunc
	switch opts.level {
	case LevelDebug:
		logFunc = logger.Debug
	case LevelInfo:
		logFunc = logger.Info
	case LevelWarn:
		logFunc = logger.Warn
	case LevelError:
		logFunc = logger.Error
	default:
		logFunc = logger.Info
	}

	return logFunc
}
