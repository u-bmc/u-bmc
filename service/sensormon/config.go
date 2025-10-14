// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultServiceName               = "sensormon"
	DefaultServiceDescription        = "Sensor monitoring service for BMC systems"
	DefaultServiceVersion            = "1.0.0"
	DefaultHwmonPath                 = "/sys/class/hwmon"
	DefaultGPIOChipPath              = "/dev/gpiochip0"
	DefaultMonitoringInterval        = 1 * time.Second
	DefaultThresholdCheckInterval    = 5 * time.Second
	DefaultSensorTimeout             = 5 * time.Second
	DefaultMaxConcurrentReads        = 10
	DefaultSensorDiscoveryTimeout    = 10 * time.Second
	DefaultTemperatureUpdateInterval = 5 * time.Second
	DefaultEmergencyResponseDelay    = 5 * time.Second
	DefaultCriticalTempThreshold     = 85.0
	DefaultWarningTempThreshold      = 75.0
)

type config struct {
	serviceName               string
	serviceDescription        string
	serviceVersion            string
	hwmonPath                 string
	gpioChipPath              string
	monitoringInterval        time.Duration
	thresholdCheckInterval    time.Duration
	sensorTimeout             time.Duration
	enableHwmonSensors        bool
	enableGPIOSensors         bool
	enableMetrics             bool
	enableTracing             bool
	enableThresholdMonitoring bool
	broadcastSensorReadings   bool
	persistSensorData         bool
	streamName                string
	streamSubjects            []string
	streamRetention           time.Duration
	maxConcurrentReads        int
	sensorDiscoveryTimeout    time.Duration
	enableThermalIntegration  bool
	thermalMgrEndpoint        string
	temperatureUpdateInterval time.Duration
	enableThermalAlerts       bool
	criticalTempThreshold     float64
	warningTempThreshold      float64
	emergencyResponseDelay    time.Duration
}

type Option interface {
	apply(*config)
}

type serviceNameOption struct {
	name string
}

func (o *serviceNameOption) apply(c *config) {
	c.serviceName = o.name
}

func WithServiceName(name string) Option {
	return &serviceNameOption{name: name}
}

type serviceDescriptionOption struct {
	description string
}

func (o *serviceDescriptionOption) apply(c *config) {
	c.serviceDescription = o.description
}

func WithServiceDescription(description string) Option {
	return &serviceDescriptionOption{description: description}
}

type serviceVersionOption struct {
	version string
}

func (o *serviceVersionOption) apply(c *config) {
	c.serviceVersion = o.version
}

func WithServiceVersion(version string) Option {
	return &serviceVersionOption{version: version}
}

type hwmonPathOption struct {
	path string
}

func (o *hwmonPathOption) apply(c *config) {
	c.hwmonPath = o.path
}

func WithHwmonPath(path string) Option {
	return &hwmonPathOption{path: path}
}

type gpioChipPathOption struct {
	path string
}

func (o *gpioChipPathOption) apply(c *config) {
	c.gpioChipPath = o.path
}

func WithGPIOChipPath(path string) Option {
	return &gpioChipPathOption{path: path}
}

type monitoringIntervalOption struct {
	interval time.Duration
}

func (o *monitoringIntervalOption) apply(c *config) {
	c.monitoringInterval = o.interval
}

func WithMonitoringInterval(interval time.Duration) Option {
	return &monitoringIntervalOption{interval: interval}
}

type thresholdCheckIntervalOption struct {
	interval time.Duration
}

func (o *thresholdCheckIntervalOption) apply(c *config) {
	c.thresholdCheckInterval = o.interval
}

func WithThresholdCheckInterval(interval time.Duration) Option {
	return &thresholdCheckIntervalOption{interval: interval}
}

type sensorTimeoutOption struct {
	timeout time.Duration
}

func (o *sensorTimeoutOption) apply(c *config) {
	c.sensorTimeout = o.timeout
}

func WithSensorTimeout(timeout time.Duration) Option {
	return &sensorTimeoutOption{timeout: timeout}
}

type enableHwmonSensorsOption struct {
	enable bool
}

func (o *enableHwmonSensorsOption) apply(c *config) {
	c.enableHwmonSensors = o.enable
}

func WithHwmonSensors(enable bool) Option {
	return &enableHwmonSensorsOption{enable: enable}
}

func WithoutHwmonSensors() Option {
	return &enableHwmonSensorsOption{enable: false}
}

type enableGPIOSensorsOption struct {
	enable bool
}

func (o *enableGPIOSensorsOption) apply(c *config) {
	c.enableGPIOSensors = o.enable
}

func WithGPIOSensors(enable bool) Option {
	return &enableGPIOSensorsOption{enable: enable}
}

func WithoutGPIOSensors() Option {
	return &enableGPIOSensorsOption{enable: false}
}

type enableMetricsOption struct {
	enable bool
}

func (o *enableMetricsOption) apply(c *config) {
	c.enableMetrics = o.enable
}

func WithMetrics(enable bool) Option {
	return &enableMetricsOption{enable: enable}
}

func WithoutMetrics() Option {
	return &enableMetricsOption{enable: false}
}

type enableTracingOption struct {
	enable bool
}

func (o *enableTracingOption) apply(c *config) {
	c.enableTracing = o.enable
}

func WithTracing(enable bool) Option {
	return &enableTracingOption{enable: enable}
}

func WithoutTracing() Option {
	return &enableTracingOption{enable: false}
}

type enableThresholdMonitoringOption struct {
	enable bool
}

func (o *enableThresholdMonitoringOption) apply(c *config) {
	c.enableThresholdMonitoring = o.enable
}

func WithThresholdMonitoring(enable bool) Option {
	return &enableThresholdMonitoringOption{enable: enable}
}

func WithoutThresholdMonitoring() Option {
	return &enableThresholdMonitoringOption{enable: false}
}

type broadcastSensorReadingsOption struct {
	enable bool
}

func (o *broadcastSensorReadingsOption) apply(c *config) {
	c.broadcastSensorReadings = o.enable
}

func WithBroadcastSensorReadings(enable bool) Option {
	return &broadcastSensorReadingsOption{enable: enable}
}

func WithoutBroadcastSensorReadings() Option {
	return &broadcastSensorReadingsOption{enable: false}
}

type persistSensorDataOption struct {
	enable bool
}

func (o *persistSensorDataOption) apply(c *config) {
	c.persistSensorData = o.enable
}

func WithPersistSensorData(enable bool) Option {
	return &persistSensorDataOption{enable: enable}
}

func WithoutPersistSensorData() Option {
	return &persistSensorDataOption{enable: false}
}

type streamNameOption struct {
	name string
}

func (o *streamNameOption) apply(c *config) {
	c.streamName = o.name
}

func WithStreamName(name string) Option {
	return &streamNameOption{name: name}
}

type streamSubjectsOption struct {
	subjects []string
}

func (o *streamSubjectsOption) apply(c *config) {
	c.streamSubjects = o.subjects
}

func WithStreamSubjects(subjects ...string) Option {
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

func (o *streamRetentionOption) apply(c *config) {
	c.streamRetention = o.retention
}

func WithStreamRetention(retention time.Duration) Option {
	return &streamRetentionOption{retention: retention}
}

type maxConcurrentReadsOption struct {
	maxVal int
}

func (o *maxConcurrentReadsOption) apply(c *config) {
	c.maxConcurrentReads = o.maxVal
}

func WithMaxConcurrentReads(maxVal int) Option {
	return &maxConcurrentReadsOption{maxVal: maxVal}
}

type sensorDiscoveryTimeoutOption struct {
	timeout time.Duration
}

func (o *sensorDiscoveryTimeoutOption) apply(c *config) {
	c.sensorDiscoveryTimeout = o.timeout
}

func WithSensorDiscoveryTimeout(timeout time.Duration) Option {
	return &sensorDiscoveryTimeoutOption{timeout: timeout}
}

type enableThermalIntegrationOption struct {
	enable bool
}

func (o *enableThermalIntegrationOption) apply(c *config) {
	c.enableThermalIntegration = o.enable
}

func WithThermalIntegration(enable bool) Option {
	return &enableThermalIntegrationOption{enable: enable}
}

func WithoutThermalIntegration() Option {
	return &enableThermalIntegrationOption{enable: false}
}

type thermalMgrEndpointOption struct {
	endpoint string
}

func (o *thermalMgrEndpointOption) apply(c *config) {
	c.thermalMgrEndpoint = o.endpoint
}

func WithThermalMgrEndpoint(endpoint string) Option {
	return &thermalMgrEndpointOption{endpoint: endpoint}
}

type temperatureUpdateIntervalOption struct {
	interval time.Duration
}

func (o *temperatureUpdateIntervalOption) apply(c *config) {
	c.temperatureUpdateInterval = o.interval
}

func WithTemperatureUpdateInterval(interval time.Duration) Option {
	return &temperatureUpdateIntervalOption{interval: interval}
}

type enableThermalAlertsOption struct {
	enable bool
}

func (o *enableThermalAlertsOption) apply(c *config) {
	c.enableThermalAlerts = o.enable
}

func WithThermalAlerts(enable bool) Option {
	return &enableThermalAlertsOption{enable: enable}
}

func WithoutThermalAlerts() Option {
	return &enableThermalAlertsOption{enable: false}
}

type thermalThresholdsOption struct {
	warning  float64
	critical float64
}

func (o *thermalThresholdsOption) apply(c *config) {
	c.warningTempThreshold = o.warning
	c.criticalTempThreshold = o.critical
}

func WithThermalThresholds(warning, critical float64) Option {
	return &thermalThresholdsOption{warning: warning, critical: critical}
}

type emergencyResponseDelayOption struct {
	delay time.Duration
}

func (o *emergencyResponseDelayOption) apply(c *config) {
	c.emergencyResponseDelay = o.delay
}

func WithEmergencyResponseDelay(delay time.Duration) Option {
	return &emergencyResponseDelayOption{delay: delay}
}

func (c *config) Validate() error {
	if c.serviceName == "" {
		return fmt.Errorf("%w: service name cannot be empty", ErrInvalidConfiguration)
	}

	if c.serviceVersion == "" {
		return fmt.Errorf("%w: service version cannot be empty", ErrInvalidConfiguration)
	}

	if c.hwmonPath == "" {
		return fmt.Errorf("%w: hwmon path cannot be empty", ErrInvalidConfiguration)
	}

	if c.gpioChipPath == "" && c.enableGPIOSensors {
		return fmt.Errorf("%w: GPIO chip path cannot be empty when GPIO sensors are enabled", ErrInvalidConfiguration)
	}

	if c.monitoringInterval <= 0 {
		return fmt.Errorf("%w: monitoring interval must be positive", ErrInvalidConfiguration)
	}

	if c.thresholdCheckInterval <= 0 {
		return fmt.Errorf("%w: threshold check interval must be positive", ErrInvalidConfiguration)
	}

	if c.sensorTimeout <= 0 {
		return fmt.Errorf("%w: sensor timeout must be positive", ErrInvalidConfiguration)
	}

	if !c.enableHwmonSensors && !c.enableGPIOSensors {
		return fmt.Errorf("%w: at least one sensor type must be enabled", ErrInvalidConfiguration)
	}

	if c.persistSensorData {
		if c.streamName == "" {
			return fmt.Errorf("%w: stream name cannot be empty when sensor data persistence is enabled", ErrInvalidConfiguration)
		}

		if len(c.streamSubjects) == 0 {
			return fmt.Errorf("%w: at least one stream subject must be configured when sensor data persistence is enabled", ErrInvalidConfiguration)
		}

		for _, s := range c.streamSubjects {
			if len(s) == 0 {
				return fmt.Errorf("%w: stream subject cannot be empty", ErrInvalidConfiguration)
			}
		}

		if c.streamRetention < 0 {
			return fmt.Errorf("%w: stream retention cannot be negative", ErrInvalidConfiguration)
		}
	}

	if c.maxConcurrentReads <= 0 {
		return fmt.Errorf("%w: maximum concurrent reads must be positive", ErrInvalidConfiguration)
	}

	if c.sensorDiscoveryTimeout <= 0 {
		return fmt.Errorf("%w: sensor discovery timeout must be positive", ErrInvalidConfiguration)
	}

	if c.enableThermalIntegration {
		if c.thermalMgrEndpoint == "" {
			return fmt.Errorf("%w: thermal manager endpoint cannot be empty when thermal integration is enabled", ErrInvalidConfiguration)
		}

		if c.temperatureUpdateInterval <= 0 {
			return fmt.Errorf("%w: temperature update interval must be positive", ErrInvalidConfiguration)
		}

		if c.criticalTempThreshold <= c.warningTempThreshold {
			return fmt.Errorf("%w: critical temperature threshold must be greater than warning threshold", ErrInvalidConfiguration)
		}

		if c.emergencyResponseDelay <= 0 {
			return fmt.Errorf("%w: emergency response delay must be positive", ErrInvalidConfiguration)
		}
	}

	return nil
}
