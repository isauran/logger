package slogger

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
)

const (
	LevelDebug string = "DEBUG"
	LevelInfo  string = "INFO"
	LevelWarn  string = "WARN"
	LevelError string = "ERROR"
)

type Option func(*loggerOptions)

type loggerOptions struct {
	json       bool
	level      slog.Level
	timeFormat string
}

func WithJSON(json bool) Option {
	return func(opts *loggerOptions) {
		opts.json = json
	}
}

func WithLevel(level string) Option {
	return func(opts *loggerOptions) {
		if strings.Contains(strings.ToUpper(level), LevelDebug) {
			opts.level = slog.LevelDebug
		}
		if strings.Contains(strings.ToUpper(level), LevelInfo) {
			opts.level = slog.LevelInfo
		}
		if strings.Contains(strings.ToUpper(level), LevelWarn) {
			opts.level = slog.LevelWarn
		}
		if strings.Contains(strings.ToUpper(level), LevelError) {
			opts.level = slog.LevelError
		}
	}
}

func WithTimeFormat(layout string) Option {
	return func(opts *loggerOptions) {
		opts.timeFormat = layout
	}
}

// logger := slogger.NewLogger(os.Stdout, slogger.WithJSON(true))
func NewLogger(w io.Writer, options ...Option) *slog.Logger {
	opts := &loggerOptions{
		json:       false,
		level:      slog.LevelInfo,
		timeFormat: time.RFC3339,
	}

	for _, opt := range options {
		opt(opts)
	}

	hOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     opts.level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				s, _ := a.Value.Any().(*slog.Source)
				if s != nil {
					return slog.String("caller", fmt.Sprintf("%s:%d", filepath.Base(s.File), s.Line))
				}
			}
			if a.Key == slog.TimeKey {
				return slog.String("time", time.Now().Format(opts.timeFormat))
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
