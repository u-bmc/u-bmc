// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/metric"
)

// ExampleMandatoryUsage demonstrates the mandatory telemetry usage pattern for u-bmc services.
// ALL u-bmc services MUST generate telemetry data and send it to the central telemetry collector.
func ExampleMandatoryUsage() error {
	ctx := context.Background()

	// MANDATORY: Initialize telemetry to send data to central collector
	// Services generate telemetry data which is sent to the central collector via NATS
	shutdown, err := Setup(ctx,
		WithServiceName("example-service"), // REQUIRED: Service identification
		WithServiceVersion("1.0.0"),
		WithMetrics(true), // REQUIRED: At least one signal must be enabled
		WithTraces(true),
		WithLogs(true),
		// No direct OTLP endpoint - data is sent to central telemetry collector
		// The collector decides whether to export (for debugging) or drop (for performance)
	)
	if err != nil {
		// Telemetry setup failure should be treated as a critical error
		return fmt.Errorf("CRITICAL: telemetry setup failed - service cannot start: %w", err)
	}
	defer shutdown(ctx)

	// Get logger with automatic telemetry context integration
	logger := GetLogger("example-service")

	// Create metrics - data automatically sent to central collector
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

	// Example service operation with mandatory telemetry generation
	return ExampleServiceOperation(ctx, logger, requestCounter, requestDuration)
}

// ExampleServiceOperation demonstrates a typical u-bmc service operation with mandatory telemetry.
// This shows the enforced pattern that all services must follow for observability.
func ExampleServiceOperation(ctx context.Context, logger *slog.Logger, counter metric.Int64Counter, histogram metric.Float64Histogram) error {
	// MANDATORY: All operations must be traced - use abstracted span functions
	return WithSpan(ctx, "example-service", "process_request", func(spanCtx context.Context) error {
		start := time.Now()

		// REQUIRED: Add meaningful span attributes for observability
		SetSpanAttributes(spanCtx,
			StringAttr("operation", "process_request"),
			StringAttr("user_id", "user123"),
			StringAttr("service.component", "bmc"),
		)

		// MANDATORY: Log with telemetry context (automatic trace/span ID inclusion)
		InfoWithContext(spanCtx, logger, "Processing request",
			slog.String("operation", "process_request"),
			slog.String("user_id", "user123"),
		)

		// REQUIRED: All operations must be metered
		IncrementCounter(spanCtx, counter, 1,
			StringAttr("method", "POST"),
			StringAttr("endpoint", "/api/v1/example"),
			StringAttr("service", "example-service"),
		)

		// Perform business logic with telemetry context
		err := performWork(spanCtx, logger)
		if err != nil {
			// MANDATORY: All errors must be recorded in telemetry
			RecordError(spanCtx, err, "Work failed")
			ErrorWithContext(spanCtx, logger, "Operation failed", err,
				slog.String("operation", "process_request"),
			)
			return err
		}

		// REQUIRED: Set span status for successful operations
		SetSpanStatus(spanCtx, StatusOK(), "Request processed successfully")
		AddSpanEvent(spanCtx, "request_completed",
			StringAttr("result", "success"),
		)

		// MANDATORY: Record operation duration for performance monitoring
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

// performWork simulates BMC work operations with mandatory telemetry patterns.
func performWork(ctx context.Context, logger *slog.Logger) error {
	// MANDATORY: Create child spans for all sub-operations
	return WithSpan(ctx, "example-service", "perform_work", func(workCtx context.Context) error {
		DebugWithContext(workCtx, logger, "Starting BMC work operation",
			slog.String("step", "initialization"),
		)

		// Simulate BMC work steps with full telemetry coverage
		for i := 0; i < 3; i++ {
			stepName := fmt.Sprintf("bmc_operation_%d", i+1)

			// REQUIRED: Record operation start events
			AddSpanEvent(workCtx, "bmc_operation_started",
				StringAttr("step", stepName),
				IntAttr("step_number", i+1),
				StringAttr("component", "bmc"),
			)

			// Simulate BMC processing time
			time.Sleep(10 * time.Millisecond)

			// REQUIRED: Record operation completion events
			AddSpanEvent(workCtx, "bmc_operation_completed",
				StringAttr("step", stepName),
				BoolAttr("success", true),
				IntAttr("duration_ms", 10),
			)

			DebugWithContext(workCtx, logger, "Completed BMC operation step",
				slog.String("step", stepName),
				slog.Int("step_number", i+1),
			)
		}

		return nil
	})
}

// ExampleCentralCollectorPattern demonstrates how services send telemetry to the central collector.
// This shows the normal pattern where services generate data but don't directly export.
func ExampleCentralCollectorPattern() error {
	ctx := context.Background()

	// STANDARD: Services generate telemetry and send to central collector
	// The central collector decides whether to export or drop for performance
	shutdown, err := Setup(ctx,
		WithServiceName("bmc-power-service"),
		WithServiceVersion("1.2.0"),
		// No direct exporter configuration - data goes to central collector
	)
	if err != nil {
		return fmt.Errorf("failed to setup telemetry: %w", err)
	}
	defer shutdown(ctx)

	// All telemetry data is generated and sent to central collector
	logger := GetLogger("bmc-power-service")
	InfoWithContext(ctx, logger, "Power service generating telemetry for central collector")

	// Central collector can enable/disable export at runtime for debugging
	return nil
}

// ExampleTestingSetup demonstrates telemetry setup for testing scenarios.
func ExampleTestingSetup() error {
	ctx := context.Background()

	// TESTING: Standard setup still works in tests - data goes to central collector
	shutdown, err := Setup(ctx,
		WithServiceName("test-service"),
		// Central collector pattern works in tests too
	)
	if err != nil {
		return fmt.Errorf("failed to setup test telemetry: %w", err)
	}
	defer shutdown(ctx)

	logger := GetLogger("test-service")
	InfoWithContext(ctx, logger, "Test service with central collector telemetry")

	return nil
}

// ExampleServiceWithResourceAttributes demonstrates adding BMC-specific resource attributes.
// Services should include relevant BMC hardware and deployment information.
func ExampleServiceWithResourceAttributes() error {
	ctx := context.Background()

	// RECOMMENDED: Include BMC-specific resource attributes for better observability
	shutdown, err := Setup(ctx,
		WithServiceName("production-bmc-service"),
		WithServiceVersion("2.0.0"),
		WithResourceAttributes(map[string]string{
			"deployment.environment": "production",
			"service.namespace":      "u-bmc",
			"bmc.hardware.vendor":    "dell",
			"bmc.hardware.model":     "poweredge-r750",
			"bmc.firmware.version":   "2.85.85.85",
			"datacenter.region":      "us-east-1",
			"datacenter.zone":        "zone-a",
			"host.name":              "bmc-host-01",
		}),
		// Data sent to central collector with rich resource context
	)
	if err != nil {
		return fmt.Errorf("CRITICAL: telemetry setup failed: %w", err)
	}
	defer shutdown(ctx)

	logger := GetLogger("production-bmc-service")
	InfoWithContext(ctx, logger, "Production BMC service with resource attributes")

	return nil
}

// ExampleCentralCollectorConfiguration demonstrates how the central telemetry collector
// might be configured differently from regular services. This would typically be done
// in the telemetry service itself, not in regular u-bmc services.
func ExampleCentralCollectorConfiguration() error {
	ctx := context.Background()

	// NOTE: This configuration would be used by the telemetry service itself,
	// not by regular u-bmc services. Regular services use the standard pattern
	// and send data to the central collector.

	// Central collector can be configured with OTLP endpoints when export is needed
	resourceAttrs := map[string]string{
		"service.role":           "telemetry-collector",
		"deployment.environment": "production",
		"service.namespace":      "u-bmc-infrastructure",
	}

	// The telemetry service itself might use different configuration
	shutdown, err := Setup(ctx,
		WithServiceName("central-telemetry-collector"),
		WithServiceVersion("2.1.0"),
		WithResourceAttributes(resourceAttrs),
		// Central collector defaults to NoOp but can be configured for export
		// when debugging or monitoring is needed
	)
	if err != nil {
		return fmt.Errorf("CRITICAL: central collector setup failed: %w", err)
	}
	defer shutdown(ctx)

	logger := GetLogger("central-telemetry-collector")
	InfoWithContext(ctx, logger, "Central telemetry collector initialized")

	return nil
}

// ExampleMiddleware demonstrates mandatory tracing middleware for BMC operations.
// All BMC operations must be wrapped with telemetry for observability.
func ExampleMiddleware() error {
	ctx := context.Background()

	// MANDATORY: Setup telemetry to send data to central collector
	shutdown, err := Setup(ctx,
		WithServiceName("bmc-middleware-service"),
		// Data sent to central collector - no direct endpoint needed
	)
	if err != nil {
		return fmt.Errorf("CRITICAL: telemetry setup failed: %w", err)
	}
	defer shutdown(ctx)

	// REQUIRED: Create tracing middleware for all BMC operations
	middleware := TracingMiddleware("bmc-middleware-service")

	// MANDATORY: Wrap all BMC operations with telemetry
	bmcPowerOperation := middleware("bmc_power_control", func(ctx context.Context) error {
		// All BMC hardware operations must be traced
		SetSpanAttributes(ctx,
			StringAttr("bmc.operation", "power_control"),
			StringAttr("hardware.component", "power_supply"),
		)
		time.Sleep(50 * time.Millisecond) // Simulate BMC operation
		return nil
	})

	bmcThermalOperation := middleware("bmc_thermal_monitor", func(ctx context.Context) error {
		// Thermal monitoring operations
		SetSpanAttributes(ctx,
			StringAttr("bmc.operation", "thermal_monitor"),
			StringAttr("hardware.component", "thermal_sensor"),
		)
		time.Sleep(30 * time.Millisecond) // Simulate BMC operation
		return nil
	})

	// Execute traced BMC operations
	if err := bmcPowerOperation(ctx); err != nil {
		return fmt.Errorf("BMC power operation failed: %w", err)
	}

	if err := bmcThermalOperation(ctx); err != nil {
		return fmt.Errorf("BMC thermal operation failed: %w", err)
	}

	return nil
}
