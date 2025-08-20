// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/propagation"
)

// DefaultSetup initializes OpenTelemetry with default configuration.
// It sets up a no-op logger provider for logging and configures
// text map propagation with TraceContext and Baggage propagators
// for distributed tracing context propagation across service boundaries.
func DefaultSetup() {
	provider := noop.NewLoggerProvider()
	global.SetLoggerProvider(provider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}
