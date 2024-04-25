package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

var gormSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	gormSourceDir = sourceDir(file)
}

func sourceDir(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "gorm.io" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

// FileWithLineNum return the file name and line number of the current file
func gormLogCaller() {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, gormSourceDir) || strings.HasSuffix(file, "_test.go")) &&
			!strings.HasSuffix(file, ".gen.go") {
			CallerSource(file, line)
		}
	}
}

var _ logger.Interface = (*gormLogger)(nil)

// logger.NewLogger(os.Stdout, logger.WithJSON(true))
// gormlog := logger.NewGormLogger("info")
// gormlog.Info(context.Background(), "init", "logger", "gorm.io/gorm/logger", "format", "text")
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
		l.LogLevel = logger.Info
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
	defer ResetCallerSource()
	gormLogCaller()
	if l.LogLevel >= logger.Info {
		slog.Info(msg, data...)
	}
}

// Warn print warn messages
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	defer ResetCallerSource()
	gormLogCaller()
	if l.LogLevel >= logger.Warn {
		slog.Warn(msg, data...)
	}
}

// Error print error messages
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	defer ResetCallerSource()
	gormLogCaller()
	if l.LogLevel >= logger.Error {
		slog.Error(msg, data...)
	}
}

// Trace print sql message
//
//nolint:cyclop
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	defer ResetCallerSource()
	gormLogCaller()
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			slog.Debug(err.Error(), "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "sql", sql)
		} else {
			slog.Debug(err.Error(), "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "rows", rows, "sql", sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			slog.Debug(slowLog, "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "sql", sql)
		} else {
			slog.Debug(slowLog, "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "rows", rows, "sql", sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			slog.Debug("", "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "sql", sql)
		} else {
			slog.Debug("", "ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6), "rows", rows, "sql", sql)
		}
	}
}
