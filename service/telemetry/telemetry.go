// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Compile-time assertion that Telemetry implements service.Service.
var _ service.Service = (*Telemetry)(nil)

// Telemetry implements a telemetry collector and aggregator service.
// It collects metrics, traces, and logs from other u-bmc services and
// exports them to configured OTLP-compatible endpoints.
type Telemetry struct {
	config         config
	provider       *telemetry.Provider
	shutdownFunc   func(context.Context) error
	logger         *slog.Logger
	isRunning      bool
	collectionTick *time.Ticker
}

// New creates a new telemetry service with the given configuration options.
func New(opts ...Option) *Telemetry {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	return &Telemetry{
		config: cfg,
		logger: log.GetGlobalLogger().With("service", cfg.name),
	}
}

// Name returns the service name.
func (s *Telemetry) Name() string {
	return s.config.name
}

// Run starts the telemetry service and runs until the context is cancelled.
func (s *Telemetry) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	if s.isRunning {
		return fmt.Errorf("%w: %s", ErrServiceAlreadyRunning, s.config.name)
	}

	s.logger.InfoContext(ctx, "Starting telemetry service",
		"service", s.config.name,
		"exporter_type", s.config.exporterType,
		"http_endpoint", s.config.httpEndpoint,
		"grpc_endpoint", s.config.grpcEndpoint,
	)

	if err := s.initializeProvider(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrProviderInitializationFailed, err)
	}

	if err := s.setupCollector(ctx, ipcConn); err != nil {
		return fmt.Errorf("%w: %v", ErrCollectorSetupFailed, err)
	}

	s.isRunning = true
	defer func() {
		s.isRunning = false
	}()

	// Start collection ticker if aggregation is enabled
	if s.config.enableAggregation {
		s.collectionTick = time.NewTicker(s.config.collectionInterval)
		defer s.collectionTick.Stop()

		go s.runAggregation(ctx)
	}

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.InfoContext(ctx, "Stopping telemetry service",
		"service", s.config.name,
		"reason", ctx.Err(),
	)

	return s.shutdown(ctx)
}

// initializeProvider initializes the telemetry provider with configured options.
func (s *Telemetry) initializeProvider(ctx context.Context) error {
	opts := s.buildTelemetryOptions()

	provider, err := telemetry.NewProvider(opts...)
	if err != nil {
		return fmt.Errorf("failed to create telemetry provider: %w", err)
	}

	s.provider = provider

	// Set up shutdown function
	shutdownFunc, err := telemetry.Setup(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to setup telemetry: %w", err)
	}

	s.shutdownFunc = shutdownFunc
	return nil
}

// buildTelemetryOptions builds telemetry configuration options from service config.
func (s *Telemetry) buildTelemetryOptions() []telemetry.Option {
	var opts []telemetry.Option

	// Set exporter type and endpoints
	switch s.config.exporterType {
	case "noop":
		opts = append(opts, telemetry.WithExporterType(telemetry.NoOp))
	case "otlp-http":
		opts = append(opts, telemetry.WithOTLPHTTP(s.config.httpEndpoint))
	case "otlp-grpc":
		opts = append(opts, telemetry.WithOTLPgRPC(s.config.grpcEndpoint))
	case "dual":
		opts = append(opts, telemetry.WithDualOTLP(s.config.httpEndpoint, s.config.grpcEndpoint))
	default:
		opts = append(opts, telemetry.WithExporterType(telemetry.NoOp))
	}

	// Add service metadata
	opts = append(opts,
		telemetry.WithServiceName(s.config.serviceName),
		telemetry.WithServiceVersion(s.config.serviceVersion),
	)

	// Add headers if configured
	if len(s.config.headers) > 0 {
		opts = append(opts, telemetry.WithHeaders(s.config.headers))
	}

	// Add timeouts and batch settings
	opts = append(opts,
		telemetry.WithTimeout(s.config.timeout),
		telemetry.WithBatchTimeout(s.config.batchTimeout),
		telemetry.WithMaxExportBatch(s.config.maxExportBatch),
		telemetry.WithMaxQueueSize(s.config.maxQueueSize),
	)

	// Add feature flags
	opts = append(opts,
		telemetry.WithMetrics(s.config.enableMetrics),
		telemetry.WithTraces(s.config.enableTraces),
		telemetry.WithLogs(s.config.enableLogs),
	)

	// Add security settings
	opts = append(opts, telemetry.WithInsecure(s.config.insecure))

	// Add sampling configuration
	opts = append(opts, telemetry.WithSamplingRatio(s.config.samplingRatio))

	// Add resource attributes
	if len(s.config.resourceAttrs) > 0 {
		opts = append(opts, telemetry.WithResourceAttributes(s.config.resourceAttrs))
	}

	return opts
}

// setupCollector initializes the telemetry collector for gathering data from other services.
func (s *Telemetry) setupCollector(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	if !s.config.enableCollection {
		s.logger.InfoContext(ctx, "Telemetry collection disabled")
		return nil
	}

	// Connect to NATS for IPC communication
	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrIPCConnectionFailed, err)
	}

	// Set up telemetry collection subscriptions
	if err := s.setupMetricsCollection(ctx, nc); err != nil {
		return fmt.Errorf("%w: %v", ErrMetricsCollectionFailed, err)
	}

	if err := s.setupTracesCollection(ctx, nc); err != nil {
		return fmt.Errorf("%w: %v", ErrTracesCollectionFailed, err)
	}

	if err := s.setupLogsCollection(ctx, nc); err != nil {
		return fmt.Errorf("%w: %v", ErrLogsCollectionFailed, err)
	}

	s.logger.InfoContext(ctx, "Telemetry collector initialized successfully")
	return nil
}

// setupMetricsCollection sets up metrics collection from other services.
func (s *Telemetry) setupMetricsCollection(ctx context.Context, nc *nats.Conn) error {
	if !s.config.enableMetrics {
		return nil
	}

	subject := fmt.Sprintf("telemetry.metrics.%s", s.config.name)
	_, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		s.handleMetricsMessage(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to metrics: %w", err)
	}

	s.logger.DebugContext(ctx, "Metrics collection subscription established", "subject", subject)
	return nil
}

// setupTracesCollection sets up traces collection from other services.
func (s *Telemetry) setupTracesCollection(ctx context.Context, nc *nats.Conn) error {
	if !s.config.enableTraces {
		return nil
	}

	subject := fmt.Sprintf("telemetry.traces.%s", s.config.name)
	_, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		s.handleTracesMessage(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to traces: %w", err)
	}

	s.logger.DebugContext(ctx, "Traces collection subscription established", "subject", subject)
	return nil
}

// setupLogsCollection sets up logs collection from other services.
func (s *Telemetry) setupLogsCollection(ctx context.Context, nc *nats.Conn) error {
	if !s.config.enableLogs {
		return nil
	}

	subject := fmt.Sprintf("telemetry.logs.%s", s.config.name)
	_, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		s.handleLogsMessage(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	s.logger.DebugContext(ctx, "Logs collection subscription established", "subject", subject)
	return nil
}

// handleMetricsMessage processes incoming metrics messages.
func (s *Telemetry) handleMetricsMessage(ctx context.Context, msg *nats.Msg) {
	// Extract context from message headers if available
	msgCtx := ctx
	if msg.Header != nil && len(msg.Header) > 0 {
		msgCtx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(msg.Header))
	}

	s.logger.DebugContext(msgCtx, "Received metrics message",
		"subject", msg.Subject,
		"size", len(msg.Data),
	)

	// Apply filtering if configured
	if s.shouldFilterMessage(msgCtx, "metrics", msg) {
		return
	}

	// Process and forward metrics data
	// Implementation would depend on the specific metrics format
	// For now, we'll just log the receipt
	telemetry.AddSpanEvent(msgCtx, "metrics.received",
		telemetry.StringAttr("subject", msg.Subject),
		telemetry.IntAttr("data_size", len(msg.Data)),
	)
}

// handleTracesMessage processes incoming traces messages.
func (s *Telemetry) handleTracesMessage(ctx context.Context, msg *nats.Msg) {
	// Extract context from message headers if available
	msgCtx := ctx
	if msg.Header != nil && len(msg.Header) > 0 {
		msgCtx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(msg.Header))
	}

	s.logger.DebugContext(msgCtx, "Received traces message",
		"subject", msg.Subject,
		"size", len(msg.Data),
	)

	// Apply filtering if configured
	if s.shouldFilterMessage(msgCtx, "traces", msg) {
		return
	}

	// Process and forward traces data
	telemetry.AddSpanEvent(msgCtx, "traces.received",
		telemetry.StringAttr("subject", msg.Subject),
		telemetry.IntAttr("data_size", len(msg.Data)),
	)
}

// handleLogsMessage processes incoming logs messages.
func (s *Telemetry) handleLogsMessage(ctx context.Context, msg *nats.Msg) {
	// Extract context from message headers if available
	msgCtx := ctx
	if msg.Header != nil && len(msg.Header) > 0 {
		msgCtx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(msg.Header))
	}

	s.logger.DebugContext(msgCtx, "Received logs message",
		"subject", msg.Subject,
		"size", len(msg.Data),
	)

	// Apply filtering if configured
	if s.shouldFilterMessage(msgCtx, "logs", msg) {
		return
	}

	// Process and forward logs data
	telemetry.AddSpanEvent(msgCtx, "logs.received",
		telemetry.StringAttr("subject", msg.Subject),
		telemetry.IntAttr("data_size", len(msg.Data)),
	)
}

// shouldFilterMessage determines if a message should be filtered based on configuration.
func (s *Telemetry) shouldFilterMessage(ctx context.Context, dataType string, msg *nats.Msg) bool {
	// Apply any configured filters
	// For now, this is a placeholder implementation
	return false
}

// runAggregation runs the aggregation process in a separate goroutine.
func (s *Telemetry) runAggregation(ctx context.Context) {
	s.logger.InfoContext(ctx, "Starting telemetry aggregation",
		"interval", s.config.collectionInterval,
	)

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "Stopping telemetry aggregation")
			return
		case <-s.collectionTick.C:
			s.performAggregation(ctx)
		}
	}
}

// performAggregation performs periodic aggregation of collected telemetry data.
func (s *Telemetry) performAggregation(ctx context.Context) {
	tracer := telemetry.GetTracer("telemetry-service")
	spanCtx, span := tracer.Start(ctx, "aggregation.cycle")
	defer span.End()

	s.logger.DebugContext(spanCtx, "Performing telemetry aggregation")

	// Implement aggregation logic here
	// This would involve:
	// 1. Collecting buffered telemetry data
	// 2. Applying aggregation rules
	// 3. Exporting aggregated data

	telemetry.AddSpanEvent(spanCtx, "aggregation.completed")
}

// shutdown gracefully shuts down the telemetry service.
func (s *Telemetry) shutdown(ctx context.Context) error {
	if !s.isRunning {
		return fmt.Errorf("%w: %s", ErrServiceNotRunning, s.config.name)
	}

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.shutdownTimeout)
	defer cancel()

	var shutdownErr error
	if s.shutdownFunc != nil {
		shutdownErr = s.shutdownFunc(shutdownCtx)
	}

	if s.collectionTick != nil {
		s.collectionTick.Stop()
	}

	if shutdownErr != nil {
		return fmt.Errorf("%w: %v", ErrShutdownTimeout, shutdownErr)
	}

	s.logger.InfoContext(ctx, "Telemetry service shutdown completed")
	return ctx.Err()
}
