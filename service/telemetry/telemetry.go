// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Compile-time assertion that Telemetry implements service.Service.
var _ service.Service = (*Telemetry)(nil)

// Telemetry implements a telemetry collector and aggregator service for the u-bmc system.
// It collects metrics, traces, and logs from other u-bmc services, applies filtering and
// aggregation rules, and exports them to configured OTLP-compatible endpoints.
//
// The telemetry service acts as the central observability hub for the BMC, ensuring that
// all services generate telemetry data in OTLP format while providing runtime configuration
// capabilities for filtering, aggregation, and debug control.
//
// Key features:
//   - Mandatory telemetry collection from all u-bmc services
//   - Runtime reconfiguration via NATS messages
//   - Filtering and aggregation to reduce overhead
//   - Support for debug mode toggle during runtime
//   - OTLP export to multiple endpoints
type Telemetry struct {
	config         *Config
	provider       *telemetry.Provider
	shutdownFunc   func(context.Context) error
	logger         *slog.Logger
	tracer         trace.Tracer
	isRunning      bool
	collectionTick *time.Ticker
	configMutex    sync.RWMutex
	natsConn       *nats.Conn
}

// New creates a new telemetry service instance with the provided configuration options.
//
// The telemetry service is configured with sensible defaults but can be customized
// using the provided Option functions. The service enforces that all u-bmc services
// participate in telemetry collection and provides runtime reconfiguration capabilities.
//
// Example usage:
//
//	telemetryService := telemetry.New(
//		telemetry.WithServiceName("bmc-telemetry"),
//		telemetry.WithOTLPHTTP("http://localhost:4318"),
//		telemetry.WithCollection(true),
//	)
func New(opts ...Option) *Telemetry {
	cfg := &Config{
		ServiceName:        DefaultServiceName,
		ServiceDescription: DefaultServiceDescription,
		ServiceVersion:     DefaultServiceVersion,
		CollectorName:      DefaultCollectorName,
		ExporterType:       DefaultExporterType,
		Timeout:            DefaultTimeout,
		BatchTimeout:       DefaultBatchTimeout,
		MaxExportBatch:     DefaultMaxExportBatch,
		MaxQueueSize:       DefaultMaxQueueSize,
		EnableMetrics:      true,
		EnableTraces:       true,
		EnableLogs:         true,
		EnableCollection:   true,
		EnableAggregation:  true,
		CollectionInterval: DefaultCollectionInterval,
		ShutdownTimeout:    DefaultShutdownTimeout,
		Insecure:           false,
		SamplingRatio:      DefaultSamplingRatio,
		Headers:            make(map[string]string),
		ResourceAttrs:      make(map[string]string),
		FilterRules: &FilterConfig{
			EnableFiltering:   false,
			MetricsFilters:    make([]FilterRule, 0),
			TracesFilters:     make([]FilterRule, 0),
			LogsFilters:       make([]FilterRule, 0),
			SamplingOverrides: make(map[string]float64),
			DebugServices:     make([]string, 0),
		},
		AggregationRules: &AggregationConfig{
			EnableAggregation:  true,
			MetricsAggregation: make([]AggregationRule, 0),
			TracesAggregation:  make([]AggregationRule, 0),
			WindowSize:         30 * time.Second,
			MaxCardinality:     10000,
		},
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &Telemetry{
		config: cfg,
	}
}

// Name returns the service name as configured.
// This implements the service.Service interface.
func (s *Telemetry) Name() string {
	return s.config.ServiceName
}

// Run starts the telemetry service and runs until the context is cancelled.
//
// This method implements the service.Service interface and handles the complete
// lifecycle of the telemetry service:
//
//  1. Initializes the telemetry provider with configured options
//  2. Sets up telemetry collection from other services via NATS
//  3. Starts runtime configuration listener
//  4. Runs aggregation if enabled
//  5. Handles graceful shutdown
//
// The service will enforce telemetry collection from all u-bmc services and
// provide runtime reconfiguration capabilities through NATS messaging.
func (s *Telemetry) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	if s.isRunning {
		return fmt.Errorf("%w: %s", ErrServiceAlreadyRunning, s.config.ServiceName)
	}

	s.tracer = otel.Tracer("u-bmc/service/telemetry")

	ctx, span := s.tracer.Start(ctx, "telemetry.Run")
	defer span.End()

	s.logger = log.GetGlobalLogger().With("service", s.config.ServiceName)
	s.logger.InfoContext(ctx, "Starting telemetry service",
		"version", s.config.ServiceVersion,
		"collector_name", s.config.CollectorName,
		"exporter_type", s.config.ExporterType,
		"http_endpoint", s.config.HTTPEndpoint,
		"grpc_endpoint", s.config.GRPCEndpoint,
		"collection_enabled", s.config.EnableCollection,
		"aggregation_enabled", s.config.EnableAggregation,
	)

	if err := s.initializeProvider(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %v", ErrProviderInitializationFailed, err)
	}

	if err := s.setupCollector(ctx, ipcConn); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %v", ErrCollectorSetupFailed, err)
	}

	if err := s.setupRuntimeConfiguration(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %v", ErrCollectorSetupFailed, err)
	}

	s.isRunning = true
	defer func() {
		s.isRunning = false
	}()

	// Start aggregation ticker if enabled
	if s.config.EnableAggregation {
		s.collectionTick = time.NewTicker(s.config.CollectionInterval)
		defer s.collectionTick.Stop()

		go s.runAggregation(ctx)
	}

	s.logger.InfoContext(ctx, "Telemetry service started successfully",
		"service_name", s.config.ServiceName,
		"collector_name", s.config.CollectorName,
	)

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.InfoContext(ctx, "Stopping telemetry service",
		"service", s.config.ServiceName,
		"reason", ctx.Err(),
	)

	return s.shutdown(ctx)
}

// initializeProvider initializes the telemetry provider with configured options.
func (s *Telemetry) initializeProvider(ctx context.Context) error {
	opts := s.config.ToTelemetryOptions()

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

// setupCollector initializes the telemetry collector for gathering data from other services.
func (s *Telemetry) setupCollector(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	if !s.config.EnableCollection {
		s.logger.InfoContext(ctx, "Telemetry collection disabled")
		return nil
	}

	// Connect to NATS for IPC communication
	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrIPCConnectionFailed, err)
	}
	s.natsConn = nc

	// Set up telemetry collection subscriptions for all services
	if err := s.setupMetricsCollection(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrMetricsCollectionFailed, err)
	}

	if err := s.setupTracesCollection(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrTracesCollectionFailed, err)
	}

	if err := s.setupLogsCollection(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrLogsCollectionFailed, err)
	}

	s.logger.InfoContext(ctx, "Telemetry collector initialized successfully")
	return nil
}

// setupRuntimeConfiguration sets up the runtime configuration listener.
func (s *Telemetry) setupRuntimeConfiguration(ctx context.Context) error {
	if s.natsConn == nil {
		return fmt.Errorf("NATS connection not available for runtime configuration")
	}

	subject := fmt.Sprintf("telemetry.config.%s", s.config.ServiceName)
	_, err := s.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleRuntimeConfigUpdate(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to runtime configuration: %w", err)
	}

	// Also subscribe to global configuration updates
	globalSubject := "telemetry.config.global"
	_, err = s.natsConn.Subscribe(globalSubject, func(msg *nats.Msg) {
		s.handleRuntimeConfigUpdate(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to global configuration: %w", err)
	}

	s.logger.InfoContext(ctx, "Runtime configuration listener established",
		"subject", subject,
		"global_subject", globalSubject,
	)
	return nil
}

// setupMetricsCollection sets up metrics collection from all services.
func (s *Telemetry) setupMetricsCollection(ctx context.Context) error {
	if !s.config.EnableMetrics {
		return nil
	}

	// Subscribe to metrics from all services using wildcard
	subject := "telemetry.metrics.*"
	_, err := s.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleMetricsMessage(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to metrics: %w", err)
	}

	s.logger.DebugContext(ctx, "Metrics collection subscription established", "subject", subject)
	return nil
}

// setupTracesCollection sets up traces collection from all services.
func (s *Telemetry) setupTracesCollection(ctx context.Context) error {
	if !s.config.EnableTraces {
		return nil
	}

	// Subscribe to traces from all services using wildcard
	subject := "telemetry.traces.*"
	_, err := s.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleTracesMessage(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to traces: %w", err)
	}

	s.logger.DebugContext(ctx, "Traces collection subscription established", "subject", subject)
	return nil
}

// setupLogsCollection sets up logs collection from all services.
func (s *Telemetry) setupLogsCollection(ctx context.Context) error {
	if !s.config.EnableLogs {
		return nil
	}

	// Subscribe to logs from all services using wildcard
	subject := "telemetry.logs.*"
	_, err := s.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleLogsMessage(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	s.logger.DebugContext(ctx, "Logs collection subscription established", "subject", subject)
	return nil
}

// handleRuntimeConfigUpdate processes runtime configuration update messages.
func (s *Telemetry) handleRuntimeConfigUpdate(ctx context.Context, msg *nats.Msg) {
	// Extract context from message headers if available
	msgCtx := ctx
	if msg.Header != nil && len(msg.Header) > 0 {
		msgCtx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(msg.Header))
	}

	s.logger.InfoContext(msgCtx, "Received runtime configuration update",
		"subject", msg.Subject,
		"size", len(msg.Data),
	)

	var runtimeConfig RuntimeConfig
	if err := json.Unmarshal(msg.Data, &runtimeConfig); err != nil {
		s.logger.ErrorContext(msgCtx, "Failed to unmarshal runtime configuration", err,
			"subject", msg.Subject,
		)
		return
	}

	s.applyRuntimeConfig(msgCtx, &runtimeConfig)
}

// applyRuntimeConfig applies a runtime configuration update.
func (s *Telemetry) applyRuntimeConfig(ctx context.Context, config *RuntimeConfig) {
	s.configMutex.Lock()
	defer s.configMutex.Unlock()

	s.logger.InfoContext(ctx, "Applying runtime configuration update",
		"type", config.Type,
		"service_name", config.ServiceName,
	)

	// Apply configuration changes based on type
	switch config.Type {
	case "filter_config":
		if config.FilterConfig != nil {
			s.config.FilterRules = config.FilterConfig
			s.logger.InfoContext(ctx, "Updated filter configuration")
		}
	case "aggregation_config":
		if config.AggregationConfig != nil {
			s.config.AggregationRules = config.AggregationConfig
			s.logger.InfoContext(ctx, "Updated aggregation configuration")
		}
	case "sampling_ratio":
		if config.SamplingRatio != nil {
			s.config.SamplingRatio = *config.SamplingRatio
			s.logger.InfoContext(ctx, "Updated sampling ratio",
				"new_ratio", *config.SamplingRatio,
			)
		}
	case "debug_mode":
		if config.DebugMode != nil {
			s.applyDebugMode(ctx, *config.DebugMode, config.ServiceName)
		}
	case "exporter_config":
		if config.ExporterConfig != nil {
			s.applyExporterConfig(ctx, config.ExporterConfig)
		}
	default:
		s.logger.WarnContext(ctx, "Unknown configuration type",
			"type", config.Type,
		)
	}

	telemetry.AddSpanEvent(ctx, "runtime_config_applied",
		telemetry.StringAttr("config_type", config.Type),
		telemetry.StringAttr("service_name", config.ServiceName),
	)
}

// applyDebugMode applies debug mode configuration.
func (s *Telemetry) applyDebugMode(ctx context.Context, debug bool, serviceName string) {
	if serviceName == "" {
		// Apply to all services
		s.config.FilterRules.DebugServices = []string{"*"}
		s.logger.InfoContext(ctx, "Enabled debug mode for all services")
	} else {
		// Apply to specific service
		if debug {
			// Add service to debug list if not already present
			found := false
			for _, svc := range s.config.FilterRules.DebugServices {
				if svc == serviceName {
					found = true
					break
				}
			}
			if !found {
				s.config.FilterRules.DebugServices = append(s.config.FilterRules.DebugServices, serviceName)
			}
		} else {
			// Remove service from debug list
			for i, svc := range s.config.FilterRules.DebugServices {
				if svc == serviceName {
					s.config.FilterRules.DebugServices = append(
						s.config.FilterRules.DebugServices[:i],
						s.config.FilterRules.DebugServices[i+1:]...,
					)
					break
				}
			}
		}
		s.logger.InfoContext(ctx, "Updated debug mode for service",
			"service_name", serviceName,
			"debug", debug,
		)
	}
}

// applyExporterConfig applies exporter configuration changes.
func (s *Telemetry) applyExporterConfig(ctx context.Context, exporterConfig *ExporterConfig) {
	// Update configuration
	s.config.ExporterType = exporterConfig.ExporterType
	if exporterConfig.HTTPEndpoint != "" {
		s.config.HTTPEndpoint = exporterConfig.HTTPEndpoint
	}
	if exporterConfig.GRPCEndpoint != "" {
		s.config.GRPCEndpoint = exporterConfig.GRPCEndpoint
	}
	if exporterConfig.Headers != nil {
		s.config.Headers = exporterConfig.Headers
	}
	if exporterConfig.Timeout > 0 {
		s.config.Timeout = exporterConfig.Timeout
	}
	s.config.Insecure = exporterConfig.Insecure

	// Note: Reinitializing the provider would require stopping and restarting
	// For now, log the configuration change - full reinitialization would need
	// more complex orchestration
	s.logger.InfoContext(ctx, "Exporter configuration updated",
		"exporter_type", exporterConfig.ExporterType,
		"http_endpoint", exporterConfig.HTTPEndpoint,
		"grpc_endpoint", exporterConfig.GRPCEndpoint,
	)
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
	telemetry.AddSpanEvent(msgCtx, "metrics.received",
		telemetry.StringAttr("subject", msg.Subject),
		telemetry.IntAttr("data_size", len(msg.Data)),
	)

	// TODO: Implement actual metrics processing and forwarding
	// This would involve parsing the OTLP metrics data and applying
	// aggregation rules before forwarding to the configured endpoints
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

	// TODO: Implement actual traces processing and forwarding
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

	// TODO: Implement actual logs processing and forwarding
}

// shouldFilterMessage determines if a message should be filtered based on configuration.
func (s *Telemetry) shouldFilterMessage(ctx context.Context, dataType string, msg *nats.Msg) bool {
	s.configMutex.RLock()
	defer s.configMutex.RUnlock()

	if !s.config.FilterRules.EnableFiltering {
		return false
	}

	// Apply filtering rules based on data type
	switch dataType {
	case "metrics":
		// TODO: Implement metrics filtering logic based on s.config.FilterRules.MetricsFilters
	case "traces":
		// TODO: Implement traces filtering logic based on s.config.FilterRules.TracesFilters
	case "logs":
		// TODO: Implement logs filtering logic based on s.config.FilterRules.LogsFilters
	default:
		return false
	}

	// TODO: Implement actual filtering logic based on filter rules
	// This would involve parsing message content and applying filter conditions

	return false
}

// runAggregation runs the aggregation process in a separate goroutine.
func (s *Telemetry) runAggregation(ctx context.Context) {
	s.logger.InfoContext(ctx, "Starting telemetry aggregation",
		"interval", s.config.CollectionInterval,
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

	s.configMutex.RLock()
	enableAggregation := s.config.AggregationRules.EnableAggregation
	s.configMutex.RUnlock()

	if !enableAggregation {
		return
	}

	// TODO: Implement aggregation logic here
	// This would involve:
	// 1. Collecting buffered telemetry data
	// 2. Applying aggregation rules
	// 3. Exporting aggregated data

	telemetry.AddSpanEvent(spanCtx, "aggregation.completed")
}

// shutdown gracefully shuts down the telemetry service.
func (s *Telemetry) shutdown(ctx context.Context) error {
	if !s.isRunning {
		return fmt.Errorf("%w: %s", ErrServiceNotRunning, s.config.ServiceName)
	}

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	var shutdownErr error
	if s.shutdownFunc != nil {
		shutdownErr = s.shutdownFunc(shutdownCtx)
	}

	if s.collectionTick != nil {
		s.collectionTick.Stop()
	}

	if s.natsConn != nil {
		s.natsConn.Close()
	}

	if shutdownErr != nil {
		return fmt.Errorf("%w: %v", ErrShutdownTimeout, shutdownErr)
	}

	s.logger.InfoContext(ctx, "Telemetry service shutdown completed")
	return ctx.Err()
}
