package logger

import (
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
	level      string
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
			opts.level = LevelDebug
		}
		if strings.Contains(strings.ToUpper(level), LevelInfo) {
			opts.level = LevelInfo
		}
		if strings.Contains(strings.ToUpper(level), LevelWarn) {
			opts.level = LevelWarn
		}
		if strings.Contains(strings.ToUpper(level), LevelError) {
			opts.level = LevelError
		}
	}
}

func WithTimeFormat(layout string) Option {
	return func(opts *loggerOptions) {
		opts.timeFormat = layout
	}
}

func LoggerOptions(options ...Option) *loggerOptions {
	opts := &loggerOptions{
		json:       false,
		level:      LevelInfo,
		timeFormat: time.RFC3339,
	}

	for _, opt := range options {
		opt(opts)
	}
	return opts
}
