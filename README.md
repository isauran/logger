# Unified slog-based Logger

This project provides a unified logging solution based on Go's standard `log/slog` package. The main goal is to standardize logging across an entire project by providing adapters and integrations that allow replacing various third-party loggers with a single, consistent logging implementation.

## Overview

The core idea is to leverage Go's built-in `log/slog` package as the foundation for all logging needs, while providing:

1. A pre-configured slog handler with sensible defaults
2. Adapters for common logging interfaces (like gorm.io Logger, go-kit logger)
3. A unified way to manage log levels, formatting, and output across the entire application

## Architecture

```
┌──────────────────┐     ┌─────────────────┐
│   Application    │     │  Third-party    │
│     Code         │     │   Libraries     │
└────────┬─────────┘     └────────┬────────┘
         │                        │
         │                        │
┌────────▼────────────────────────▼───────┐
│              Unified Logger             │
├─────────────────────────────────────────┤
│        Custom slog Handler              │
└─────────────────────────────────────────┘
```

## Key Features

- Single source of truth for logging configuration
- Consistent log format across the entire application
- Built-in support for both JSON and text output formats
- Context-aware logging with source code location tracking
- Interface implementations for common third-party loggers
- Atomic writes to prevent log message interleaving
- Performance optimizations through buffer pooling

## Benefits

1. **Standardization**
   - Uniform log format across the entire application
   - Consistent logging behavior
   - Centralized configuration
   - Easier log aggregation and analysis

2. **Simplification**
   - No need to manage multiple logging libraries
   - Reduced dependency footprint
   - Simpler maintenance
   - Easier onboarding for new developers

3. **Performance**
   - Leverages slog's efficient design
   - Minimal allocations
   - Buffer pooling for improved performance
   - Atomic writes to prevent interleaving

4. **Future-Proof**
   - Built on standard library
   - No external dependencies
   - Easy to upgrade as slog evolves
   - Maintained as part of Go itself

## Implementation Details

### Core Handler

The core handler is built on top of slog.Handler with additional features:

```go
type Options struct {
    // JSON enables JSON output format
    JSON bool
    
    // Level sets the minimum logging level
    Level string
    
    // TimeFormat specifies the format for timestamps
    TimeFormat string
}
```

### Interface Adapters

Provides implementations for common logging interfaces:

1. **GORM Logger** (gorm.io/gorm/logger.Interface)
   - SQL query logging
   - Slow query detection
   - Error tracking

2. **Go-Kit Logger** (github.com/go-kit/log.Logger)
   - Key-value logging
   - Error handling
   - Level filtering

## Potential Challenges

1. **Interface Compatibility**
   - Some third-party interfaces might have unique features that are hard to map to slog
   - Need to carefully handle edge cases and special logging requirements

2. **Performance Overhead**
   - Additional abstraction layer might introduce minor overhead
   - Need to carefully implement buffer pooling and other optimizations

3. **Migration Effort**
   - Existing code might need updates to use the new logging system
   - Need to provide clear migration guides and tools

## Usage Examples

Standard logging:
```go
logger.NewLogger(os.Stdout, logger.WithJSON(true))
slog.Info("message", "key", "value")
```

GORM integration:
```go
db.Logger = logger.NewGormLogger("info")
```

Go-Kit integration:
```go
kitLogger := logger.NewGoKitLogger("info")
```

## Configuration

The logger can be configured through options:

```go
logger.NewLogger(
    os.Stdout,
    logger.WithJSON(true),
    logger.WithLevel("debug"),
    logger.WithTimeFormat(time.RFC3339),
)
```

## Best Practices

1. Always use structured logging with consistent key names
2. Include relevant context in log messages
3. Use appropriate log levels
4. Configure source code location tracking for development
5. Use JSON format in production for better log processing

## Conclusion

This unified logging approach provides a clean, efficient, and maintainable solution for application-wide logging. By leveraging Go's standard library slog package and providing necessary adapters, it eliminates the need for multiple logging libraries while ensuring consistent logging behavior across the entire application.