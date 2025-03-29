package std

import (
	"log"
	"log/slog"
)

// StdLogger implements standard log.Logger interface using slog
type StdLogger struct {
	level  slog.Level
	logger *slog.Logger
}

// NewStdLogger creates a new standard logger adapter
func NewStdLogger(logger *slog.Logger, lvl slog.Level) *log.Logger {
	sl := &StdLogger{
		level:  lvl,
		logger: logger,
	}

	return log.New(sl, "", 0)
}

// Write implements io.Writer for log.Logger compatibility
func (l *StdLogger) Write(p []byte) (n int, err error) {
	msg := string(p)
	switch l.level {
	case slog.LevelDebug:
		l.logger.Debug(msg)
	case slog.LevelWarn:
		l.logger.Warn(msg)
	case slog.LevelError:
		l.logger.Error(msg)
	default:
		l.logger.Info(msg)
	}
	return len(p), nil
}

// Helper function to create a logger with prefix
func NewPrefixLogger(logger *slog.Logger, lvl slog.Level, prefix string) *log.Logger {
	sl := &StdLogger{
		level:  lvl,
		logger: logger.With("prefix", prefix),
	}

	return log.New(sl, "", 0)
}
