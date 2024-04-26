package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

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

	hOpts := &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if s, ok := a.Value.Any().(*slog.Source); ok {
					if s != nil {
						return slog.String("caller", fmt.Sprintf("%s/%s:%d", filepath.Base(filepath.Dir(s.File)), filepath.Base(s.File), s.Line))
					}
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
		h = slog.NewJSONHandler(w, hOpts)
	} else {
		h = slog.NewTextHandler(w, hOpts)
	}

	keys := []any{
		sourceKey{}, 
	}

	var l *slog.Logger
	if opts.json {
		enc := h.(*slog.JSONHandler)
		h := ContextHandler{enc, keys}
		l = slog.New(h)
	} else {
		enc := h.(*slog.TextHandler)
		h := ContextHandler{enc, keys}
		l = slog.New(h)
	}

	slog.SetDefault(l)
	return l
}

type ContextHandler struct {
	slog.Handler
	keys []any
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx.Value(sourceKey{}) == nil {
		r.Add(slog.SourceKey, CallerSource(4))
	}
	r.AddAttrs(h.observe(ctx)...)
	return h.Handler.Handle(ctx, r)
}

func (h ContextHandler) observe(ctx context.Context) (as []slog.Attr) {
	for _, k := range h.keys {
		a, ok := ctx.Value(k).(slog.Attr)
		if !ok {
			continue
		}
		a.Value = a.Value.Resolve()
		as = append(as, a)
	}
	return
}

func SourceContext(ctx context.Context, s *slog.Source) context.Context {
	return context.WithValue(ctx, sourceKey{}, slog.Any(slog.SourceKey, s))
}

func CallerSource(skip int) *slog.Source {
	_, file, line, _ := runtime.Caller(skip)
	return &slog.Source{File: file, Line: line}
}

type sourceKey struct{}
