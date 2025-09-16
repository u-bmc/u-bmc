// SPDX-License-Identifier: BSD-3-Clause

package thermalmgr

import (
	"fmt"
	"time"
)

// Config holds configuration for the thermal manager service.
type Config struct {
	ServiceName        string
	ServiceVersion     string
	ServiceDescription string

	// Thermal management settings
	EnableThermalControl   bool
	ThermalControlInterval time.Duration
	EmergencyCheckInterval time.Duration
	DefaultPIDSampleTime   time.Duration
	MaxThermalZones        int
	MaxCoolingDevices      int

	// Hardware paths
	HwmonPath            string
	EnableHwmonDiscovery bool

	// Temperature thresholds (Celsius)
	DefaultWarningTemp    float64
	DefaultCriticalTemp   float64
	EmergencyShutdownTemp float64

	// PID defaults
	DefaultPIDKp     float64
	DefaultPIDKi     float64
	DefaultPIDKd     float64
	DefaultOutputMin float64
	DefaultOutputMax float64

	// Communication settings
	SensormonEndpoint       string
	PowermgrEndpoint        string
	EnableSensorIntegration bool
	EnablePowerIntegration  bool

	// Persistence settings
	PersistThermalData bool
	StreamName         string
	StreamSubjects     []string
	StreamRetention    time.Duration

	// Safety settings
	EnableEmergencyResponse bool
	EmergencyResponseDelay  time.Duration
	FailsafeCoolingLevel    float64
}

// Option configures the thermal manager service.
type Option func(*Config)

// NewConfig creates a new Config with default values and applies the provided options.
func NewConfig(opts ...Option) *Config {
	config := &Config{
		ServiceName:        "thermalmgr",
		ServiceVersion:     "1.0.0",
		ServiceDescription: "Thermal Management Service",

		EnableThermalControl:   true,
		ThermalControlInterval: time.Second,
		EmergencyCheckInterval: 500 * time.Millisecond,
		DefaultPIDSampleTime:   time.Second,
		MaxThermalZones:        16,
		MaxCoolingDevices:      64,

		HwmonPath:            "/sys/class/hwmon",
		EnableHwmonDiscovery: true,

		DefaultWarningTemp:    75.0,
		DefaultCriticalTemp:   85.0,
		EmergencyShutdownTemp: 95.0,

		DefaultPIDKp:     1.0,
		DefaultPIDKi:     0.1,
		DefaultPIDKd:     0.05,
		DefaultOutputMin: 0.0,
		DefaultOutputMax: 100.0,

		SensormonEndpoint:       "sensormon",
		PowermgrEndpoint:        "powermgr",
		EnableSensorIntegration: true,
		EnablePowerIntegration:  true,

		PersistThermalData: false,
		StreamName:         "THERMALMGR",
		StreamSubjects:     []string{"thermalmgr.>"},
		StreamRetention:    24 * time.Hour,

		EnableEmergencyResponse: true,
		EmergencyResponseDelay:  2 * time.Second,
		FailsafeCoolingLevel:    100.0,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if c.ThermalControlInterval <= 0 {
		return fmt.Errorf("thermal control interval must be positive")
	}

	if c.EmergencyCheckInterval <= 0 {
		return fmt.Errorf("emergency check interval must be positive")
	}

	if c.DefaultPIDSampleTime <= 0 {
		return fmt.Errorf("PID sample time must be positive")
	}

	if c.MaxThermalZones <= 0 {
		return fmt.Errorf("max thermal zones must be positive")
	}

	if c.MaxCoolingDevices <= 0 {
		return fmt.Errorf("max cooling devices must be positive")
	}

	if c.DefaultWarningTemp >= c.DefaultCriticalTemp {
		return fmt.Errorf("warning temperature must be less than critical temperature")
	}

	if c.DefaultCriticalTemp >= c.EmergencyShutdownTemp {
		return fmt.Errorf("critical temperature must be less than emergency shutdown temperature")
	}

	if c.DefaultOutputMin >= c.DefaultOutputMax {
		return fmt.Errorf("output minimum must be less than output maximum")
	}

	if c.FailsafeCoolingLevel < 0 || c.FailsafeCoolingLevel > 100 {
		return fmt.Errorf("failsafe cooling level must be between 0 and 100")
	}

	if c.StreamRetention <= 0 {
		return fmt.Errorf("stream retention must be positive")
	}

	return nil
}

// WithServiceName sets the service name.
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithServiceVersion sets the service version.
func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.ServiceVersion = version
	}
}

// WithServiceDescription sets the service description.
func WithServiceDescription(description string) Option {
	return func(c *Config) {
		c.ServiceDescription = description
	}
}

// WithThermalControlInterval sets the thermal control loop interval.
func WithThermalControlInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.ThermalControlInterval = interval
	}
}

// WithEmergencyCheckInterval sets the emergency check interval.
func WithEmergencyCheckInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.EmergencyCheckInterval = interval
	}
}

// WithHwmonPath sets the hwmon filesystem path.
func WithHwmonPath(path string) Option {
	return func(c *Config) {
		c.HwmonPath = path
	}
}

// WithDefaultPIDConfig sets the default PID configuration.
func WithDefaultPIDConfig(kp, ki, kd float64) Option {
	return func(c *Config) {
		c.DefaultPIDKp = kp
		c.DefaultPIDKi = ki
		c.DefaultPIDKd = kd
	}
}

// WithTemperatureThresholds sets the default temperature thresholds.
func WithTemperatureThresholds(warning, critical, emergency float64) Option {
	return func(c *Config) {
		c.DefaultWarningTemp = warning
		c.DefaultCriticalTemp = critical
		c.EmergencyShutdownTemp = emergency
	}
}

// WithSensormonEndpoint sets the sensormon service endpoint.
func WithSensormonEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.SensormonEndpoint = endpoint
	}
}

// WithPowermgrEndpoint sets the powermgr service endpoint.
func WithPowermgrEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.PowermgrEndpoint = endpoint
	}
}

// WithPersistence enables or disables thermal data persistence.
func WithPersistence(enabled bool, streamName string, retention time.Duration) Option {
	return func(c *Config) {
		c.PersistThermalData = enabled
		c.StreamName = streamName
		c.StreamRetention = retention
	}
}

// WithEmergencyResponse configures emergency response settings.
func WithEmergencyResponse(enabled bool, delay time.Duration, failsafeLevel float64) Option {
	return func(c *Config) {
		c.EnableEmergencyResponse = enabled
		c.EmergencyResponseDelay = delay
		c.FailsafeCoolingLevel = failsafeLevel
	}
}

// WithDiscovery enables or disables hardware discovery.
func WithDiscovery(enableHwmon bool) Option {
	return func(c *Config) {
		c.EnableHwmonDiscovery = enableHwmon
	}
}

// WithIntegration configures service integration settings.
func WithIntegration(enableSensor, enablePower bool) Option {
	return func(c *Config) {
		c.EnableSensorIntegration = enableSensor
		c.EnablePowerIntegration = enablePower
	}
}
