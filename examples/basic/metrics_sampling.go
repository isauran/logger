package basic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/isauran/logger/core/handler"
)

func ExampleMetricsAndSampling() {
	// Initialize base handler
	baseHandler := handler.New(os.Stdout, &handler.Options{
		JSON:       true,
		Level:      slog.LevelDebug,
		AddSource:  true,
		BufferPool: true,
	})

	// Add metrics collection
	metricsHandler := handler.NewMetricsHandler(baseHandler)

	// Add sampling for high-volume logs
	samplingHandler := handler.NewSamplingHandler(metricsHandler, time.Second, 10)

	// Create logger with all features
	logger := slog.New(samplingHandler)

	// Simulate some logging activity
	ctx := context.Background()

	// Generate some sample logs
	for i := 0; i < 100; i++ {
		logger.InfoContext(ctx, "high volume log",
			"iteration", i,
			"status", "processing",
		)
		time.Sleep(time.Millisecond * 10)
	}

	// Log different levels
	logger.Debug("debug message", "detail", "verbose information")
	logger.Info("info message", "detail", "normal operation")
	logger.Warn("warning message", "detail", "potential issue")
	logger.Error("error message", "detail", "operation failed")

	// Print metrics
	metrics := metricsHandler.GetMetrics().Snapshot()
	metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
	fmt.Printf("\nLogging Metrics:\n%s\n", metricsJSON)
}
