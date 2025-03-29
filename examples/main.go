package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/isauran/logger/examples/adapters"
	"github.com/isauran/logger/examples/basic"
	"github.com/isauran/logger/examples/builder"
	"github.com/isauran/logger/examples/file"
)

func runExample(name string, fn func()) {
	fmt.Printf("\n=== Running %s Example ===\n", name)
	fn()
}

func main() {
	fmt.Println("Starting logger examples...")

	// Set up the default slog logger with JSON handler
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Basic examples
	runExample("Basic Metrics and Sampling", basic.ExampleMetricsAndSampling)

	// File handling examples
	runExample("File Rotation", file.ExampleFileRotation)

	// Builder pattern example
	runExample("Builder Pattern", builder.ExampleBuilder)

	// Adapter examples
	runExample("GORM Logger Adapter", adapters.ExampleGormLogger)
	runExample("GoKit Logger Adapter", adapters.ExampleGoKitLogger)

	fmt.Println("\nAll examples completed successfully")

	// Clean up test database
	os.Remove("test.db")
}
