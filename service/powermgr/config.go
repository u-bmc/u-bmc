// SPDX-License-Identifier: BSD-3-Clause

package powermgr

import (
	"fmt"
	"strings"
	"time"

	"github.com/u-bmc/u-bmc/pkg/gpio"
)

// Default configuration constants.
const (
	DefaultServiceName        = "powermgr"
	DefaultServiceDescription = "Power management service for BMC components"
	DefaultServiceVersion     = "1.0.0"
	DefaultGPIOChip           = "/dev/gpiochip0"
	DefaultOperationTimeout   = 30 * time.Second
	DefaultPowerOnDelay       = 200 * time.Millisecond
	DefaultPowerOffDelay      = 200 * time.Millisecond
	DefaultResetDelay         = 100 * time.Millisecond
	DefaultForceOffDelay      = 4 * time.Second
)

// GPIOLineConfig holds GPIO line configuration for power operations.
type GPIOLineConfig struct {
	// Line is the GPIO line name or number identifier
	Line string
	// Direction specifies the GPIO direction (input/output)
	Direction gpio.Direction
	// ActiveState specifies active high or low behavior
	ActiveState gpio.ActiveState
	// InitialValue is the initial value for output lines
	InitialValue int
	// Bias configures pull-up/pull-down resistors
	Bias gpio.Bias
}

// GPIOConfig holds GPIO configuration for a component.
type GPIOConfig struct {
	// PowerButton configures the power button GPIO line
	PowerButton GPIOLineConfig
	// ResetButton configures the reset button GPIO line
	ResetButton GPIOLineConfig
	// PowerStatus configures the power status input GPIO line
	PowerStatus GPIOLineConfig
}

// ComponentConfig holds configuration for a single component.
type ComponentConfig struct {
	// Name is the component identifier
	Name string
	// Type specifies the component type (host, chassis, bmc)
	Type string
	// Enabled indicates if the component is enabled for power management
	Enabled bool
	// GPIO holds GPIO-specific configuration
	GPIO GPIOConfig
	// OperationTimeout is the timeout for power operations
	OperationTimeout time.Duration
	// PowerOnDelay is the duration for power button press (power on)
	PowerOnDelay time.Duration
	// PowerOffDelay is the duration for power button press (soft power off)
	PowerOffDelay time.Duration
	// ResetDelay is the duration for reset button press
	ResetDelay time.Duration
	// ForceOffDelay is the duration for force power off (hard shutdown)
	ForceOffDelay time.Duration
}

// Config holds the configuration for the power manager service.
type Config struct {
	// ServiceName is the name of the service in the NATS micro framework
	ServiceName string
	// ServiceDescription provides a human-readable description of the service
	ServiceDescription string
	// ServiceVersion is the semantic version of the service
	ServiceVersion string
	// GPIOChip is the path to the GPIO chip device
	GPIOChip string
	// Components maps component names to their configuration
	Components map[string]ComponentConfig
	// EnableHostManagement enables power management for host components
	EnableHostManagement bool
	// EnableChassisManagement enables power management for chassis components
	EnableChassisManagement bool
	// EnableBMCManagement enables power management for BMC components
	EnableBMCManagement bool
	// NumHosts is the number of hosts to manage
	NumHosts int
	// NumChassis is the number of chassis to manage
	NumChassis int
	// DefaultOperationTimeout is the default timeout for power operations
	DefaultOperationTimeout time.Duration
	// EnableMetrics enables metrics collection for power operations
	EnableMetrics bool
	// EnableTracing enables distributed tracing for power operations
	EnableTracing bool
}

// Option represents a configuration option for the power manager.
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

type gpioChipOption struct {
	chip string
}

func (o *gpioChipOption) apply(c *Config) {
	c.GPIOChip = o.chip
}

// WithGPIOChip sets the GPIO chip path.
func WithGPIOChip(chip string) Option {
	return &gpioChipOption{chip: chip}
}

type componentsOption struct {
	components map[string]ComponentConfig
}

func (o *componentsOption) apply(c *Config) {
	if c.Components == nil {
		c.Components = make(map[string]ComponentConfig)
	}
	for name, config := range o.components {
		c.Components[name] = config
	}
}

// WithComponents sets the component configurations.
func WithComponents(components map[string]ComponentConfig) Option {
	return &componentsOption{components: components}
}

type enableHostManagementOption struct {
	enable bool
}

func (o *enableHostManagementOption) apply(c *Config) {
	c.EnableHostManagement = o.enable
}

// WithHostManagement enables or disables host power management.
func WithHostManagement(enable bool) Option {
	return &enableHostManagementOption{enable: enable}
}

type enableChassisManagementOption struct {
	enable bool
}

func (o *enableChassisManagementOption) apply(c *Config) {
	c.EnableChassisManagement = o.enable
}

// WithChassisManagement enables or disables chassis power management.
func WithChassisManagement(enable bool) Option {
	return &enableChassisManagementOption{enable: enable}
}

type enableBMCManagementOption struct {
	enable bool
}

func (o *enableBMCManagementOption) apply(c *Config) {
	c.EnableBMCManagement = o.enable
}

// WithBMCManagement enables or disables BMC power management.
func WithBMCManagement(enable bool) Option {
	return &enableBMCManagementOption{enable: enable}
}

type numHostsOption struct {
	num int
}

func (o *numHostsOption) apply(c *Config) {
	c.NumHosts = o.num
}

// WithNumHosts sets the number of hosts to manage.
func WithNumHosts(num int) Option {
	return &numHostsOption{num: num}
}

type numChassisOption struct {
	num int
}

func (o *numChassisOption) apply(c *Config) {
	c.NumChassis = o.num
}

// WithNumChassis sets the number of chassis to manage.
func WithNumChassis(num int) Option {
	return &numChassisOption{num: num}
}

type defaultOperationTimeoutOption struct {
	timeout time.Duration
}

func (o *defaultOperationTimeoutOption) apply(c *Config) {
	c.DefaultOperationTimeout = o.timeout
}

// WithDefaultOperationTimeout sets the default timeout for power operations.
func WithDefaultOperationTimeout(timeout time.Duration) Option {
	return &defaultOperationTimeoutOption{timeout: timeout}
}

type enableMetricsOption struct {
	enable bool
}

func (o *enableMetricsOption) apply(c *Config) {
	c.EnableMetrics = o.enable
}

// WithMetrics enables or disables metrics collection.
func WithMetrics(enable bool) Option {
	return &enableMetricsOption{enable: enable}
}

type enableTracingOption struct {
	enable bool
}

func (o *enableTracingOption) apply(c *Config) {
	c.EnableTracing = o.enable
}

// WithTracing enables or disables distributed tracing.
func WithTracing(enable bool) Option {
	return &enableTracingOption{enable: enable}
}

// NewConfig creates a new power manager configuration with default values.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		ServiceName:             DefaultServiceName,
		ServiceDescription:      DefaultServiceDescription,
		ServiceVersion:          DefaultServiceVersion,
		GPIOChip:                DefaultGPIOChip,
		Components:              make(map[string]ComponentConfig),
		EnableHostManagement:    true,
		EnableChassisManagement: true,
		EnableBMCManagement:     true,
		NumHosts:                1,
		NumChassis:              1,
		DefaultOperationTimeout: DefaultOperationTimeout,
		EnableMetrics:           true,
		EnableTracing:           true,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

// NewComponentConfig creates a new component configuration with default values.
func NewComponentConfig(name, componentType string) ComponentConfig {
	return ComponentConfig{
		Name:             name,
		Type:             componentType,
		Enabled:          true,
		GPIO:             NewDefaultGPIOConfig(),
		OperationTimeout: DefaultOperationTimeout,
		PowerOnDelay:     DefaultPowerOnDelay,
		PowerOffDelay:    DefaultPowerOffDelay,
		ResetDelay:       DefaultResetDelay,
		ForceOffDelay:    DefaultForceOffDelay,
	}
}

// NewDefaultGPIOConfig creates a default GPIO configuration.
func NewDefaultGPIOConfig() GPIOConfig {
	return GPIOConfig{
		PowerButton: GPIOLineConfig{
			Direction:    gpio.DirectionOutput,
			ActiveState:  gpio.ActiveLow,
			InitialValue: 0,
			Bias:         gpio.BiasDisabled,
		},
		ResetButton: GPIOLineConfig{
			Direction:    gpio.DirectionOutput,
			ActiveState:  gpio.ActiveLow,
			InitialValue: 0,
			Bias:         gpio.BiasDisabled,
		},
		PowerStatus: GPIOLineConfig{
			Direction:   gpio.DirectionInput,
			ActiveState: gpio.ActiveHigh,
			Bias:        gpio.BiasPullDown,
		},
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("%w: service name cannot be empty", ErrInvalidConfiguration)
	}

	if c.ServiceVersion == "" {
		return fmt.Errorf("%w: service version cannot be empty", ErrInvalidConfiguration)
	}

	if c.GPIOChip == "" {
		return fmt.Errorf("%w: GPIO chip path cannot be empty", ErrInvalidConfiguration)
	}

	if !strings.HasPrefix(c.GPIOChip, "/dev/gpiochip") {
		return fmt.Errorf("%w: GPIO chip path must start with '/dev/gpiochip'", ErrInvalidConfiguration)
	}

	if !c.EnableHostManagement && !c.EnableChassisManagement && !c.EnableBMCManagement {
		return fmt.Errorf("%w: at least one component type must be enabled", ErrInvalidConfiguration)
	}

	if c.EnableHostManagement && c.NumHosts <= 0 {
		return fmt.Errorf("%w: number of hosts must be positive when host management is enabled", ErrInvalidConfiguration)
	}

	if c.EnableChassisManagement && c.NumChassis <= 0 {
		return fmt.Errorf("%w: number of chassis must be positive when chassis management is enabled", ErrInvalidConfiguration)
	}

	if c.DefaultOperationTimeout <= 0 {
		return fmt.Errorf("%w: default operation timeout must be positive", ErrInvalidConfiguration)
	}

	for name, component := range c.Components {
		if err := c.validateComponentConfig(name, component); err != nil {
			return err
		}
	}

	return nil
}

// validateComponentConfig validates a single component configuration.
func (c *Config) validateComponentConfig(name string, component ComponentConfig) error {
	if component.Name != name {
		return fmt.Errorf("%w: component name mismatch for '%s'", ErrInvalidConfiguration, name)
	}

	validTypes := []string{"host", "chassis", "bmc"}
	validType := false
	for _, t := range validTypes {
		if component.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("%w: invalid component type '%s' for component '%s'", ErrInvalidConfiguration, component.Type, name)
	}

	if component.OperationTimeout <= 0 {
		return fmt.Errorf("%w: operation timeout must be positive for component '%s'", ErrInvalidConfiguration, name)
	}

	if component.PowerOnDelay <= 0 {
		return fmt.Errorf("%w: power on delay must be positive for component '%s'", ErrInvalidConfiguration, name)
	}

	if component.PowerOffDelay <= 0 {
		return fmt.Errorf("%w: power off delay must be positive for component '%s'", ErrInvalidConfiguration, name)
	}

	if component.ResetDelay <= 0 {
		return fmt.Errorf("%w: reset delay must be positive for component '%s'", ErrInvalidConfiguration, name)
	}

	if component.ForceOffDelay <= 0 {
		return fmt.Errorf("%w: force off delay must be positive for component '%s'", ErrInvalidConfiguration, name)
	}

	if err := c.validateGPIOConfig(name, component.GPIO); err != nil {
		return err
	}

	return nil
}

// validateGPIOConfig validates GPIO configuration for a component.
func (c *Config) validateGPIOConfig(componentName string, gpioConfig GPIOConfig) error {
	lines := map[string]GPIOLineConfig{
		"power_button": gpioConfig.PowerButton,
		"reset_button": gpioConfig.ResetButton,
		"power_status": gpioConfig.PowerStatus,
	}

	for lineName, lineConfig := range lines {
		if lineConfig.Line == "" {
			continue // Optional GPIO lines
		}

		if lineConfig.InitialValue < 0 || lineConfig.InitialValue > 1 {
			return fmt.Errorf("%w: initial value for GPIO line '%s' of component '%s' must be 0 or 1", ErrInvalidGPIOConfiguration, lineName, componentName)
		}

		if lineConfig.Direction == gpio.DirectionOutput && lineName == "power_status" {
			return fmt.Errorf("%w: power status GPIO line for component '%s' must be input", ErrInvalidGPIOConfiguration, componentName)
		}

		if lineConfig.Direction == gpio.DirectionInput && (lineName == "power_button" || lineName == "reset_button") {
			return fmt.Errorf("%w: control GPIO line '%s' for component '%s' must be output", ErrInvalidGPIOConfiguration, lineName, componentName)
		}
	}

	return nil
}

// GetComponentConfig returns the configuration for a specific component.
func (c *Config) GetComponentConfig(name string) (ComponentConfig, bool) {
	config, exists := c.Components[name]
	return config, exists
}

// GetHostConfig returns the configuration for a host by index.
func (c *Config) GetHostConfig(index int) (ComponentConfig, bool) {
	name := fmt.Sprintf("host.%d", index)
	return c.GetComponentConfig(name)
}

// GetChassisConfig returns the configuration for a chassis by index.
func (c *Config) GetChassisConfig(index int) (ComponentConfig, bool) {
	name := fmt.Sprintf("chassis.%d", index)
	return c.GetComponentConfig(name)
}

// GetBMCConfig returns the configuration for the BMC.
func (c *Config) GetBMCConfig() (ComponentConfig, bool) {
	return c.GetComponentConfig("bmc.0")
}

// AddDefaultComponents adds default component configurations based on enabled features.
func (c *Config) AddDefaultComponents() {
	if c.Components == nil {
		c.Components = make(map[string]ComponentConfig)
	}

	if c.EnableHostManagement {
		for i := 0; i < c.NumHosts; i++ {
			name := fmt.Sprintf("host.%d", i)
			if _, exists := c.Components[name]; !exists {
				c.Components[name] = NewComponentConfig(name, "host")
			}
		}
	}

	if c.EnableChassisManagement {
		for i := 0; i < c.NumChassis; i++ {
			name := fmt.Sprintf("chassis.%d", i)
			if _, exists := c.Components[name]; !exists {
				c.Components[name] = NewComponentConfig(name, "chassis")
			}
		}
	}

	if c.EnableBMCManagement {
		name := "bmc.0"
		if _, exists := c.Components[name]; !exists {
			c.Components[name] = NewComponentConfig(name, "bmc")
		}
	}
}
