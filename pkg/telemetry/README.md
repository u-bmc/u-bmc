# Telemetry Package

The telemetry package provides reusable abstractions around the OpenTelemetry SDK for Go, specifically designed for u-bmc services. It offers simplified configuration and setup for metrics, traces, and logs collection with support for various exporters including OTLP HTTP, OTLP gRPC, and no-op providers.

## Features

- **Multiple Exporter Types**: Support for NoOp, OTLP HTTP, OTLP gRPC, and dual export modes
- **Simple Configuration**: Functional options pattern for easy configuration
- **Stateless Design**: Functions over methods where possible
- **Context Propagation**: Built-in support for distributed tracing context propagation
- **Resource Management**: Automatic resource attribute management and service metadata
- **Performance Optimized**: Configurable sampling, batching, and queue sizes
- **Production Ready**: Comprehensive error handling and graceful shutdown

## Operation Modes

### NoOp Mode
Discards all telemetry data with minimal overhead. Perfect for environments where telemetry is not needed.

```go
shutdown, err := telemetry.Setup(ctx, telemetry.WithNoOp())
```

### OTLP HTTP Export
Exports telemetry data via OTLP over HTTP to compatible endpoints like Jaeger, Zipkin, or OpenTelemetry Collector.

```go
shutdown, err := telemetry.Setup(ctx, 
    telemetry.WithOTLPHTTP("http://localhost:4318"))
```

### OTLP gRPC Export
Exports telemetry data via OTLP over gRPC for high-performance scenarios.

```go
shutdown, err := telemetry.Setup(ctx, 
    telemetry.WithOTLPgRPC("localhost:4317"))
```

### Dual Export
Exports telemetry data via both HTTP and gRPC protocols simultaneously.

```go
shutdown, err := telemetry.Setup(ctx, 
    telemetry.WithDualOTLP("http://localhost:4318", "localhost:4317"))
```

## Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/u-bmc/u-bmc/pkg/telemetry"
)

func main() {
    ctx := context.Background()
    
    // Initialize telemetry
    shutdown, err := telemetry.Setup(ctx,
        telemetry.WithOTLPHTTP("http://localhost:4318"),
        telemetry.WithServiceName("my-service"),
        telemetry.WithServiceVersion("1.0.0"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)
    
    // Your service code here
    runService(ctx)
}
```

### Using Traces

```go
func processRequest(ctx context.Context) error {
    return telemetry.WithSpan(ctx, "my-service", "process_request", func(spanCtx context.Context) error {
        // Add span attributes
        telemetry.SetSpanAttributes(spanCtx,
            telemetry.StringAttr("user_id", "user123"),
            telemetry.StringAttr("operation", "process"),
        )
        
        // Perform work
        result, err := doWork(spanCtx)
        if err != nil {
            telemetry.RecordError(spanCtx, err, "Work failed")
            return err
        }
        
        // Add events
        telemetry.AddSpanEvent(spanCtx, "work_completed",
            telemetry.StringAttr("result", result),
        )
        
        return nil
    })
}
```

### Using Metrics

```go
func setupMetrics() error {
    // Create metrics
    requestCounter, err := telemetry.Counter("my-service", "requests_total",
        "Total number of requests", "1")
    if err != nil {
        return err
    }
    
    requestDuration, err := telemetry.Histogram("my-service", "request_duration_seconds",
        "Request duration in seconds", "s")
    if err != nil {
        return err
    }
    
    return nil
}

func handleRequest(ctx context.Context, counter metric.Int64Counter, histogram metric.Float64Histogram) {
    start := time.Now()
    
    // Increment counter
    telemetry.IncrementCounter(ctx, counter, 1,
        telemetry.StringAttr("method", "GET"),
        telemetry.StringAttr("endpoint", "/api/health"),
    )
    
    // Record duration
    duration := time.Since(start).Seconds()
    telemetry.RecordDuration(ctx, histogram, duration,
        telemetry.StringAttr("method", "GET"),
        telemetry.StringAttr("status", "success"),
    )
}
```

### Using Structured Logging with Context

```go
func businessLogic(ctx context.Context) {
    logger := telemetry.GetLogger("my-service")
    
    // Log with telemetry context (includes trace/span IDs automatically)
    telemetry.InfoWithContext(ctx, logger, "Processing started",
        slog.String("operation", "business_logic"),
        slog.String("user_id", "user123"),
    )
    
    // Error logging with telemetry context
    if err := someOperation(); err != nil {
        telemetry.ErrorWithContext(ctx, logger, "Operation failed", err,
            slog.String("operation", "some_operation"),
        )
    }
}
```

## Configuration Options

### Service Configuration
- `WithServiceName(name)` - Set service name for telemetry data
- `WithServiceVersion(version)` - Set service version
- `WithResourceAttributes(attrs)` - Add custom resource attributes

### Export Configuration
- `WithOTLPHTTP(endpoint)` - Configure OTLP HTTP endpoint
- `WithOTLPgRPC(endpoint)` - Configure OTLP gRPC endpoint
- `WithDualOTLP(httpEndpoint, grpcEndpoint)` - Configure dual export
- `WithHeaders(headers)` - Add custom headers for authentication
- `WithInsecure(bool)` - Enable/disable insecure connections
- `WithTimeout(duration)` - Set operation timeout

### Performance Configuration
- `WithBatchTimeout(duration)` - Set batch export timeout
- `WithMaxExportBatch(size)` - Set maximum items per batch
- `WithMaxQueueSize(size)` - Set maximum queue size
- `WithSamplingRatio(ratio)` - Set trace sampling ratio (0.0-1.0)

### Feature Flags
- `WithMetrics(enabled)` - Enable/disable metrics collection
- `WithTraces(enabled)` - Enable/disable trace collection
- `WithLogs(enabled)` - Enable/disable log collection

## Advanced Usage

### Custom Configuration

```go
shutdown, err := telemetry.Setup(ctx,
    telemetry.WithOTLPHTTP("https://telemetry.example.com/v1/traces"),
    telemetry.WithServiceName("production-service"),
    telemetry.WithServiceVersion("2.1.0"),
    telemetry.WithResourceAttributes(map[string]string{
        "deployment.environment": "production",
        "service.namespace":      "u-bmc",
        "host.name":              "bmc-host-01",
    }),
    telemetry.WithHeaders(map[string]string{
        "Authorization": "Bearer " + token,
        "X-API-Key":     apiKey,
    }),
    telemetry.WithSamplingRatio(0.1), // Sample 10% of traces
    telemetry.WithTimeout(60*time.Second),
    telemetry.WithBatchTimeout(5*time.Second),
    telemetry.WithMaxExportBatch(1024),
)
```

### Tracing Middleware

```go
// Create reusable tracing middleware
middleware := telemetry.TracingMiddleware("my-service")

// Wrap operations
tracedOperation := middleware("database_query", func(ctx context.Context) error {
    return database.Query(ctx, "SELECT * FROM users")
})

// Execute with automatic tracing
err := tracedOperation(ctx)
```

### Manual Span Management

```go
func complexOperation(ctx context.Context) error {
    spanCtx, span := telemetry.StartSpan(ctx, "my-service", "complex_operation")
    defer span.End()
    
    // Phase 1
    telemetry.AddSpanEvent(spanCtx, "phase1_started")
    if err := phase1(spanCtx); err != nil {
        telemetry.RecordError(spanCtx, err, "Phase 1 failed")
        return err
    }
    
    // Phase 2
    telemetry.AddSpanEvent(spanCtx, "phase2_started")
    if err := phase2(spanCtx); err != nil {
        telemetry.RecordError(spanCtx, err, "Phase 2 failed")
        return err
    }
    
    telemetry.SetSpanStatus(spanCtx, codes.Ok, "Operation completed successfully")
    return nil
}
```

## Helper Functions

### Attribute Helpers
- `StringAttr(key, value)` - Create string attribute
- `IntAttr(key, value)` - Create integer attribute
- `Int64Attr(key, value)` - Create int64 attribute
- `Float64Attr(key, value)` - Create float64 attribute
- `BoolAttr(key, value)` - Create boolean attribute
- `StringSliceAttr(key, value)` - Create string slice attribute
- `IntSliceAttr(key, value)` - Create integer slice attribute

### Span Helpers
- `StartSpan(ctx, tracerName, spanName, opts...)` - Start a new span
- `WithSpan(ctx, tracerName, spanName, fn, opts...)` - Execute function in span
- `SetSpanAttributes(ctx, attrs...)` - Set span attributes
- `AddSpanEvent(ctx, name, attrs...)` - Add span event
- `SetSpanStatus(ctx, code, description)` - Set span status
- `RecordError(ctx, err, description)` - Record error in span

### Metric Helpers
- `Counter(meterName, name, description, unit)` - Create counter
- `Histogram(meterName, name, description, unit)` - Create histogram
- `Gauge(meterName, name, description, unit)` - Create gauge
- `IncrementCounter(ctx, counter, value, attrs...)` - Increment counter
- `RecordDuration(ctx, histogram, duration, attrs...)` - Record duration

### Logging Helpers
- `GetLogger(name)` - Get logger with component name
- `InfoWithContext(ctx, logger, msg, attrs...)` - Log info with context
- `WarnWithContext(ctx, logger, msg, attrs...)` - Log warning with context
- `ErrorWithContext(ctx, logger, msg, err, attrs...)` - Log error with context
- `DebugWithContext(ctx, logger, msg, attrs...)` - Log debug with context

## Best Practices

### 1. Use Structured Logging
Always use structured logging with slog attributes for better observability:

```go
telemetry.InfoWithContext(ctx, logger, "User action performed",
    slog.String("user_id", userID),
    slog.String("action", "login"),
    slog.Duration("duration", duration),
)
```

### 2. Add Meaningful Span Attributes
Include relevant context in span attributes:

```go
telemetry.SetSpanAttributes(ctx,
    telemetry.StringAttr("user_id", userID),
    telemetry.StringAttr("operation_type", "create"),
    telemetry.StringAttr("resource_id", resourceID),
)
```

### 3. Use Appropriate Sampling
Set sampling ratios based on your environment:
- Development: 1.0 (100%)
- Staging: 0.5 (50%)
- Production: 0.1 (10%) or lower for high-traffic services

### 4. Handle Errors Properly
Always record errors in spans and logs:

```go
if err != nil {
    telemetry.RecordError(ctx, err, "Operation failed")
    telemetry.ErrorWithContext(ctx, logger, "Failed to process request", err)
    return err
}
```

### 5. Use No-Op Mode When Appropriate
For testing or environments where telemetry isn't needed:

```go
if testing.Testing() {
    shutdown, _ := telemetry.Setup(ctx, telemetry.WithNoOp())
    defer shutdown(ctx)
}
```

## Error Handling

The package defines standard errors in `errors.go`:

- `ErrInvalidExporterType` - Unsupported exporter type
- `ErrMissingEndpoint` - Required endpoint not provided
- `ErrProviderNotInitialized` - Provider not initialized
- `ErrInvalidConfiguration` - Invalid configuration
- `ErrShutdownFailed` - Provider shutdown failed
- `ErrExporterSetupFailed` - Exporter initialization failed
- `ErrConnectionFailed` - Connection to OTLP endpoint failed

Always check for errors and handle them appropriately:

```go
shutdown, err := telemetry.Setup(ctx, telemetry.WithOTLPHTTP(endpoint))
if err != nil {
    if errors.Is(err, telemetry.ErrConnectionFailed) {
        // Handle connection failure (maybe fallback to no-op)
        log.Warn("Telemetry endpoint unavailable, using no-op mode")
        shutdown, _ = telemetry.Setup(ctx, telemetry.WithNoOp())
    } else {
        return fmt.Errorf("failed to setup telemetry: %w", err)
    }
}
```

## Performance Considerations

1. **Sampling**: Use appropriate sampling ratios for production
2. **Batching**: Configure batch sizes based on your traffic patterns
3. **Queue Sizes**: Set queue sizes to handle traffic spikes
4. **Timeouts**: Configure reasonable timeouts for export operations
5. **No-Op Mode**: Use for testing or when telemetry is not required

## Integration with u-bmc Services

The telemetry package is designed to integrate seamlessly with u-bmc services:

```go
func (s *MyService) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
    // Setup telemetry for this service
    shutdown, err := telemetry.Setup(ctx,
        telemetry.WithOTLPHTTP(s.config.TelemetryEndpoint),
        telemetry.WithServiceName(s.Name()),
    )
    if err != nil {
        return fmt.Errorf("failed to setup telemetry: %w", err)
    }
    defer shutdown(ctx)
    
    // Service implementation with telemetry
    return s.runWithTelemetry(ctx, ipcConn)
}
```

## Examples

See `example.go` for comprehensive usage examples including:
- Basic setup and usage
- Advanced configuration
- Middleware patterns
- Error handling
- Performance optimization