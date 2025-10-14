// SPDX-License-Identifier: BSD-3-Clause

package thermalmgr

import (
	"fmt"
	"strings"
	"time"
)

// ThermalEvent represents different thermal management events.
type ThermalEvent string

const (
	EventTemperatureWarning     ThermalEvent = "temperature_warning"
	EventTemperatureCritical    ThermalEvent = "temperature_critical"
	EventTemperatureNormal      ThermalEvent = "temperature_normal"
	EventEmergencyShutdown      ThermalEvent = "emergency_shutdown"
	EventCoolingEngaged         ThermalEvent = "cooling_engaged"
	EventCoolingDisengaged      ThermalEvent = "cooling_disengaged"
	EventThermalZoneCreated     ThermalEvent = "thermal_zone_created"
	EventCoolingDeviceConnected ThermalEvent = "cooling_device_connected"
	EventPIDControllerUpdated   ThermalEvent = "pid_controller_updated"
)

// ThermalEventCallback is called when thermal management events occur.
type ThermalEventCallback func(zoneName string, event ThermalEvent, data interface{}) error

// ThermalCallbacks holds callback functions for thermal management events.
type ThermalCallbacks struct {
	OnTemperatureWarning     ThermalEventCallback `json:"-"` // called when temperature exceeds warning threshold
	OnTemperatureCritical    ThermalEventCallback `json:"-"` // called when temperature exceeds critical threshold
	OnTemperatureNormal      ThermalEventCallback `json:"-"` // called when temperature returns to normal
	OnEmergencyShutdown      ThermalEventCallback `json:"-"` // called during emergency shutdown
	OnCoolingEngaged         ThermalEventCallback `json:"-"` // called when cooling is engaged
	OnCoolingDisengaged      ThermalEventCallback `json:"-"` // called when cooling is disengaged
	OnThermalZoneCreated     ThermalEventCallback `json:"-"` // called when thermal zone is created
	OnCoolingDeviceConnected ThermalEventCallback `json:"-"` // called when cooling device is connected
	OnPIDControllerUpdated   ThermalEventCallback `json:"-"` // called when PID controller is updated
}

// CoolingDeviceType represents the type of cooling device.
type CoolingDeviceType string

const (
	CoolingDeviceTypeFan    CoolingDeviceType = "fan"
	CoolingDeviceTypePump   CoolingDeviceType = "pump"
	CoolingDeviceTypeValve  CoolingDeviceType = "valve"
	CoolingDeviceTypeMock   CoolingDeviceType = "mock"
	CoolingDeviceTypeCustom CoolingDeviceType = "custom"
)

// CoolingDeviceConfig represents configuration for a cooling device.
type CoolingDeviceConfig struct {
	ID               string            `json:"id"`                          // unique device identifier
	Name             string            `json:"name"`                        // human-readable name
	Type             CoolingDeviceType `json:"type"`                        // device type
	Enabled          bool              `json:"enabled"`                     // whether device is enabled
	MinSpeed         float64           `json:"min_speed"`                   // minimum speed/output (0-100%)
	MaxSpeed         float64           `json:"max_speed"`                   // maximum speed/output (0-100%)
	InitialSpeed     float64           `json:"initial_speed"`               // initial speed on startup
	HwmonPath        string            `json:"hwmon_path,omitempty"`        // hwmon path for fan control
	PWMChannel       int               `json:"pwm_channel,omitempty"`       // PWM channel number
	CustomAttributes map[string]string `json:"custom_attributes,omitempty"` // additional attributes
}

// PIDConfig represents PID controller configuration.
type PIDConfig struct {
	Kp         float64       `json:"kp"`          // proportional gain
	Ki         float64       `json:"ki"`          // integral gain
	Kd         float64       `json:"kd"`          // derivative gain
	SampleTime time.Duration `json:"sample_time"` // PID sample time
	OutputMin  float64       `json:"output_min"`  // minimum output value
	OutputMax  float64       `json:"output_max"`  // maximum output value
}

// ThermalZoneConfig represents configuration for a thermal zone.
type ThermalZoneConfig struct {
	ID               string            `json:"id"`                          // unique zone identifier
	Name             string            `json:"name"`                        // human-readable name
	Description      string            `json:"description,omitempty"`       // zone description
	Enabled          bool              `json:"enabled"`                     // whether zone is enabled
	SensorIDs        []string          `json:"sensor_ids"`                  // sensor IDs to monitor
	CoolingDeviceIDs []string          `json:"cooling_device_ids"`          // cooling device IDs to control
	WarningTemp      float64           `json:"warning_temp"`                // warning temperature threshold
	CriticalTemp     float64           `json:"critical_temp"`               // critical temperature threshold
	EmergencyTemp    float64           `json:"emergency_temp,omitempty"`    // emergency temperature threshold
	TargetTemp       float64           `json:"target_temp"`                 // target temperature for PID control
	PIDConfig        *PIDConfig        `json:"pid_config,omitempty"`        // PID controller configuration
	CustomAttributes map[string]string `json:"custom_attributes,omitempty"` // additional attributes
}

const (
	DefaultServiceName            = "thermalmgr"
	DefaultServiceDescription     = "Thermal management service for BMC components"
	DefaultServiceVersion         = "1.0.0"
	DefaultThermalControlInterval = 2 * time.Second
	DefaultEmergencyCheckInterval = 500 * time.Millisecond
	DefaultDefaultPIDSampleTime   = 1 * time.Second
	DefaultMaxThermalZones        = 16
	DefaultMaxCoolingDevices      = 64
	DefaultHwmonPath              = "/sys/class/hwmon"
	DefaultDefaultWarningTemp     = 75.0
	DefaultDefaultCriticalTemp    = 85.0
	DefaultEmergencyShutdownTemp  = 95.0
	DefaultDefaultPIDKp           = 1.0
	DefaultDefaultPIDKi           = 0.1
	DefaultDefaultPIDKd           = 0.05
	DefaultDefaultOutputMin       = 0.0
	DefaultDefaultOutputMax       = 100.0
	DefaultEmergencyResponseDelay = 2 * time.Second
	DefaultFailsafeCoolingLevel   = 100.0
)

type config struct {
	serviceName             string
	serviceDescription      string
	serviceVersion          string
	enableThermalControl    bool
	thermalControlInterval  time.Duration
	emergencyCheckInterval  time.Duration
	defaultPIDSampleTime    time.Duration
	maxThermalZones         int
	maxCoolingDevices       int
	hwmonPath               string
	enableHwmonDiscovery    bool
	defaultWarningTemp      float64
	defaultCriticalTemp     float64
	emergencyShutdownTemp   float64
	defaultPIDKp            float64
	defaultPIDKi            float64
	defaultPIDKd            float64
	defaultOutputMin        float64
	defaultOutputMax        float64
	sensormonEndpoint       string
	powermgrEndpoint        string
	enableSensorIntegration bool
	enablePowerIntegration  bool
	persistThermalData      bool
	streamName              string
	streamSubjects          []string
	streamRetention         time.Duration
	enableEmergencyResponse bool
	emergencyResponseDelay  time.Duration
	failsafeCoolingLevel    float64

	// Enhanced configuration
	thermalZones   []ThermalZoneConfig
	coolingDevices []CoolingDeviceConfig
	callbacks      ThermalCallbacks
	enableMockMode bool
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

type enableThermalControlOption struct {
	enable bool
}

func (o *enableThermalControlOption) apply(c *config) {
	c.enableThermalControl = o.enable
}

func WithThermalControl(enable bool) Option {
	return &enableThermalControlOption{enable: enable}
}

func WithoutThermalControl() Option {
	return &enableThermalControlOption{enable: false}
}

type thermalControlIntervalOption struct {
	interval time.Duration
}

func (o *thermalControlIntervalOption) apply(c *config) {
	c.thermalControlInterval = o.interval
}

func WithThermalControlInterval(interval time.Duration) Option {
	return &thermalControlIntervalOption{interval: interval}
}

type emergencyCheckIntervalOption struct {
	interval time.Duration
}

func (o *emergencyCheckIntervalOption) apply(c *config) {
	c.emergencyCheckInterval = o.interval
}

func WithEmergencyCheckInterval(interval time.Duration) Option {
	return &emergencyCheckIntervalOption{interval: interval}
}

type defaultPIDSampleTimeOption struct {
	sampleTime time.Duration
}

func (o *defaultPIDSampleTimeOption) apply(c *config) {
	c.defaultPIDSampleTime = o.sampleTime
}

func WithDefaultPIDSampleTime(sampleTime time.Duration) Option {
	return &defaultPIDSampleTimeOption{sampleTime: sampleTime}
}

type maxThermalZonesOption struct {
	maxVal int
}

func (o *maxThermalZonesOption) apply(c *config) {
	c.maxThermalZones = o.maxVal
}

func WithMaxThermalZones(maxVal int) Option {
	return &maxThermalZonesOption{maxVal: maxVal}
}

type maxCoolingDevicesOption struct {
	maxVal int
}

func (o *maxCoolingDevicesOption) apply(c *config) {
	c.maxCoolingDevices = o.maxVal
}

func WithMaxCoolingDevices(maxVal int) Option {
	return &maxCoolingDevicesOption{maxVal: maxVal}
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

type enableHwmonDiscoveryOption struct {
	enable bool
}

func (o *enableHwmonDiscoveryOption) apply(c *config) {
	c.enableHwmonDiscovery = o.enable
}

func WithDiscovery(enable bool) Option {
	return &enableHwmonDiscoveryOption{enable: enable}
}

func WithoutDiscovery() Option {
	return &enableHwmonDiscoveryOption{enable: false}
}

type temperatureThresholdsOption struct {
	warning   float64
	critical  float64
	emergency float64
}

func (o *temperatureThresholdsOption) apply(c *config) {
	c.defaultWarningTemp = o.warning
	c.defaultCriticalTemp = o.critical
	c.emergencyShutdownTemp = o.emergency
}

func WithTemperatureThresholds(warning, critical, emergency float64) Option {
	return &temperatureThresholdsOption{warning: warning, critical: critical, emergency: emergency}
}

type defaultPIDConfigOption struct {
	kp float64
	ki float64
	kd float64
}

func (o *defaultPIDConfigOption) apply(c *config) {
	c.defaultPIDKp = o.kp
	c.defaultPIDKi = o.ki
	c.defaultPIDKd = o.kd
}

func WithDefaultPIDConfig(kp, ki, kd float64) Option {
	return &defaultPIDConfigOption{kp: kp, ki: ki, kd: kd}
}

type outputRangeOption struct {
	minVal float64
	maxVal float64
}

func (o *outputRangeOption) apply(c *config) {
	c.defaultOutputMin = o.minVal
	c.defaultOutputMax = o.maxVal
}

func WithOutputRange(minVal, maxVal float64) Option {
	return &outputRangeOption{minVal: minVal, maxVal: maxVal}
}

type sensormonEndpointOption struct {
	endpoint string
}

func (o *sensormonEndpointOption) apply(c *config) {
	c.sensormonEndpoint = o.endpoint
}

func WithSensormonEndpoint(endpoint string) Option {
	return &sensormonEndpointOption{endpoint: endpoint}
}

type powermgrEndpointOption struct {
	endpoint string
}

func (o *powermgrEndpointOption) apply(c *config) {
	c.powermgrEndpoint = o.endpoint
}

func WithPowermgrEndpoint(endpoint string) Option {
	return &powermgrEndpointOption{endpoint: endpoint}
}

type enableSensorIntegrationOption struct {
	enable bool
}

func (o *enableSensorIntegrationOption) apply(c *config) {
	c.enableSensorIntegration = o.enable
}

func WithSensorIntegration(enable bool) Option {
	return &enableSensorIntegrationOption{enable: enable}
}

func WithoutSensorIntegration() Option {
	return &enableSensorIntegrationOption{enable: false}
}

type enablePowerIntegrationOption struct {
	enable bool
}

func (o *enablePowerIntegrationOption) apply(c *config) {
	c.enablePowerIntegration = o.enable
}

func WithPowerIntegration(enable bool) Option {
	return &enablePowerIntegrationOption{enable: enable}
}

func WithoutPowerIntegration() Option {
	return &enablePowerIntegrationOption{enable: false}
}

type integrationOption struct {
	enableSensor bool
	enablePower  bool
}

func (o *integrationOption) apply(c *config) {
	c.enableSensorIntegration = o.enableSensor
	c.enablePowerIntegration = o.enablePower
}

func WithIntegration(enableSensor, enablePower bool) Option {
	return &integrationOption{enableSensor: enableSensor, enablePower: enablePower}
}

type persistThermalDataOption struct {
	enable bool
}

func (o *persistThermalDataOption) apply(c *config) {
	c.persistThermalData = o.enable
}

func WithPersistThermalData(enable bool) Option {
	return &persistThermalDataOption{enable: enable}
}

func WithoutPersistThermalData() Option {
	return &persistThermalDataOption{enable: false}
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

type persistenceOption struct {
	enabled    bool
	streamName string
	retention  time.Duration
}

func (o *persistenceOption) apply(c *config) {
	c.persistThermalData = o.enabled
	c.streamName = o.streamName
	c.streamRetention = o.retention
}

func WithPersistence(enabled bool, streamName string, retention time.Duration) Option {
	return &persistenceOption{enabled: enabled, streamName: streamName, retention: retention}
}

type enableEmergencyResponseOption struct {
	enable bool
}

func (o *enableEmergencyResponseOption) apply(c *config) {
	c.enableEmergencyResponse = o.enable
}

func WithEmergencyResponse(enable bool) Option {
	return &enableEmergencyResponseOption{enable: enable}
}

func WithoutEmergencyResponse() Option {
	return &enableEmergencyResponseOption{enable: false}
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

type failsafeCoolingLevelOption struct {
	level float64
}

func (o *failsafeCoolingLevelOption) apply(c *config) {
	c.failsafeCoolingLevel = o.level
}

func WithFailsafeCoolingLevel(level float64) Option {
	return &failsafeCoolingLevelOption{level: level}
}

type emergencyResponseConfigOption struct {
	enabled       bool
	delay         time.Duration
	failsafeLevel float64
}

func (o *emergencyResponseConfigOption) apply(c *config) {
	c.enableEmergencyResponse = o.enabled
	c.emergencyResponseDelay = o.delay
	c.failsafeCoolingLevel = o.failsafeLevel
}

func WithEmergencyResponseConfig(enabled bool, delay time.Duration, failsafeLevel float64) Option {
	return &emergencyResponseConfigOption{enabled: enabled, delay: delay, failsafeLevel: failsafeLevel}
}

type thermalZonesOption struct {
	zones []ThermalZoneConfig
}

func (o *thermalZonesOption) apply(c *config) {
	c.thermalZones = o.zones
}

func WithThermalZones(zones ...ThermalZoneConfig) Option {
	return &thermalZonesOption{zones: zones}
}

type coolingDevicesOption struct {
	devices []CoolingDeviceConfig
}

func (o *coolingDevicesOption) apply(c *config) {
	c.coolingDevices = o.devices
}

func WithCoolingDevices(devices ...CoolingDeviceConfig) Option {
	return &coolingDevicesOption{devices: devices}
}

type callbacksOption struct {
	callbacks ThermalCallbacks
}

func (o *callbacksOption) apply(c *config) {
	c.callbacks = o.callbacks
}

func WithCallbacks(callbacks ThermalCallbacks) Option {
	return &callbacksOption{callbacks: callbacks}
}

type enableMockModeOption struct {
	enable bool
}

func (o *enableMockModeOption) apply(c *config) {
	c.enableMockMode = o.enable
}

func WithMockMode(enable bool) Option {
	return &enableMockModeOption{enable: enable}
}

func WithoutMockMode() Option {
	return &enableMockModeOption{enable: false}
}

func (c *config) Validate() error {
	if c.serviceName == "" {
		return fmt.Errorf("%w: service name cannot be empty", ErrInvalidConfiguration)
	}

	if c.serviceVersion == "" {
		return fmt.Errorf("%w: service version cannot be empty", ErrInvalidConfiguration)
	}

	if c.thermalControlInterval <= 0 {
		return fmt.Errorf("%w: thermal control interval must be positive", ErrInvalidConfiguration)
	}

	if c.emergencyCheckInterval <= 0 {
		return fmt.Errorf("%w: emergency check interval must be positive", ErrInvalidConfiguration)
	}

	if c.defaultPIDSampleTime <= 0 {
		return fmt.Errorf("%w: PID sample time must be positive", ErrInvalidConfiguration)
	}

	if c.maxThermalZones <= 0 {
		return fmt.Errorf("%w: max thermal zones must be positive", ErrInvalidConfiguration)
	}

	if c.maxCoolingDevices <= 0 {
		return fmt.Errorf("%w: max cooling devices must be positive", ErrInvalidConfiguration)
	}

	if c.defaultWarningTemp >= c.defaultCriticalTemp {
		return fmt.Errorf("%w: warning temperature must be less than critical temperature", ErrInvalidConfiguration)
	}

	if c.defaultCriticalTemp >= c.emergencyShutdownTemp {
		return fmt.Errorf("%w: critical temperature must be less than emergency shutdown temperature", ErrInvalidConfiguration)
	}

	if c.defaultOutputMin >= c.defaultOutputMax {
		return fmt.Errorf("%w: output minimum must be less than output maximum", ErrInvalidConfiguration)
	}

	if c.failsafeCoolingLevel < 0 || c.failsafeCoolingLevel > 100 {
		return fmt.Errorf("%w: failsafe cooling level must be between 0 and 100", ErrInvalidConfiguration)
	}

	if c.persistThermalData {
		if c.streamName == "" {
			return fmt.Errorf("%w: stream name cannot be empty when thermal data persistence is enabled", ErrInvalidConfiguration)
		}

		if len(c.streamSubjects) == 0 {
			return fmt.Errorf("%w: at least one stream subject must be configured when thermal data persistence is enabled", ErrInvalidConfiguration)
		}

		for _, s := range c.streamSubjects {
			if len(s) == 0 {
				return fmt.Errorf("%w: stream subject cannot be empty", ErrInvalidConfiguration)
			}
		}

		if c.streamRetention <= 0 {
			return fmt.Errorf("%w: stream retention must be positive when thermal data persistence is enabled", ErrInvalidConfiguration)
		}
	}

	if c.enableSensorIntegration && c.sensormonEndpoint == "" {
		return fmt.Errorf("%w: sensormon endpoint cannot be empty when sensor integration is enabled", ErrInvalidConfiguration)
	}

	if c.enablePowerIntegration && c.powermgrEndpoint == "" {
		return fmt.Errorf("%w: powermgr endpoint cannot be empty when power integration is enabled", ErrInvalidConfiguration)
	}

	if c.enableEmergencyResponse && c.emergencyResponseDelay <= 0 {
		return fmt.Errorf("%w: emergency response delay must be positive when emergency response is enabled", ErrInvalidConfiguration)
	}

	// Validate thermal zones
	zoneIDs := make(map[string]bool)
	for i, zone := range c.thermalZones {
		if zone.ID == "" {
			return fmt.Errorf("%w: thermal zone %d has empty ID", ErrInvalidConfiguration, i)
		}

		if zoneIDs[zone.ID] {
			return fmt.Errorf("%w: duplicate thermal zone ID '%s'", ErrInvalidConfiguration, zone.ID)
		}
		zoneIDs[zone.ID] = true

		if zone.Name == "" {
			return fmt.Errorf("%w: thermal zone '%s' has empty name", ErrInvalidConfiguration, zone.ID)
		}

		if len(zone.SensorIDs) == 0 {
			return fmt.Errorf("%w: thermal zone '%s' has no sensors", ErrInvalidConfiguration, zone.ID)
		}

		if zone.WarningTemp >= zone.CriticalTemp {
			return fmt.Errorf("%w: thermal zone '%s' warning temperature must be less than critical temperature", ErrInvalidConfiguration, zone.ID)
		}

		if zone.EmergencyTemp > 0 && zone.CriticalTemp >= zone.EmergencyTemp {
			return fmt.Errorf("%w: thermal zone '%s' critical temperature must be less than emergency temperature", ErrInvalidConfiguration, zone.ID)
		}

		if zone.PIDConfig != nil {
			if zone.PIDConfig.SampleTime <= 0 {
				return fmt.Errorf("%w: thermal zone '%s' PID sample time must be positive", ErrInvalidConfiguration, zone.ID)
			}
			if zone.PIDConfig.OutputMin >= zone.PIDConfig.OutputMax {
				return fmt.Errorf("%w: thermal zone '%s' PID output minimum must be less than maximum", ErrInvalidConfiguration, zone.ID)
			}
		}
	}

	// Validate cooling devices
	deviceIDs := make(map[string]bool)
	for i, device := range c.coolingDevices {
		if device.ID == "" {
			return fmt.Errorf("%w: cooling device %d has empty ID", ErrInvalidConfiguration, i)
		}

		if deviceIDs[device.ID] {
			return fmt.Errorf("%w: duplicate cooling device ID '%s'", ErrInvalidConfiguration, device.ID)
		}
		deviceIDs[device.ID] = true

		if device.Name == "" {
			return fmt.Errorf("%w: cooling device '%s' has empty name", ErrInvalidConfiguration, device.ID)
		}

		if device.MinSpeed < 0 || device.MinSpeed > 100 {
			return fmt.Errorf("%w: cooling device '%s' minimum speed must be between 0 and 100", ErrInvalidConfiguration, device.ID)
		}

		if device.MaxSpeed < 0 || device.MaxSpeed > 100 {
			return fmt.Errorf("%w: cooling device '%s' maximum speed must be between 0 and 100", ErrInvalidConfiguration, device.ID)
		}

		if device.MinSpeed >= device.MaxSpeed {
			return fmt.Errorf("%w: cooling device '%s' minimum speed must be less than maximum speed", ErrInvalidConfiguration, device.ID)
		}

		if device.InitialSpeed < device.MinSpeed || device.InitialSpeed > device.MaxSpeed {
			return fmt.Errorf("%w: cooling device '%s' initial speed must be between minimum and maximum speed", ErrInvalidConfiguration, device.ID)
		}

		if device.Type == CoolingDeviceTypeFan && device.HwmonPath == "" && !c.enableMockMode {
			return fmt.Errorf("%w: cooling device '%s' of type fan must have hwmon_path in non-mock mode", ErrInvalidConfiguration, device.ID)
		}
	}

	return nil
}

// Helper functions for creating thermal configurations

// NewThermalZone creates a new thermal zone configuration.
func NewThermalZone(id, name string, sensorIDs []string, coolingDeviceIDs []string, targetTemp, warningTemp, criticalTemp float64) ThermalZoneConfig {
	return ThermalZoneConfig{
		ID:               id,
		Name:             name,
		Enabled:          true,
		SensorIDs:        sensorIDs,
		CoolingDeviceIDs: coolingDeviceIDs,
		TargetTemp:       targetTemp,
		WarningTemp:      warningTemp,
		CriticalTemp:     criticalTemp,
		PIDConfig: &PIDConfig{
			Kp:         DefaultDefaultPIDKp,
			Ki:         DefaultDefaultPIDKi,
			Kd:         DefaultDefaultPIDKd,
			SampleTime: DefaultDefaultPIDSampleTime,
			OutputMin:  DefaultDefaultOutputMin,
			OutputMax:  DefaultDefaultOutputMax,
		},
	}
}

// NewCoolingDevice creates a new cooling device configuration.
func NewCoolingDevice(id, name string, deviceType CoolingDeviceType, minSpeed, maxSpeed, initialSpeed float64) CoolingDeviceConfig {
	return CoolingDeviceConfig{
		ID:           id,
		Name:         name,
		Type:         deviceType,
		Enabled:      true,
		MinSpeed:     minSpeed,
		MaxSpeed:     maxSpeed,
		InitialSpeed: initialSpeed,
	}
}

// NewFanDevice creates a new fan cooling device configuration.
func NewFanDevice(id, name, hwmonPath string, pwmChannel int, minSpeed, maxSpeed, initialSpeed float64) CoolingDeviceConfig {
	return CoolingDeviceConfig{
		ID:           id,
		Name:         name,
		Type:         CoolingDeviceTypeFan,
		Enabled:      true,
		MinSpeed:     minSpeed,
		MaxSpeed:     maxSpeed,
		InitialSpeed: initialSpeed,
		HwmonPath:    hwmonPath,
		PWMChannel:   pwmChannel,
	}
}

// NewMockCoolingDevice creates a new mock cooling device configuration for testing.
func NewMockCoolingDevice(id, name string, minSpeed, maxSpeed, initialSpeed float64) CoolingDeviceConfig {
	return CoolingDeviceConfig{
		ID:           id,
		Name:         name,
		Type:         CoolingDeviceTypeMock,
		Enabled:      true,
		MinSpeed:     minSpeed,
		MaxSpeed:     maxSpeed,
		InitialSpeed: initialSpeed,
	}
}

// NewPIDConfig creates a new PID controller configuration.
func NewPIDConfig(kp, ki, kd float64, sampleTime time.Duration, outputMin, outputMax float64) *PIDConfig {
	return &PIDConfig{
		Kp:         kp,
		Ki:         ki,
		Kd:         kd,
		SampleTime: sampleTime,
		OutputMin:  outputMin,
		OutputMax:  outputMax,
	}
}
