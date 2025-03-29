package handler

import (
	"context"
	"log/slog"
	"sync"
)

// MultiHandler combines multiple handlers into one
type MultiHandler struct {
	handlers   []slog.Handler
	errHandler func(error)
	mu         sync.RWMutex
}

// NewMultiHandler creates a new multi-handler
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// WithErrorHandler sets a custom error handler
func (h *MultiHandler) WithErrorHandler(f func(error)) *MultiHandler {
	h.errHandler = f
	return h
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// If any handler is enabled for this level, we're enabled
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var lastErr error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r.Clone()); err != nil {
			lastErr = err
			if h.errHandler != nil {
				h.errHandler(err)
			}
		}
	}
	return lastErr
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}

	return &MultiHandler{
		handlers:   newHandlers,
		errHandler: h.errHandler,
	}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}

	return &MultiHandler{
		handlers:   newHandlers,
		errHandler: h.errHandler,
	}
}
