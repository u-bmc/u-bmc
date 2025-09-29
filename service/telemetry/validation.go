// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ValidateConfig validates the entire telemetry service configuration
// and ensures all mandatory requirements are met for OTLP enforcement.
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate basic service information
	if err := validateServiceInfo(config); err != nil {
		return fmt.Errorf("service information validation failed: %w", err)
	}

	// Validate exporter configuration
	if err := validateExporterConfig(config); err != nil {
		return fmt.Errorf("exporter configuration validation failed: %w", err)
	}

	// Validate timing and performance settings
	if err := validateTimingConfig(config); err != nil {
		return fmt.Errorf("timing configuration validation failed: %w", err)
	}

	// Validate telemetry signal configuration
	if err := validateSignalConfig(config); err != nil {
		return fmt.Errorf("signal configuration validation failed: %w", err)
	}

	// Validate collection and aggregation settings
	if err := validateCollectionConfig(config); err != nil {
		return fmt.Errorf("collection configuration validation failed: %w", err)
	}

	// Validate runtime configuration settings
	if err := validateRuntimeConfig(config); err != nil {
		return fmt.Errorf("runtime configuration validation failed: %w", err)
	}

	return nil
}

// validateServiceInfo validates basic service information.
func validateServiceInfo(config *Config) error {
	if config.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if !isValidServiceName(config.ServiceName) {
		return fmt.Errorf("invalid service name format: %s", config.ServiceName)
	}

	if config.ServiceVersion == "" {
		return fmt.Errorf("service version cannot be empty")
	}

	if !isValidVersion(config.ServiceVersion) {
		return fmt.Errorf("invalid service version format: %s", config.ServiceVersion)
	}

	if config.CollectorName == "" {
		return fmt.Errorf("collector name cannot be empty")
	}

	if !isValidServiceName(config.CollectorName) {
		return fmt.Errorf("invalid collector name format: %s", config.CollectorName)
	}

	return nil
}

// validateExporterConfig validates exporter configuration for the telemetry service.
// NoOp is the default and recommended mode to minimize overhead.
func validateExporterConfig(config *Config) error {
	// Allow NoOp as default - telemetry service drops data by default for minimal overhead
	validExporterTypes := []string{"noop", "otlp-http", "otlp-grpc", "dual"}
	if !containsString(validExporterTypes, config.ExporterType) {
		return fmt.Errorf("invalid exporter type '%s' - must be one of: %s",
			config.ExporterType, strings.Join(validExporterTypes, ", "))
	}

	// Validate endpoints based on exporter type (only when not NoOp)
	switch config.ExporterType {
	case "noop":
		// NoOp is valid and default - no endpoint validation needed
	case "otlp-http":
		if config.HTTPEndpoint == "" {
			return fmt.Errorf("HTTP endpoint is required for OTLP HTTP exporter")
		}
		if err := validateHTTPEndpoint(config.HTTPEndpoint); err != nil {
			return fmt.Errorf("invalid HTTP endpoint: %w", err)
		}

	case "otlp-grpc":
		if config.GRPCEndpoint == "" {
			return fmt.Errorf("gRPC endpoint is required for OTLP gRPC exporter")
		}
		if err := validateGRPCEndpoint(config.GRPCEndpoint); err != nil {
			return fmt.Errorf("invalid gRPC endpoint: %w", err)
		}

	case "dual":
		if config.HTTPEndpoint == "" {
			return fmt.Errorf("HTTP endpoint is required for dual exporter")
		}
		if config.GRPCEndpoint == "" {
			return fmt.Errorf("gRPC endpoint is required for dual exporter")
		}
		if err := validateHTTPEndpoint(config.HTTPEndpoint); err != nil {
			return fmt.Errorf("invalid HTTP endpoint for dual exporter: %w", err)
		}
		if err := validateGRPCEndpoint(config.GRPCEndpoint); err != nil {
			return fmt.Errorf("invalid gRPC endpoint for dual exporter: %w", err)
		}
	}

	// Validate headers if provided
	if err := validateHeaders(config.Headers); err != nil {
		return fmt.Errorf("invalid headers: %w", err)
	}

	return nil
}

// validateTimingConfig validates timing and performance configuration.
func validateTimingConfig(config *Config) error {
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", config.Timeout)
	}

	if config.Timeout > 5*time.Minute {
		return fmt.Errorf("timeout too large (max 5 minutes), got %v", config.Timeout)
	}

	if config.BatchTimeout <= 0 {
		return fmt.Errorf("batch timeout must be positive, got %v", config.BatchTimeout)
	}

	if config.BatchTimeout > config.Timeout {
		return fmt.Errorf("batch timeout cannot be larger than timeout")
	}

	if config.MaxExportBatch <= 0 {
		return fmt.Errorf("max export batch size must be positive, got %d", config.MaxExportBatch)
	}

	if config.MaxExportBatch > 10000 {
		return fmt.Errorf("max export batch size too large (max 10000), got %d", config.MaxExportBatch)
	}

	if config.MaxQueueSize <= 0 {
		return fmt.Errorf("max queue size must be positive, got %d", config.MaxQueueSize)
	}

	if config.MaxQueueSize < config.MaxExportBatch {
		return fmt.Errorf("max queue size must be at least as large as max export batch size")
	}

	if config.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive, got %v", config.ShutdownTimeout)
	}

	if config.CollectionInterval <= 0 {
		return fmt.Errorf("collection interval must be positive, got %v", config.CollectionInterval)
	}

	return nil
}

// validateSignalConfig validates telemetry signal configuration.
func validateSignalConfig(config *Config) error {
	// Enforce that at least one signal type is enabled
	if !config.EnableMetrics && !config.EnableTraces && !config.EnableLogs {
		return fmt.Errorf("at least one telemetry signal (metrics, traces, or logs) must be enabled")
	}

	// Validate sampling ratio
	if config.SamplingRatio < 0.0 || config.SamplingRatio > 1.0 {
		return fmt.Errorf("sampling ratio must be between 0.0 and 1.0, got %f", config.SamplingRatio)
	}

	// Validate resource attributes
	if err := validateResourceAttributes(config.ResourceAttrs); err != nil {
		return fmt.Errorf("invalid resource attributes: %w", err)
	}

	return nil
}

// validateCollectionConfig validates collection and aggregation configuration.
func validateCollectionConfig(config *Config) error {
	// Collection should be enabled for telemetry service to collect from other services
	// Allow disabling for testing scenarios
	if !config.EnableCollection {
		// Allow but warn - collection is typically enabled for the telemetry service
	}

	// Validate filter configuration if provided
	if config.FilterRules != nil {
		if err := validateFilterConfig(config.FilterRules); err != nil {
			return fmt.Errorf("invalid filter configuration: %w", err)
		}
	}

	// Validate aggregation configuration if provided
	if config.AggregationRules != nil {
		if err := validateAggregationConfig(config.AggregationRules); err != nil {
			return fmt.Errorf("invalid aggregation configuration: %w", err)
		}
	}

	return nil
}

// validateRuntimeConfig validates runtime configuration settings.
func validateRuntimeConfig(config *Config) error {
	// Validate filter rules if provided
	if config.FilterRules != nil {
		for i, rule := range config.FilterRules.MetricsFilters {
			if err := validateFilterRule(&rule); err != nil {
				return fmt.Errorf("invalid metrics filter rule %d: %w", i, err)
			}
		}

		for i, rule := range config.FilterRules.TracesFilters {
			if err := validateFilterRule(&rule); err != nil {
				return fmt.Errorf("invalid traces filter rule %d: %w", i, err)
			}
		}

		for i, rule := range config.FilterRules.LogsFilters {
			if err := validateFilterRule(&rule); err != nil {
				return fmt.Errorf("invalid logs filter rule %d: %w", i, err)
			}
		}

		// Validate sampling overrides
		for service, ratio := range config.FilterRules.SamplingOverrides {
			if !isValidServiceName(service) {
				return fmt.Errorf("invalid service name in sampling override: %s", service)
			}
			if ratio < 0.0 || ratio > 1.0 {
				return fmt.Errorf("invalid sampling ratio %f for service %s", ratio, service)
			}
		}

		// Validate debug services
		for _, service := range config.FilterRules.DebugServices {
			if service != "*" && !isValidServiceName(service) {
				return fmt.Errorf("invalid service name in debug services: %s", service)
			}
		}
	}

	return nil
}

// Validation helper functions

// validateHTTPEndpoint validates an HTTP endpoint URL.
func validateHTTPEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("endpoint must use http or https scheme, got %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("endpoint must include host")
	}

	return nil
}

// validateGRPCEndpoint validates a gRPC endpoint.
func validateGRPCEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}

	// gRPC endpoints can be host:port or full URLs
	if strings.Contains(endpoint, "://") {
		u, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("invalid URL format: %w", err)
		}
		if u.Host == "" {
			return fmt.Errorf("endpoint must include host")
		}
	} else {
		// Validate host:port format
		parts := strings.Split(endpoint, ":")
		if len(parts) != 2 {
			return fmt.Errorf("gRPC endpoint must be in host:port format")
		}
		if parts[0] == "" {
			return fmt.Errorf("host cannot be empty")
		}
		if parts[1] == "" {
			return fmt.Errorf("port cannot be empty")
		}
	}

	return nil
}

// validateHeaders validates HTTP headers.
func validateHeaders(headers map[string]string) error {
	for key, value := range headers {
		if key == "" {
			return fmt.Errorf("header key cannot be empty")
		}
		if !isValidHeaderName(key) {
			return fmt.Errorf("invalid header name: %s", key)
		}
		if value == "" {
			return fmt.Errorf("header value cannot be empty for key %s", key)
		}
	}
	return nil
}

// validateResourceAttributes validates resource attributes.
func validateResourceAttributes(attrs map[string]string) error {
	for key, value := range attrs {
		if key == "" {
			return fmt.Errorf("resource attribute key cannot be empty")
		}
		if !isValidAttributeKey(key) {
			return fmt.Errorf("invalid resource attribute key: %s", key)
		}
		if value == "" {
			return fmt.Errorf("resource attribute value cannot be empty for key %s", key)
		}
	}
	return nil
}

// validateFilterConfig validates filter configuration.
func validateFilterConfig(config *FilterConfig) error {
	if config == nil {
		return fmt.Errorf("filter configuration cannot be nil")
	}

	// Validate sampling overrides
	for service, ratio := range config.SamplingOverrides {
		if !isValidServiceName(service) && service != "*" {
			return fmt.Errorf("invalid service name in sampling override: %s", service)
		}
		if ratio < 0.0 || ratio > 1.0 {
			return fmt.Errorf("invalid sampling ratio %f for service %s", ratio, service)
		}
	}

	return nil
}

// validateAggregationConfig validates aggregation configuration.
func validateAggregationConfig(config *AggregationConfig) error {
	if config == nil {
		return fmt.Errorf("aggregation configuration cannot be nil")
	}

	if config.WindowSize <= 0 {
		return fmt.Errorf("aggregation window size must be positive")
	}

	if config.MaxCardinality <= 0 {
		return fmt.Errorf("max cardinality must be positive")
	}

	if config.MaxCardinality > 1000000 {
		return fmt.Errorf("max cardinality too large (max 1,000,000)")
	}

	return nil
}

// validateFilterRule validates a single filter rule.
func validateFilterRule(rule *FilterRule) error {
	if rule.Name == "" {
		return fmt.Errorf("filter rule name cannot be empty")
	}

	validTypes := []string{"drop", "sample", "transform"}
	if !containsString(validTypes, rule.Type) {
		return fmt.Errorf("invalid filter rule type: %s", rule.Type)
	}

	// Validate action
	switch rule.Action.Type {
	case "drop":
		// No additional validation needed
	case "sample":
		if rule.Action.SampleRate < 0.0 || rule.Action.SampleRate > 1.0 {
			return fmt.Errorf("invalid sample rate: %f", rule.Action.SampleRate)
		}
	case "modify":
		if len(rule.Action.Modifications) == 0 {
			return fmt.Errorf("modify action requires at least one modification")
		}
	default:
		return fmt.Errorf("invalid action type: %s", rule.Action.Type)
	}

	return nil
}

// Validation utility functions

// isValidServiceName validates service name format.
func isValidServiceName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	// Service names should follow DNS label format
	matched, _ := regexp.MatchString(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, name)
	return matched
}

// isValidVersion validates semantic version format.
func isValidVersion(version string) bool {
	if version == "" {
		return false
	}
	// Basic semantic version validation
	matched, _ := regexp.MatchString(`^v?[0-9]+\.[0-9]+\.[0-9]+`, version)
	return matched
}

// isValidHeaderName validates HTTP header name format.
func isValidHeaderName(name string) bool {
	if name == "" {
		return false
	}
	// HTTP header names should be valid tokens
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, name)
	return matched
}

// isValidAttributeKey validates OpenTelemetry attribute key format.
func isValidAttributeKey(key string) bool {
	if key == "" || len(key) > 256 {
		return false
	}
	// Attribute keys should be valid identifiers
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_.-]*$`, key)
	return matched
}

// containsString checks if a slice contains a specific string.
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SanitizeConfig sanitizes and normalizes configuration values.
func SanitizeConfig(config *Config) {
	if config == nil {
		return
	}

	// Normalize service names to lowercase
	config.ServiceName = strings.ToLower(strings.TrimSpace(config.ServiceName))
	config.CollectorName = strings.ToLower(strings.TrimSpace(config.CollectorName))

	// Normalize exporter type to lowercase
	config.ExporterType = strings.ToLower(strings.TrimSpace(config.ExporterType))

	// Trim whitespace from endpoints
	config.HTTPEndpoint = strings.TrimSpace(config.HTTPEndpoint)
	config.GRPCEndpoint = strings.TrimSpace(config.GRPCEndpoint)

	// Ensure resource attributes don't have empty values
	if config.ResourceAttrs != nil {
		cleanAttrs := make(map[string]string)
		for k, v := range config.ResourceAttrs {
			if k != "" && v != "" {
				cleanAttrs[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
		config.ResourceAttrs = cleanAttrs
	}

	// Ensure headers don't have empty values
	if config.Headers != nil {
		cleanHeaders := make(map[string]string)
		for k, v := range config.Headers {
			if k != "" && v != "" {
				cleanHeaders[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
		config.Headers = cleanHeaders
	}

	// Clamp sampling ratio to valid range
	if config.SamplingRatio < 0.0 {
		config.SamplingRatio = 0.0
	} else if config.SamplingRatio > 1.0 {
		config.SamplingRatio = 1.0
	}

	// Set minimum values for timing configuration
	if config.Timeout < time.Second {
		config.Timeout = time.Second
	}
	if config.BatchTimeout < 100*time.Millisecond {
		config.BatchTimeout = 100 * time.Millisecond
	}
	if config.ShutdownTimeout < time.Second {
		config.ShutdownTimeout = time.Second
	}
	if config.CollectionInterval < time.Second {
		config.CollectionInterval = time.Second
	}

	// Set minimum values for batch and queue sizes
	if config.MaxExportBatch < 1 {
		config.MaxExportBatch = 1
	}
	if config.MaxQueueSize < config.MaxExportBatch {
		config.MaxQueueSize = config.MaxExportBatch * 2
	}
}

// GetConfigWarnings returns a list of configuration warnings (non-fatal issues).
func GetConfigWarnings(config *Config) []string {
	var warnings []string

	if config == nil {
		return []string{"configuration is nil"}
	}

	// Check for potentially problematic settings
	if config.SamplingRatio == 0.0 && config.ExporterType != "noop" {
		warnings = append(warnings, "sampling ratio is 0.0 - no traces will be exported")
	}

	if config.Timeout > 60*time.Second {
		warnings = append(warnings, "timeout is very large (>60s) - may cause performance issues")
	}

	if config.MaxExportBatch > 1000 {
		warnings = append(warnings, "max export batch size is very large (>1000) - may cause memory issues")
	}

	if config.MaxQueueSize > 10000 {
		warnings = append(warnings, "max queue size is very large (>10000) - may cause memory issues")
	}

	if !config.EnableMetrics && !config.EnableTraces && !config.EnableLogs {
		warnings = append(warnings, "all telemetry signals are disabled - no data will be collected")
	}

	if config.Insecure && (strings.Contains(config.HTTPEndpoint, "https://") ||
		strings.Contains(config.GRPCEndpoint, "tls://")) {
		warnings = append(warnings, "insecure mode enabled but endpoints suggest TLS")
	}

	return warnings
}
