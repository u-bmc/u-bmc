// SPDX-License-Identifier: BSD-3-Clause

//nolint:goconst
package hwmon

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// SensorType represents the type of hardware sensor.
type SensorType int

const (
	// SensorTypeTemperature represents temperature sensors (temp*).
	SensorTypeTemperature SensorType = iota
	// SensorTypeVoltage represents voltage sensors (in*).
	SensorTypeVoltage
	// SensorTypeFan represents fan sensors (fan*).
	SensorTypeFan
	// SensorTypePower represents power sensors (power*).
	SensorTypePower
	// SensorTypeCurrent represents current sensors (curr*).
	SensorTypeCurrent
	// SensorTypeHumidity represents humidity sensors (humidity*).
	SensorTypeHumidity
	// SensorTypePressure represents pressure sensors (pressure*).
	SensorTypePressure
	// SensorTypePWM represents PWM outputs (pwm*).
	SensorTypePWM
	// SensorTypeGeneric represents generic sensors or custom types.
	SensorTypeGeneric
)

// String returns the string representation of the sensor type.
func (st SensorType) String() string {
	switch st {
	case SensorTypeTemperature:
		return "temperature"
	case SensorTypeVoltage:
		return "voltage"
	case SensorTypeFan:
		return "fan"
	case SensorTypePower:
		return "power"
	case SensorTypeCurrent:
		return "current"
	case SensorTypeHumidity:
		return "humidity"
	case SensorTypePressure:
		return "pressure"
	case SensorTypePWM:
		return "pwm"
	case SensorTypeGeneric:
		return "generic"
	default:
		return "unknown"
	}
}

// Prefix returns the hwmon file prefix for the sensor type.
func (st SensorType) Prefix() string {
	switch st {
	case SensorTypeTemperature:
		return "temp"
	case SensorTypeVoltage:
		return "in"
	case SensorTypeFan:
		return "fan"
	case SensorTypePower:
		return "power"
	case SensorTypeCurrent:
		return "curr"
	case SensorTypeHumidity:
		return "humidity"
	case SensorTypePressure:
		return "pressure"
	case SensorTypePWM:
		return "pwm"
	default:
		return ""
	}
}

// SensorAttribute represents different sensor attributes available in hwmon.
type SensorAttribute int

const (
	// AttributeInput represents the current sensor reading (*_input).
	AttributeInput SensorAttribute = iota
	// AttributeLabel represents the sensor label (*_label).
	AttributeLabel
	// AttributeMin represents the minimum threshold (*_min).
	AttributeMin
	// AttributeMax represents the maximum threshold (*_max).
	AttributeMax
	// AttributeCrit represents the critical threshold (*_crit).
	AttributeCrit
	// AttributeAlarm represents the alarm status (*_alarm).
	AttributeAlarm
	// AttributeEnable represents the enable/disable control (*_enable).
	AttributeEnable
	// AttributeTarget represents the target value (*_target).
	AttributeTarget
	// AttributeFault represents the fault status (*_fault).
	AttributeFault
	// AttributeBeep represents the beep enable (*_beep).
	AttributeBeep
	// AttributeOffset represents the sensor offset (*_offset).
	AttributeOffset
	// AttributeType represents the sensor type (*_type).
	AttributeType
)

// String returns the string representation of the sensor attribute.
func (sa SensorAttribute) String() string {
	switch sa {
	case AttributeInput:
		return "input"
	case AttributeLabel:
		return "label"
	case AttributeMin:
		return "min"
	case AttributeMax:
		return "max"
	case AttributeCrit:
		return "crit"
	case AttributeAlarm:
		return "alarm"
	case AttributeEnable:
		return "enable"
	case AttributeTarget:
		return "target"
	case AttributeFault:
		return "fault"
	case AttributeBeep:
		return "beep"
	case AttributeOffset:
		return "offset"
	case AttributeType:
		return "type"
	default:
		return "unknown"
	}
}

// IsWritable returns true if the attribute is typically writable.
func (sa SensorAttribute) IsWritable() bool {
	switch sa {
	case AttributeMin, AttributeMax, AttributeCrit, AttributeEnable,
		AttributeTarget, AttributeBeep, AttributeOffset:
		return true
	default:
		return false
	}
}

// Config holds the configuration for hwmon sensor access.
type Config struct {
	// Device name or hwmon device identifier
	Device string
	// SensorLabel is the human-readable sensor label
	SensorLabel string
	// SensorIndex is the numeric sensor index (e.g., 1 for temp1, 2 for fan2)
	SensorIndex int
	// SensorType specifies the type of sensor
	SensorType SensorType
	// Attribute specifies which sensor attribute to access
	Attribute SensorAttribute
	// BasePath is the base hwmon path
	BasePath string
	// Timeout for read/write operations
	Timeout time.Duration
	// RetryCount for failed operations
	RetryCount int
	// RetryDelay between retry attempts
	RetryDelay time.Duration
	// Writable indicates if this sensor supports writing
	Writable bool
	// ValidationEnabled enables value validation
	ValidationEnabled bool
	// CachingEnabled enables value caching
	CachingEnabled bool
	// CacheTTL is the cache time-to-live
	CacheTTL time.Duration
	// MinValue for validation (optional)
	MinValue *float64
	// MaxValue for validation (optional)
	MaxValue *float64
	// CustomPath allows specifying a custom sysfs path
	CustomPath string
	// UseIndex forces using sensor index instead of label
	UseIndex bool
	// StrictValidation enables strict path and value validation
	StrictValidation bool
}

// Option represents a configuration option for hwmon sensors.
type Option interface {
	apply(*Config)
}

type deviceOption struct {
	device string
}

func (o *deviceOption) apply(c *Config) {
	c.Device = o.device
}

// WithDevice sets the hwmon device name or identifier.
func WithDevice(device string) Option {
	return &deviceOption{device: device}
}

type sensorLabelOption struct {
	label string
}

func (o *sensorLabelOption) apply(c *Config) {
	c.SensorLabel = o.label
}

// WithSensorLabel sets the sensor label for identification.
func WithSensorLabel(label string) Option {
	return &sensorLabelOption{label: label}
}

type sensorIndexOption struct {
	index int
}

func (o *sensorIndexOption) apply(c *Config) {
	c.SensorIndex = o.index
}

// WithSensorIndex sets the sensor index (e.g., 1 for temp1).
func WithSensorIndex(index int) Option {
	return &sensorIndexOption{index: index}
}

type sensorTypeOption struct {
	sensorType SensorType
}

func (o *sensorTypeOption) apply(c *Config) {
	c.SensorType = o.sensorType
}

// WithSensorType sets the type of sensor.
func WithSensorType(sensorType SensorType) Option {
	return &sensorTypeOption{sensorType: sensorType}
}

type attributeOption struct {
	attribute SensorAttribute
}

func (o *attributeOption) apply(c *Config) {
	c.Attribute = o.attribute
}

// WithAttribute sets the sensor attribute to access.
func WithAttribute(attribute SensorAttribute) Option {
	return &attributeOption{attribute: attribute}
}

type basePathOption struct {
	basePath string
}

func (o *basePathOption) apply(c *Config) {
	c.BasePath = o.basePath
}

// WithBasePath sets the base hwmon sysfs path.
func WithBasePath(basePath string) Option {
	return &basePathOption{basePath: basePath}
}

type timeoutOption struct {
	timeout time.Duration
}

func (o *timeoutOption) apply(c *Config) {
	c.Timeout = o.timeout
}

// WithTimeout sets the timeout for read/write operations.
func WithTimeout(timeout time.Duration) Option {
	return &timeoutOption{timeout: timeout}
}

type retryOption struct {
	count int
	delay time.Duration
}

func (o *retryOption) apply(c *Config) {
	c.RetryCount = o.count
	c.RetryDelay = o.delay
}

// WithRetry sets the retry count and delay for failed operations.
func WithRetry(count int, delay time.Duration) Option {
	return &retryOption{count: count, delay: delay}
}

type writableOption struct {
	writable bool
}

func (o *writableOption) apply(c *Config) {
	c.Writable = o.writable
}

// WithWritable sets whether the sensor supports writing.
func WithWritable(writable bool) Option {
	return &writableOption{writable: writable}
}

type validationOption struct {
	enabled bool
}

func (o *validationOption) apply(c *Config) {
	c.ValidationEnabled = o.enabled
}

// WithValidation enables or disables value validation.
func WithValidation(enabled bool) Option {
	return &validationOption{enabled: enabled}
}

type cachingOption struct {
	enabled bool
	ttl     time.Duration
}

func (o *cachingOption) apply(c *Config) {
	c.CachingEnabled = o.enabled
	c.CacheTTL = o.ttl
}

// WithCaching enables value caching with the specified TTL.
func WithCaching(enabled bool, ttl time.Duration) Option {
	return &cachingOption{enabled: enabled, ttl: ttl}
}

type valueRangeOption struct {
	minVal *float64
	maxVal *float64
}

func (o *valueRangeOption) apply(c *Config) {
	c.MinValue = o.minVal
	c.MaxValue = o.maxVal
}

// WithValueRange sets the valid value range for validation.
func WithValueRange(minVal, maxVal *float64) Option {
	return &valueRangeOption{minVal: minVal, maxVal: maxVal}
}

type customPathOption struct {
	path string
}

func (o *customPathOption) apply(c *Config) {
	c.CustomPath = o.path
}

// WithCustomPath sets a custom sysfs path instead of auto-discovery.
func WithCustomPath(path string) Option {
	return &customPathOption{path: path}
}

type useIndexOption struct {
	useIndex bool
}

func (o *useIndexOption) apply(c *Config) {
	c.UseIndex = o.useIndex
}

// WithUseIndex forces using sensor index instead of label for identification.
func WithUseIndex(useIndex bool) Option {
	return &useIndexOption{useIndex: useIndex}
}

type strictValidationOption struct {
	strict bool
}

func (o *strictValidationOption) apply(c *Config) {
	c.StrictValidation = o.strict
}

// WithStrictValidation enables strict path and value validation.
func WithStrictValidation(strict bool) Option {
	return &strictValidationOption{strict: strict}
}

// NewConfig creates a new hwmon configuration with sensible defaults.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Device:            "",
		SensorLabel:       "",
		SensorIndex:       1,
		SensorType:        SensorTypeGeneric,
		Attribute:         AttributeInput,
		BasePath:          "/sys/class/hwmon",
		Timeout:           5 * time.Second,
		RetryCount:        3,
		RetryDelay:        100 * time.Millisecond,
		Writable:          false,
		ValidationEnabled: true,
		CachingEnabled:    false,
		CacheTTL:          1 * time.Second,
		MinValue:          nil,
		MaxValue:          nil,
		CustomPath:        "",
		UseIndex:          false,
		StrictValidation:  true,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.BasePath == "" {
		return fmt.Errorf("%w: base path cannot be empty", ErrInvalidConfig)
	}

	if c.CustomPath == "" {
		if c.Device == "" {
			return fmt.Errorf("%w: device name cannot be empty when custom path is not set", ErrInvalidConfig)
		}

		if !c.UseIndex && c.SensorLabel == "" {
			return fmt.Errorf("%w: sensor label cannot be empty when not using index", ErrInvalidConfig)
		}

		if c.UseIndex && c.SensorIndex < 1 {
			return fmt.Errorf("%w: sensor index must be positive", ErrInvalidConfig)
		}

		if c.SensorType == SensorTypeGeneric && c.SensorType.Prefix() == "" {
			return fmt.Errorf("%w: sensor type must be specified or use custom path", ErrInvalidConfig)
		}
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("%w: timeout must be positive", ErrInvalidConfig)
	}

	if c.RetryCount < 0 {
		return fmt.Errorf("%w: retry count cannot be negative", ErrInvalidConfig)
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf("%w: retry delay cannot be negative", ErrInvalidConfig)
	}

	if c.CachingEnabled && c.CacheTTL <= 0 {
		return fmt.Errorf("%w: cache TTL must be positive when caching is enabled", ErrInvalidConfig)
	}

	if c.MinValue != nil && c.MaxValue != nil && *c.MinValue > *c.MaxValue {
		return fmt.Errorf("%w: minimum value cannot be greater than maximum value", ErrInvalidConfig)
	}

	if c.Writable && !c.Attribute.IsWritable() {
		return fmt.Errorf("%w: attribute %s is not writable", ErrInvalidConfig, c.Attribute.String())
	}

	return nil
}

// GetSensorPath constructs the full sysfs path for the sensor.
func (c *Config) GetSensorPath() string {
	if c.CustomPath != "" {
		return c.CustomPath
	}

	if c.SensorType == SensorTypeGeneric {
		return ""
	}

	filename := BuildSensorFilename(c.SensorType, c.SensorIndex, c.Attribute)
	if filename == "" {
		return ""
	}

	return filepath.Join(c.BasePath, c.Device, filename)
}

// GetDevicePath returns the path to the hwmon device directory.
func (c *Config) GetDevicePath() string {
	if c.CustomPath != "" {
		return filepath.Dir(c.CustomPath)
	}
	return filepath.Join(c.BasePath, c.Device)
}

// IsReadOnly returns true if the sensor is configured as read-only.
func (c *Config) IsReadOnly() bool {
	return !c.Writable
}

// IsWriteOnly returns true if the sensor is configured as write-only.
func (c *Config) IsWriteOnly() bool {
	return c.Writable && c.Attribute != AttributeInput
}

// HasValueRange returns true if a value range is configured for validation.
func (c *Config) HasValueRange() bool {
	return c.MinValue != nil || c.MaxValue != nil
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	clone := *c

	if c.MinValue != nil {
		val := *c.MinValue
		clone.MinValue = &val
	}

	if c.MaxValue != nil {
		val := *c.MaxValue
		clone.MaxValue = &val
	}

	return &clone
}

// String returns a string representation of the configuration.
func (c *Config) String() string {
	var parts []string

	if c.CustomPath != "" {
		parts = append(parts, fmt.Sprintf("path=%s", c.CustomPath))
	} else {
		parts = append(parts, fmt.Sprintf("device=%s", c.Device))
		if c.UseIndex {
			parts = append(parts, fmt.Sprintf("index=%d", c.SensorIndex))
		} else {
			parts = append(parts, fmt.Sprintf("label=%s", c.SensorLabel))
		}
		parts = append(parts, fmt.Sprintf("type=%s", c.SensorType.String()))
		parts = append(parts, fmt.Sprintf("attr=%s", c.Attribute.String()))
	}

	if c.Writable {
		parts = append(parts, "writable=true")
	}

	if c.CachingEnabled {
		parts = append(parts, fmt.Sprintf("cache=%v", c.CacheTTL))
	}

	return fmt.Sprintf("Config{%s}", strings.Join(parts, ", "))
}
