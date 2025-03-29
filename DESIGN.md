# Alternative Logger Design Approaches

This document explores different design approaches for the slog-based unified logger, focusing on initialization patterns and configuration options.

## 1. Builder Pattern Approach

Instead of using functional options, we could use a builder pattern for more fluent configuration:

```go
type LoggerBuilder struct {
    writer io.Writer
    config Config
}

type Config struct {
    JSON       bool
    Level      slog.Level
    TimeFormat string
    AddSource  bool
    BufferPool bool
    AddCaller  bool
    // Additional configuration options
}

// Usage example:
logger := NewLoggerBuilder(os.Stdout).
    WithJSON().
    WithLevel(slog.LevelDebug).
    WithTimeFormat(time.RFC3339).
    WithSourceInfo().
    WithBufferPool().
    Build()
```

### Benefits
- Fluent, chainable API
- Clear configuration visualization
- Type-safe configuration
- Easy to extend with new options
- Self-documenting code

### Drawbacks
- More boilerplate code
- Slightly more complex implementation
- Need to maintain builder state

## 2. Configuration-Based Approach

Use configuration files (YAML/JSON) for logger setup:

```yaml
logger:
  format: json
  level: info
  time_format: "2006-01-02T15:04:05Z07:00"
  source:
    enabled: true
    short_path: true
  buffer:
    enabled: true
    size: 4096
    pool_size: 32
  handlers:
    - type: file
      path: /var/log/app.log
      rotate:
        max_size: 100
        max_age: 7
        max_backups: 5
    - type: stdout
      format: text
```

```go
type LoggerConfig struct {
    Format     string         `yaml:"format"`
    Level      string         `yaml:"level"`
    TimeFormat string         `yaml:"time_format"`
    Source     SourceConfig   `yaml:"source"`
    Buffer     BufferConfig   `yaml:"buffer"`
    Handlers   []HandlerConfig `yaml:"handlers"`
}

logger, err := NewLoggerFromConfig("config.yaml")
```

### Benefits
- External configuration
- Easy to change settings without recompiling
- Support for multiple output handlers
- Environment-specific configurations
- Clearer separation of concerns

### Drawbacks
- Runtime configuration parsing
- Potential configuration errors
- More complex error handling

## 3. Context-Based Logger

Focus on context propagation and enrichment:

```go
type ContextLogger struct {
    base   *slog.Logger
    fields []slog.Attr
}

// Usage example:
logger := NewContextLogger(os.Stdout)

// Add request-specific context
reqLogger := logger.With(
    slog.String("request_id", id),
    slog.String("user_id", userID),
)

// Use in handlers
func Handler(ctx context.Context) {
    logger := FromContext(ctx)
    logger.Info("processing request")
}
```

### Benefits
- Better context propagation
- Request-scoped logging
- Easy to add/remove context
- Clean handler integration
- Trace correlation support

### Drawbacks
- Need to manage context properly
- Potential memory overhead
- More complex context management

## 4. Async Logger Implementation

Optimize for performance with async logging:

```go
type AsyncLogger struct {
    queue  chan *LogEntry
    done   chan struct{}
    wg     sync.WaitGroup
    config AsyncConfig
}

type AsyncConfig struct {
    QueueSize    int
    Workers      int
    BatchSize    int
    FlushInterval time.Duration
}

// Usage example:
logger := NewAsyncLogger(AsyncConfig{
    QueueSize:    1000,
    Workers:      2,
    BatchSize:    100,
    FlushInterval: time.Second,
})
defer logger.Close()
```

### Benefits
- Better performance
- Batched writing
- Non-blocking logging
- Controlled resource usage
- Buffer overflow protection

### Drawbacks
- Potential log loss on crash
- More complex shutdown
- Memory overhead for queue
- Need for careful tuning

## 5. Interface-First Design

Focus on interfaces for better testing and flexibility:

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    With(args ...any) Logger
}

type Handler interface {
    Handle(r Record) error
    WithAttrs(attrs []Attr) Handler
    WithGroup(name string) Handler
}

// Implementation can be swapped easily:
var logger Logger = NewProductionLogger()
// For tests:
var logger Logger = NewTestLogger()
```

### Benefits
- Easy to mock for testing
- Clear contract
- Simpler integration
- Easy to extend
- Better separation of concerns

### Drawbacks
- More interfaces to maintain
- Potential interface bloat
- Additional abstraction layer

## 6. Structured Event Logger

Focus on structured events rather than free-form messages:

```go
type Event struct {
    Name    string
    Level   slog.Level
    Fields  map[string]interface{}
    Time    time.Time
}

// Usage example:
logger.LogEvent(Event{
    Name:  "UserLogin",
    Level: slog.LevelInfo,
    Fields: map[string]interface{}{
        "user_id": "123",
        "ip":      "192.168.1.1",
        "success": true,
    },
})
```

### Benefits
- Consistent structured data
- Better analytics support
- Clear event semantics
- Easy to process
- Strong typing possible

### Drawbacks
- More rigid structure
- More setup required
- Less flexibility

## Recommendations

1. **For Small Projects**
   - Use the current functional options approach
   - Keep it simple with direct slog usage
   - Focus on essential features

2. **For Medium Projects**
   - Consider the Builder pattern
   - Add basic context support
   - Implement simple configuration

3. **For Large Projects**
   - Use configuration-based approach
   - Implement async logging
   - Add full context support
   - Consider structured events

4. **For High-Performance Requirements**
   - Implement async logging
   - Use buffer pools
   - Consider structured events
   - Optimize for minimal allocations

5. **For Microservices**
   - Focus on context propagation
   - Use structured events
   - Implement tracing support
   - Consider distributed logging

## Migration Strategy

When adopting a new design:

1. Create new interfaces/types alongside existing ones
2. Provide adapters for backward compatibility
3. Update documentation and examples
4. Add migration guides
5. Consider providing migration tools
6. Plan for gradual adoption

## Conclusion

Each approach has its strengths and ideal use cases. Consider your specific requirements:

- Performance needs
- Operational complexity
- Team expertise
- Project scale
- Maintenance requirements
- Integration requirements

Choose the approach that best balances these factors for your specific use case.