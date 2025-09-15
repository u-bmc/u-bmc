// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
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
