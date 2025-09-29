// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"time"

	"github.com/u-bmc/u-bmc/pkg/telemetry"
)

// Default configuration constants.
const (
	DefaultServiceName        = "telemetry"
	DefaultServiceDescription = "Telemetry collector and aggregator service for BMC observability"
	DefaultServiceVersion     = "1.0.0"
	DefaultCollectorName      = "u-bmc-telemetry"
	DefaultExporterType       = "noop"
	DefaultTimeout            = 30 * time.Second
	DefaultBatchTimeout       = 5 * time.Second
	DefaultMaxExportBatch     = 512
	DefaultMaxQueueSize       = 2048
	DefaultCollectionInterval = 30 * time.Second
	DefaultShutdownTimeout    = 10 * time.Second
	DefaultSamplingRatio      = 1.0
)

// Config holds the configuration for the telemetry service.
type Config struct {
	// ServiceName is the name of the service in the u-bmc framework
	ServiceName string
	// ServiceDescription provides a human-readable description of the service
	ServiceDescription string
	// ServiceVersion is the semantic version of the service
	ServiceVersion string
	// CollectorName is the name used for the telemetry collector instance
	CollectorName string
	// ExporterType determines the type of exporter to use
	ExporterType string
	// HTTPEndpoint is the OTLP HTTP endpoint for telemetry export
	HTTPEndpoint string
	// GRPCEndpoint is the OTLP gRPC endpoint for telemetry export
	GRPCEndpoint string
	// Headers contains additional headers for OTLP exporters
	Headers map[string]string
	// Timeout is the timeout for telemetry operations
	Timeout time.Duration
	// BatchTimeout is the timeout for batch exports
	BatchTimeout time.Duration
	// MaxExportBatch is the maximum number of items in an export batch
	MaxExportBatch int
	// MaxQueueSize is the maximum queue size for pending exports
	MaxQueueSize int
	// EnableMetrics controls metrics collection
	EnableMetrics bool
	// EnableTraces controls trace collection
	EnableTraces bool
	// EnableLogs controls log collection
	EnableLogs bool
	// EnableCollection controls telemetry collection from other services
	EnableCollection bool
	// EnableAggregation controls telemetry aggregation
	EnableAggregation bool
	// CollectionInterval is the interval for telemetry collection
	CollectionInterval time.Duration
	// ShutdownTimeout is the maximum time to wait for graceful shutdown
	ShutdownTimeout time.Duration
	// Insecure enables insecure connections to OTLP endpoints
	Insecure bool
	// SamplingRatio is the sampling ratio for traces (0.0 to 1.0)
	SamplingRatio float64
	// ResourceAttrs contains additional resource attributes
	ResourceAttrs map[string]string
	// FilterRules contains runtime-configurable filtering rules
	FilterRules *FilterConfig
	// AggregationRules contains runtime-configurable aggregation rules
	AggregationRules *AggregationConfig
}

// FilterConfig contains runtime-configurable filtering rules for telemetry data.
type FilterConfig struct {
	// EnableFiltering controls whether filtering is active
	EnableFiltering bool `json:"enable_filtering"`
	// MetricsFilters contains filtering rules for metrics
	MetricsFilters []FilterRule `json:"metrics_filters"`
	// TracesFilters contains filtering rules for traces
	TracesFilters []FilterRule `json:"traces_filters"`
	// LogsFilters contains filtering rules for logs
	LogsFilters []FilterRule `json:"logs_filters"`
	// SamplingOverrides allows runtime sampling ratio adjustments per service
	SamplingOverrides map[string]float64 `json:"sampling_overrides"`
	// DebugServices lists services that should have debug-level telemetry
	DebugServices []string `json:"debug_services"`
}

// FilterRule defines a rule for filtering telemetry data.
type FilterRule struct {
	// Name is a human-readable name for the filter rule
	Name string `json:"name"`
	// Type is the type of filter (drop, sample, transform)
	Type string `json:"type"`
	// Condition is the condition to match for applying the filter
	Condition FilterCondition `json:"condition"`
	// Action is the action to take when the condition matches
	Action FilterAction `json:"action"`
	// Enabled controls whether the rule is active
	Enabled bool `json:"enabled"`
}

// FilterCondition defines conditions for matching telemetry data.
type FilterCondition struct {
	// ServiceName matches the service name
	ServiceName string `json:"service_name,omitempty"`
	// AttributeMatches contains attribute key-value pairs to match
	AttributeMatches map[string]string `json:"attribute_matches,omitempty"`
	// MetricName matches the metric name
	MetricName string `json:"metric_name,omitempty"`
	// SpanName matches the span name
	SpanName string `json:"span_name,omitempty"`
	// LogLevel matches the log level
	LogLevel string `json:"log_level,omitempty"`
	// ResourceMatches contains resource attribute key-value pairs to match
	ResourceMatches map[string]string `json:"resource_matches,omitempty"`
}

// FilterAction defines the action to take when a filter condition matches.
type FilterAction struct {
	// Type is the action type (drop, sample, modify)
	Type string `json:"type"`
	// SampleRate is the sampling rate for sample actions (0.0 to 1.0)
	SampleRate float64 `json:"sample_rate,omitempty"`
	// Modifications contains attribute modifications for modify actions
	Modifications map[string]string `json:"modifications,omitempty"`
}

// AggregationConfig contains runtime-configurable aggregation rules.
type AggregationConfig struct {
	// EnableAggregation controls whether aggregation is active
	EnableAggregation bool `json:"enable_aggregation"`
	// MetricsAggregation contains aggregation rules for metrics
	MetricsAggregation []AggregationRule `json:"metrics_aggregation"`
	// TracesAggregation contains aggregation rules for traces
	TracesAggregation []AggregationRule `json:"traces_aggregation"`
	// WindowSize is the time window for aggregation
	WindowSize time.Duration `json:"window_size"`
	// MaxCardinality limits the number of unique metric series
	MaxCardinality int `json:"max_cardinality"`
}

// AggregationRule defines a rule for aggregating telemetry data.
type AggregationRule struct {
	// Name is a human-readable name for the aggregation rule
	Name string `json:"name"`
	// Type is the type of aggregation (sum, avg, count, histogram)
	Type string `json:"type"`
	// GroupBy contains attribute names to group by
	GroupBy []string `json:"group_by"`
	// Condition is the condition to match for applying the aggregation
	Condition FilterCondition `json:"condition"`
	// Enabled controls whether the rule is active
	Enabled bool `json:"enabled"`
}

// RuntimeConfig represents a runtime configuration update message.
type RuntimeConfig struct {
	// Type identifies the type of configuration update
	Type string `json:"type"`
	// ServiceName is the target service name (empty for all services)
	ServiceName string `json:"service_name,omitempty"`
	// FilterConfig contains updated filtering configuration
	FilterConfig *FilterConfig `json:"filter_config,omitempty"`
	// AggregationConfig contains updated aggregation configuration
	AggregationConfig *AggregationConfig `json:"aggregation_config,omitempty"`
	// SamplingRatio updates the global sampling ratio
	SamplingRatio *float64 `json:"sampling_ratio,omitempty"`
	// DebugMode enables/disables debug mode
	DebugMode *bool `json:"debug_mode,omitempty"`
	// ExporterConfig contains updated exporter configuration
	ExporterConfig *ExporterConfig `json:"exporter_config,omitempty"`
}

// ExporterConfig contains runtime-configurable exporter settings.
type ExporterConfig struct {
	// ExporterType is the type of exporter to use
	ExporterType string `json:"exporter_type"`
	// HTTPEndpoint is the OTLP HTTP endpoint
	HTTPEndpoint string `json:"http_endpoint,omitempty"`
	// GRPCEndpoint is the OTLP gRPC endpoint
	GRPCEndpoint string `json:"grpc_endpoint,omitempty"`
	// Headers contains additional headers
	Headers map[string]string `json:"headers,omitempty"`
	// Timeout is the operation timeout
	Timeout time.Duration `json:"timeout,omitempty"`
	// Insecure enables insecure connections
	Insecure bool `json:"insecure"`
}

// Option represents a configuration option for the telemetry service.
type Option interface {
	apply(*Config)
}

type serviceNameOption struct {
	name string
}

func (o *serviceNameOption) apply(c *Config) {
	c.ServiceName = o.name
}

// WithServiceName sets the name of the service.
func WithServiceName(name string) Option {
	return &serviceNameOption{name: name}
}

type serviceDescriptionOption struct {
	description string
}

func (o *serviceDescriptionOption) apply(c *Config) {
	c.ServiceDescription = o.description
}

// WithServiceDescription sets the description of the service.
func WithServiceDescription(description string) Option {
	return &serviceDescriptionOption{description: description}
}

type serviceVersionOption struct {
	version string
}

func (o *serviceVersionOption) apply(c *Config) {
	c.ServiceVersion = o.version
}

// WithServiceVersion sets the version of the service.
func WithServiceVersion(version string) Option {
	return &serviceVersionOption{version: version}
}

type collectorNameOption struct {
	name string
}

func (o *collectorNameOption) apply(c *Config) {
	c.CollectorName = o.name
}

// WithCollectorName sets the telemetry collector name.
func WithCollectorName(name string) Option {
	return &collectorNameOption{name: name}
}

type exporterTypeOption struct {
	exporterType string
}

func (o *exporterTypeOption) apply(c *Config) {
	c.ExporterType = o.exporterType
}

// WithExporterType sets the exporter type (noop, otlp-http, otlp-grpc, dual).
func WithExporterType(exporterType string) Option {
	return &exporterTypeOption{exporterType: exporterType}
}

type httpEndpointOption struct {
	endpoint string
}

func (o *httpEndpointOption) apply(c *Config) {
	c.HTTPEndpoint = o.endpoint
}

// WithHTTPEndpoint sets the HTTP endpoint for OTLP export.
func WithHTTPEndpoint(endpoint string) Option {
	return &httpEndpointOption{endpoint: endpoint}
}

type grpcEndpointOption struct {
	endpoint string
}

func (o *grpcEndpointOption) apply(c *Config) {
	c.GRPCEndpoint = o.endpoint
}

// WithGRPCEndpoint sets the gRPC endpoint for OTLP export.
func WithGRPCEndpoint(endpoint string) Option {
	return &grpcEndpointOption{endpoint: endpoint}
}

type headersOption struct {
	headers map[string]string
}

func (o *headersOption) apply(c *Config) {
	c.Headers = o.headers
}

// WithHeaders sets additional headers for OTLP exporters.
func WithHeaders(headers map[string]string) Option {
	return &headersOption{headers: headers}
}

type timeoutOption struct {
	timeout time.Duration
}

func (o *timeoutOption) apply(c *Config) {
	c.Timeout = o.timeout
}

// WithTimeout sets the timeout for telemetry operations.
func WithTimeout(timeout time.Duration) Option {
	return &timeoutOption{timeout: timeout}
}

type batchTimeoutOption struct {
	timeout time.Duration
}

func (o *batchTimeoutOption) apply(c *Config) {
	c.BatchTimeout = o.timeout
}

// WithBatchTimeout sets the timeout for batch exports.
func WithBatchTimeout(timeout time.Duration) Option {
	return &batchTimeoutOption{timeout: timeout}
}

type maxExportBatchOption struct {
	size int
}

func (o *maxExportBatchOption) apply(c *Config) {
	c.MaxExportBatch = o.size
}

// WithMaxExportBatch sets the maximum number of items in an export batch.
func WithMaxExportBatch(size int) Option {
	return &maxExportBatchOption{size: size}
}

type maxQueueSizeOption struct {
	size int
}

func (o *maxQueueSizeOption) apply(c *Config) {
	c.MaxQueueSize = o.size
}

// WithMaxQueueSize sets the maximum queue size for pending exports.
func WithMaxQueueSize(size int) Option {
	return &maxQueueSizeOption{size: size}
}

type metricsOption struct {
	enabled bool
}

func (o *metricsOption) apply(c *Config) {
	c.EnableMetrics = o.enabled
}

// WithMetrics enables or disables metrics collection.
func WithMetrics(enabled bool) Option {
	return &metricsOption{enabled: enabled}
}

type tracesOption struct {
	enabled bool
}

func (o *tracesOption) apply(c *Config) {
	c.EnableTraces = o.enabled
}

// WithTraces enables or disables trace collection.
func WithTraces(enabled bool) Option {
	return &tracesOption{enabled: enabled}
}

type logsOption struct {
	enabled bool
}

func (o *logsOption) apply(c *Config) {
	c.EnableLogs = o.enabled
}

// WithLogs enables or disables log collection.
func WithLogs(enabled bool) Option {
	return &logsOption{enabled: enabled}
}

type collectionOption struct {
	enabled bool
}

func (o *collectionOption) apply(c *Config) {
	c.EnableCollection = o.enabled
}

// WithCollection enables or disables telemetry collection from other services.
func WithCollection(enabled bool) Option {
	return &collectionOption{enabled: enabled}
}

type aggregationOption struct {
	enabled bool
}

func (o *aggregationOption) apply(c *Config) {
	c.EnableAggregation = o.enabled
}

// WithAggregation enables or disables telemetry aggregation.
func WithAggregation(enabled bool) Option {
	return &aggregationOption{enabled: enabled}
}

type collectionIntervalOption struct {
	interval time.Duration
}

func (o *collectionIntervalOption) apply(c *Config) {
	c.CollectionInterval = o.interval
}

// WithCollectionInterval sets the interval for telemetry collection.
func WithCollectionInterval(interval time.Duration) Option {
	return &collectionIntervalOption{interval: interval}
}

type shutdownTimeoutOption struct {
	timeout time.Duration
}

func (o *shutdownTimeoutOption) apply(c *Config) {
	c.ShutdownTimeout = o.timeout
}

// WithShutdownTimeout sets the timeout for service shutdown.
func WithShutdownTimeout(timeout time.Duration) Option {
	return &shutdownTimeoutOption{timeout: timeout}
}

type insecureOption struct {
	insecure bool
}

func (o *insecureOption) apply(c *Config) {
	c.Insecure = o.insecure
}

// WithInsecure enables or disables insecure connections to OTLP endpoints.
func WithInsecure(insecure bool) Option {
	return &insecureOption{insecure: insecure}
}

type samplingRatioOption struct {
	ratio float64
}

func (o *samplingRatioOption) apply(c *Config) {
	if o.ratio < 0.0 {
		c.SamplingRatio = 0.0
	} else if o.ratio > 1.0 {
		c.SamplingRatio = 1.0
	} else {
		c.SamplingRatio = o.ratio
	}
}

// WithSamplingRatio sets the sampling ratio for traces (0.0 to 1.0).
func WithSamplingRatio(ratio float64) Option {
	return &samplingRatioOption{ratio: ratio}
}

type resourceAttributesOption struct {
	attrs map[string]string
}

func (o *resourceAttributesOption) apply(c *Config) {
	c.ResourceAttrs = o.attrs
}

// WithResourceAttributes sets additional resource attributes for telemetry data.
func WithResourceAttributes(attrs map[string]string) Option {
	return &resourceAttributesOption{attrs: attrs}
}

type filterConfigOption struct {
	config *FilterConfig
}

func (o *filterConfigOption) apply(c *Config) {
	c.FilterRules = o.config
}

// WithFilterConfig sets the filtering configuration.
func WithFilterConfig(config *FilterConfig) Option {
	return &filterConfigOption{config: config}
}

type aggregationConfigOption struct {
	config *AggregationConfig
}

func (o *aggregationConfigOption) apply(c *Config) {
	c.AggregationRules = o.config
}

// WithAggregationConfig sets the aggregation configuration.
func WithAggregationConfig(config *AggregationConfig) Option {
	return &aggregationConfigOption{config: config}
}

// WithOTLPHTTP is a convenience function that configures OTLP HTTP export.
func WithOTLPHTTP(endpoint string) Option {
	return &exporterTypeOption{exporterType: "otlp-http"}
}

// WithOTLPGRPC is a convenience function that configures OTLP gRPC export.
func WithOTLPGRPC(endpoint string) Option {
	return &exporterTypeOption{exporterType: "otlp-grpc"}
}

// WithDualOTLP is a convenience function that configures dual OTLP export.
func WithDualOTLP(httpEndpoint, grpcEndpoint string) Option {
	return &exporterTypeOption{exporterType: "dual"}
}

// WithNoOp is a convenience function that configures no-op operation.
func WithNoOp() Option {
	return &exporterTypeOption{exporterType: "noop"}
}

// WithName is a backward compatibility alias for WithServiceName.
// Deprecated: Use WithServiceName instead.
func WithName(name string) Option {
	return WithServiceName(name)
}

// ToTelemetryOptions converts the service configuration to telemetry package options.
// The telemetry service defaults to NoOp to minimize overhead and only exports when configured.
func (c *Config) ToTelemetryOptions() []telemetry.Option {
	var opts []telemetry.Option

	// Set exporter type and endpoints - NoOp is the default for the telemetry service
	switch c.ExporterType {
	case "noop":
		opts = append(opts, telemetry.WithExporterType(telemetry.NoOp))
	case "otlp-http":
		if c.HTTPEndpoint != "" {
			opts = append(opts, telemetry.WithOTLPHTTP(c.HTTPEndpoint))
		} else {
			opts = append(opts, telemetry.WithExporterType(telemetry.NoOp))
		}
	case "otlp-grpc":
		if c.GRPCEndpoint != "" {
			opts = append(opts, telemetry.WithOTLPgRPC(c.GRPCEndpoint))
		} else {
			opts = append(opts, telemetry.WithExporterType(telemetry.NoOp))
		}
	case "dual":
		if c.HTTPEndpoint != "" && c.GRPCEndpoint != "" {
			opts = append(opts, telemetry.WithDualOTLP(c.HTTPEndpoint, c.GRPCEndpoint))
		} else {
			opts = append(opts, telemetry.WithExporterType(telemetry.NoOp))
		}
	default:
		// Default to NoOp for the telemetry service to minimize overhead
		opts = append(opts, telemetry.WithExporterType(telemetry.NoOp))
	}

	// Add service metadata
	opts = append(opts,
		telemetry.WithServiceName(c.CollectorName),
		telemetry.WithServiceVersion(c.ServiceVersion),
	)

	// Add headers if configured
	if len(c.Headers) > 0 {
		opts = append(opts, telemetry.WithHeaders(c.Headers))
	}

	// Add timeouts and batch settings
	opts = append(opts,
		telemetry.WithTimeout(c.Timeout),
		telemetry.WithBatchTimeout(c.BatchTimeout),
		telemetry.WithMaxExportBatch(c.MaxExportBatch),
		telemetry.WithMaxQueueSize(c.MaxQueueSize),
	)

	// Add feature flags
	opts = append(opts,
		telemetry.WithMetrics(c.EnableMetrics),
		telemetry.WithTraces(c.EnableTraces),
		telemetry.WithLogs(c.EnableLogs),
	)

	// Add security settings
	opts = append(opts, telemetry.WithInsecure(c.Insecure))

	// Add sampling configuration
	opts = append(opts, telemetry.WithSamplingRatio(c.SamplingRatio))

	// Add resource attributes
	if len(c.ResourceAttrs) > 0 {
		opts = append(opts, telemetry.WithResourceAttributes(c.ResourceAttrs))
	}

	return opts
}
