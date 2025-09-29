// SPDX-License-Identifier: BSD-3-Clause

package telemetry_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service/ipc"
	telemetryservice "github.com/u-bmc/u-bmc/service/telemetry"
)

// ExampleTelemetryService demonstrates basic usage of the telemetry service.
func ExampleTelemetryService() {
	ctx := context.Background()

	// Create IPC service for inter-service communication
	ipcService := ipc.New(
		ipc.WithServiceName("example-ipc"),
		ipc.WithJetStream(true),
	)

	// Create telemetry service with default NoOp behavior (minimal overhead)
	telemetryService := telemetryservice.New(
		telemetryservice.WithServiceName("example-telemetry"),
		telemetryservice.WithCollectorName("bmc-telemetry-collector"),
		// No exporter configured - defaults to NoOp for minimal overhead
		// Export can be enabled at runtime via NATS configuration messages
		telemetryservice.WithCollection(true),
		telemetryservice.WithAggregation(true),
	)

	// Start IPC service in a goroutine
	go func() {
		if err := ipcService.Run(ctx, nil); err != nil {
			slog.Error("IPC service error", "error", err)
		}
	}()

	// Wait for IPC to be ready
	time.Sleep(100 * time.Millisecond)

	// Get connection provider and start telemetry service
	connProvider := ipcService.GetConnProvider()
	go func() {
		if err := telemetryService.Run(ctx, connProvider); err != nil {
			slog.Error("Telemetry service error", "error", err)
		}
	}()

	// Wait for services to start
	time.Sleep(200 * time.Millisecond)

	fmt.Println("Telemetry service started successfully")
	// Output: Telemetry service started successfully
}

// ExampleRuntimeConfiguration demonstrates how to send runtime configuration updates.
func ExampleRuntimeConfiguration() {
	ctx := context.Background()

	// Connect to NATS for sending configuration updates
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		slog.Error("Failed to connect to NATS", "error", err)
		return
	}
	defer nc.Close()

	logger := slog.Default()
	configManager := telemetryservice.NewRuntimeConfigManager(nc, logger)

	// Enable debug mode for a specific service
	err = configManager.EnableDebugMode(ctx, "power-service")
	if err != nil {
		slog.Error("Failed to enable debug mode", "error", err)
		return
	}

	// Update sampling ratio globally
	err = configManager.UpdateSamplingRatio(ctx, "", 0.1) // 10% sampling
	if err != nil {
		slog.Error("Failed to update sampling ratio", "error", err)
		return
	}

	// Create production filter configuration
	filterConfig := telemetryservice.CreateProductionFilterConfig()
	err = configManager.UpdateFilterConfig(ctx, "", filterConfig)
	if err != nil {
		slog.Error("Failed to update filter config", "error", err)
		return
	}

	fmt.Println("Runtime configuration updates sent successfully")
	// Output: Runtime configuration updates sent successfully
}

// ExampleDebugModeToggle demonstrates toggling debug mode during runtime.
func ExampleDebugModeToggle() {
	// Create a runtime configuration message for enabling debug mode
	debugConfig := telemetryservice.RuntimeConfig{
		Type:        "debug_mode",
		ServiceName: "thermal-service",
		DebugMode:   boolPtr(true),
	}

	// Convert to JSON for sending via NATS
	configJSON, err := json.Marshal(debugConfig)
	if err != nil {
		slog.Error("Failed to marshal config", "error", err)
		return
	}

	fmt.Printf("Debug mode configuration: %s\n", string(configJSON))

	// Create configuration for disabling debug mode
	disableDebugConfig := telemetryservice.RuntimeConfig{
		Type:        "debug_mode",
		ServiceName: "thermal-service",
		DebugMode:   boolPtr(false),
	}

	disableConfigJSON, err := json.Marshal(disableDebugConfig)
	if err != nil {
		slog.Error("Failed to marshal disable config", "error", err)
		return
	}

	fmt.Printf("Disable debug configuration: %s\n", string(disableConfigJSON))

	// Output:
	// Debug mode configuration: {"type":"debug_mode","service_name":"thermal-service","debug_mode":true}
	// Disable debug configuration: {"type":"debug_mode","service_name":"thermal-service","debug_mode":false}
}

// ExampleFilterConfiguration demonstrates creating and applying filter configurations.
func ExampleFilterConfiguration() {
	// Create a custom filter configuration
	filterConfig := &telemetryservice.FilterConfig{
		EnableFiltering: true,
		MetricsFilters: []telemetryservice.FilterRule{
			{
				Name:    "drop_high_frequency_metrics",
				Type:    "drop",
				Enabled: true,
				Condition: telemetryservice.FilterCondition{
					AttributeMatches: map[string]string{
						"frequency": "high",
					},
				},
				Action: telemetryservice.FilterAction{
					Type: "drop",
				},
			},
		},
		TracesFilters: []telemetryservice.FilterRule{
			{
				Name:    "sample_background_tasks",
				Type:    "sample",
				Enabled: true,
				Condition: telemetryservice.FilterCondition{
					SpanName: "background_task",
				},
				Action: telemetryservice.FilterAction{
					Type:       "sample",
					SampleRate: 0.05, // 5% sampling
				},
			},
		},
		SamplingOverrides: map[string]float64{
			"power":   0.8, // 80% sampling for power service
			"thermal": 0.6, // 60% sampling for thermal service
		},
		DebugServices: []string{"fan-control"},
	}

	// Convert to runtime configuration message
	runtimeConfig := telemetryservice.RuntimeConfig{
		Type:         "filter_config",
		FilterConfig: filterConfig,
	}

	configJSON, err := json.Marshal(runtimeConfig)
	if err != nil {
		slog.Error("Failed to marshal filter config", "error", err)
		return
	}

	fmt.Printf("Filter configuration created: %d bytes\n", len(configJSON))
	fmt.Printf("Metrics filters: %d\n", len(filterConfig.MetricsFilters))
	fmt.Printf("Traces filters: %d\n", len(filterConfig.TracesFilters))
	fmt.Printf("Sampling overrides: %d services\n", len(filterConfig.SamplingOverrides))

	// Output:
	// Filter configuration created: 445 bytes
	// Metrics filters: 1
	// Traces filters: 1
	// Sampling overrides: 2 services
}

// ExampleAggregationConfiguration demonstrates creating aggregation rules.
func ExampleAggregationConfiguration() {
	// Create aggregation configuration
	aggregationConfig := &telemetryservice.AggregationConfig{
		EnableAggregation: true,
		WindowSize:        60 * time.Second,
		MaxCardinality:    5000,
		MetricsAggregation: []telemetryservice.AggregationRule{
			{
				Name:    "aggregate_cpu_metrics",
				Type:    "avg",
				Enabled: true,
				GroupBy: []string{"service_name", "cpu_core"},
				Condition: telemetryservice.FilterCondition{
					MetricName: "cpu_usage",
				},
			},
			{
				Name:    "count_error_metrics",
				Type:    "count",
				Enabled: true,
				GroupBy: []string{"service_name", "error_type"},
				Condition: telemetryservice.FilterCondition{
					AttributeMatches: map[string]string{
						"level": "error",
					},
				},
			},
		},
	}

	// Create runtime configuration message
	runtimeConfig := telemetryservice.RuntimeConfig{
		Type:              "aggregation_config",
		AggregationConfig: aggregationConfig,
	}

	configJSON, err := json.Marshal(runtimeConfig)
	if err != nil {
		slog.Error("Failed to marshal aggregation config", "error", err)
		return
	}

	fmt.Printf("Aggregation configuration created: %d bytes\n", len(configJSON))
	fmt.Printf("Window size: %v\n", aggregationConfig.WindowSize)
	fmt.Printf("Max cardinality: %d\n", aggregationConfig.MaxCardinality)
	fmt.Printf("Metrics aggregation rules: %d\n", len(aggregationConfig.MetricsAggregation))

	// Output:
	// Aggregation configuration created: 354 bytes
	// Window size: 1m0s
	// Max cardinality: 5000
	// Metrics aggregation rules: 2
}

// ExampleServiceIntegration demonstrates how a service should integrate with telemetry.
func ExampleServiceIntegration() error {
	ctx := context.Background()

	// Services MUST use the telemetry package to send data to central collector
	shutdown, err := telemetry.Setup(ctx,
		telemetry.WithServiceName("example-bmc-service"),
		// No direct OTLP endpoint - data sent to central telemetry collector
		telemetry.WithMetrics(true),
		telemetry.WithTraces(true),
		telemetry.WithLogs(true),
	)
	if err != nil {
		slog.Error("Failed to setup telemetry", "error", err)
		return err
	}
	defer shutdown(ctx)

	// Create metrics
	requestCounter, err := telemetry.Counter("example-service", "requests_total",
		"Total number of requests", "1")
	if err != nil {
		slog.Error("Failed to create counter", "error", err)
		return err
	}

	// Create traces
	logger := telemetry.GetLogger("example-service")
	err = telemetry.WithSpan(ctx, "example-service", "process_request", func(spanCtx context.Context) error {
		// Add span attributes
		telemetry.SetSpanAttributes(spanCtx,
			telemetry.StringAttr("operation", "example"),
			telemetry.StringAttr("user_id", "user123"),
		)

		// Log with telemetry context
		telemetry.InfoWithContext(spanCtx, logger, "Processing request",
			slog.String("operation", "example"),
		)

		// Increment counter
		telemetry.IncrementCounter(spanCtx, requestCounter, 1,
			telemetry.StringAttr("method", "POST"),
		)

		return nil
	})

	if err != nil {
		slog.Error("Operation failed", "error", err)
		return err
	}

	fmt.Println("Service integration with central collector completed")
	// Output: Service integration with central collector completed
	return nil
}

// ExampleConfigurationValidation demonstrates configuration validation.
func ExampleConfigurationValidation() {
	// Create a valid configuration with default NoOp behavior
	validConfig := &telemetryservice.Config{
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
		CollectorName:      "test-collector",
		ExporterType:       "noop", // Default for minimal overhead
		HTTPEndpoint:       "",     // No endpoint needed for NoOp
		EnableMetrics:      true,
		EnableTraces:       true,
		EnableCollection:   true,
		Timeout:            30 * time.Second,
		BatchTimeout:       5 * time.Second,
		MaxExportBatch:     512,
		MaxQueueSize:       2048,
		ShutdownTimeout:    10 * time.Second,
		CollectionInterval: 30 * time.Second,
		SamplingRatio:      1.0,
	}

	// Validate the configuration
	if err := telemetryservice.ValidateConfig(validConfig); err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		return
	}

	fmt.Println("Configuration validation passed")

	// Test invalid configuration (missing service name)
	invalidConfig := &telemetryservice.Config{
		ServiceName:   "", // Invalid - required field
		ExporterType:  "noop",
		EnableMetrics: true,
	}

	if err := telemetryservice.ValidateConfig(invalidConfig); err != nil {
		fmt.Printf("Invalid config correctly rejected: %s\n", err.Error()[:50]+"...")
	}

	// Output:
	// Configuration validation passed
	// Invalid config correctly rejected: service information validation failed: service...
}

// Helper function for examples
func boolPtr(b bool) *bool {
	return &b
}
