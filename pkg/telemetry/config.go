// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"time"
)

// ExporterType defines the type of telemetry exporter to use.
type ExporterType int

const (
	// NoOp discards all telemetry data with minimal overhead.
	NoOp ExporterType = iota
	// OTLPHTTP exports telemetry data via OTLP over HTTP.
	OTLPHTTP
	// OTLPgRPC exports telemetry data via OTLP over gRPC.
	OTLPgRPC
	// Dual exports telemetry data via both HTTP and gRPC.
	Dual
)

// Config holds the configuration for telemetry providers.
type Config struct {
	exporterType   ExporterType
	httpEndpoint   string
	grpcEndpoint   string
	headers        map[string]string
	timeout        time.Duration
	batchTimeout   time.Duration
	maxExportBatch int
	maxQueueSize   int
	serviceName    string
	serviceVersion string
	enableMetrics  bool
	enableTraces   bool
	enableLogs     bool
	insecure       bool
	samplingRatio  float64
	resourceAttrs  map[string]string
}

// DefaultConfig returns a default configuration for telemetry providers.
// Services generate telemetry data and send it to the central telemetry collector.
func DefaultConfig() *Config {
	return &Config{
		exporterType:   NoOp, // Services send to central collector, not direct export
		timeout:        30 * time.Second,
		batchTimeout:   5 * time.Second,
		maxExportBatch: 512,
		maxQueueSize:   2048,
		serviceName:    "u-bmc",
		serviceVersion: "1.0.0",
		enableMetrics:  true,
		enableTraces:   true,
		enableLogs:     true,
		insecure:       false,
		samplingRatio:  1.0,
		headers:        make(map[string]string),
		resourceAttrs:  make(map[string]string),
	}
}

// Option defines a function that modifies the telemetry configuration.
type Option func(*Config)

// WithExporterType sets the exporter type for telemetry data.
func WithExporterType(exporterType ExporterType) Option {
	return func(c *Config) {
		c.exporterType = exporterType
	}
}

// WithHTTPEndpoint sets the HTTP endpoint for OTLP export.
func WithHTTPEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.httpEndpoint = endpoint
	}
}

// WithgRPCEndpoint sets the gRPC endpoint for OTLP export.
func WithgRPCEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.grpcEndpoint = endpoint
	}
}

// WithHeaders sets additional headers for OTLP exporters.
func WithHeaders(headers map[string]string) Option {
	return func(c *Config) {
		c.headers = headers
	}
}

// WithTimeout sets the timeout for telemetry operations.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.timeout = timeout
	}
}

// WithBatchTimeout sets the timeout for batch exports.
func WithBatchTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.batchTimeout = timeout
	}
}

// WithMaxExportBatch sets the maximum number of items in an export batch.
func WithMaxExportBatch(size int) Option {
	return func(c *Config) {
		c.maxExportBatch = size
	}
}

// WithMaxQueueSize sets the maximum queue size for pending exports.
func WithMaxQueueSize(size int) Option {
	return func(c *Config) {
		c.maxQueueSize = size
	}
}

// WithServiceName sets the service name for telemetry data.
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.serviceName = name
	}
}

// WithServiceVersion sets the service version for telemetry data.
func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.serviceVersion = version
	}
}

// WithMetrics enables or disables metrics collection.
func WithMetrics(enabled bool) Option {
	return func(c *Config) {
		c.enableMetrics = enabled
	}
}

// WithTraces enables or disables trace collection.
func WithTraces(enabled bool) Option {
	return func(c *Config) {
		c.enableTraces = enabled
	}
}

// WithLogs enables or disables log collection.
func WithLogs(enabled bool) Option {
	return func(c *Config) {
		c.enableLogs = enabled
	}
}

// WithInsecure enables or disables insecure connections to OTLP endpoints.
func WithInsecure(insecure bool) Option {
	return func(c *Config) {
		c.insecure = insecure
	}
}

// WithSamplingRatio sets the sampling ratio for traces (0.0 to 1.0).
func WithSamplingRatio(ratio float64) Option {
	return func(c *Config) {
		if ratio < 0.0 {
			ratio = 0.0
		}
		if ratio > 1.0 {
			ratio = 1.0
		}
		c.samplingRatio = ratio
	}
}

// WithResourceAttributes sets additional resource attributes for telemetry data.
func WithResourceAttributes(attrs map[string]string) Option {
	return func(c *Config) {
		c.resourceAttrs = attrs
	}
}

// WithOTLPHTTP is a convenience function that configures OTLP HTTP export.
func WithOTLPHTTP(endpoint string) Option {
	return func(c *Config) {
		c.exporterType = OTLPHTTP
		c.httpEndpoint = endpoint
	}
}

// WithOTLPgRPC is a convenience function that configures OTLP gRPC export.
func WithOTLPgRPC(endpoint string) Option {
	return func(c *Config) {
		c.exporterType = OTLPgRPC
		c.grpcEndpoint = endpoint
	}
}

// WithDualOTLP is a convenience function that configures dual OTLP export.
func WithDualOTLP(httpEndpoint, grpcEndpoint string) Option {
	return func(c *Config) {
		c.exporterType = Dual
		c.httpEndpoint = httpEndpoint
		c.grpcEndpoint = grpcEndpoint
	}
}
