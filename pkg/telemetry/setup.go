// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	lognoop "go.opentelemetry.io/otel/log/noop"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// DefaultSetup initializes OpenTelemetry with default no-op configuration.
// It sets up no-op providers for tracing, metrics, and logging, and configures
// text map propagation with TraceContext and Baggage propagators
// for distributed tracing context propagation across service boundaries.
func DefaultSetup() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetTracerProvider(tracenoop.NewTracerProvider())
	otel.SetMeterProvider(metricnoop.NewMeterProvider())
	global.SetLoggerProvider(lognoop.NewLoggerProvider())
}
