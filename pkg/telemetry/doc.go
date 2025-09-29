// SPDX-License-Identifier: BSD-3-Clause

// Package telemetry provides mandatory telemetry abstractions for u-bmc services.
// It ensures all services generate telemetry data and send it to the central telemetry
// collector, which handles export decisions to minimize overhead and enable runtime debugging.
//
// # Mandatory Telemetry Generation
//
// ALL u-bmc services MUST use this package to generate telemetry data. This ensures:
//   - Consistent telemetry data format across all services
//   - All services can be debugged at runtime when needed
//   - Proper context propagation between services
//   - Integration with u-bmc logging infrastructure
//   - Central collection point for filtering and aggregation
//
// Services send telemetry data to the central telemetry collector via NATS messaging.
// The collector decides whether to export (for debugging) or drop (for performance).
//
// # Central Collection Architecture
//
// The package uses a central collection model:
//   - Services generate telemetry data using this package
//   - Data is sent to the central telemetry collector via NATS
//   - The collector defaults to NoOp (dropping data) for minimal overhead
//   - Export can be enabled at runtime for debugging or monitoring
//   - No direct OTLP configuration by individual services
//
// # Service Integration
//
// Services must initialize telemetry during startup:
//
//	import "github.com/u-bmc/u-bmc/pkg/telemetry"
//
//	// In service Run method - connects to central collector
//	shutdown, err := telemetry.Setup(ctx,
//		telemetry.WithServiceName("my-service"),
//		// No direct OTLP endpoint - sends to central collector
//	)
//	if err != nil {
//		return fmt.Errorf("telemetry setup failed: %w", err)
//	}
//	defer shutdown(ctx)
//
// # Telemetry Usage
//
// Use the provided helper functions to create spans, metrics, and logs:
//
//	// Create spans - automatically sent to central collector
//	err := telemetry.WithSpan(ctx, "service-name", "operation", func(spanCtx context.Context) error {
//		telemetry.SetSpanAttributes(spanCtx,
//			telemetry.StringAttr("user_id", "123"),
//		)
//		return performOperation(spanCtx)
//	})
//
//	// Create metrics - automatically sent to central collector
//	counter, err := telemetry.Counter("service-name", "requests_total", "desc", "1")
//	telemetry.IncrementCounter(ctx, counter, 1)
//
//	// Structured logging with telemetry context
//	logger := telemetry.GetLogger("service-name")
//	telemetry.InfoWithContext(ctx, logger, "Operation completed")
//
// # Runtime Configuration
//
// The central telemetry collector provides runtime configuration capabilities through NATS.
// This allows dynamic control without service restarts:
//   - Enable/disable export for debugging (NoOp by default)
//   - Debug mode toggling for specific services
//   - Sampling ratio adjustments
//   - Filtering rule updates
//   - Aggregation configuration changes
//
// # Abstraction Layer
//
// This package abstracts all OpenTelemetry types and functions to:
//   - Prevent services from directly configuring OTLP exporters
//   - Ensure all telemetry data flows through the central collector
//   - Enable centralized export control (NoOp by default, export when needed)
//   - Provide consistent telemetry APIs across all services
package telemetry
