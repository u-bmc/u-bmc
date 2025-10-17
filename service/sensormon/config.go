// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"fmt"
	"strings"
	"time"

	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
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

// SensorBackendType represents the backend type for sensor reading.
type SensorBackendType string

const (
	BackendTypeHwmon SensorBackendType = "hwmon"
	BackendTypeGPIO  SensorBackendType = "gpio"
	BackendTypeMock  SensorBackendType = "mock"
)

// MockSensorBehavior defines how mock sensors behave.
type MockSensorBehavior string

const (
	MockBehaviorFixed     MockSensorBehavior = "fixed"     // Return fixed value
	MockBehaviorRandomize MockSensorBehavior = "randomize" // Return randomized values around base
	MockBehaviorSine      MockSensorBehavior = "sine"      // Sine wave pattern
	MockBehaviorStep      MockSensorBehavior = "step"      // Step increases/decreases
)

// Threshold represents sensor threshold configuration.
type Threshold struct {
	Warning  *float64 // warning threshold value
	Critical *float64 // critical threshold value
}

// Location represents sensor physical location.
type Location struct {
	Zone        string            // Thermal zone (e.g., "cpu", "memory", "psu")
	Position    string            // Physical position (e.g., "inlet", "outlet", "center")
	Component   string            // Component name (e.g., "CPU0", "DIMM_A1")
	Coordinates map[string]string // Additional location data
}

// HwmonSensorConfig represents hwmon-specific sensor configuration.
type HwmonSensorConfig struct {
	DevicePath     string   // hwmon device path (e.g., "/sys/class/hwmon/hwmon0")
	AttributeName  string   // attribute name (e.g., "temp1_input")
	LabelAttribute string   // label attribute (e.g., "temp1_label")
	ScaleFactor    int      // scaling factor for raw values
	MatchPattern   string   // regex pattern to match device
	RequiredFiles  []string // files that must exist for detection
}

// GPIOSensorConfig represents GPIO-specific sensor configuration.
type GPIOSensorConfig struct {
	ChipPath     string            // GPIO chip path (e.g., "/dev/gpiochip0")
	Line         int               // GPIO line number
	ActiveState  string            // "high" or "low"
	PullResistor string            // "up", "down", or "none"
	DebounceTime time.Duration     // debounce time for discrete sensors
	ValueMapping map[string]string // map GPIO values to sensor states
}

// MockSensorConfig represents mock sensor configuration for testing.
type MockSensorConfig struct {
	Behavior      MockSensorBehavior // how the sensor behaves
	BaseValue     float64            // base value for calculations
	Variance      float64            // variance for randomization
	Period        time.Duration      // period for periodic behaviors
	StepSize      float64            // step size for step behavior
	MinValue      float64            // minimum value
	MaxValue      float64            // maximum value
	FailureRate   float64            // probability of read failure (0.0-1.0)
	FailurePeriod time.Duration      // duration of failure periods
}

// SensorDefinition represents a complete sensor configuration.
type SensorDefinition struct {
	ID               string                 // unique sensor identifier
	Name             string                 // human-readable name
	Description      string                 // sensor description
	Context          v1alpha1.SensorContext // sensor type/context
	Unit             v1alpha1.SensorUnit    // measurement unit
	Backend          SensorBackendType      // backend type
	Location         Location               // physical location
	UpperThresholds  *Threshold             // upper thresholds
	LowerThresholds  *Threshold             // lower thresholds
	Enabled          bool                   // whether sensor is enabled
	ReadOnly         bool                   // whether sensor is read-only
	CustomAttributes map[string]string      // additional attributes

	// Backend-specific configurations
	HwmonConfig *HwmonSensorConfig // hwmon configuration
	GPIOConfig  *GPIOSensorConfig  // GPIO configuration
	MockConfig  *MockSensorConfig  // mock configuration
}

// SensorEventCallback is called when sensor events occur.
type SensorEventCallback func(sensorID string, event SensorEvent, data interface{}) error

// SensorEvent represents different sensor events.
type SensorEvent string

const (
	EventSensorRead         SensorEvent = "sensor_read"
	EventThresholdWarning   SensorEvent = "threshold_warning"
	EventThresholdCritical  SensorEvent = "threshold_critical"
	EventThresholdNormal    SensorEvent = "threshold_normal"
	EventSensorError        SensorEvent = "sensor_error"
	EventSensorDiscovered   SensorEvent = "sensor_discovered"
	EventSensorStatusChange SensorEvent = "sensor_status_change"
)

// SensorCallbacks holds callback functions for sensor events.
type SensorCallbacks struct {
	OnSensorRead        SensorEventCallback // called after each sensor read
	OnThresholdWarning  SensorEventCallback // called when warning threshold exceeded
	OnThresholdCritical SensorEventCallback // called when critical threshold exceeded
	OnThresholdNormal   SensorEventCallback // called when sensor returns to normal
	OnSensorError       SensorEventCallback // called when sensor read fails
	OnSensorDiscovered  SensorEventCallback // called when new sensor is discovered
	OnStatusChange      SensorEventCallback // called when sensor status changes
}

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
	enableMockSensors         bool
	enableThresholdMonitoring bool
	enableAutoDiscovery       bool
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

	// Enhanced configuration
	sensorDefinitions      []SensorDefinition
	callbacks              SensorCallbacks
	mockFailureSimulation  bool
	mockFailureRate        float64
	customBackendFactories map[string]SensorBackendFactory
}

// SensorBackendFactory creates sensor backend instances.
type SensorBackendFactory func(config interface{}) (SensorBackend, error)

// SensorBackend interface for different sensor implementations.
type SensorBackend interface {
	ReadValue() (interface{}, error)
	GetStatus() (v1alpha1.SensorStatus, error)
	Configure(config interface{}) error
	Close() error
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

type enableMockSensorsOption struct {
	enable bool
}

func (o *enableMockSensorsOption) apply(c *config) {
	c.enableMockSensors = o.enable
}

func WithMockSensors(enable bool) Option {
	return &enableMockSensorsOption{enable: enable}
}

func WithoutMockSensors() Option {
	return &enableMockSensorsOption{enable: false}
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

type enableAutoDiscoveryOption struct {
	enable bool
}

func (o *enableAutoDiscoveryOption) apply(c *config) {
	c.enableAutoDiscovery = o.enable
}

func WithAutoDiscovery(enable bool) Option {
	return &enableAutoDiscoveryOption{enable: enable}
}

func WithoutAutoDiscovery() Option {
	return &enableAutoDiscoveryOption{enable: false}
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
	max int
}

func (o *maxConcurrentReadsOption) apply(c *config) {
	c.maxConcurrentReads = o.max
}

func WithMaxConcurrentReads(max int) Option {
	return &maxConcurrentReadsOption{max: max}
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

type criticalTempThresholdOption struct {
	threshold float64
}

func (o *criticalTempThresholdOption) apply(c *config) {
	c.criticalTempThreshold = o.threshold
}

func WithCriticalTempThreshold(threshold float64) Option {
	return &criticalTempThresholdOption{threshold: threshold}
}

type warningTempThresholdOption struct {
	threshold float64
}

func (o *warningTempThresholdOption) apply(c *config) {
	c.warningTempThreshold = o.threshold
}

func WithWarningTempThreshold(threshold float64) Option {
	return &warningTempThresholdOption{threshold: threshold}
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

// Enhanced configuration options
type sensorDefinitionsOption struct {
	definitions []SensorDefinition
}

func (o *sensorDefinitionsOption) apply(c *config) {
	c.sensorDefinitions = o.definitions
}

func WithSensorDefinitions(definitions ...SensorDefinition) Option {
	return &sensorDefinitionsOption{definitions: definitions}
}

type callbacksOption struct {
	callbacks SensorCallbacks
}

func (o *callbacksOption) apply(c *config) {
	c.callbacks = o.callbacks
}

func WithCallbacks(callbacks SensorCallbacks) Option {
	return &callbacksOption{callbacks: callbacks}
}

type mockFailureSimulationOption struct {
	enable bool
	rate   float64
}

func (o *mockFailureSimulationOption) apply(c *config) {
	c.mockFailureSimulation = o.enable
	c.mockFailureRate = o.rate
}

func WithMockFailureSimulation(enable bool, rate float64) Option {
	return &mockFailureSimulationOption{enable: enable, rate: rate}
}

func WithoutMockFailureSimulation() Option {
	return &mockFailureSimulationOption{enable: false, rate: 0.0}
}

type customBackendFactoriesOption struct {
	factories map[string]SensorBackendFactory
}

func (o *customBackendFactoriesOption) apply(c *config) {
	if c.customBackendFactories == nil {
		c.customBackendFactories = make(map[string]SensorBackendFactory)
	}
	for name, factory := range o.factories {
		c.customBackendFactories[name] = factory
	}
}

func WithCustomBackendFactories(factories map[string]SensorBackendFactory) Option {
	return &customBackendFactoriesOption{factories: factories}
}

func WithCustomBackendFactory(name string, factory SensorBackendFactory) Option {
	return &customBackendFactoriesOption{factories: map[string]SensorBackendFactory{name: factory}}
}

func (c *config) Validate() error {
	if c.serviceName == "" {
		return fmt.Errorf("%w: service name cannot be empty", ErrInvalidConfiguration)
	}

	if c.serviceVersion == "" {
		return fmt.Errorf("%w: service version cannot be empty", ErrInvalidConfiguration)
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

	if c.maxConcurrentReads <= 0 {
		return fmt.Errorf("%w: max concurrent reads must be positive", ErrInvalidConfiguration)
	}

	if c.sensorDiscoveryTimeout <= 0 {
		return fmt.Errorf("%w: sensor discovery timeout must be positive", ErrInvalidConfiguration)
	}

	if c.enableThermalAlerts {
		if c.warningTempThreshold >= c.criticalTempThreshold {
			return fmt.Errorf("%w: warning temperature threshold must be less than critical threshold", ErrInvalidConfiguration)
		}
	}

	if c.enableThermalIntegration && c.thermalMgrEndpoint == "" {
		return fmt.Errorf("%w: thermal manager endpoint cannot be empty when thermal integration is enabled", ErrInvalidConfiguration)
	}

	if c.temperatureUpdateInterval <= 0 {
		return fmt.Errorf("%w: temperature update interval must be positive", ErrInvalidConfiguration)
	}

	if c.emergencyResponseDelay <= 0 {
		return fmt.Errorf("%w: emergency response delay must be positive", ErrInvalidConfiguration)
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

		if c.streamRetention <= 0 {
			return fmt.Errorf("%w: stream retention must be positive when sensor data persistence is enabled", ErrInvalidConfiguration)
		}
	}

	// Validate sensor definitions
	sensorIDs := make(map[string]bool)
	for i, def := range c.sensorDefinitions {
		if def.ID == "" {
			return fmt.Errorf("%w: sensor definition %d has empty ID", ErrInvalidConfiguration, i)
		}

		if sensorIDs[def.ID] {
			return fmt.Errorf("%w: duplicate sensor ID '%s'", ErrInvalidConfiguration, def.ID)
		}
		sensorIDs[def.ID] = true

		if def.Name == "" {
			return fmt.Errorf("%w: sensor definition '%s' has empty name", ErrInvalidConfiguration, def.ID)
		}

		// Validate backend-specific configurations
		switch def.Backend {
		case BackendTypeHwmon:
			if def.HwmonConfig == nil {
				return fmt.Errorf("%w: sensor '%s' has hwmon backend but no hwmon config", ErrInvalidConfiguration, def.ID)
			}
			if def.HwmonConfig.DevicePath == "" && def.HwmonConfig.MatchPattern == "" {
				return fmt.Errorf("%w: sensor '%s' hwmon config must have either device_path or match_pattern", ErrInvalidConfiguration, def.ID)
			}
			if def.HwmonConfig.AttributeName == "" {
				return fmt.Errorf("%w: sensor '%s' hwmon config must have attribute_name", ErrInvalidConfiguration, def.ID)
			}
		case BackendTypeGPIO:
			if def.GPIOConfig == nil {
				return fmt.Errorf("%w: sensor '%s' has gpio backend but no gpio config", ErrInvalidConfiguration, def.ID)
			}
			if def.GPIOConfig.ChipPath == "" {
				return fmt.Errorf("%w: sensor '%s' gpio config must have chip_path", ErrInvalidConfiguration, def.ID)
			}
			if def.GPIOConfig.Line < 0 {
				return fmt.Errorf("%w: sensor '%s' gpio config must have valid line number", ErrInvalidConfiguration, def.ID)
			}
		case BackendTypeMock:
			if def.MockConfig == nil {
				return fmt.Errorf("%w: sensor '%s' has mock backend but no mock config", ErrInvalidConfiguration, def.ID)
			}
		default:
			// Check if it's a custom backend
			if c.customBackendFactories == nil || c.customBackendFactories[string(def.Backend)] == nil {
				return fmt.Errorf("%w: sensor '%s' has unknown backend type '%s'", ErrInvalidConfiguration, def.ID, def.Backend)
			}
		}

		// Validate thresholds
		if def.UpperThresholds != nil {
			if def.UpperThresholds.Warning != nil && def.UpperThresholds.Critical != nil {
				if *def.UpperThresholds.Warning >= *def.UpperThresholds.Critical {
					return fmt.Errorf("%w: sensor '%s' upper warning threshold must be less than critical threshold", ErrInvalidConfiguration, def.ID)
				}
			}
		}

		if def.LowerThresholds != nil {
			if def.LowerThresholds.Warning != nil && def.LowerThresholds.Critical != nil {
				if *def.LowerThresholds.Warning <= *def.LowerThresholds.Critical {
					return fmt.Errorf("%w: sensor '%s' lower warning threshold must be greater than critical threshold", ErrInvalidConfiguration, def.ID)
				}
			}
		}
	}

	if c.mockFailureSimulation {
		if c.mockFailureRate < 0.0 || c.mockFailureRate > 1.0 {
			return fmt.Errorf("%w: mock failure rate must be between 0.0 and 1.0", ErrInvalidConfiguration)
		}
	}

	return nil
}

// Helper functions for creating sensor definitions
func NewTemperatureSensor(id, name string, backend SensorBackendType) SensorDefinition {
	context := v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE
	unit := v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS
	return SensorDefinition{
		ID:      id,
		Name:    name,
		Context: context,
		Unit:    unit,
		Backend: backend,
		Enabled: true,
	}
}

func NewVoltageSensor(id, name string, backend SensorBackendType) SensorDefinition {
	context := v1alpha1.SensorContext_SENSOR_CONTEXT_VOLTAGE
	unit := v1alpha1.SensorUnit_SENSOR_UNIT_VOLTS
	return SensorDefinition{
		ID:      id,
		Name:    name,
		Context: context,
		Unit:    unit,
		Backend: backend,
		Enabled: true,
	}
}

func NewFanSensor(id, name string, backend SensorBackendType) SensorDefinition {
	context := v1alpha1.SensorContext_SENSOR_CONTEXT_TACH
	unit := v1alpha1.SensorUnit_SENSOR_UNIT_RPM
	return SensorDefinition{
		ID:      id,
		Name:    name,
		Context: context,
		Unit:    unit,
		Backend: backend,
		Enabled: true,
	}
}

func NewPowerSensor(id, name string, backend SensorBackendType) SensorDefinition {
	context := v1alpha1.SensorContext_SENSOR_CONTEXT_POWER
	unit := v1alpha1.SensorUnit_SENSOR_UNIT_WATTS
	return SensorDefinition{
		ID:      id,
		Name:    name,
		Context: context,
		Unit:    unit,
		Backend: backend,
		Enabled: true,
	}
}

func NewCurrentSensor(id, name string, backend SensorBackendType) SensorDefinition {
	context := v1alpha1.SensorContext_SENSOR_CONTEXT_CURRENT
	unit := v1alpha1.SensorUnit_SENSOR_UNIT_AMPS
	return SensorDefinition{
		ID:      id,
		Name:    name,
		Context: context,
		Unit:    unit,
		Backend: backend,
		Enabled: true,
	}
}
