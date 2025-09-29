// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan creates a new span with the given name and options.
// It returns the span and a context containing the span.
func StartSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := GetTracer(tracerName)
	return tracer.Start(ctx, spanName, opts...)
}

// RecordError records an error on the span in the given context.
// If no span is found in the context, this is a no-op.
func RecordError(ctx context.Context, err error, description string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.description", description),
		))
		span.SetStatus(codes.Error, description)
	}
}

// SetSpanAttributes sets attributes on the span in the given context.
// If no span is found in the context, this is a no-op.
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds an event to the span in the given context.
// If no span is found in the context, this is a no-op.
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetSpanStatus sets the status of the span in the given context.
// If no span is found in the context, this is a no-op.
func SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(code, description)
	}
}

// WithSpan executes a function within a new span context.
// The span is automatically finished when the function returns.
func WithSpan(ctx context.Context, tracerName, spanName string, fn func(context.Context) error, opts ...trace.SpanStartOption) error {
	spanCtx, span := StartSpan(ctx, tracerName, spanName, opts...)
	defer span.End()

	if err := fn(spanCtx); err != nil {
		RecordError(spanCtx, err, "operation failed")
		return err
	}

	return nil
}

// Counter creates or retrieves a counter metric with the given name.
func Counter(meterName, name, description, unit string) (metric.Int64Counter, error) {
	meter := GetMeter(meterName)
	return meter.Int64Counter(name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
}

// Histogram creates or retrieves a histogram metric with the given name.
func Histogram(meterName, name, description, unit string) (metric.Float64Histogram, error) {
	meter := GetMeter(meterName)
	return meter.Float64Histogram(name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
}

// Gauge creates or retrieves a gauge metric with the given name.
func Gauge(meterName, name, description, unit string) (metric.Int64ObservableGauge, error) {
	meter := GetMeter(meterName)
	return meter.Int64ObservableGauge(name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
}

// LogWithContext logs a message with telemetry context information.
// It extracts trace information from the context and includes it in the log.
func LogWithContext(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, attrs ...slog.Attr) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		attrs = append(attrs,
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	logger.LogAttrs(ctx, level, msg, attrs...)
}

// InfoWithContext logs an info message with telemetry context.
func InfoWithContext(ctx context.Context, logger *slog.Logger, msg string, attrs ...slog.Attr) {
	LogWithContext(ctx, logger, slog.LevelInfo, msg, attrs...)
}

// WarnWithContext logs a warning message with telemetry context.
func WarnWithContext(ctx context.Context, logger *slog.Logger, msg string, attrs ...slog.Attr) {
	LogWithContext(ctx, logger, slog.LevelWarn, msg, attrs...)
}

// ErrorWithContext logs an error message with telemetry context.
func ErrorWithContext(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	LogWithContext(ctx, logger, slog.LevelError, msg, attrs...)
}

// DebugWithContext logs a debug message with telemetry context.
func DebugWithContext(ctx context.Context, logger *slog.Logger, msg string, attrs ...slog.Attr) {
	LogWithContext(ctx, logger, slog.LevelDebug, msg, attrs...)
}

// RecordDuration records the duration of an operation as a histogram metric.
func RecordDuration(ctx context.Context, histogram metric.Float64Histogram, duration float64, attrs ...attribute.KeyValue) {
	opt := metric.WithAttributes(attrs...)
	histogram.Record(ctx, duration, opt)
}

// IncrementCounter increments a counter metric by the given value.
func IncrementCounter(ctx context.Context, counter metric.Int64Counter, value int64, attrs ...attribute.KeyValue) {
	opt := metric.WithAttributes(attrs...)
	counter.Add(ctx, value, opt)
}

// StringAttr creates a string attribute for telemetry.
func StringAttr(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// IntAttr creates an integer attribute for telemetry.
func IntAttr(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// Int64Attr creates an int64 attribute for telemetry.
func Int64Attr(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// Float64Attr creates a float64 attribute for telemetry.
func Float64Attr(key string, value float64) attribute.KeyValue {
	return attribute.Float64(key, value)
}

// BoolAttr creates a boolean attribute for telemetry.
func BoolAttr(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}

// StringSliceAttr creates a string slice attribute for telemetry.
func StringSliceAttr(key string, value []string) attribute.KeyValue {
	return attribute.StringSlice(key, value)
}

// IntSliceAttr creates an integer slice attribute for telemetry.
func IntSliceAttr(key string, value []int) attribute.KeyValue {
	return attribute.IntSlice(key, value)
}

// TracingMiddleware returns a middleware function that creates spans for operations.
func TracingMiddleware(tracerName string) func(operation string, fn func(context.Context) error) func(context.Context) error {
	return func(operation string, fn func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error {
			return WithSpan(ctx, tracerName, operation, fn)
		}
	}
}

// MustCreateCounter creates a counter metric and panics on error.
// This should only be used during initialization when errors are not recoverable.
func MustCreateCounter(meterName, name, description, unit string) metric.Int64Counter {
	counter, err := Counter(meterName, name, description, unit)
	if err != nil {
		panic(fmt.Sprintf("failed to create counter %s: %v", name, err))
	}
	return counter
}

// MustCreateHistogram creates a histogram metric and panics on error.
// This should only be used during initialization when errors are not recoverable.
func MustCreateHistogram(meterName, name, description, unit string) metric.Float64Histogram {
	histogram, err := Histogram(meterName, name, description, unit)
	if err != nil {
		panic(fmt.Sprintf("failed to create histogram %s: %v", name, err))
	}
	return histogram
}

// InjectContext injects telemetry context into a propagation.TextMapCarrier.
// This abstracts OpenTelemetry's context injection to prevent direct usage.
func InjectContext(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

// ExtractContext extracts telemetry context from a propagation.TextMapCarrier.
// This abstracts OpenTelemetry's context extraction to prevent direct usage.
func ExtractContext(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// GetCurrentSpan returns the current span from the context.
// This abstracts OpenTelemetry's span retrieval to prevent direct usage.
func GetCurrentSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// IsSpanRecording returns true if the span in the context is recording.
// This abstracts OpenTelemetry's span recording check to prevent direct usage.
func IsSpanRecording(ctx context.Context) bool {
	span := trace.SpanFromContext(ctx)
	return span.IsRecording()
}

// GetTraceID returns the trace ID from the current span context.
// This abstracts OpenTelemetry's trace ID retrieval to prevent direct usage.
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from the current span context.
// This abstracts OpenTelemetry's span ID retrieval to prevent direct usage.
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// GetBaggage retrieves baggage value for a key from the context.
// This abstracts OpenTelemetry's baggage retrieval to prevent direct usage.
func GetBaggage(ctx context.Context, key string) string {
	// Implementation would use otel/baggage package when needed
	return ""
}

// SetBaggage sets baggage key-value pair in the context.
// This abstracts OpenTelemetry's baggage setting to prevent direct usage.
func SetBaggage(ctx context.Context, key, value string) context.Context {
	// Implementation would use otel/baggage package when needed
	return ctx
}

// CreateAttributeSet creates an attribute set from key-value pairs.
// This abstracts OpenTelemetry's attribute set creation to prevent direct usage.
func CreateAttributeSet(attrs ...attribute.KeyValue) attribute.Set {
	return attribute.NewSet(attrs...)
}

// SpanKindServer returns the server span kind for span creation.
// This abstracts OpenTelemetry's span kind constants to prevent direct usage.
func SpanKindServer() trace.SpanStartOption {
	return trace.WithSpanKind(trace.SpanKindServer)
}

// SpanKindClient returns the client span kind for span creation.
// This abstracts OpenTelemetry's span kind constants to prevent direct usage.
func SpanKindClient() trace.SpanStartOption {
	return trace.WithSpanKind(trace.SpanKindClient)
}

// SpanKindInternal returns the internal span kind for span creation.
// This abstracts OpenTelemetry's span kind constants to prevent direct usage.
func SpanKindInternal() trace.SpanStartOption {
	return trace.WithSpanKind(trace.SpanKindInternal)
}

// SpanKindProducer returns the producer span kind for span creation.
// This abstracts OpenTelemetry's span kind constants to prevent direct usage.
func SpanKindProducer() trace.SpanStartOption {
	return trace.WithSpanKind(trace.SpanKindProducer)
}

// SpanKindConsumer returns the consumer span kind for span creation.
// This abstracts OpenTelemetry's span kind constants to prevent direct usage.
func SpanKindConsumer() trace.SpanStartOption {
	return trace.WithSpanKind(trace.SpanKindConsumer)
}

// WithSpanLinks creates span start option with links to other spans.
// This abstracts OpenTelemetry's span linking to prevent direct usage.
func WithSpanLinks(links ...trace.Link) trace.SpanStartOption {
	return trace.WithLinks(links...)
}

// CreateSpanLink creates a link to another span.
// This abstracts OpenTelemetry's span link creation to prevent direct usage.
func CreateSpanLink(spanContext trace.SpanContext, attrs ...attribute.KeyValue) trace.Link {
	return trace.Link{
		SpanContext: spanContext,
		Attributes:  attrs,
	}
}

// StatusOK returns the OK status code for spans.
// This abstracts OpenTelemetry's status codes to prevent direct usage.
func StatusOK() codes.Code {
	return codes.Ok
}

// StatusError returns the Error status code for spans.
// This abstracts OpenTelemetry's status codes to prevent direct usage.
func StatusError() codes.Code {
	return codes.Error
}

// StatusUnset returns the Unset status code for spans.
// This abstracts OpenTelemetry's status codes to prevent direct usage.
func StatusUnset() codes.Code {
	return codes.Unset
}

// WithMetricAttributes creates metric option with attributes.
// This abstracts OpenTelemetry's metric attributes to prevent direct usage.
func WithMetricAttributes(attrs ...attribute.KeyValue) metric.RecordOption {
	return metric.WithAttributes(attrs...)
}

// WithHistogramBuckets creates histogram option with explicit buckets.
// This abstracts OpenTelemetry's histogram configuration to prevent direct usage.
func WithHistogramBuckets(buckets ...float64) metric.Float64HistogramOption {
	return metric.WithExplicitBucketBoundaries(buckets...)
}
