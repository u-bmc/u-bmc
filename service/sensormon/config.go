// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"fmt"
	"strings"
	"time"
)

// Default configuration constants.
const (
	DefaultServiceName            = "sensormon"
	DefaultServiceDescription     = "Sensor monitoring service for BMC systems"
	DefaultServiceVersion         = "1.0.0"
	DefaultHwmonPath              = "/sys/class/hwmon"
	DefaultGPIOChipPath           = "/dev/gpiochip0"
	DefaultMonitoringInterval     = 1 * time.Second
	DefaultThresholdCheckInterval = 5 * time.Second
	DefaultSensorTimeout          = 5 * time.Second
)

// Config holds the configuration for the sensor monitoring service.
type Config struct {
	// ServiceName is the name of the service in the NATS micro framework
	ServiceName string
	// ServiceDescription provides a human-readable description of the service
	ServiceDescription string
	// ServiceVersion is the semantic version of the service
	ServiceVersion string
	// HwmonPath is the path to hwmon devices in sysfs
	HwmonPath string
	// GPIOChipPath is the path to the GPIO chip device
	GPIOChipPath string
	// MonitoringInterval is the interval between sensor readings
	MonitoringInterval time.Duration
	// ThresholdCheckInterval is the interval between threshold checks
	ThresholdCheckInterval time.Duration
	// SensorTimeout is the maximum time to wait for sensor operations
	SensorTimeout time.Duration
	// EnableHwmonSensors enables hardware monitoring sensors
	EnableHwmonSensors bool
	// EnableGPIOSensors enables GPIO-based sensors
	EnableGPIOSensors bool
	// EnableMetrics enables metrics collection for sensor operations
	EnableMetrics bool
	// EnableTracing enables distributed tracing for sensor operations
	EnableTracing bool
	// EnableThresholdMonitoring enables automatic threshold monitoring
	EnableThresholdMonitoring bool
	// BroadcastSensorReadings enables broadcasting sensor readings via NATS
	BroadcastSensorReadings bool
	// PersistSensorData enables persisting sensor data to JetStream
	PersistSensorData bool
	// StreamName is the name of the JetStream stream for sensor data persistence
	StreamName string
	// StreamSubjects are the subjects the stream will listen on
	StreamSubjects []string
	// StreamRetention defines how long to retain sensor data
	StreamRetention time.Duration
	// MaxConcurrentReads limits the number of concurrent sensor reads
	MaxConcurrentReads int
	// SensorDiscoveryTimeout is the timeout for sensor discovery operations
	SensorDiscoveryTimeout time.Duration
	// EnableThermalIntegration enables thermal management integration
	EnableThermalIntegration bool
	// ThermalMgrEndpoint is the thermal manager service endpoint
	ThermalMgrEndpoint string
	// TemperatureUpdateInterval is the interval for sending temperature updates to thermal manager
	TemperatureUpdateInterval time.Duration
	// EnableThermalAlerts enables thermal threshold alerting
	EnableThermalAlerts bool
	// CriticalTempThreshold is the critical temperature threshold in Celsius
	CriticalTempThreshold float64
	// WarningTempThreshold is the warning temperature threshold in Celsius
	WarningTempThreshold float64
	// EmergencyResponseDelay is the delay before sending emergency notifications
	EmergencyResponseDelay time.Duration
}

// Option represents a configuration option for the sensor monitoring service.
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

type hwmonPathOption struct {
	path string
}

func (o *hwmonPathOption) apply(c *Config) {
	c.HwmonPath = o.path
}

// WithHwmonPath sets the path to hwmon devices in sysfs.
func WithHwmonPath(path string) Option {
	return &hwmonPathOption{path: path}
}

type gpioChipPathOption struct {
	path string
}

func (o *gpioChipPathOption) apply(c *Config) {
	c.GPIOChipPath = o.path
}

// WithGPIOChipPath sets the path to the GPIO chip device.
func WithGPIOChipPath(path string) Option {
	return &gpioChipPathOption{path: path}
}

type monitoringIntervalOption struct {
	interval time.Duration
}

func (o *monitoringIntervalOption) apply(c *Config) {
	c.MonitoringInterval = o.interval
}

// WithMonitoringInterval sets the interval between sensor readings.
func WithMonitoringInterval(interval time.Duration) Option {
	return &monitoringIntervalOption{interval: interval}
}

type thresholdCheckIntervalOption struct {
	interval time.Duration
}

func (o *thresholdCheckIntervalOption) apply(c *Config) {
	c.ThresholdCheckInterval = o.interval
}

// WithThresholdCheckInterval sets the interval between threshold checks.
func WithThresholdCheckInterval(interval time.Duration) Option {
	return &thresholdCheckIntervalOption{interval: interval}
}

type sensorTimeoutOption struct {
	timeout time.Duration
}

func (o *sensorTimeoutOption) apply(c *Config) {
	c.SensorTimeout = o.timeout
}

// WithSensorTimeout sets the maximum time to wait for sensor operations.
func WithSensorTimeout(timeout time.Duration) Option {
	return &sensorTimeoutOption{timeout: timeout}
}

type enableHwmonSensorsOption struct {
	enable bool
}

func (o *enableHwmonSensorsOption) apply(c *Config) {
	c.EnableHwmonSensors = o.enable
}

// WithEnableHwmonSensors enables or disables hardware monitoring sensors.
func WithEnableHwmonSensors(enable bool) Option {
	return &enableHwmonSensorsOption{enable: enable}
}

type enableGPIOSensorsOption struct {
	enable bool
}

func (o *enableGPIOSensorsOption) apply(c *Config) {
	c.EnableGPIOSensors = o.enable
}

// WithEnableGPIOSensors enables or disables GPIO-based sensors.
func WithEnableGPIOSensors(enable bool) Option {
	return &enableGPIOSensorsOption{enable: enable}
}

type enableMetricsOption struct {
	enable bool
}

func (o *enableMetricsOption) apply(c *Config) {
	c.EnableMetrics = o.enable
}

// WithEnableMetrics enables or disables metrics collection.
func WithEnableMetrics(enable bool) Option {
	return &enableMetricsOption{enable: enable}
}

type enableTracingOption struct {
	enable bool
}

func (o *enableTracingOption) apply(c *Config) {
	c.EnableTracing = o.enable
}

// WithEnableTracing enables or disables distributed tracing.
func WithEnableTracing(enable bool) Option {
	return &enableTracingOption{enable: enable}
}

type enableThresholdMonitoringOption struct {
	enable bool
}

func (o *enableThresholdMonitoringOption) apply(c *Config) {
	c.EnableThresholdMonitoring = o.enable
}

// WithEnableThresholdMonitoring enables or disables automatic threshold monitoring.
func WithEnableThresholdMonitoring(enable bool) Option {
	return &enableThresholdMonitoringOption{enable: enable}
}

type broadcastSensorReadingsOption struct {
	enable bool
}

func (o *broadcastSensorReadingsOption) apply(c *Config) {
	c.BroadcastSensorReadings = o.enable
}

// WithBroadcastSensorReadings enables or disables broadcasting sensor readings via NATS.
func WithBroadcastSensorReadings(enable bool) Option {
	return &broadcastSensorReadingsOption{enable: enable}
}

type persistSensorDataOption struct {
	enable bool
}

func (o *persistSensorDataOption) apply(c *Config) {
	c.PersistSensorData = o.enable
}

// WithPersistSensorData enables or disables persisting sensor data to JetStream.
func WithPersistSensorData(enable bool) Option {
	return &persistSensorDataOption{enable: enable}
}

type streamNameOption struct {
	name string
}

func (o *streamNameOption) apply(c *Config) {
	c.StreamName = o.name
}

// WithStreamName sets the JetStream stream name for sensor data persistence.
func WithStreamName(name string) Option {
	return &streamNameOption{name: name}
}

type streamSubjectsOption struct {
	subjects []string
}

func (o *streamSubjectsOption) apply(c *Config) {
	c.StreamSubjects = o.subjects
}

// WithStreamSubjects sets the subjects the JetStream stream will listen on.
func WithStreamSubjects(subjects ...string) Option {
	// sanitize
	set := make(map[string]struct{}, len(subjects))
	out := make([]string, 0, len(subjects))
	for _, s := range subjects {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := set[s]; ok {
			continue
		}
		set[s] = struct{}{}
		out = append(out, s)
	}
	return &streamSubjectsOption{subjects: out}
}

type streamRetentionOption struct {
	retention time.Duration
}

func (o *streamRetentionOption) apply(c *Config) {
	c.StreamRetention = o.retention
}

// WithStreamRetention sets how long to retain sensor data in JetStream.
func WithStreamRetention(retention time.Duration) Option {
	return &streamRetentionOption{retention: retention}
}

type maxConcurrentReadsOption struct {
	max int
}

func (o *maxConcurrentReadsOption) apply(c *Config) {
	c.MaxConcurrentReads = o.max
}

// WithMaxConcurrentReads sets the maximum number of concurrent sensor reads.
func WithMaxConcurrentReads(max int) Option {
	return &maxConcurrentReadsOption{max: max}
}

type sensorDiscoveryTimeoutOption struct {
	timeout time.Duration
}

func (o *sensorDiscoveryTimeoutOption) apply(c *Config) {
	c.SensorDiscoveryTimeout = o.timeout
}

// WithSensorDiscoveryTimeout sets the timeout for sensor discovery operations.
func WithSensorDiscoveryTimeout(timeout time.Duration) Option {
	return &sensorDiscoveryTimeoutOption{timeout: timeout}
}

type enableThermalIntegrationOption struct {
	enable bool
}

func (o *enableThermalIntegrationOption) apply(c *Config) {
	c.EnableThermalIntegration = o.enable
}

// WithEnableThermalIntegration enables or disables thermal management integration.
func WithEnableThermalIntegration(enable bool) Option {
	return &enableThermalIntegrationOption{enable: enable}
}

type thermalMgrEndpointOption struct {
	endpoint string
}

func (o *thermalMgrEndpointOption) apply(c *Config) {
	c.ThermalMgrEndpoint = o.endpoint
}

// WithThermalMgrEndpoint sets the thermal manager service endpoint.
func WithThermalMgrEndpoint(endpoint string) Option {
	return &thermalMgrEndpointOption{endpoint: endpoint}
}

type temperatureUpdateIntervalOption struct {
	interval time.Duration
}

func (o *temperatureUpdateIntervalOption) apply(c *Config) {
	c.TemperatureUpdateInterval = o.interval
}

// WithTemperatureUpdateInterval sets the interval for sending temperature updates to thermal manager.
func WithTemperatureUpdateInterval(interval time.Duration) Option {
	return &temperatureUpdateIntervalOption{interval: interval}
}

type enableThermalAlertsOption struct {
	enable bool
}

func (o *enableThermalAlertsOption) apply(c *Config) {
	c.EnableThermalAlerts = o.enable
}

// WithEnableThermalAlerts enables or disables thermal threshold alerting.
func WithEnableThermalAlerts(enable bool) Option {
	return &enableThermalAlertsOption{enable: enable}
}

type thermalThresholdsOption struct {
	warning  float64
	critical float64
}

func (o *thermalThresholdsOption) apply(c *Config) {
	c.WarningTempThreshold = o.warning
	c.CriticalTempThreshold = o.critical
}

// WithThermalThresholds sets the warning and critical temperature thresholds in Celsius.
func WithThermalThresholds(warning, critical float64) Option {
	return &thermalThresholdsOption{warning: warning, critical: critical}
}

type emergencyResponseDelayOption struct {
	delay time.Duration
}

func (o *emergencyResponseDelayOption) apply(c *Config) {
	c.EmergencyResponseDelay = o.delay
}

// WithEmergencyResponseDelay sets the delay before sending emergency thermal notifications.
func WithEmergencyResponseDelay(delay time.Duration) Option {
	return &emergencyResponseDelayOption{delay: delay}
}

// NewConfig creates a new sensor monitoring configuration with default values.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		ServiceName:               DefaultServiceName,
		ServiceDescription:        DefaultServiceDescription,
		ServiceVersion:            DefaultServiceVersion,
		HwmonPath:                 DefaultHwmonPath,
		GPIOChipPath:              DefaultGPIOChipPath,
		MonitoringInterval:        DefaultMonitoringInterval,
		ThresholdCheckInterval:    DefaultThresholdCheckInterval,
		SensorTimeout:             DefaultSensorTimeout,
		EnableHwmonSensors:        true,
		EnableGPIOSensors:         false,
		EnableMetrics:             true,
		EnableTracing:             true,
		EnableThresholdMonitoring: true,
		BroadcastSensorReadings:   false,
		PersistSensorData:         false,
		StreamName:                "SENSORMON",
		StreamSubjects:            []string{"sensormon.data.>", "sensormon.events.>"},
		StreamRetention:           24 * time.Hour,
		MaxConcurrentReads:        10,
		SensorDiscoveryTimeout:    10 * time.Second,
		EnableThermalIntegration:  false,
		ThermalMgrEndpoint:        "thermalmgr",
		TemperatureUpdateInterval: 5 * time.Second,
		EnableThermalAlerts:       false,
		CriticalTempThreshold:     85.0,
		WarningTempThreshold:      75.0,
		EmergencyResponseDelay:    5 * time.Second,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if c.ServiceVersion == "" {
		return fmt.Errorf("service version cannot be empty")
	}

	if c.HwmonPath == "" {
		return fmt.Errorf("hwmon path cannot be empty")
	}

	if c.GPIOChipPath == "" && c.EnableGPIOSensors {
		return fmt.Errorf("GPIO chip path cannot be empty when GPIO sensors are enabled")
	}

	if c.MonitoringInterval <= 0 {
		return fmt.Errorf("monitoring interval must be positive")
	}

	if c.ThresholdCheckInterval <= 0 {
		return fmt.Errorf("threshold check interval must be positive")
	}

	if c.SensorTimeout <= 0 {
		return fmt.Errorf("sensor timeout must be positive")
	}

	if !c.EnableHwmonSensors && !c.EnableGPIOSensors {
		return fmt.Errorf("at least one sensor type must be enabled")
	}

	if c.PersistSensorData {
		if c.StreamName == "" {
			return fmt.Errorf("stream name cannot be empty when sensor data persistence is enabled")
		}

		if len(c.StreamSubjects) == 0 {
			return fmt.Errorf("at least one stream subject must be configured when sensor data persistence is enabled")
		}

		for _, s := range c.StreamSubjects {
			if len(s) == 0 {
				return fmt.Errorf("stream subject cannot be empty")
			}
		}

		if c.StreamRetention < 0 {
			return fmt.Errorf("stream retention cannot be negative")
		}
	}

	if c.MaxConcurrentReads <= 0 {
		return fmt.Errorf("maximum concurrent reads must be positive")
	}

	if c.SensorDiscoveryTimeout <= 0 {
		return fmt.Errorf("sensor discovery timeout must be positive")
	}

	if c.EnableThermalIntegration {
		if c.ThermalMgrEndpoint == "" {
			return fmt.Errorf("thermal manager endpoint cannot be empty when thermal integration is enabled")
		}

		if c.TemperatureUpdateInterval <= 0 {
			return fmt.Errorf("temperature update interval must be positive")
		}

		if c.CriticalTempThreshold <= c.WarningTempThreshold {
			return fmt.Errorf("critical temperature threshold must be greater than warning threshold")
		}

		if c.EmergencyResponseDelay <= 0 {
			return fmt.Errorf("emergency response delay must be positive")
		}
	}

	return nil
}
