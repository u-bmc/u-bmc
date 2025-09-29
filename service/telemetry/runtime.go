// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"go.opentelemetry.io/otel/propagation"
)

// RuntimeConfigManager handles runtime configuration updates for the telemetry service.
// It provides a high-level interface for sending configuration updates to telemetry
// services and validating configuration changes.
type RuntimeConfigManager struct {
	natsConn *nats.Conn
	logger   *slog.Logger
}

// NewRuntimeConfigManager creates a new runtime configuration manager.
func NewRuntimeConfigManager(natsConn *nats.Conn, logger *slog.Logger) *RuntimeConfigManager {
	return &RuntimeConfigManager{
		natsConn: natsConn,
		logger:   logger,
	}
}

// EnableDebugMode enables debug mode for a specific service or all services.
// If serviceName is empty or "*", debug mode is enabled for all services.
func (r *RuntimeConfigManager) EnableDebugMode(ctx context.Context, serviceName string) error {
	config := RuntimeConfig{
		Type:        "debug_mode",
		ServiceName: serviceName,
		DebugMode:   boolPtr(true),
	}

	subject := r.getConfigSubject(serviceName)
	return r.sendConfig(ctx, subject, &config)
}

// DisableDebugMode disables debug mode for a specific service or all services.
// If serviceName is empty or "*", debug mode is disabled for all services.
func (r *RuntimeConfigManager) DisableDebugMode(ctx context.Context, serviceName string) error {
	config := RuntimeConfig{
		Type:        "debug_mode",
		ServiceName: serviceName,
		DebugMode:   boolPtr(false),
	}

	subject := r.getConfigSubject(serviceName)
	return r.sendConfig(ctx, subject, &config)
}

// UpdateSamplingRatio updates the sampling ratio for traces.
// If serviceName is empty or "*", the ratio is applied globally.
func (r *RuntimeConfigManager) UpdateSamplingRatio(ctx context.Context, serviceName string, ratio float64) error {
	if ratio < 0.0 || ratio > 1.0 {
		return fmt.Errorf("sampling ratio must be between 0.0 and 1.0, got %f", ratio)
	}

	config := RuntimeConfig{
		Type:          "sampling_ratio",
		ServiceName:   serviceName,
		SamplingRatio: &ratio,
	}

	subject := r.getConfigSubject(serviceName)
	return r.sendConfig(ctx, subject, &config)
}

// UpdateFilterConfig updates the filtering configuration.
func (r *RuntimeConfigManager) UpdateFilterConfig(ctx context.Context, serviceName string, filterConfig *FilterConfig) error {
	if err := r.validateFilterConfig(filterConfig); err != nil {
		return fmt.Errorf("invalid filter configuration: %w", err)
	}

	config := RuntimeConfig{
		Type:         "filter_config",
		ServiceName:  serviceName,
		FilterConfig: filterConfig,
	}

	subject := r.getConfigSubject(serviceName)
	return r.sendConfig(ctx, subject, &config)
}

// UpdateAggregationConfig updates the aggregation configuration.
func (r *RuntimeConfigManager) UpdateAggregationConfig(ctx context.Context, serviceName string, aggregationConfig *AggregationConfig) error {
	if err := r.validateAggregationConfig(aggregationConfig); err != nil {
		return fmt.Errorf("invalid aggregation configuration: %w", err)
	}

	config := RuntimeConfig{
		Type:              "aggregation_config",
		ServiceName:       serviceName,
		AggregationConfig: aggregationConfig,
	}

	subject := r.getConfigSubject(serviceName)
	return r.sendConfig(ctx, subject, &config)
}

// UpdateExporterConfig updates the exporter configuration.
func (r *RuntimeConfigManager) UpdateExporterConfig(ctx context.Context, serviceName string, exporterConfig *ExporterConfig) error {
	if err := r.validateExporterConfig(exporterConfig); err != nil {
		return fmt.Errorf("invalid exporter configuration: %w", err)
	}

	config := RuntimeConfig{
		Type:           "exporter_config",
		ServiceName:    serviceName,
		ExporterConfig: exporterConfig,
	}

	subject := r.getConfigSubject(serviceName)
	return r.sendConfig(ctx, subject, &config)
}

// getConfigSubject returns the appropriate NATS subject for configuration updates.
func (r *RuntimeConfigManager) getConfigSubject(serviceName string) string {
	if serviceName == "" || serviceName == "*" {
		return "telemetry.config.global"
	}
	return fmt.Sprintf("telemetry.config.%s", serviceName)
}

// sendConfig sends a configuration update message via NATS.
func (r *RuntimeConfigManager) sendConfig(ctx context.Context, subject string, config *RuntimeConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Add telemetry context to message headers
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  make(nats.Header),
	}

	// Inject telemetry context into headers
	carrier := propagation.HeaderCarrier(msg.Header)
	telemetry.InjectContext(ctx, carrier)

	if err := r.natsConn.PublishMsg(msg); err != nil {
		return fmt.Errorf("failed to publish configuration update: %w", err)
	}

	r.logger.InfoContext(ctx, "Sent runtime configuration update",
		"subject", subject,
		"config_type", config.Type,
		"service_name", config.ServiceName,
	)

	return nil
}

// validateFilterConfig validates a filter configuration.
func (r *RuntimeConfigManager) validateFilterConfig(config *FilterConfig) error {
	if config == nil {
		return fmt.Errorf("filter configuration cannot be nil")
	}

	// Validate sampling overrides
	for service, ratio := range config.SamplingOverrides {
		if ratio < 0.0 || ratio > 1.0 {
			return fmt.Errorf("invalid sampling ratio %f for service %s", ratio, service)
		}
	}

	// Validate filter rules
	for i, rule := range config.MetricsFilters {
		if err := r.validateFilterRule(&rule); err != nil {
			return fmt.Errorf("invalid metrics filter rule %d: %w", i, err)
		}
	}

	for i, rule := range config.TracesFilters {
		if err := r.validateFilterRule(&rule); err != nil {
			return fmt.Errorf("invalid traces filter rule %d: %w", i, err)
		}
	}

	for i, rule := range config.LogsFilters {
		if err := r.validateFilterRule(&rule); err != nil {
			return fmt.Errorf("invalid logs filter rule %d: %w", i, err)
		}
	}

	return nil
}

// validateFilterRule validates a single filter rule.
func (r *RuntimeConfigManager) validateFilterRule(rule *FilterRule) error {
	if rule.Name == "" {
		return fmt.Errorf("filter rule name cannot be empty")
	}

	validTypes := []string{"drop", "sample", "transform"}
	if !containsStringSlice(validTypes, rule.Type) {
		return fmt.Errorf("invalid filter rule type: %s", rule.Type)
	}

	// Validate action based on type
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

// validateAggregationConfig validates an aggregation configuration.
func (r *RuntimeConfigManager) validateAggregationConfig(config *AggregationConfig) error {
	if config == nil {
		return fmt.Errorf("aggregation configuration cannot be nil")
	}

	if config.WindowSize <= 0 {
		return fmt.Errorf("window size must be positive")
	}

	if config.MaxCardinality <= 0 {
		return fmt.Errorf("max cardinality must be positive")
	}

	// Validate aggregation rules
	for i, rule := range config.MetricsAggregation {
		if err := r.validateAggregationRule(&rule); err != nil {
			return fmt.Errorf("invalid metrics aggregation rule %d: %w", i, err)
		}
	}

	for i, rule := range config.TracesAggregation {
		if err := r.validateAggregationRule(&rule); err != nil {
			return fmt.Errorf("invalid traces aggregation rule %d: %w", i, err)
		}
	}

	return nil
}

// validateAggregationRule validates a single aggregation rule.
func (r *RuntimeConfigManager) validateAggregationRule(rule *AggregationRule) error {
	if rule.Name == "" {
		return fmt.Errorf("aggregation rule name cannot be empty")
	}

	validTypes := []string{"sum", "avg", "count", "histogram"}
	if !containsStringSlice(validTypes, rule.Type) {
		return fmt.Errorf("invalid aggregation rule type: %s", rule.Type)
	}

	if len(rule.GroupBy) == 0 {
		return fmt.Errorf("aggregation rule must specify at least one group-by attribute")
	}

	return nil
}

// validateExporterConfig validates an exporter configuration.
func (r *RuntimeConfigManager) validateExporterConfig(config *ExporterConfig) error {
	if config == nil {
		return fmt.Errorf("exporter configuration cannot be nil")
	}

	validTypes := []string{"noop", "otlp-http", "otlp-grpc", "dual"}
	if !containsStringSlice(validTypes, config.ExporterType) {
		return fmt.Errorf("invalid exporter type: %s", config.ExporterType)
	}

	// Validate endpoints based on exporter type
	switch config.ExporterType {
	case "otlp-http":
		if config.HTTPEndpoint == "" {
			return fmt.Errorf("HTTP endpoint required for otlp-http exporter")
		}
	case "otlp-grpc":
		if config.GRPCEndpoint == "" {
			return fmt.Errorf("gRPC endpoint required for otlp-grpc exporter")
		}
	case "dual":
		if config.HTTPEndpoint == "" || config.GRPCEndpoint == "" {
			return fmt.Errorf("both HTTP and gRPC endpoints required for dual exporter")
		}
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// CreateDefaultFilterConfig creates a default filter configuration.
func CreateDefaultFilterConfig() *FilterConfig {
	return &FilterConfig{
		EnableFiltering:   false,
		MetricsFilters:    make([]FilterRule, 0),
		TracesFilters:     make([]FilterRule, 0),
		LogsFilters:       make([]FilterRule, 0),
		SamplingOverrides: make(map[string]float64),
		DebugServices:     make([]string, 0),
	}
}

// CreateDefaultAggregationConfig creates a default aggregation configuration.
func CreateDefaultAggregationConfig() *AggregationConfig {
	return &AggregationConfig{
		EnableAggregation:  true,
		MetricsAggregation: make([]AggregationRule, 0),
		TracesAggregation:  make([]AggregationRule, 0),
		WindowSize:         30 * time.Second,
		MaxCardinality:     10000,
	}
}

// CreateProductionFilterConfig creates a filter configuration optimized for production.
func CreateProductionFilterConfig() *FilterConfig {
	return &FilterConfig{
		EnableFiltering: true,
		MetricsFilters: []FilterRule{
			{
				Name:    "drop_debug_metrics",
				Type:    "drop",
				Enabled: true,
				Condition: FilterCondition{
					AttributeMatches: map[string]string{
						"level": "debug",
					},
				},
				Action: FilterAction{
					Type: "drop",
				},
			},
		},
		TracesFilters: []FilterRule{
			{
				Name:    "sample_routine_traces",
				Type:    "sample",
				Enabled: true,
				Condition: FilterCondition{
					AttributeMatches: map[string]string{
						"operation.type": "routine",
					},
				},
				Action: FilterAction{
					Type:       "sample",
					SampleRate: 0.1, // Sample 10% of routine traces
				},
			},
		},
		LogsFilters: []FilterRule{
			{
				Name:    "drop_trace_logs",
				Type:    "drop",
				Enabled: true,
				Condition: FilterCondition{
					LogLevel: "trace",
				},
				Action: FilterAction{
					Type: "drop",
				},
			},
		},
		SamplingOverrides: map[string]float64{
			"power":   0.5, // 50% sampling for power service
			"thermal": 0.3, // 30% sampling for thermal service
		},
		DebugServices: make([]string, 0),
	}
}

// CreateDebugFilterConfig creates a filter configuration optimized for debugging.
func CreateDebugFilterConfig() *FilterConfig {
	return &FilterConfig{
		EnableFiltering:   false, // Disable filtering in debug mode
		MetricsFilters:    make([]FilterRule, 0),
		TracesFilters:     make([]FilterRule, 0),
		LogsFilters:       make([]FilterRule, 0),
		SamplingOverrides: map[string]float64{}, // No sampling overrides
		DebugServices:     []string{"*"},        // All services in debug mode
	}
}

// Helper functions

// boolPtr returns a pointer to a boolean value.
func boolPtr(b bool) *bool {
	return &b
}

// containsStringSlice checks if a slice contains a specific string.
func containsStringSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// isDebugService checks if a service is in debug mode.
func (c *FilterConfig) isDebugService(serviceName string) bool {
	for _, debugService := range c.DebugServices {
		if debugService == "*" || debugService == serviceName {
			return true
		}
	}
	return false
}

// getSamplingRatio returns the sampling ratio for a service.
func (c *FilterConfig) getSamplingRatio(serviceName string, defaultRatio float64) float64 {
	if ratio, exists := c.SamplingOverrides[serviceName]; exists {
		return ratio
	}
	return defaultRatio
}

// MatchesCondition checks if telemetry data matches a filter condition.
func (c *FilterCondition) MatchesCondition(serviceName string, attributes map[string]string, resourceAttrs map[string]string) bool {
	// Check service name
	if c.ServiceName != "" && c.ServiceName != serviceName {
		return false
	}

	// Check attribute matches
	for key, expectedValue := range c.AttributeMatches {
		if actualValue, exists := attributes[key]; !exists || actualValue != expectedValue {
			return false
		}
	}

	// Check resource matches
	for key, expectedValue := range c.ResourceMatches {
		if actualValue, exists := resourceAttrs[key]; !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

// ApplyFilter applies a filter rule to determine the action to take.
func (r *FilterRule) ApplyFilter(data interface{}) (bool, interface{}) {
	if !r.Enabled {
		return false, data // Don't filter if rule is disabled
	}

	// TODO: Implement actual condition matching based on data type
	// This would require parsing the telemetry data and checking conditions

	switch r.Action.Type {
	case "drop":
		return true, nil // Drop the data
	case "sample":
		// TODO: Implement probabilistic sampling based on SampleRate
		return false, data // For now, don't drop
	case "modify":
		// TODO: Implement data modification based on Modifications
		return false, data // For now, return unmodified
	default:
		return false, data
	}
}

// String returns a string representation of the filter condition.
func (c *FilterCondition) String() string {
	var parts []string

	if c.ServiceName != "" {
		parts = append(parts, fmt.Sprintf("service=%s", c.ServiceName))
	}

	if c.MetricName != "" {
		parts = append(parts, fmt.Sprintf("metric=%s", c.MetricName))
	}

	if c.SpanName != "" {
		parts = append(parts, fmt.Sprintf("span=%s", c.SpanName))
	}

	if c.LogLevel != "" {
		parts = append(parts, fmt.Sprintf("level=%s", c.LogLevel))
	}

	for key, value := range c.AttributeMatches {
		parts = append(parts, fmt.Sprintf("attr.%s=%s", key, value))
	}

	for key, value := range c.ResourceMatches {
		parts = append(parts, fmt.Sprintf("resource.%s=%s", key, value))
	}

	return strings.Join(parts, " AND ")
}
