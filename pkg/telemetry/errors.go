// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import "errors"

var (
	// ErrTelemetryInitialization indicates a failure during telemetry system initialization.
	ErrTelemetryInitialization = errors.New("failed to initialize telemetry system")
	// ErrTracerProvider indicates a failure with the OpenTelemetry tracer provider.
	ErrTracerProvider = errors.New("tracer provider error")
	// ErrLoggerProvider indicates a failure with the OpenTelemetry logger provider.
	ErrLoggerProvider = errors.New("logger provider error")
	// ErrMeterProvider indicates a failure with the OpenTelemetry meter provider.
	ErrMeterProvider = errors.New("meter provider error")
	// ErrContextPropagation indicates a failure during context propagation operations.
	ErrContextPropagation = errors.New("context propagation error")
	// ErrSpanCreation indicates a failure to create a new span.
	ErrSpanCreation = errors.New("failed to create span")
	// ErrSpanAttributes indicates a failure to set span attributes.
	ErrSpanAttributes = errors.New("failed to set span attributes")
	// ErrTraceExport indicates a failure during trace export operations.
	ErrTraceExport = errors.New("trace export error")
	// ErrMetricExport indicates a failure during metric export operations.
	ErrMetricExport = errors.New("metric export error")
	// ErrLogExport indicates a failure during log export operations.
	ErrLogExport = errors.New("log export error")
	// ErrInvalidConfiguration indicates an invalid telemetry configuration.
	ErrInvalidConfiguration = errors.New("invalid telemetry configuration")
	// ErrResourceDetection indicates a failure during resource detection.
	ErrResourceDetection = errors.New("resource detection error")
	// ErrInstrumentationSetup indicates a failure during instrumentation setup.
	ErrInstrumentationSetup = errors.New("instrumentation setup error")
	// ErrSamplingConfiguration indicates an invalid sampling configuration.
	ErrSamplingConfiguration = errors.New("invalid sampling configuration")
	// ErrExporterConfiguration indicates an invalid exporter configuration.
	ErrExporterConfiguration = errors.New("invalid exporter configuration")
	// ErrHeaderExtraction indicates a failure to extract trace context from headers.
	ErrHeaderExtraction = errors.New("failed to extract trace context from headers")
	// ErrHeaderInjection indicates a failure to inject trace context into headers.
	ErrHeaderInjection = errors.New("failed to inject trace context into headers")
)
