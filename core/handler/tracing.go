package handler

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingHandler adds OpenTelemetry trace context to log entries
type TracingHandler struct {
	handler slog.Handler
	tp      trace.TracerProvider
}

// NewTracingHandler creates a new tracing handler
func NewTracingHandler(handler slog.Handler) *TracingHandler {
	return &TracingHandler{
		handler: handler,
		tp:      otel.GetTracerProvider(),
	}
}

func (h *TracingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *TracingHandler) Handle(ctx context.Context, r slog.Record) error {
	// Get current span from context
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return h.handler.Handle(ctx, r)
	}

	// Add trace context to log record
	spanCtx := span.SpanContext()
	if spanCtx.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)

		if spanCtx.IsSampled() {
			r.AddAttrs(slog.Bool("sampled", true))
		}
	}

	// Add log record as span event
	attrs := make([]attribute.KeyValue, 0, r.NumAttrs())
	r.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attributeFromAttr(attr))
		return true
	})

	opts := []trace.EventOption{
		trace.WithAttributes(attrs...),
	}

	if r.Level >= slog.LevelError {
		span.SetStatus(codes.Error, r.Message)
	}

	span.AddEvent(r.Message, opts...)

	return h.handler.Handle(ctx, r)
}

func (h *TracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTracingHandler(h.handler.WithAttrs(attrs))
}

func (h *TracingHandler) WithGroup(name string) slog.Handler {
	return NewTracingHandler(h.handler.WithGroup(name))
}

// attributeFromAttr converts a slog.Attr to a trace attribute
func attributeFromAttr(attr slog.Attr) attribute.KeyValue {
	key := string(attr.Key)
	switch attr.Value.Kind() {
	case slog.KindBool:
		return attribute.Bool(key, attr.Value.Bool())
	case slog.KindDuration:
		return attribute.Int64(key, int64(attr.Value.Duration()))
	case slog.KindFloat64:
		return attribute.Float64(key, attr.Value.Float64())
	case slog.KindInt64:
		return attribute.Int64(key, attr.Value.Int64())
	case slog.KindString:
		return attribute.String(key, attr.Value.String())
	case slog.KindTime:
		return attribute.Int64(key, attr.Value.Time().UnixNano())
	default:
		return attribute.String(key, attr.Value.String())
	}
}
