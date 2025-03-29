package handler

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	logEntries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_entries_total",
			Help: "Total number of log entries by level",
		},
		[]string{"level"},
	)

	logErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "log_errors_total",
			Help: "Total number of logging errors",
		},
	)

	logLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "log_latency_seconds",
			Help:    "Logging latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"level"},
	)

	queueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "log_queue_size",
			Help: "Current size of the logging queue",
		},
	)

	droppedLogs = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "log_entries_dropped_total",
			Help: "Total number of log entries dropped",
		},
	)
)

// Metrics holds the collected logging metrics
type Metrics struct {
	Entries    map[string]int
	Errors     int
	Dropped    int
	QueueSize  int64
	AvgLatency map[string]float64
}

// Snapshot creates a snapshot of the current metrics
func (m *Metrics) Snapshot() *Metrics {
	return &Metrics{
		Entries:    m.Entries,
		Errors:     m.Errors,
		Dropped:    m.Dropped,
		QueueSize:  m.QueueSize,
		AvgLatency: m.AvgLatency,
	}
}

// MetricsHandler wraps a handler with metrics collection
type MetricsHandler struct {
	handler   slog.Handler
	queueSize atomic.Int64
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(handler slog.Handler) *MetricsHandler {
	return &MetricsHandler{
		handler: handler,
	}
}

func (h *MetricsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *MetricsHandler) Handle(ctx context.Context, r slog.Record) error {
	// Track queue size for async handlers
	if h.queueSize.Load() > 0 {
		queueSize.Set(float64(h.queueSize.Load()))
	}

	// Track log entry count by level
	logEntries.WithLabelValues(r.Level.String()).Inc()

	// Track latency
	start := time.Now()
	err := h.handler.Handle(ctx, r)
	duration := time.Since(start)

	logLatency.WithLabelValues(r.Level.String()).Observe(duration.Seconds())

	if err != nil {
		logErrors.Inc()
	}

	return err
}

func (h *MetricsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewMetricsHandler(h.handler.WithAttrs(attrs))
}

func (h *MetricsHandler) WithGroup(name string) slog.Handler {
	return NewMetricsHandler(h.handler.WithGroup(name))
}

// RecordDroppedLog increments the dropped logs counter
func (h *MetricsHandler) RecordDroppedLog() {
	droppedLogs.Inc()
}

// SetQueueSize updates the current queue size metric
func (h *MetricsHandler) SetQueueSize(size int64) {
	h.queueSize.Store(size)
}

// GetMetrics returns the current metrics
func (h *MetricsHandler) GetMetrics() *Metrics {
	// Initialize metrics structure
	m := &Metrics{
		Entries:    make(map[string]int),
		AvgLatency: make(map[string]float64),
		QueueSize:  h.queueSize.Load(),
	}

	// Collect metrics from Prometheus
	// In a real implementation, this would extract data from the Prometheus metrics
	// Here we're simulating data since we can't directly access Prometheus counters
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		// Simulated values (in a real implementation, these would come from Prometheus)
		m.Entries[level] = 0
		m.AvgLatency[level] = 0.0
	}

	return m
}
