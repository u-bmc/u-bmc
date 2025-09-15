# Telemetry Service

The telemetry service acts as a collector and aggregator for OpenTelemetry data within the u-bmc system. It implements the u-bmc service interface and provides centralized telemetry data collection, filtering, and export capabilities.

## Overview

The telemetry service can operate in several modes:
- **NoOp**: Discards all telemetry data with minimal overhead
- **OTLP HTTP**: Exports telemetry data via OTLP over HTTP
- **OTLP gRPC**: Exports telemetry data via OTLP over gRPC  
- **Dual**: Exports telemetry data via both HTTP and gRPC protocols

## Features

- **Data Collection**: Collects metrics, traces, and logs from other u-bmc services
- **Aggregation**: Applies filtering and aggregation as supported by OTLP
- **Multiple Exporters**: Support for HTTP, gRPC, and dual export modes
- **IPC Integration**: Uses NATS for inter-service communication
- **Configurable**: Extensive configuration options for performance tuning
- **Graceful Shutdown**: Proper cleanup and data flushing on shutdown

## Configuration

The service supports extensive configuration through functional options:

### Basic Configuration
```go
service := telemetry.New(
    telemetry.WithName("telemetry"),
    telemetry.WithServiceName("u-bmc-telemetry"),
    telemetry.WithServiceVersion("1.0.0"),
)
```

### Export Configuration
```go
service := telemetry.New(
    telemetry.WithExporterType("otlp-http"),
    telemetry.WithHTTPEndpoint("http://localhost:4318"),
    telemetry.WithGRPCEndpoint("localhost:4317"),
    telemetry.WithHeaders(map[string]string{
        "Authorization": "Bearer token",
    }),
    telemetry.WithInsecure(true), // For development only
)
```

### Performance Tuning
```go
service := telemetry.New(
    telemetry.WithTimeout(30*time.Second),
    telemetry.WithBatchTimeout(5*time.Second),
    telemetry.WithMaxExportBatch(512),
    telemetry.WithMaxQueueSize(2048),
    telemetry.WithSamplingRatio(0.1), // 10% sampling
)
```

### Feature Control
```go
service := telemetry.New(
    telemetry.WithMetrics(true),
    telemetry.WithTraces(true),
    telemetry.WithLogs(true),
    telemetry.WithCollection(true),
    telemetry.WithAggregation(true),
    telemetry.WithCollectionInterval(30*time.Second),
)
```

## Usage

### Creating the Service
```go
import "github.com/u-bmc/u-bmc/service/telemetry"

// NoOp mode for testing
service := telemetry.New(telemetry.WithNoOp())

// OTLP HTTP export
service := telemetry.New(
    telemetry.WithOTLPHTTP("http://localhost:4318"),
    telemetry.WithServiceName("u-bmc-telemetry"),
)

// Dual export with authentication
service := telemetry.New(
    telemetry.WithDualOTLP("http://localhost:4318", "localhost:4317"),
    telemetry.WithHeaders(map[string]string{
        "Authorization": "Bearer " + token,
        "X-API-Key":     apiKey,
    }),
)
```

### Running the Service
```go
func main() {
    ctx := context.Background()
    
    service := telemetry.New(
        telemetry.WithOTLPHTTP("http://localhost:4318"),
        telemetry.WithServiceName("u-bmc-telemetry"),
    )
    
    // The service will be started by the u-bmc service manager
    err := service.Run(ctx, ipcConn)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Architecture

### Data Flow
1. **Collection**: Other u-bmc services send telemetry data via NATS IPC
2. **Processing**: The service receives and processes the data
3. **Filtering**: Applies any configured filters to the data
4. **Aggregation**: Performs aggregation if enabled
5. **Export**: Sends processed data to configured OTLP endpoints

### IPC Integration
The service uses NATS subjects for telemetry data collection:
- `telemetry.metrics.<service_name>` - Metrics data
- `telemetry.traces.<service_name>` - Trace data  
- `telemetry.logs.<service_name>` - Log data

### Context Propagation
The service automatically extracts and propagates OpenTelemetry context from NATS message headers, maintaining distributed tracing across service boundaries.

## Configuration Options

### Service Options
- `WithName(name)` - Set service name
- `WithServiceName(serviceName)` - Set telemetry service name
- `WithServiceVersion(version)` - Set service version

### Export Options  
- `WithExporterType(type)` - Set exporter type ("noop", "otlp-http", "otlp-grpc", "dual")
- `WithHTTPEndpoint(endpoint)` - Set HTTP endpoint
- `WithGRPCEndpoint(endpoint)` - Set gRPC endpoint
- `WithHeaders(headers)` - Set additional headers
- `WithInsecure(bool)` - Enable/disable insecure connections
- `WithTimeout(duration)` - Set operation timeout

### Performance Options
- `WithBatchTimeout(duration)` - Set batch export timeout
- `WithMaxExportBatch(size)` - Set maximum items per batch
- `WithMaxQueueSize(size)` - Set maximum queue size
- `WithSamplingRatio(ratio)` - Set trace sampling ratio (0.0-1.0)

### Feature Options
- `WithMetrics(enabled)` - Enable/disable metrics collection
- `WithTraces(enabled)` - Enable/disable trace collection  
- `WithLogs(enabled)` - Enable/disable log collection
- `WithCollection(enabled)` - Enable/disable data collection
- `WithAggregation(enabled)` - Enable/disable data aggregation
- `WithCollectionInterval(duration)` - Set aggregation interval
- `WithShutdownTimeout(duration)` - Set shutdown timeout

### Resource Options
- `WithResourceAttributes(attrs)` - Set resource attributes

### Convenience Options
- `WithOTLPHTTP(endpoint)` - Configure OTLP HTTP export
- `WithOTLPGRPC(endpoint)` - Configure OTLP gRPC export
- `WithDualOTLP(httpEndpoint, grpcEndpoint)` - Configure dual export
- `WithNoOp()` - Configure no-op operation

## Deployment

### Production Configuration
```go
service := telemetry.New(
    telemetry.WithDualOTLP(
        "https://telemetry.company.com/v1/traces",
        "telemetry.company.com:443",
    ),
    telemetry.WithServiceName("u-bmc-telemetry"),
    telemetry.WithServiceVersion("1.0.0"),
    telemetry.WithHeaders(map[string]string{
        "Authorization": "Bearer " + os.Getenv("TELEMETRY_TOKEN"),
    }),
    telemetry.WithResourceAttributes(map[string]string{
        "deployment.environment": "production",
        "service.namespace":      "u-bmc", 
        "host.name":              hostname,
    }),
    telemetry.WithSamplingRatio(0.1),
    telemetry.WithTimeout(60*time.Second),
    telemetry.WithBatchTimeout(10*time.Second),
)
```

### Development Configuration
```go
service := telemetry.New(
    telemetry.WithOTLPHTTP("http://localhost:4318"),
    telemetry.WithServiceName("u-bmc-telemetry-dev"),
    telemetry.WithInsecure(true),
    telemetry.WithSamplingRatio(1.0), // 100% sampling for dev
)
```

### Testing Configuration
```go
service := telemetry.New(telemetry.WithNoOp())
```

## Monitoring

The telemetry service itself generates telemetry data:
- **Metrics**: Collection rates, export success/failure rates, queue sizes
- **Traces**: Processing spans for received telemetry data
- **Logs**: Service lifecycle, errors, and debug information

## Error Handling

The service defines standard errors in `errors.go`:
- `ErrServiceNotConfigured` - Service not properly configured
- `ErrProviderInitializationFailed` - Telemetry provider failed to initialize
- `ErrCollectorSetupFailed` - Collector setup failed
- `ErrIPCConnectionFailed` - IPC connection failed
- `ErrShutdownTimeout` - Service shutdown timed out

## Best Practices

### 1. Use Appropriate Sampling
Set sampling ratios based on environment and traffic:
```go
// Production: Low sampling for high-traffic services
telemetry.WithSamplingRatio(0.01)

// Development: High sampling for debugging
telemetry.WithSamplingRatio(1.0)
```

### 2. Configure Proper Timeouts
Set timeouts appropriate for your network conditions:
```go
telemetry.WithTimeout(30*time.Second),
telemetry.WithBatchTimeout(5*time.Second),
telemetry.WithShutdownTimeout(10*time.Second),
```

### 3. Secure Your Endpoints
Always use authentication for production endpoints:
```go
telemetry.WithHeaders(map[string]string{
    "Authorization": "Bearer " + token,
    "X-API-Key":     apiKey,
}),
```

### 4. Monitor Resource Usage
Configure queue and batch sizes based on your traffic:
```go
telemetry.WithMaxQueueSize(4096),      // Higher for traffic spikes
telemetry.WithMaxExportBatch(1024),    // Larger batches for efficiency
```

### 5. Use Resource Attributes
Add meaningful resource attributes for better observability:
```go
telemetry.WithResourceAttributes(map[string]string{
    "deployment.environment": env,
    "service.namespace":      "u-bmc",
    "host.name":              hostname,
    "service.instance.id":    instanceID,
}),
```

## Troubleshooting

### Common Issues

**Service won't start**
- Check endpoint connectivity
- Verify authentication credentials
- Review configuration for typos

**High memory usage**  
- Reduce queue sizes
- Increase batch export frequency
- Lower sampling ratios

**Missing telemetry data**
- Check sampling configuration
- Verify endpoint is receiving data
- Review service logs for errors

**Connection failures**
- Verify endpoint URLs
- Check network connectivity
- Review security settings (TLS/insecure)

### Debug Mode
Enable debug logging for troubleshooting:
```go
// The service automatically uses the global logger with debug level
// Check service logs for detailed information
```

## Integration with OpenTelemetry Ecosystem

The service is compatible with:
- **Jaeger** - Distributed tracing backend
- **Zipkin** - Distributed tracing system  
- **Prometheus** - Metrics collection and alerting
- **OpenTelemetry Collector** - Vendor-agnostic telemetry pipeline
- **Grafana** - Observability and monitoring platform
- **DataDog, New Relic, Honeycomb** - Commercial observability platforms

Any OTLP-compatible backend can receive data from this service.