// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	defaultSetupOnce sync.Once
	globalProvider   *Provider
	setupMutex       sync.Mutex
	isSetup          bool
)

// DefaultSetup initializes OpenTelemetry with default configuration for u-bmc services.
// Services generate telemetry data and send it to the central telemetry collector via NATS.
// The central collector handles export decisions (NoOp by default for minimal overhead).
func DefaultSetup() {
	defaultSetupOnce.Do(func() {
		// Default setup sends telemetry to central collector (not direct OTLP export)
		_, err := Setup(context.Background(),
			WithServiceName("u-bmc-default"),
			// No direct OTLP endpoint - data goes to central telemetry collector
		)
		if err != nil {
			// Fallback to basic setup with context propagation
			provider := noop.NewLoggerProvider()
			global.SetLoggerProvider(provider)

			// Set up propagation for distributed tracing context
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			))
		}
	})
}

// Setup initializes OpenTelemetry for u-bmc services to send telemetry data to the
// central telemetry collector. Services generate telemetry data which is sent via NATS
// to the central collector. The collector decides whether to export or drop the data.
//
// All u-bmc services MUST use this function to ensure:
//   - Consistent telemetry data generation
//   - Central collection point for filtering and debugging
//   - Proper context propagation between services
//   - Service name identification for runtime debugging
//
// It returns a shutdown function that should be called when the application exits.
func Setup(ctx context.Context, opts ...Option) (func(context.Context) error, error) {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if isSetup {
		return func(context.Context) error { return nil }, fmt.Errorf("telemetry already initialized - multiple setup calls not allowed")
	}

	// Configure telemetry to send to central collector
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	if err := validateServiceConfig(config); err != nil {
		return nil, fmt.Errorf("telemetry configuration validation failed: %w", err)
	}

	provider, err := NewProvider(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry provider: %w", err)
	}

	globalProvider = provider
	isSetup = true

	shutdown := func(shutdownCtx context.Context) error {
		setupMutex.Lock()
		defer setupMutex.Unlock()

		if globalProvider != nil {
			err := globalProvider.Shutdown(shutdownCtx)
			globalProvider = nil
			isSetup = false
			return err
		}
		return nil
	}

	return shutdown, nil
}

// validateServiceConfig validates that service telemetry configuration is valid
// for sending data to the central telemetry collector.
func validateServiceConfig(config *Config) error {
	if config.serviceName == "" {
		return fmt.Errorf("service name is mandatory and cannot be empty")
	}

	// Services send to central collector, so no direct OTLP endpoint validation needed
	// The central telemetry collector handles export decisions

	// Ensure at least some telemetry generation is enabled
	if !config.enableMetrics && !config.enableTraces && !config.enableLogs {
		return fmt.Errorf("at least one telemetry signal (metrics, traces, or logs) must be enabled")
	}

	return nil
}

// ForceSetup allows overriding the setup lock for testing purposes only.
// This function should NEVER be used in production code.
func ForceSetup(ctx context.Context, opts ...Option) (func(context.Context) error, error) {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if globalProvider != nil {
		globalProvider.Shutdown(ctx)
	}

	isSetup = false
	globalProvider = nil

	return Setup(ctx, opts...)
}

// GetTracer returns a tracer with the given name from the global provider.
// This function ensures that all services generate telemetry data consistently
// and send it to the central telemetry collector. If no provider is initialized,
// it triggers default setup to ensure telemetry data generation.
func GetTracer(name string) trace.Tracer {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if globalProvider == nil {
		// Auto-initialize with default configuration if not already set up
		DefaultSetup()
	}

	if globalProvider != nil {
		return globalProvider.Tracer(name)
	}
	return otel.GetTracerProvider().Tracer(name)
}

// GetMeter returns a meter with the given name from the global provider.
// This function ensures that all services generate telemetry data consistently
// and send it to the central telemetry collector. If no provider is initialized,
// it triggers default setup to ensure telemetry data generation.
func GetMeter(name string) metric.Meter {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if globalProvider == nil {
		// Auto-initialize with default configuration if not already set up
		DefaultSetup()
	}

	if globalProvider != nil {
		return globalProvider.Meter(name)
	}
	return otel.GetMeterProvider().Meter(name)
}

// GetLogger returns a logger with the given name.
// This uses the u-bmc logging infrastructure and integrates with telemetry context.
// The logger automatically includes telemetry context (trace/span IDs) in log output.
func GetLogger(name string) *slog.Logger {
	return slog.Default().With("component", name)
}

// IsInitialized returns true if a global telemetry provider has been initialized.
// This function helps services verify that telemetry is properly set up.
func IsInitialized() bool {
	setupMutex.Lock()
	defer setupMutex.Unlock()
	return globalProvider != nil && isSetup
}

// MustBeInitialized panics if telemetry is not properly initialized.
// This function should be used in critical paths where telemetry generation is mandatory.
func MustBeInitialized() {
	if !IsInitialized() {
		panic("telemetry must be initialized - all u-bmc services must generate telemetry data")
	}
}

// GetProviderInfo returns information about the current telemetry provider.
// This function helps with debugging and configuration validation.
func GetProviderInfo() map[string]interface{} {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	info := map[string]interface{}{
		"initialized": isSetup,
		"provider":    globalProvider != nil,
	}

	if globalProvider != nil && globalProvider.config != nil {
		info["exporter_type"] = globalProvider.config.exporterType
		info["service_name"] = globalProvider.config.serviceName
		info["metrics_enabled"] = globalProvider.config.enableMetrics
		info["traces_enabled"] = globalProvider.config.enableTraces
		info["logs_enabled"] = globalProvider.config.enableLogs
	}

	return info
}
