package logger

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

var source *slog.Source

func CallerSource(file string, line int) {
	if file != "" {
		source = &slog.Source{
			File: file,
			Line: line,
		}
	}
}

func DefaultCallerSource() {
	_, file, line, _ := runtime.Caller(2)
	CallerSource(file, line)
}

func ResetCallerSource() {
	source = nil
}

// logger.NewLogger(os.Stdout, logger.WithJSON(true))
// slog.Info("init", "logger", "log/slog", "format", "json")
func NewLogger(w io.Writer, options ...Option) *slog.Logger {
	opts := LoggerOptions(options...)

	var level slog.Level
	switch opts.level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo
	}

	hOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if source != nil {
					return slog.String("caller", fmt.Sprintf("%s/%s:%d", filepath.Base(filepath.Dir(source.File)), filepath.Base(source.File), source.Line))
				}

				s, _ := a.Value.Any().(*slog.Source)
				if s != nil {
					return slog.String("caller", fmt.Sprintf("%s/%s:%d", filepath.Base(filepath.Dir(s.File)), filepath.Base(s.File), s.Line))
				}
			}
			if a.Key == slog.TimeKey {
				return slog.String("time", time.Now().Format(opts.timeFormat))
			}
			if a.Key == slog.MessageKey {
				if len(a.Value.String()) == 0 {
					return slog.Attr{}
				}
			}
			return a
		},
	}

	var h interface{}
	if opts.json {
		h = slog.NewJSONHandler(w, &hOpts)
	} else {
		h = slog.NewTextHandler(w, &hOpts)
	}

	var l *slog.Logger
	if opts.json {
		l = slog.New(h.(*slog.JSONHandler))
	} else {
		l = slog.New(h.(*slog.TextHandler))
	}

	slog.SetDefault(l)
	return l
}
