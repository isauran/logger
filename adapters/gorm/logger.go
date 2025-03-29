package gorm

import (
	stdctx "context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/isauran/logger/internal/context" // internal context package for source info
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Logger struct {
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
}

func NewLogger(lvl string) logger.Interface {
	gormLogger := &Logger{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: true,
	}

	switch lvl {
	case "debug":
		gormLogger.LogLevel = logger.Info // Changed from Silent to Info to allow debug logs
	case "info":
		gormLogger.LogLevel = logger.Info
	case "warn", "warning":
		gormLogger.LogLevel = logger.Warn
	case "error":
		gormLogger.LogLevel = logger.Error
	default:
		gormLogger.LogLevel = logger.Info
	}

	return gormLogger
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *Logger) Info(ctx stdctx.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Info {
		ctx = context.WithSource(ctx, 2)
		slog.InfoContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *Logger) Warn(ctx stdctx.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Warn {
		ctx = context.WithSource(ctx, 2)
		slog.WarnContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *Logger) Error(ctx stdctx.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Error {
		ctx = context.WithSource(ctx, 2)
		slog.ErrorContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *Logger) Trace(ctx stdctx.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		if rows == -1 {
			slog.ErrorContext(ctx, err.Error(),
				"elapsed", elapsed.String(),
				"sql", sql,
			)
		} else {
			slog.ErrorContext(ctx, err.Error(),
				"elapsed", elapsed.String(),
				"rows", rows,
				"sql", sql,
			)
		}

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			slog.WarnContext(ctx, slowLog,
				"elapsed", elapsed.String(),
				"sql", sql,
			)
		} else {
			slog.WarnContext(ctx, slowLog,
				"elapsed", elapsed.String(),
				"rows", rows,
				"sql", sql,
			)
		}

	case l.LogLevel >= logger.Info:
		if rows == -1 {
			slog.InfoContext(ctx, "SQL Query",
				"elapsed", elapsed.String(),
				"sql", sql,
			)
		} else {
			slog.InfoContext(ctx, "SQL Query",
				"elapsed", elapsed.String(),
				"rows", rows,
				"sql", sql,
			)
		}
	}
}
