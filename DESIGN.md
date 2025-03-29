# Logger Library Design

## Overview

This logger library provides a flexible, high-performance structured logging solution for Go applications. It builds upon Go's standard `log/slog` package, extending it with additional features like custom log levels, context awareness, enhanced error handling, file rotation, asynchronous logging, and adapters for popular Go libraries.

## Design Principles

1. **Performance-first approach**: Minimize allocations and optimize for high-throughput applications
2. **Structured logging**: All logs are structured for better parsing and analysis
3. **Extensibility**: Supports custom handlers, formatters, and levels
4. **Context awareness**: Leverages Go's context system for better tracing and metadata
5. **Adapters for ecosystem**: Seamless integration with popular libraries (GORM, Go-kit, etc.)
6. **Configuration flexibility**: Support for file, environment, and code-based configuration

## Architecture

The library is organized around several key components:

### Core Components

1. **Handler**: Enhanced implementation of `slog.Handler` with additional features:
   - Buffer pooling for improved performance
   - Global attributes and groups
   - Customizable formatting
   - Context awareness
   - Error handling with stack traces

2. **Level System**: Extended level support beyond standard slog levels:
   - Custom named levels
   - Level registry for organization-specific levels
   - Color and icon support for terminal output
   - Level-based filtering

3. **Field System**: Type-safe field creation with automatic inference:
   - Structured field creation
   - Automatic type handling
   - Group support
   - Error chain extraction
   - Stack trace generation

4. **Configuration**: Flexible configuration from multiple sources:
   - YAML/JSON file support
   - Environment variables
   - Programmatic configuration
   - Default sensible settings

### Handler Chain

Handlers can be composed in a chain to provide multiple features:

```
Input → ContextHandler → ErrorHandler → AsyncHandler → SamplingHandler → FileHandler → Output
```

Each handler in the chain adds specific functionality:

- **ContextHandler**: Extracts and adds context values to logs
- **ErrorHandler**: Enhances error logs with stack traces and cause chains
- **AsyncHandler**: Moves logging to background goroutines
- **SamplingHandler**: Reduces log volume for repetitive messages
- **FileHandler**: Manages log file output and rotation

### Builder Pattern

A fluent builder API allows for intuitive logger construction:

```go
logger := logger.NewBuilder().
    WithJSON().
    WithLevel(slog.LevelInfo).
    WithSource().
    WithFile("app.log", 100, 7).
    WithAsync(1000, 2).
    Build()
```

## Adapters

The library provides adapters for popular Go libraries:

1. **GORM**: Adapter for GORM's logger interface
2. **Go-kit**: Adapter for Go-kit's logging system
3. **Standard Logger**: Adapter for the standard library's log.Logger

## Performance Considerations

Several techniques are used to optimize performance:

1. **Buffer Pooling**: Reuses buffers to reduce garbage collection pressure
2. **Sampling**: Reduces log volume for repetitive messages
3. **Asynchronous Logging**: Moves logging off the critical path
4. **Minimal Allocations**: Careful design to minimize heap allocations
5. **Efficient Formatting**: Optimized text and JSON formatting

## Usage Patterns

### Basic Usage

```go
import (
    "github.com/isauran/logger"
    "log/slog"
)

func main() {
    // Create default logger
    logger.NewLogger(os.Stdout, logger.WithJSON(true))
    
    // Log with standard slog
    slog.Info("application started", "version", "1.0.0")
}
```

### Advanced Configuration

```go
import (
    "github.com/isauran/logger"
    "github.com/isauran/logger/core/handler"
    "log/slog"
    "os"
    "time"
)

func main() {
    // Create configured handler
    h, _ := handler.NewBuilder().
        WithJSON().
        WithLevel(slog.LevelDebug).
        WithFile("app.log", 100, 7).
        WithAsync(1000, 2).
        WithSampling(time.Second, 10).
        Build()
    
    // Create and set logger
    logger := slog.New(h)
    slog.SetDefault(logger)
    
    // Log using global slog
    slog.Info("application started", 
        "version", "1.0.0",
        "environment", "production")
}
```

### Context-Aware Logging

```go
func ProcessRequest(ctx context.Context, req Request) {
    // Add request ID to context
    ctx = context.WithValue(ctx, "request_id", req.ID)
    
    // Log with context
    slog.InfoContext(ctx, "processing request", 
        "method", req.Method,
        "path", req.Path)
        
    // Context ID will be automatically included in log
}
```

### Error Handling

```go
func ProcessFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        // Log error with stack trace
        slog.Error("failed to read file", 
            "path", path,
            "error", err)
        return fmt.Errorf("read file %s: %w", path, err)
    }
    
    // Process data...
    return nil
}
```

## Future Work

1. **Metrics Integration**: Better integration with Prometheus and OpenTelemetry
2. **More Adapters**: Support for additional logging libraries and frameworks
3. **Enhanced Web Integration**: Better support for HTTP middleware logging
4. **Profiling Tools**: Integrated logging performance analysis
5. **CLI Tools**: Command-line utilities for log analysis and management

## References

- [Go slog Documentation](https://pkg.go.dev/log/slog)
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging-best-practices)
- [OpenTelemetry Logging](https://opentelemetry.io/docs/specs/otel/logs/)