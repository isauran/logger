package handler

import (
	"context"
	"log/slog"
)

// ContextExtractor is a function that extracts attributes from a context
type ContextExtractor func(context.Context) []slog.Attr

// ContextHandler automatically adds context information to log records
type ContextHandler struct {
	handler    slog.Handler
	extractors []ContextExtractor
}

// NewContextHandler creates a new context handler
func NewContextHandler(handler slog.Handler, extractors ...ContextExtractor) *ContextHandler {
	return &ContextHandler{
		handler:    handler,
		extractors: extractors,
	}
}

// AddExtractor adds a new context extractor
func (h *ContextHandler) AddExtractor(extractor ContextExtractor) {
	h.extractors = append(h.extractors, extractor)
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract context attributes
	var attrs []slog.Attr
	for _, extractor := range h.extractors {
		if ctxAttrs := extractor(ctx); len(ctxAttrs) > 0 {
			attrs = append(attrs, ctxAttrs...)
		}
	}

	// Add extracted attributes to the record
	if len(attrs) > 0 {
		r2 := cloneRecord(r)
		for _, attr := range attrs {
			r2.AddAttrs(attr)
		}
		return h.handler.Handle(ctx, r2)
	}

	return h.handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{
		handler:    h.handler.WithAttrs(attrs),
		extractors: h.extractors,
	}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{
		handler:    h.handler.WithGroup(name),
		extractors: h.extractors,
	}
}

// cloneRecord creates a copy of a slog.Record
func cloneRecord(r slog.Record) slog.Record {
	clone := slog.Record{
		Time:    r.Time,
		Level:   r.Level,
		Message: r.Message,
	}

	r.Attrs(func(attr slog.Attr) bool {
		clone.AddAttrs(attr)
		return true
	})

	return clone
}

// Common context extractors

// RequestIDExtractor extracts a request ID from context
func RequestIDExtractor(key any) ContextExtractor {
	return func(ctx context.Context) []slog.Attr {
		if id := ctx.Value(key); id != nil {
			return []slog.Attr{slog.Any("request_id", id)}
		}
		return nil
	}
}

// TraceIDExtractor extracts a trace ID from context
func TraceIDExtractor(key any) ContextExtractor {
	return func(ctx context.Context) []slog.Attr {
		if id := ctx.Value(key); id != nil {
			return []slog.Attr{slog.Any("trace_id", id)}
		}
		return nil
	}
}

// UserIDExtractor extracts a user ID from context
func UserIDExtractor(key any) ContextExtractor {
	return func(ctx context.Context) []slog.Attr {
		if id := ctx.Value(key); id != nil {
			return []slog.Attr{slog.Any("user_id", id)}
		}
		return nil
	}
}

// MultiExtractor combines multiple extractors into one
func MultiExtractor(extractors ...ContextExtractor) ContextExtractor {
	return func(ctx context.Context) []slog.Attr {
		var attrs []slog.Attr
		for _, e := range extractors {
			if extracted := e(ctx); len(extracted) > 0 {
				attrs = append(attrs, extracted...)
			}
		}
		return attrs
	}
}
