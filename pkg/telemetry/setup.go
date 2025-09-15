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
)

// DefaultSetup initializes OpenTelemetry with default no-op configuration.
// It sets up a no-op logger provider for logging and configures
// text map propagation with TraceContext and Baggage propagators
// for distributed tracing context propagation across service boundaries.
func DefaultSetup() {
	defaultSetupOnce.Do(func() {
		provider := noop.NewLoggerProvider()
		global.SetLoggerProvider(provider)

		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	})
}

// Setup initializes OpenTelemetry with the provided configuration options.
// This function should be called once at application startup.
// It returns a shutdown function that should be called when the application exits.
func Setup(ctx context.Context, opts ...Option) (func(context.Context) error, error) {
	provider, err := NewProvider(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry provider: %w", err)
	}

	globalProvider = provider

	shutdown := func(shutdownCtx context.Context) error {
		if globalProvider != nil {
			return globalProvider.Shutdown(shutdownCtx)
		}
		return nil
	}

	return shutdown, nil
}

// GetTracer returns a tracer with the given name from the global provider.
// If no global provider is set, it returns a no-op tracer.
func GetTracer(name string) trace.Tracer {
	if globalProvider != nil {
		return globalProvider.Tracer(name)
	}
	return otel.GetTracerProvider().Tracer(name)
}

// GetMeter returns a meter with the given name from the global provider.
// If no global provider is set, it returns a no-op meter.
func GetMeter(name string) metric.Meter {
	if globalProvider != nil {
		return globalProvider.Meter(name)
	}
	return otel.GetMeterProvider().Meter(name)
}

// GetLogger returns a logger with the given name.
// This uses slog's default logger which can be configured via the log package.
func GetLogger(name string) *slog.Logger {
	return slog.Default().With("component", name)
}

// IsInitialized returns true if a global telemetry provider has been initialized.
func IsInitialized() bool {
	return globalProvider != nil
}
