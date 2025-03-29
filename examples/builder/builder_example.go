package builder

import (
	"log/slog"
	"os"
	"time"

	"github.com/isauran/logger/core/handler"
)

// Example demonstrating how to use the builder pattern to create a logger
func ExampleBuilder() {
	// Configure the handler with a fluent API
	h, err := handler.NewBuilder().
		WithJSON().                        // Use JSON format for logs
		WithLevel(slog.LevelDebug).        // Set minimum log level to debug
		WithSource().                      // Include source file and line in logs
		WithTimeFormat(time.RFC3339Nano).  // Set precise time format
		WithWriter(os.Stdout).             // Log to standard output
		WithSampling(time.Second, 10).     // Sample logs (1 per 10 in each second)
		WithMetrics().                     // Enable metrics collection
		WithErrorHandler(func(err error) { // Custom error handler
			slog.Error("logging error", "error", err)
		}).
		Build()

	if err != nil {
		slog.Error("failed to build logger", "error", err)
		return
	}

	// Create a new logger with the configured handler
	logger := slog.New(h)

	// Use the logger
	logger.Info("application started",
		"version", "1.0.0",
		"environment", "production",
	)

	logger.Debug("debug information",
		"config", map[string]interface{}{
			"timeout":  30,
			"retries":  3,
			"features": []string{"auth", "notifications", "reports"},
		},
	)

	logger.Warn("potential issue detected",
		"component", "database",
		"latency_ms", 250,
		"threshold_ms", 200,
	)

	logger.Error("operation failed",
		"operation", "data_sync",
		"user_id", 12345,
		"attempt", 3,
		"error", "connection timeout",
	)

	// Add a small delay to ensure async handler processes the messages
	time.Sleep(100 * time.Millisecond)
}
