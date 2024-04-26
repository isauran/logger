package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var _ logger.Interface = (*gormLogger)(nil)

// logger.NewLogger(os.Stdout, logger.WithJSON(true))
// logger := logger.NewGormLogger("info")
func NewGormLogger(level string) logger.Interface {
	l := &gormLogger{}

	switch {
	case strings.EqualFold(level, LevelDebug):
		l.LogLevel = logger.Info
	case strings.EqualFold(level, LevelInfo):
		l.LogLevel = logger.Info
	case strings.EqualFold(level, LevelWarn):
		l.LogLevel = logger.Warn
	case strings.EqualFold(level, LevelError):
		l.LogLevel = logger.Error
	default:
		l.LogLevel = logger.Silent
	}

	return l
}

type gormLogger struct {
	logger.Config
}

// LogMode log mode
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		defer ResetCallerSource()
		DefaultCallerSource()
		fileLine := filepath.Base(utils.FileWithLineNum())
		parts := strings.Split(fileLine, ":")
		if len(parts) == 2 {
			file := parts[0]
			line, _ := strconv.Atoi(parts[1])
			CallerSource(file, line)
		}

		slog.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn print warn messages
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		defer ResetCallerSource()
		DefaultCallerSource()
		fileLine := filepath.Base(utils.FileWithLineNum())
		parts := strings.Split(fileLine, ":")
		if len(parts) == 2 {
			file := parts[0]
			line, _ := strconv.Atoi(parts[1])
			CallerSource(file, line)
		}

		slog.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error print error messages
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		defer ResetCallerSource()
		DefaultCallerSource()
		fileLine := filepath.Base(utils.FileWithLineNum())
		parts := strings.Split(fileLine, ":")
		if len(parts) == 2 {
			file := parts[0]
			line, _ := strconv.Atoi(parts[1])
			CallerSource(file, line)
		}
		
		slog.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace print sql message
//
//nolint:cyclop
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	defer ResetCallerSource()
	DefaultCallerSource()
	fileLine := filepath.Base(utils.FileWithLineNum())
	parts := strings.Split(fileLine, ":")
	if len(parts) == 2 {
		file := parts[0]
		line, _ := strconv.Atoi(parts[1])
		CallerSource(file, line)
	}
	
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			slog.Error(err.Error(), "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "sql", sql)
		} else {
			slog.Error(err.Error(), "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "rows", rows, "sql", sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			slog.Warn(slowLog, "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "sql", sql)
		} else {
			slog.Warn(slowLog, "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "rows", rows, "sql", sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			slog.Info("", "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "sql", sql)
		} else {
			slog.Info("", "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "rows", rows, "sql", sql)
		}
	}
}
