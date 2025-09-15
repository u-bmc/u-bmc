// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"time"
)

// config holds the configuration for the telemetry service.
type config struct {
	name               string
	serviceName        string
	serviceVersion     string
	exporterType       string
	httpEndpoint       string
	grpcEndpoint       string
	headers            map[string]string
	timeout            time.Duration
	batchTimeout       time.Duration
	maxExportBatch     int
	maxQueueSize       int
	enableMetrics      bool
	enableTraces       bool
	enableLogs         bool
	enableCollection   bool
	enableAggregation  bool
	collectionInterval time.Duration
	shutdownTimeout    time.Duration
	insecure           bool
	samplingRatio      float64
	resourceAttrs      map[string]string
}

// defaultConfig returns a default configuration for the telemetry service.
func defaultConfig() config {
	return config{
		name:               "telemetry",
		serviceName:        "u-bmc-telemetry",
		serviceVersion:     "1.0.0",
		exporterType:       "noop",
		timeout:            30 * time.Second,
		batchTimeout:       5 * time.Second,
		maxExportBatch:     512,
		maxQueueSize:       2048,
		enableMetrics:      true,
		enableTraces:       true,
		enableLogs:         true,
		enableCollection:   true,
		enableAggregation:  true,
		collectionInterval: 30 * time.Second,
		shutdownTimeout:    10 * time.Second,
		insecure:           false,
		samplingRatio:      1.0,
		headers:            make(map[string]string),
		resourceAttrs:      make(map[string]string),
	}
}

// Option defines a function that modifies the telemetry service configuration.
type Option interface {
	apply(*config)
}

// nameOption configures the service name.
type nameOption struct {
	name string
}

func (o *nameOption) apply(c *config) {
	c.name = o.name
}

// WithName sets the service name.
func WithName(name string) Option {
	return &nameOption{
		name: name,
	}
}

// serviceNameOption configures the service name for telemetry.
type serviceNameOption struct {
	serviceName string
}

func (o *serviceNameOption) apply(c *config) {
	c.serviceName = o.serviceName
}

// WithServiceName sets the service name for telemetry data.
func WithServiceName(serviceName string) Option {
	return &serviceNameOption{
		serviceName: serviceName,
	}
}

// serviceVersionOption configures the service version.
type serviceVersionOption struct {
	serviceVersion string
}

func (o *serviceVersionOption) apply(c *config) {
	c.serviceVersion = o.serviceVersion
}

// WithServiceVersion sets the service version for telemetry data.
func WithServiceVersion(serviceVersion string) Option {
	return &serviceVersionOption{
		serviceVersion: serviceVersion,
	}
}

// exporterTypeOption configures the exporter type.
type exporterTypeOption struct {
	exporterType string
}

func (o *exporterTypeOption) apply(c *config) {
	c.exporterType = o.exporterType
}

// WithExporterType sets the exporter type (noop, otlp-http, otlp-grpc, dual).
func WithExporterType(exporterType string) Option {
	return &exporterTypeOption{
		exporterType: exporterType,
	}
}

// httpEndpointOption configures the HTTP endpoint.
type httpEndpointOption struct {
	endpoint string
}

func (o *httpEndpointOption) apply(c *config) {
	c.httpEndpoint = o.endpoint
}

// WithHTTPEndpoint sets the HTTP endpoint for OTLP export.
func WithHTTPEndpoint(endpoint string) Option {
	return &httpEndpointOption{
		endpoint: endpoint,
	}
}

// grpcEndpointOption configures the gRPC endpoint.
type grpcEndpointOption struct {
	endpoint string
}

func (o *grpcEndpointOption) apply(c *config) {
	c.grpcEndpoint = o.endpoint
}

// WithGRPCEndpoint sets the gRPC endpoint for OTLP export.
func WithGRPCEndpoint(endpoint string) Option {
	return &grpcEndpointOption{
		endpoint: endpoint,
	}
}

// headersOption configures additional headers.
type headersOption struct {
	headers map[string]string
}

func (o *headersOption) apply(c *config) {
	c.headers = o.headers
}

// WithHeaders sets additional headers for OTLP exporters.
func WithHeaders(headers map[string]string) Option {
	return &headersOption{
		headers: headers,
	}
}

// timeoutOption configures the timeout.
type timeoutOption struct {
	timeout time.Duration
}

func (o *timeoutOption) apply(c *config) {
	c.timeout = o.timeout
}

// WithTimeout sets the timeout for telemetry operations.
func WithTimeout(timeout time.Duration) Option {
	return &timeoutOption{
		timeout: timeout,
	}
}

// batchTimeoutOption configures the batch timeout.
type batchTimeoutOption struct {
	timeout time.Duration
}

func (o *batchTimeoutOption) apply(c *config) {
	c.batchTimeout = o.timeout
}

// WithBatchTimeout sets the timeout for batch exports.
func WithBatchTimeout(timeout time.Duration) Option {
	return &batchTimeoutOption{
		timeout: timeout,
	}
}

// maxExportBatchOption configures the maximum export batch size.
type maxExportBatchOption struct {
	size int
}

func (o *maxExportBatchOption) apply(c *config) {
	c.maxExportBatch = o.size
}

// WithMaxExportBatch sets the maximum number of items in an export batch.
func WithMaxExportBatch(size int) Option {
	return &maxExportBatchOption{
		size: size,
	}
}

// maxQueueSizeOption configures the maximum queue size.
type maxQueueSizeOption struct {
	size int
}

func (o *maxQueueSizeOption) apply(c *config) {
	c.maxQueueSize = o.size
}

// WithMaxQueueSize sets the maximum queue size for pending exports.
func WithMaxQueueSize(size int) Option {
	return &maxQueueSizeOption{
		size: size,
	}
}

// metricsOption configures metrics collection.
type metricsOption struct {
	enabled bool
}

func (o *metricsOption) apply(c *config) {
	c.enableMetrics = o.enabled
}

// WithMetrics enables or disables metrics collection.
func WithMetrics(enabled bool) Option {
	return &metricsOption{
		enabled: enabled,
	}
}

// tracesOption configures traces collection.
type tracesOption struct {
	enabled bool
}

func (o *tracesOption) apply(c *config) {
	c.enableTraces = o.enabled
}

// WithTraces enables or disables trace collection.
func WithTraces(enabled bool) Option {
	return &tracesOption{
		enabled: enabled,
	}
}

// logsOption configures logs collection.
type logsOption struct {
	enabled bool
}

func (o *logsOption) apply(c *config) {
	c.enableLogs = o.enabled
}

// WithLogs enables or disables log collection.
func WithLogs(enabled bool) Option {
	return &logsOption{
		enabled: enabled,
	}
}

// collectionOption configures collection functionality.
type collectionOption struct {
	enabled bool
}

func (o *collectionOption) apply(c *config) {
	c.enableCollection = o.enabled
}

// WithCollection enables or disables telemetry collection from other services.
func WithCollection(enabled bool) Option {
	return &collectionOption{
		enabled: enabled,
	}
}

// aggregationOption configures aggregation functionality.
type aggregationOption struct {
	enabled bool
}

func (o *aggregationOption) apply(c *config) {
	c.enableAggregation = o.enabled
}

// WithAggregation enables or disables telemetry aggregation.
func WithAggregation(enabled bool) Option {
	return &aggregationOption{
		enabled: enabled,
	}
}

// collectionIntervalOption configures the collection interval.
type collectionIntervalOption struct {
	interval time.Duration
}

func (o *collectionIntervalOption) apply(c *config) {
	c.collectionInterval = o.interval
}

// WithCollectionInterval sets the interval for telemetry collection.
func WithCollectionInterval(interval time.Duration) Option {
	return &collectionIntervalOption{
		interval: interval,
	}
}

// shutdownTimeoutOption configures the shutdown timeout.
type shutdownTimeoutOption struct {
	timeout time.Duration
}

func (o *shutdownTimeoutOption) apply(c *config) {
	c.shutdownTimeout = o.timeout
}

// WithShutdownTimeout sets the timeout for service shutdown.
func WithShutdownTimeout(timeout time.Duration) Option {
	return &shutdownTimeoutOption{
		timeout: timeout,
	}
}

// insecureOption configures insecure connections.
type insecureOption struct {
	insecure bool
}

func (o *insecureOption) apply(c *config) {
	c.insecure = o.insecure
}

// WithInsecure enables or disables insecure connections to OTLP endpoints.
func WithInsecure(insecure bool) Option {
	return &insecureOption{
		insecure: insecure,
	}
}

// samplingRatioOption configures the sampling ratio.
type samplingRatioOption struct {
	ratio float64
}

func (o *samplingRatioOption) apply(c *config) {
	if o.ratio < 0.0 {
		c.samplingRatio = 0.0
	} else if o.ratio > 1.0 {
		c.samplingRatio = 1.0
	} else {
		c.samplingRatio = o.ratio
	}
}

// WithSamplingRatio sets the sampling ratio for traces (0.0 to 1.0).
func WithSamplingRatio(ratio float64) Option {
	return &samplingRatioOption{
		ratio: ratio,
	}
}

// resourceAttributesOption configures resource attributes.
type resourceAttributesOption struct {
	attrs map[string]string
}

func (o *resourceAttributesOption) apply(c *config) {
	c.resourceAttrs = o.attrs
}

// WithResourceAttributes sets additional resource attributes for telemetry data.
func WithResourceAttributes(attrs map[string]string) Option {
	return &resourceAttributesOption{
		attrs: attrs,
	}
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
