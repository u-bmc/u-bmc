// SPDX-License-Identifier: BSD-3-Clause

// Package telemetry provides a telemetry collector and aggregator service for the u-bmc system.
// It acts as the central observability hub for all BMC components, ensuring mandatory telemetry
// collection while providing runtime configuration capabilities for filtering, aggregation, and
// debug control.
//
// # Core Architecture
//
// The telemetry service enforces that all u-bmc services generate telemetry data and send it
// to the central telemetry collector. By default, the telemetry service operates in NoOp mode
// (dropping all data) to minimize overhead in production. The service can be dynamically
// reconfigured at runtime to enable OTLP export when debugging or monitoring is needed.
//
// # Key Features
//
//   - Mandatory telemetry generation from all u-bmc services (sent to central collector)
//   - NoOp default behavior to minimize production overhead
//   - Runtime reconfiguration via NATS messaging for dynamic export control
//   - Filtering and aggregation when export is enabled
//   - Support for debug mode toggle during runtime without restarts
//   - OTLP export to multiple endpoints when enabled (HTTP, gRPC, or dual)
//   - Integration with u-bmc logging and IPC infrastructure
//
// # Operation Modes
//
// The telemetry service supports multiple exporter configurations:
//
//   - NoOp: Default mode that discards telemetry data for minimal overhead
//   - OTLP HTTP: Exports to OTLP-compatible endpoints via HTTP (when configured)
//   - OTLP gRPC: Exports to OTLP-compatible endpoints via gRPC (when configured)
//   - Dual: Exports to both HTTP and gRPC endpoints simultaneously (when configured)
//
// # Runtime Configuration
//
// The service supports runtime reconfiguration through JSON messages sent via NATS.
// Configuration updates can control:
//
//   - Filtering rules for metrics, traces, and logs
//   - Aggregation rules and time windows
//   - Sampling ratios per service or globally
//   - Debug mode for specific services or all services
//   - Exporter endpoint configuration
//
// Example runtime configuration message:
//
//	{
//	  "type": "debug_mode",
//	  "service_name": "bmc-power",
//	  "debug_mode": true
//	}
//
// # Usage
//
// The telemetry service is typically started early in the BMC system lifecycle:
//
//	telemetryService := telemetry.New(
//		telemetry.WithServiceName("bmc-telemetry"),
//		// Note: No exporter configured - defaults to NoOp for minimal overhead
//		telemetry.WithCollection(true),
//		telemetry.WithAggregation(true),
//	)
//
//	// Start the service - will collect but not export by default
//	err := telemetryService.Run(ctx, ipcConnProvider)
//
// # Message Subjects
//
// The service uses the following NATS subjects for communication:
//
//   - telemetry.metrics.* - Metrics collection from all services
//   - telemetry.traces.* - Trace collection from all services
//   - telemetry.logs.* - Log collection from all services
//   - telemetry.config.{service-name} - Service-specific configuration
//   - telemetry.config.global - Global configuration updates
//
// # Integration with u-bmc Services
//
// All u-bmc services must use the centralized telemetry package (pkg/telemetry)
// to generate and send telemetry data to the central collector. This ensures:
//
//   - Consistent telemetry data format across all services
//   - Centralized control over export behavior (NoOp by default)
//   - Proper context propagation between services
//   - Integration with the global logging infrastructure
//   - Ability to enable debugging on any service at runtime
//
// Services should initialize telemetry using the telemetry package:
//
//	import "github.com/u-bmc/u-bmc/pkg/telemetry"
//
//	// In service initialization - all services send data to central collector
//	shutdown, err := telemetry.Setup(ctx,
//		telemetry.WithServiceName("my-service"),
//		// Services send to central collector, not direct OTLP endpoints
//	)
//	if err != nil {
//		return fmt.Errorf("failed to setup telemetry: %w", err)
//	}
//	defer shutdown(ctx)
//
// # Configuration
//
// The telemetry service supports extensive configuration options:
//
//   - WithServiceName: Set the service name
//   - WithCollectorName: Set the telemetry collector name
//   - WithExporterType: Configure the exporter type
//   - WithHTTPEndpoint/WithGRPCEndpoint: Set OTLP endpoints
//   - WithCollection: Enable/disable collection from other services
//   - WithAggregation: Enable/disable aggregation
//   - WithFilterConfig: Set initial filtering rules
//   - WithAggregationConfig: Set initial aggregation rules
//
// # Performance Considerations
//
// The service is designed to handle high-throughput telemetry data with:
//
//   - NoOp default behavior for minimal production overhead
//   - Configurable batch sizes and timeouts when export is enabled
//   - Runtime sampling ratio adjustments
//   - Filtering to drop unnecessary data
//   - Aggregation to reduce data volume
//   - Efficient NATS-based message passing
//   - Dynamic export enable/disable for debugging
//
// # Security
//
// The service supports secure connections to OTLP endpoints with:
//
//   - TLS encryption for gRPC and HTTP endpoints
//   - Custom headers for authentication
//   - Configurable timeouts and connection limits
//
// The runtime configuration system uses NATS subject-based authorization
// to control which services can modify telemetry behavior.
package telemetry
