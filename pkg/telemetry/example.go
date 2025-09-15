// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

// ExampleUsage demonstrates how to use the telemetry package in a u-bmc service.
func ExampleUsage() error {
	ctx := context.Background()

	// Initialize telemetry with OTLP HTTP export
	shutdown, err := Setup(ctx,
		WithOTLPHTTP("http://localhost:4318"),
		WithServiceName("example-service"),
		WithServiceVersion("1.0.0"),
		WithMetrics(true),
		WithTraces(true),
		WithLogs(true),
		WithSamplingRatio(1.0),
	)
	if err != nil {
		return fmt.Errorf("failed to setup telemetry: %w", err)
	}
	defer shutdown(ctx)

	// Get logger with telemetry context
	logger := GetLogger("example-service")

	// Create metrics
	requestCounter, err := Counter("example-service", "requests_total",
		"Total number of requests", "1")
	if err != nil {
		return fmt.Errorf("failed to create counter: %w", err)
	}

	requestDuration, err := Histogram("example-service", "request_duration_seconds",
		"Request duration in seconds", "s")
	if err != nil {
		return fmt.Errorf("failed to create histogram: %w", err)
	}

	// Example service operation with tracing and metrics
	return ExampleServiceOperation(ctx, logger, requestCounter, requestDuration)
}

// ExampleServiceOperation demonstrates a typical service operation with full telemetry.
func ExampleServiceOperation(ctx context.Context, logger *slog.Logger, counter metric.Int64Counter, histogram metric.Float64Histogram) error {
	// Start a span for the operation
	return WithSpan(ctx, "example-service", "process_request", func(spanCtx context.Context) error {
		start := time.Now()

		// Add span attributes
		SetSpanAttributes(spanCtx,
			StringAttr("operation", "process_request"),
			StringAttr("user_id", "user123"),
		)

		// Log with telemetry context
		InfoWithContext(spanCtx, logger, "Processing request",
			slog.String("operation", "process_request"),
			slog.String("user_id", "user123"),
		)

		// Increment request counter
		IncrementCounter(spanCtx, counter, 1,
			StringAttr("method", "POST"),
			StringAttr("endpoint", "/api/v1/example"),
		)

		// Simulate some work
		err := performWork(spanCtx, logger)
		if err != nil {
			// Record error in span and logs
			RecordError(spanCtx, err, "Work failed")
			ErrorWithContext(spanCtx, logger, "Operation failed", err,
				slog.String("operation", "process_request"),
			)
			return err
		}

		// Record successful completion
		SetSpanStatus(spanCtx, codes.Ok, "Request processed successfully")
		AddSpanEvent(spanCtx, "request_completed",
			StringAttr("result", "success"),
		)

		// Record request duration
		duration := time.Since(start).Seconds()
		RecordDuration(spanCtx, histogram, duration,
			StringAttr("method", "POST"),
			StringAttr("endpoint", "/api/v1/example"),
			StringAttr("status", "success"),
		)

		InfoWithContext(spanCtx, logger, "Request processed successfully",
			slog.Duration("duration", time.Since(start)),
			slog.String("result", "success"),
		)

		return nil
	})
}

// performWork simulates some work that might fail.
func performWork(ctx context.Context, logger *slog.Logger) error {
	// Create a child span for the work
	return WithSpan(ctx, "example-service", "perform_work", func(workCtx context.Context) error {
		DebugWithContext(workCtx, logger, "Starting work",
			slog.String("step", "initialization"),
		)

		// Simulate work steps
		for i := 0; i < 3; i++ {
			stepName := fmt.Sprintf("step_%d", i+1)

			AddSpanEvent(workCtx, "work_step_started",
				StringAttr("step", stepName),
				IntAttr("step_number", i+1),
			)

			// Simulate some processing time
			time.Sleep(10 * time.Millisecond)

			AddSpanEvent(workCtx, "work_step_completed",
				StringAttr("step", stepName),
				BoolAttr("success", true),
			)

			DebugWithContext(workCtx, logger, "Completed work step",
				slog.String("step", stepName),
				slog.Int("step_number", i+1),
			)
		}

		return nil
	})
}

// ExampleNoOpSetup demonstrates how to set up telemetry in no-op mode for minimal overhead.
func ExampleNoOpSetup() error {
	ctx := context.Background()

	// Initialize telemetry with no-op providers
	shutdown, err := Setup(ctx, WithExporterType(NoOp))
	if err != nil {
		return fmt.Errorf("failed to setup no-op telemetry: %w", err)
	}
	defer shutdown(ctx)

	// All telemetry operations will be no-ops but the API remains the same
	logger := GetLogger("example-service")
	InfoWithContext(ctx, logger, "Service started with no-op telemetry")

	return nil
}

// ExampleDualExport demonstrates how to export telemetry data to both HTTP and gRPC endpoints.
func ExampleDualExport() error {
	ctx := context.Background()

	// Initialize telemetry with dual export
	shutdown, err := Setup(ctx,
		WithDualOTLP("http://localhost:4318", "localhost:4317"),
		WithServiceName("dual-export-service"),
		WithHeaders(map[string]string{
			"Authorization": "Bearer token123",
			"X-API-Key":     "api-key-456",
		}),
		WithTimeout(30*time.Second),
		WithBatchTimeout(5*time.Second),
		WithMaxExportBatch(256),
		WithInsecure(true), // For development only
	)
	if err != nil {
		return fmt.Errorf("failed to setup dual export telemetry: %w", err)
	}
	defer shutdown(ctx)

	logger := GetLogger("dual-export-service")
	InfoWithContext(ctx, logger, "Service started with dual export telemetry")

	return nil
}

// ExampleCustomConfiguration demonstrates advanced telemetry configuration.
func ExampleCustomConfiguration() error {
	ctx := context.Background()

	// Custom resource attributes
	resourceAttrs := map[string]string{
		"deployment.environment": "production",
		"service.namespace":      "u-bmc",
		"service.instance.id":    "instance-123",
		"host.name":              "bmc-host-01",
	}

	// Initialize telemetry with custom configuration
	shutdown, err := Setup(ctx,
		WithOTLPHTTP("https://telemetry.example.com/v1/traces"),
		WithServiceName("custom-service"),
		WithServiceVersion("2.1.0"),
		WithResourceAttributes(resourceAttrs),
		WithSamplingRatio(0.1), // Sample 10% of traces
		WithBatchTimeout(2*time.Second),
		WithMaxExportBatch(1024),
		WithMaxQueueSize(4096),
		WithTimeout(60*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to setup custom telemetry: %w", err)
	}
	defer shutdown(ctx)

	logger := GetLogger("custom-service")
	InfoWithContext(ctx, logger, "Service started with custom telemetry configuration")

	return nil
}

// ExampleMiddleware demonstrates how to create and use tracing middleware.
func ExampleMiddleware() error {
	ctx := context.Background()

	// Setup telemetry
	shutdown, err := Setup(ctx, WithOTLPHTTP("http://localhost:4318"))
	if err != nil {
		return fmt.Errorf("failed to setup telemetry: %w", err)
	}
	defer shutdown(ctx)

	// Create tracing middleware
	middleware := TracingMiddleware("example-service")

	// Wrap an operation with tracing
	operation := middleware("database_query", func(ctx context.Context) error {
		// Simulate database query
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	// Execute the traced operation
	return operation(ctx)
}
