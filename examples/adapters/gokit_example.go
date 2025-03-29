package adapters

import (
	"context"
	"time"

	"github.com/isauran/logger/adapters/gokit"
)

func ExampleGoKitLogger() {
	// Initialize GoKit logger adapter
	kitLogger := gokit.NewLogger("info")

	// Log simple message
	kitLogger.Log("msg", "application started")

	// Log with multiple key-value pairs
	kitLogger.Log(
		"msg", "service health check",
		"status", "healthy",
		"timestamp", time.Now().Format(time.RFC3339),
	)

	// Log with context
	ctx := context.Background()
	kitLogger.Log(
		"msg", "processing request",
		"trace_id", ctx.Value("trace_id"),
		"duration_ms", 150,
		"status", "success",
	)

	// Log errors
	err := context.DeadlineExceeded
	kitLogger.Log(
		"msg", "request failed",
		"error", err.Error(),
		"severity", "error",
		"component", "api",
	)

	// Log structured data
	kitLogger.Log(
		"msg", "user action",
		"action", "login",
		"user_id", 123,
		"ip_address", "192.168.1.1",
		"browser", "Chrome",
		"success", true,
	)
}
