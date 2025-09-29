// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Provider encapsulates OpenTelemetry providers for metrics, traces, and logs.
type Provider struct {
	config        *Config
	traceProvider *trace.TracerProvider
	meterProvider *sdkmetric.MeterProvider
	logProvider   *log.LoggerProvider
	resource      *resource.Resource
}

// NewProvider creates a new telemetry provider with the given configuration options.
func NewProvider(opts ...Option) (*Provider, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfiguration, err)
	}

	res, err := createResource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := &Provider{
		config:   config,
		resource: res,
	}

	if err := provider.setupProviders(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExporterSetupFailed, err)
	}

	provider.setGlobalProviders()
	setupTextMapPropagator()

	return provider, nil
}

// Tracer returns a tracer with the given name.
func (p *Provider) Tracer(name string) oteltrace.Tracer {
	if p.traceProvider == nil {
		return tracenoop.NewTracerProvider().Tracer(name)
	}
	return p.traceProvider.Tracer(name)
}

// Meter returns a meter with the given name.
func (p *Provider) Meter(name string) metric.Meter {
	if p.meterProvider == nil {
		return metricnoop.NewMeterProvider().Meter(name)
	}
	return p.meterProvider.Meter(name)
}

// Logger returns a logger with the given name.
func (p *Provider) Logger(name string) *slog.Logger {
	return slog.Default()
}

// Shutdown gracefully shuts down all providers.
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error

	if p.traceProvider != nil {
		if err := p.traceProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("trace provider shutdown: %w", err))
		}
	}

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
		}
	}

	if p.logProvider != nil {
		if err := p.logProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("log provider shutdown: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %v", ErrShutdownFailed, errs)
	}

	return nil
}

// validateConfig validates the telemetry configuration.
func validateConfig(config *Config) error {
	switch config.exporterType {
	case NoOp:
		// No validation needed for NoOp
	case OTLPHTTP:
		if config.httpEndpoint == "" {
			return ErrMissingEndpoint
		}
	case OTLPgRPC:
		if config.grpcEndpoint == "" {
			return ErrMissingEndpoint
		}
	case Dual:
		if config.httpEndpoint == "" || config.grpcEndpoint == "" {
			return ErrMissingEndpoint
		}
	default:
		return ErrInvalidExporterType
	}

	if config.samplingRatio < 0.0 || config.samplingRatio > 1.0 {
		return fmt.Errorf("sampling ratio must be between 0.0 and 1.0, got %f", config.samplingRatio)
	}

	return nil
}

// createResource creates an OpenTelemetry resource with service information.
func createResource(config *Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(config.serviceName),
		semconv.ServiceVersion(config.serviceVersion),
	}

	for key, value := range config.resourceAttrs {
		attrs = append(attrs, attribute.String(key, value))
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			attrs...,
		),
	)
}

// setupProviders initializes the trace, metric, and log providers based on configuration.
func (p *Provider) setupProviders() error {
	if p.config.enableTraces {
		if err := p.setupTraceProvider(); err != nil {
			return fmt.Errorf("failed to setup trace provider: %w", err)
		}
	}

	if p.config.enableMetrics {
		if err := p.setupMeterProvider(); err != nil {
			return fmt.Errorf("failed to setup meter provider: %w", err)
		}
	}

	if p.config.enableLogs {
		if err := p.setupLogProvider(); err != nil {
			return fmt.Errorf("failed to setup log provider: %w", err)
		}
	}

	return nil
}

// setupTraceProvider initializes the trace provider.
func (p *Provider) setupTraceProvider() error {
	if p.config.exporterType == NoOp {
		p.traceProvider = trace.NewTracerProvider()
		return nil
	}

	exporters, err := p.createTraceExporters()
	if err != nil {
		return err
	}

	opts := []trace.TracerProviderOption{
		trace.WithResource(p.resource),
		trace.WithSampler(trace.TraceIDRatioBased(p.config.samplingRatio)),
	}

	for _, exporter := range exporters {
		opts = append(opts, trace.WithBatcher(exporter,
			trace.WithBatchTimeout(p.config.batchTimeout),
			trace.WithMaxExportBatchSize(p.config.maxExportBatch),
			trace.WithMaxQueueSize(p.config.maxQueueSize),
		))
	}

	p.traceProvider = trace.NewTracerProvider(opts...)
	return nil
}

// setupMeterProvider initializes the meter provider.
func (p *Provider) setupMeterProvider() error {
	if p.config.exporterType == NoOp {
		p.meterProvider = sdkmetric.NewMeterProvider()
		return nil
	}

	readers, err := p.createMetricReaders()
	if err != nil {
		return err
	}

	opts := []sdkmetric.Option{
		sdkmetric.WithResource(p.resource),
	}

	for _, reader := range readers {
		opts = append(opts, sdkmetric.WithReader(reader))
	}

	p.meterProvider = sdkmetric.NewMeterProvider(opts...)
	return nil
}

// setupLogProvider initializes the log provider.
func (p *Provider) setupLogProvider() error {
	if p.config.exporterType == NoOp {
		p.logProvider = log.NewLoggerProvider()
		return nil
	}

	processors, err := p.createLogProcessors()
	if err != nil {
		return err
	}

	opts := []log.LoggerProviderOption{
		log.WithResource(p.resource),
	}

	for _, processor := range processors {
		opts = append(opts, log.WithProcessor(processor))
	}

	p.logProvider = log.NewLoggerProvider(opts...)
	return nil
}

// createTraceExporters creates trace exporters based on configuration.
func (p *Provider) createTraceExporters() ([]trace.SpanExporter, error) {
	var exporters []trace.SpanExporter

	switch p.config.exporterType {
	case OTLPHTTP, Dual:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(p.config.httpEndpoint),
			otlptracehttp.WithHeaders(p.config.headers),
			otlptracehttp.WithTimeout(p.config.timeout),
		}
		if p.config.insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exporter, err := otlptracehttp.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP trace exporter: %w", err)
		}
		exporters = append(exporters, exporter)
	}

	switch p.config.exporterType {
	case OTLPgRPC, Dual:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(p.config.grpcEndpoint),
			otlptracegrpc.WithHeaders(p.config.headers),
			otlptracegrpc.WithTimeout(p.config.timeout),
		}
		if p.config.insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err := otlptracegrpc.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC trace exporter: %w", err)
		}
		exporters = append(exporters, exporter)
	}

	return exporters, nil
}

// createMetricReaders creates metric readers based on configuration.
func (p *Provider) createMetricReaders() ([]sdkmetric.Reader, error) {
	var readers []sdkmetric.Reader

	switch p.config.exporterType {
	case OTLPHTTP, Dual:
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(p.config.httpEndpoint),
			otlpmetrichttp.WithHeaders(p.config.headers),
			otlpmetrichttp.WithTimeout(p.config.timeout),
		}
		if p.config.insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		exporter, err := otlpmetrichttp.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP metric exporter: %w", err)
		}
		readers = append(readers, sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(p.config.batchTimeout),
		))
	}

	switch p.config.exporterType {
	case OTLPgRPC, Dual:
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(p.config.grpcEndpoint),
			otlpmetricgrpc.WithHeaders(p.config.headers),
			otlpmetricgrpc.WithTimeout(p.config.timeout),
		}
		if p.config.insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		exporter, err := otlpmetricgrpc.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC metric exporter: %w", err)
		}
		readers = append(readers, sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(p.config.batchTimeout),
		))
	}

	return readers, nil
}

// createLogProcessors creates log processors based on configuration.
func (p *Provider) createLogProcessors() ([]log.Processor, error) {
	var processors []log.Processor

	switch p.config.exporterType {
	case OTLPHTTP, Dual:
		opts := []otlploghttp.Option{
			otlploghttp.WithEndpoint(p.config.httpEndpoint),
			otlploghttp.WithHeaders(p.config.headers),
			otlploghttp.WithTimeout(p.config.timeout),
		}
		if p.config.insecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		exporter, err := otlploghttp.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP log exporter: %w", err)
		}
		processors = append(processors, log.NewBatchProcessor(exporter))
	}

	switch p.config.exporterType {
	case OTLPgRPC, Dual:
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(p.config.grpcEndpoint),
			otlploggrpc.WithHeaders(p.config.headers),
			otlploggrpc.WithTimeout(p.config.timeout),
		}
		if p.config.insecure {
			opts = append(opts, otlploggrpc.WithInsecure())
		}
		exporter, err := otlploggrpc.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC log exporter: %w", err)
		}
		processors = append(processors, log.NewBatchProcessor(exporter))
	}

	return processors, nil
}

// setGlobalProviders sets the global OpenTelemetry providers.
func (p *Provider) setGlobalProviders() {
	if p.traceProvider != nil {
		otel.SetTracerProvider(p.traceProvider)
	}

	if p.meterProvider != nil {
		otel.SetMeterProvider(p.meterProvider)
	}

	if p.logProvider != nil {
		global.SetLoggerProvider(p.logProvider)
	} else {
		global.SetLoggerProvider(noop.NewLoggerProvider())
	}
}

// setupTextMapPropagator configures the global text map propagator.
func setupTextMapPropagator() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}
