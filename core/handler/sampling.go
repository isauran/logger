package handler

import (
	"context"
	"hash/fnv"
	"log/slog"
	"sync"
	"time"
)

// SamplingHandler implements intelligent log sampling
type SamplingHandler struct {
	handler     slog.Handler
	period      time.Duration
	sampleRate  uint32
	threshold   uint32
	mu          sync.RWMutex
	counters    map[uint64]uint32
	lastCleanup time.Time
}

// SamplingOption configures the sampling handler
type SamplingOption func(*SamplingHandler)

// NewSamplingHandler creates a new sampling handler
func NewSamplingHandler(handler slog.Handler, period time.Duration, sampleRate uint32, opts ...SamplingOption) *SamplingHandler {
	h := &SamplingHandler{
		handler:     handler,
		period:      period,
		sampleRate:  sampleRate,
		threshold:   1,
		counters:    make(map[uint64]uint32),
		lastCleanup: time.Now(),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// WithThreshold sets the sampling threshold
func WithThreshold(threshold uint32) SamplingOption {
	return func(h *SamplingHandler) {
		h.threshold = threshold
	}
}

func (h *SamplingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *SamplingHandler) Handle(ctx context.Context, r slog.Record) error {
	// Always log errors and higher severity
	if r.Level >= slog.LevelError {
		return h.handler.Handle(ctx, r)
	}

	// Generate hash key from message and attributes
	key := h.hashRecord(r)

	h.mu.Lock()
	defer h.mu.Unlock()

	// Cleanup old counters if needed
	h.cleanupIfNeeded()

	// Get current counter
	count := h.counters[key]
	h.counters[key] = count + 1

	// Determine if this record should be sampled
	if count < h.threshold || count%h.sampleRate == 0 {
		// Add sampling metadata
		r2 := cloneRecord(r)
		r2.AddAttrs(
			slog.Int("sampling.count", int(count+1)),
			slog.Int("sampling.rate", int(h.sampleRate)),
		)
		return h.handler.Handle(ctx, r2)
	}

	return nil
}

func (h *SamplingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SamplingHandler{
		handler:     h.handler.WithAttrs(attrs),
		period:      h.period,
		sampleRate:  h.sampleRate,
		threshold:   h.threshold,
		counters:    h.counters,
		lastCleanup: h.lastCleanup,
	}
}

func (h *SamplingHandler) WithGroup(name string) slog.Handler {
	return &SamplingHandler{
		handler:     h.handler.WithGroup(name),
		period:      h.period,
		sampleRate:  h.sampleRate,
		threshold:   h.threshold,
		counters:    h.counters,
		lastCleanup: h.lastCleanup,
	}
}

func (h *SamplingHandler) hashRecord(r slog.Record) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(r.Message))
	r.Attrs(func(attr slog.Attr) bool {
		hasher.Write([]byte(attr.Key))
		hasher.Write([]byte(attr.Value.String()))
		return true
	})
	return hasher.Sum64()
}

func (h *SamplingHandler) cleanupIfNeeded() {
	now := time.Now()
	if now.Sub(h.lastCleanup) >= h.period {
		h.counters = make(map[uint64]uint32)
		h.lastCleanup = now
	}
}
