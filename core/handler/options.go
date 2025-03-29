package handler

import (
	"fmt"
	"log/slog"
	"time"
)

// Options configures handler behavior
type Options struct {
	// General options
	JSON       bool
	Level      slog.Leveler // Changed to slog.Leveler to support LevelVar
	TimeFormat string
	AddSource  bool
	BufferPool bool

	// ReplaceAttr allows customizing how attributes are logged
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr

	// Performance options
	BufferSize int
	PoolSize   int

	// Sampling options
	SamplingEnabled bool
	SampleInterval  time.Duration
	SampleRate      int

	// File options
	FileEnabled bool
	FilePath    string
	MaxFileSize int64
	MaxAge      int
	MaxBackups  int
	RotateEvery time.Duration

	// Metrics options
	MetricsEnabled bool

	// Tracing options
	TracingEnabled bool
}

// Validate checks if the options are valid
func (o *Options) Validate() error {
	if o.BufferPool && (o.BufferSize <= 0 || o.PoolSize <= 0) {
		return fmt.Errorf("invalid buffer pool configuration: size=%d, pool=%d", o.BufferSize, o.PoolSize)
	}

	if o.SamplingEnabled && (o.SampleInterval <= 0 || o.SampleRate <= 0) {
		return fmt.Errorf("invalid sampling configuration: interval=%v, rate=%d", o.SampleInterval, o.SampleRate)
	}

	if o.FileEnabled {
		if o.FilePath == "" {
			return fmt.Errorf("file path is required for file handler")
		}
		if o.MaxFileSize <= 0 {
			return fmt.Errorf("invalid max file size: %d", o.MaxFileSize)
		}
		if o.MaxAge <= 0 {
			return fmt.Errorf("invalid max age: %d", o.MaxAge)
		}
		if o.MaxBackups <= 0 {
			return fmt.Errorf("invalid max backups: %d", o.MaxBackups)
		}
		if o.RotateEvery <= 0 {
			return fmt.Errorf("invalid rotate interval: %v", o.RotateEvery)
		}
	}

	return nil
}

// NewOptions creates Options with default values
func NewOptions() *Options {
	return &Options{
		// General defaults
		JSON:       false,
		Level:      slog.LevelInfo,
		TimeFormat: time.RFC3339,
		AddSource:  true,
		BufferPool: true,

		// Performance defaults
		BufferSize: 4096,
		PoolSize:   32,

		// Sampling defaults
		SamplingEnabled: false,
		SampleInterval:  time.Second,
		SampleRate:      10,

		// File defaults
		FileEnabled: false,
		MaxFileSize: 100, // MB
		MaxAge:      7,   // days
		MaxBackups:  5,
		RotateEvery: 24 * time.Hour,

		// Feature flags
		MetricsEnabled: false,
		TracingEnabled: false,
	}
}

// Option is a function that modifies Options
type Option func(*Options)

// WithReplaceAttr sets a function to customize how attributes are logged
func WithReplaceAttr(fn func(groups []string, a slog.Attr) slog.Attr) Option {
	return func(o *Options) {
		o.ReplaceAttr = fn
	}
}

// WithJSON sets JSON output format
func WithJSON(enabled bool) Option {
	return func(o *Options) {
		o.JSON = enabled
	}
}

// WithLevel sets the minimum log level
func WithLevel(level slog.Level) Option {
	return func(o *Options) {
		o.Level = level
	}
}

// WithTimeFormat sets the time format
func WithTimeFormat(format string) Option {
	return func(o *Options) {
		o.TimeFormat = format
	}
}

// WithSource enables source location
func WithSource(enabled bool) Option {
	return func(o *Options) {
		o.AddSource = enabled
	}
}

// WithBufferPool enables buffer pooling
func WithBufferPool(enabled bool, size, poolSize int) Option {
	return func(o *Options) {
		o.BufferPool = enabled
		if size > 0 {
			o.BufferSize = size
		}
		if poolSize > 0 {
			o.PoolSize = poolSize
		}
	}
}

// WithSampling enables log sampling
func WithSampling(interval time.Duration, rate int) Option {
	return func(o *Options) {
		o.SamplingEnabled = true
		o.SampleInterval = interval
		o.SampleRate = rate
	}
}

// WithFile enables file output
func WithFile(path string, maxSize int64, maxAge, maxBackups int, rotateEvery time.Duration) Option {
	return func(o *Options) {
		o.FileEnabled = true
		o.FilePath = path
		o.MaxFileSize = maxSize
		o.MaxAge = maxAge
		o.MaxBackups = maxBackups
		o.RotateEvery = rotateEvery
	}
}

// WithMetrics enables metrics collection
func WithMetrics(enabled bool) Option {
	return func(o *Options) {
		o.MetricsEnabled = enabled
	}
}

// WithTracing enables distributed tracing
func WithTracing(enabled bool) Option {
	return func(o *Options) {
		o.TracingEnabled = enabled
	}
}
