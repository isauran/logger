package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"
)

// ErrorHandler adds enhanced error handling capabilities
type ErrorHandler struct {
	handler    slog.Handler
	stackTrace bool
	errorAttrs []string
	skipFrames int
	errorHook  func(error)
	mu         sync.RWMutex // Added mutex for thread safety
}

// ErrorHandlerOption configures the error handler
type ErrorHandlerOption func(*ErrorHandler)

// NewErrorHandler creates a new error handler
func NewErrorHandler(handler slog.Handler, opts ...ErrorHandlerOption) *ErrorHandler {
	if handler == nil {
		panic("handler is required")
	}

	h := &ErrorHandler{
		handler:    handler,
		stackTrace: true,
		errorAttrs: []string{"error", "err"},
		skipFrames: 2,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// WithStackTrace enables or disables stack trace collection
func WithStackTrace(enabled bool) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.stackTrace = enabled
	}
}

// WithErrorAttributes sets custom error attribute names to look for
func WithErrorAttributes(attrs []string) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.errorAttrs = attrs
	}
}

// WithErrorHook sets a hook function to be called for each error
func WithErrorHook(hook func(error)) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.errorHook = hook
	}
}

func (h *ErrorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ErrorHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var foundErr error
	var attrs []slog.Attr

	// Pre-allocate attrs slice to avoid reallocations
	attrs = make([]slog.Attr, 0, r.NumAttrs()+4)

	// Process existing attributes first
	r.Attrs(func(attr slog.Attr) bool {
		if err := h.extractError(attr); err != nil && foundErr == nil {
			foundErr = err
		}
		attrs = append(attrs, attr)
		return true
	})

	if foundErr != nil {
		// Add error context efficiently
		attrs = h.errorAttrsFromError(foundErr, attrs)

		// Call error hook if configured
		if h.errorHook != nil {
			h.errorHook(foundErr)
		}
	}

	// Create new record with all attributes
	enhanced := slog.Record{
		Time:    r.Time,
		Level:   r.Level,
		Message: r.Message,
		PC:      r.PC,
	}
	enhanced.AddAttrs(attrs...)

	return h.handler.Handle(ctx, enhanced)
}

func (h *ErrorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewErrorHandler(h.handler.WithAttrs(attrs),
		WithStackTrace(h.stackTrace),
		WithErrorAttributes(h.errorAttrs),
		WithErrorHook(h.errorHook),
	)
}

func (h *ErrorHandler) WithGroup(name string) slog.Handler {
	return NewErrorHandler(h.handler.WithGroup(name),
		WithStackTrace(h.stackTrace),
		WithErrorAttributes(h.errorAttrs),
		WithErrorHook(h.errorHook),
	)
}

func (h *ErrorHandler) extractError(attr slog.Attr) error {
	for _, name := range h.errorAttrs {
		if attr.Key == name {
			switch v := attr.Value.Any().(type) {
			case error:
				return v
			case string:
				return errors.New(v)
			}
		}
	}
	return nil
}

func (h *ErrorHandler) errorAttrsFromError(err error, attrs []slog.Attr) []slog.Attr {
	attrs = append(attrs,
		slog.String("error.type", fmt.Sprintf("%T", err)),
		slog.String("error.message", err.Error()),
	)

	if h.stackTrace {
		attrs = append(attrs, slog.String("error.stack", h.captureStack()))
	}

	// Handle wrapped errors efficiently
	for unwrapped := errors.Unwrap(err); unwrapped != nil; unwrapped = errors.Unwrap(unwrapped) {
		attrs = append(attrs,
			slog.String("error.cause.type", fmt.Sprintf("%T", unwrapped)),
			slog.String("error.cause.message", unwrapped.Error()),
		)
	}

	return attrs
}

func (h *ErrorHandler) captureStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(h.skipFrames, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		if frame.Function != "" {
			fmt.Fprintf(&builder, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		}
	}
	return builder.String()
}
