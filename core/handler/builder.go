package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/isauran/logger/internal/buffer"
	"go.opentelemetry.io/otel/trace"
)

// Builder provides a fluent API for constructing logger handlers
type Builder struct {
	options    *Options
	writers    []io.Writer
	errHandler func(error)
	levelVar   *slog.LevelVar // Added support for dynamic level changes
}

// NewBuilder creates a new logger builder
func NewBuilder() *Builder {
	return &Builder{
		options: NewOptions(),
		writers: []io.Writer{os.Stdout},
	}
}

// WithJSON enables JSON formatting
func (b *Builder) WithJSON() *Builder {
	b.options.JSON = true
	return b
}

// WithLevel sets the minimum log level
func (b *Builder) WithLevel(level slog.Level) *Builder {
	b.options.Level = level
	return b
}

// WithLevelVar sets a dynamic log level
func (b *Builder) WithLevelVar(lv *slog.LevelVar) *Builder {
	b.levelVar = lv
	if lv != nil {
		b.options.Level = lv
	}
	return b
}

// WithTimeFormat sets the time format
func (b *Builder) WithTimeFormat(format string) *Builder {
	b.options.TimeFormat = format
	return b
}

// WithSource enables source location
func (b *Builder) WithSource() *Builder {
	b.options.AddSource = true
	return b
}

// WithWriter adds an output writer
func (b *Builder) WithWriter(w io.Writer) *Builder {
	if w != nil {
		b.writers = append(b.writers, w)
	}
	return b
}

// WithFile adds a file output
func (b *Builder) WithFile(path string, maxSize int64, maxAge int) *Builder {
	b.options.FileEnabled = true
	b.options.FilePath = path
	b.options.MaxFileSize = maxSize
	b.options.MaxAge = maxAge
	return b
}

// WithSampling enables log sampling
func (b *Builder) WithSampling(interval time.Duration, rate int) *Builder {
	b.options.SamplingEnabled = true
	if interval > 0 {
		b.options.SampleInterval = interval
	}
	if rate > 0 {
		b.options.SampleRate = rate
	}
	return b
}

// WithMetrics enables metrics collection
func (b *Builder) WithMetrics() *Builder {
	b.options.MetricsEnabled = true
	return b
}

// WithTracing enables OpenTelemetry tracing
func (b *Builder) WithTracing(tp trace.TracerProvider) *Builder {
	if tp != nil {
		b.options.TracingEnabled = true
	}
	return b
}

// WithErrorHandler sets a custom error handler
func (b *Builder) WithErrorHandler(f func(error)) *Builder {
	b.errHandler = f
	return b
}

// WithReplaceAttr sets a function to customize how attributes are logged
func (b *Builder) WithReplaceAttr(fn func(groups []string, a slog.Attr) slog.Attr) *Builder {
	b.options.ReplaceAttr = fn
	return b
}

// Build constructs the final handler
func (b *Builder) Build() (slog.Handler, error) {
	// Validate options
	if err := b.options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Create base handler for each writer
	var handlers []slog.Handler
	for _, w := range b.writers {
		h := newBaseHandler(w, b.options)
		handlers = append(handlers, h)
	}

	// Add file handler if enabled
	if b.options.FileEnabled {
		fh, err := NewFileHandler(FileOptions{
			Path:       b.options.FilePath,
			MaxSize:    b.options.MaxFileSize,
			MaxAge:     b.options.MaxAge,
			MaxBackups: b.options.MaxBackups,
			Interval:   b.options.RotateEvery,
		})
		if err != nil {
			return nil, fmt.Errorf("create file handler: %w", err)
		}
		handlers = append(handlers, fh)
	}

	// Create multi-handler if we have multiple outputs
	var handler slog.Handler
	if len(handlers) > 1 {
		mh := NewMultiHandler(handlers...)
		if b.errHandler != nil {
			mh.WithErrorHandler(b.errHandler)
		}
		handler = mh
	} else {
		handler = handlers[0]
	}

	// Add optional handlers in order
	if b.options.SamplingEnabled {
		handler = NewSamplingHandler(handler, b.options.SampleInterval, uint32(b.options.SampleRate))
	}

	if b.options.MetricsEnabled {
		handler = NewMetricsHandler(handler)
	}

	if b.options.TracingEnabled {
		handler = NewTracingHandler(handler)
	}

	// Add context handler as the outermost wrapper
	handler = NewContextHandler(handler)

	return handler, nil
}

// BaseHandler implements slog.Handler with additional features
type BaseHandler struct {
	out          io.Writer
	opts         *Options
	mu           sync.Mutex
	pool         *buffer.Pool
	attrs        []slog.Attr
	groups       []string
	globalAttrs  []slog.Attr
	globalGroups []string
}

// newBaseHandler creates a new enhanced slog handler
func newBaseHandler(out io.Writer, opts *Options) *BaseHandler {
	if opts == nil {
		opts = NewOptions()
	}

	h := &BaseHandler{
		out:  out,
		opts: opts,
	}

	if opts.BufferPool {
		h.pool = buffer.NewPool(int64(opts.BufferSize), int64(buffer.DefaultMaxBuffers))
	}

	return h
}

// AddGlobalAttrs adds attributes that will be included in all log entries
func (h *BaseHandler) AddGlobalAttrs(attrs ...slog.Attr) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.globalAttrs = append(h.globalAttrs, attrs...)
}

// AddGlobalGroup adds a group name that will be applied to all log entries
func (h *BaseHandler) AddGlobalGroup(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.globalGroups = append(h.globalGroups, name)
}

// formatJSON implements JSON formatting with global attributes and groups
func (h *BaseHandler) formatJSON(buf []byte, r slog.Record) []byte {
	// Use an ordered map to ensure "time" always comes first in JSON
	m := make(map[string]interface{})

	// Add timestamp first
	m["time"] = r.Time.Format(h.opts.TimeFormat)

	// Add level
	m["level"] = r.Level.String()

	// Add message
	if r.Message != "" {
		m["msg"] = r.Message
	}

	// Add source if enabled
	if h.opts.AddSource {
		var pcs [1]uintptr
		if runtime.Callers(3, pcs[:]) == 1 {
			fs := runtime.CallersFrames(pcs[:])
			if frame, _ := fs.Next(); frame.File != "" {
				m["source"] = fmt.Sprintf("%s:%d",
					filepath.Join(filepath.Base(filepath.Dir(frame.File)), filepath.Base(frame.File)),
					frame.Line,
				)
			}
		}
	}

	// Add global attributes first
	for _, attr := range h.globalAttrs {
		addAttrToMap(m, h.globalGroups, attr)
	}

	// Add record attributes
	r.Attrs(func(a slog.Attr) bool {
		groups := append(h.globalGroups, h.groups...)
		addAttrToMap(m, groups, a)
		return true
	})

	// Add handler attributes
	for _, attr := range h.attrs {
		groups := append(h.globalGroups, h.groups...)
		addAttrToMap(m, groups, attr)
	}

	// Marshal to JSON
	b, err := json.Marshal(m)
	if err != nil {
		return append(buf, []byte(fmt.Sprintf("error marshaling JSON: %v", err))...)
	}

	return append(buf, b...)
}

// formatText implements text formatting with global attributes and groups
func (h *BaseHandler) formatText(buf []byte, r slog.Record) []byte {
	// Add timestamp
	buf = append(buf, r.Time.Format(h.opts.TimeFormat)...)
	buf = append(buf, ' ')

	// Add level
	buf = append(buf, r.Level.String()...)
	buf = append(buf, ' ')

	// Add message
	if r.Message != "" {
		buf = append(buf, r.Message...)
		buf = append(buf, ' ')
	}

	// Add source if enabled
	if h.opts.AddSource {
		var pcs [1]uintptr
		if runtime.Callers(3, pcs[:]) == 1 {
			fs := runtime.CallersFrames(pcs[:])
			if frame, _ := fs.Next(); frame.File != "" {
				buf = append(buf, "source="...)
				buf = append(buf, fmt.Sprintf("%s:%d",
					filepath.Join(filepath.Base(filepath.Dir(frame.File)), filepath.Base(frame.File)),
					frame.Line,
				)...)
				buf = append(buf, ' ')
			}
		}
	}

	// Add global attributes first
	for _, attr := range h.globalAttrs {
		buf = appendAttr(buf, h.globalGroups, attr)
	}

	// Add record attributes
	r.Attrs(func(a slog.Attr) bool {
		groups := append(h.globalGroups, h.groups...)
		buf = appendAttr(buf, groups, a)
		return true
	})

	// Add handler attributes
	for _, attr := range h.attrs {
		groups := append(h.globalGroups, h.groups...)
		buf = appendAttr(buf, groups, attr)
	}

	return buf
}

// Helper functions
func addAttrToMap(m map[string]interface{}, groups []string, a slog.Attr) {
	if !a.Equal(slog.Attr{}) {
		key := a.Key
		if len(groups) > 0 {
			for _, g := range groups {
				m = addGroup(m, g)
			}
		}
		m[key] = a.Value.Any()
	}
}

// Enabled implements slog.Handler.Enabled method
func (h *BaseHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func addGroup(m map[string]interface{}, name string) map[string]interface{} {
	if v, ok := m[name]; ok {
		if vm, ok := v.(map[string]interface{}); ok {
			return vm
		}
	}
	nm := make(map[string]interface{})
	m[name] = nm
	return nm
}

func appendAttr(buf []byte, groups []string, a slog.Attr) []byte {
	if !a.Equal(slog.Attr{}) {
		if len(groups) > 0 {
			for _, g := range groups {
				buf = append(buf, g...)
				buf = append(buf, '.')
			}
		}
		buf = append(buf, a.Key...)
		buf = append(buf, '=')
		buf = append(buf, fmt.Sprint(a.Value.Any())...)
		buf = append(buf, ' ')
	}
	return buf
}

// Handle implements slog.Handler.Handle method
func (h *BaseHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	var buf []byte
	if h.pool != nil {
		bufPtr := h.pool.Get()
		defer h.pool.Put(bufPtr)
		buf = (*bufPtr)[:0]
	} else {
		buf = make([]byte, 0, 1024)
	}

	// Format the record according to the configured format
	if h.opts.JSON {
		buf = h.formatJSON(buf, r)
		buf = append(buf, '\n')
	} else {
		buf = h.formatText(buf, r)
		buf = append(buf, '\n')
	}

	// Write to output
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf)
	return err
}

// WithAttrs implements slog.Handler.WithAttrs method
func (h *BaseHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := &BaseHandler{
		out:          h.out,
		opts:         h.opts,
		pool:         h.pool,
		attrs:        append(h.attrs[:], attrs...),
		groups:       h.groups[:],
		globalAttrs:  h.globalAttrs[:],
		globalGroups: h.globalGroups[:],
	}
	return h2
}

// WithGroup implements slog.Handler.WithGroup method
func (h *BaseHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h2 := &BaseHandler{
		out:          h.out,
		opts:         h.opts,
		pool:         h.pool,
		attrs:        h.attrs[:],
		groups:       append(h.groups[:], name),
		globalAttrs:  h.globalAttrs[:],
		globalGroups: h.globalGroups[:],
	}
	return h2
}

// For backward compatibility, expose New function that calls newBaseHandler
func New(out io.Writer, opts *Options) *BaseHandler {
	return newBaseHandler(out, opts)
}
