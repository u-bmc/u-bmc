// SPDX-License-Identifier: BSD-3-Clause

package powermgr

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultServiceName        = "powermgr"
	DefaultServiceDescription = "Power management service for BMC components"
	DefaultServiceVersion     = "1.0.0"
	DefaultGPIOChip           = "/dev/gpiochip0"
	DefaultI2CDevice          = "/dev/i2c-0"
	DefaultOperationTimeout   = 30 * time.Second
	DefaultPowerOnDelay       = 200 * time.Millisecond
	DefaultPowerOffDelay      = 200 * time.Millisecond
	DefaultResetDelay         = 100 * time.Millisecond
	DefaultForceOffDelay      = 4 * time.Second
)

type BackendType string

const (
	BackendTypeGPIO BackendType = "gpio"
	BackendTypeI2C  BackendType = "i2c"
)

type GPIOActiveState int

const (
	ActiveHigh GPIOActiveState = iota
	ActiveLow
)

type GPIODirection int

const (
	DirectionInput GPIODirection = iota
	DirectionOutput
)

type GPIOBias int

const (
	BiasDisabled GPIOBias = iota
	BiasPullUp
	BiasPullDown
)

type GPIOLineConfig struct {
	Line         string
	Direction    GPIODirection
	ActiveState  GPIOActiveState
	InitialValue int
	Bias         GPIOBias
}

type GPIOConfig struct {
	PowerButton GPIOLineConfig
	ResetButton GPIOLineConfig
	PowerStatus GPIOLineConfig
}

type I2CConfig struct {
	DevicePath    string
	SlaveAddress  uint8
	PowerOnReg    uint8
	PowerOffReg   uint8
	ResetReg      uint8
	StatusReg     uint8
	PowerOnValue  uint8
	PowerOffValue uint8
	ResetValue    uint8
}

type ComponentConfig struct {
	Name             string
	Type             string
	Enabled          bool
	Backend          BackendType
	GPIO             GPIOConfig
	I2C              I2CConfig
	OperationTimeout time.Duration
	PowerOnDelay     time.Duration
	PowerOffDelay    time.Duration
	ResetDelay       time.Duration
	ForceOffDelay    time.Duration
}

type config struct {
	serviceName                 string
	serviceDescription          string
	serviceVersion              string
	gpioChip                    string
	i2cDevice                   string
	defaultBackend              BackendType
	components                  map[string]ComponentConfig
	enableHostManagement        bool
	enableChassisManagement     bool
	enableBMCManagement         bool
	numHosts                    int
	numChassis                  int
	defaultOperationTimeout     time.Duration
	enableMetrics               bool
	enableTracing               bool
	enableStateReporting        bool
	stateReportingSubjectPrefix string
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

type gpioChipOption struct {
	chip string
}

func (o *gpioChipOption) apply(c *config) {
	c.gpioChip = o.chip
}

func WithGPIOChip(chip string) Option {
	return &gpioChipOption{chip: chip}
}

type i2cDeviceOption struct {
	device string
}

func (o *i2cDeviceOption) apply(c *config) {
	c.i2cDevice = o.device
}

func WithI2CDevice(device string) Option {
	return &i2cDeviceOption{device: device}
}

type defaultBackendOption struct {
	backend BackendType
}

func (o *defaultBackendOption) apply(c *config) {
	c.defaultBackend = o.backend
}

func WithDefaultBackend(backend BackendType) Option {
	return &defaultBackendOption{backend: backend}
}

type componentsOption struct {
	components map[string]ComponentConfig
}

func (o *componentsOption) apply(c *config) {
	if c.components == nil {
		c.components = make(map[string]ComponentConfig)
	}
	for name, config := range o.components {
		c.components[name] = config
	}
}

func WithComponents(components map[string]ComponentConfig) Option {
	return &componentsOption{components: components}
}

type enableHostManagementOption struct {
	enable bool
}

func (o *enableHostManagementOption) apply(c *config) {
	c.enableHostManagement = o.enable
}

func WithHostManagement(enable bool) Option {
	return &enableHostManagementOption{enable: enable}
}

type enableChassisManagementOption struct {
	enable bool
}

func (o *enableChassisManagementOption) apply(c *config) {
	c.enableChassisManagement = o.enable
}

func WithChassisManagement(enable bool) Option {
	return &enableChassisManagementOption{enable: enable}
}

type enableBMCManagementOption struct {
	enable bool
}

func (o *enableBMCManagementOption) apply(c *config) {
	c.enableBMCManagement = o.enable
}

func WithBMCManagement(enable bool) Option {
	return &enableBMCManagementOption{enable: enable}
}

type numHostsOption struct {
	num int
}

func (o *numHostsOption) apply(c *config) {
	c.numHosts = o.num
}

func WithNumHosts(num int) Option {
	return &numHostsOption{num: num}
}

type numChassisOption struct {
	num int
}

func (o *numChassisOption) apply(c *config) {
	c.numChassis = o.num
}

func WithNumChassis(num int) Option {
	return &numChassisOption{num: num}
}

type defaultOperationTimeoutOption struct {
	timeout time.Duration
}

func (o *defaultOperationTimeoutOption) apply(c *config) {
	c.defaultOperationTimeout = o.timeout
}

func WithDefaultOperationTimeout(timeout time.Duration) Option {
	return &defaultOperationTimeoutOption{timeout: timeout}
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

type enableTracingOption struct {
	enable bool
}

func (o *enableTracingOption) apply(c *config) {
	c.enableTracing = o.enable
}

func WithTracing(enable bool) Option {
	return &enableTracingOption{enable: enable}
}

type enableStateReportingOption struct {
	enable bool
}

func (o *enableStateReportingOption) apply(c *config) {
	c.enableStateReporting = o.enable
}

func WithStateReporting(enable bool) Option {
	return &enableStateReportingOption{enable: enable}
}

type stateReportingSubjectPrefixOption struct {
	prefix string
}

func (o *stateReportingSubjectPrefixOption) apply(c *config) {
	c.stateReportingSubjectPrefix = o.prefix
}

func WithStateReportingSubjectPrefix(prefix string) Option {
	return &stateReportingSubjectPrefixOption{prefix: prefix}
}

func (c *config) Validate() error {
	if c.serviceName == "" {
		return fmt.Errorf("%w: service name cannot be empty", ErrInvalidConfiguration)
	}

	if c.serviceVersion == "" {
		return fmt.Errorf("%w: service version cannot be empty", ErrInvalidConfiguration)
	}

	if c.defaultBackend == BackendTypeGPIO && c.gpioChip == "" {
		return fmt.Errorf("%w: GPIO chip path cannot be empty", ErrInvalidConfiguration)
	}

	if c.defaultBackend == BackendTypeGPIO && !strings.HasPrefix(c.gpioChip, "/dev/gpiochip") {
		return fmt.Errorf("%w: GPIO chip path must start with '/dev/gpiochip'", ErrInvalidConfiguration)
	}

	if c.defaultBackend == BackendTypeI2C && c.i2cDevice == "" {
		return fmt.Errorf("%w: I2C device path cannot be empty", ErrInvalidConfiguration)
	}

	if c.defaultBackend == BackendTypeI2C && !strings.HasPrefix(c.i2cDevice, "/dev/i2c") {
		return fmt.Errorf("%w: I2C device path must start with '/dev/i2c'", ErrInvalidConfiguration)
	}

	if !c.enableHostManagement && !c.enableChassisManagement && !c.enableBMCManagement {
		return fmt.Errorf("%w: at least one component type must be enabled", ErrInvalidConfiguration)
	}

	if c.enableHostManagement && c.numHosts <= 0 {
		return fmt.Errorf("%w: number of hosts must be positive when host management is enabled", ErrInvalidConfiguration)
	}

	if c.enableChassisManagement && c.numChassis <= 0 {
		return fmt.Errorf("%w: number of chassis must be positive when chassis management is enabled", ErrInvalidConfiguration)
	}

	if c.defaultOperationTimeout <= 0 {
		return fmt.Errorf("%w: default operation timeout must be positive", ErrInvalidConfiguration)
	}

	for name, component := range c.components {
		if err := c.validateComponentConfig(name, component); err != nil {
			return err
		}
	}

	return nil
}

func (c *config) validateComponentConfig(name string, component ComponentConfig) error {
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

	if component.Backend != BackendTypeGPIO && component.Backend != BackendTypeI2C {
		return fmt.Errorf("%w: invalid backend type '%s' for component '%s'", ErrInvalidConfiguration, component.Backend, name)
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

	if component.Backend == BackendTypeGPIO {
		if err := c.validateGPIOConfig(name, component.GPIO); err != nil {
			return err
		}
	}

	if component.Backend == BackendTypeI2C {
		if err := c.validateI2CConfig(name, component.I2C); err != nil {
			return err
		}
	}

	return nil
}

func (c *config) validateGPIOConfig(componentName string, gpioConfig GPIOConfig) error {
	lines := map[string]GPIOLineConfig{
		"power_button": gpioConfig.PowerButton,
		"reset_button": gpioConfig.ResetButton,
		"power_status": gpioConfig.PowerStatus,
	}

	for lineName, lineConfig := range lines {
		if lineConfig.Line == "" {
			continue
		}

		if lineConfig.InitialValue < 0 || lineConfig.InitialValue > 1 {
			return fmt.Errorf("%w: initial value for GPIO line '%s' of component '%s' must be 0 or 1", ErrInvalidGPIOConfiguration, lineName, componentName)
		}

		if lineConfig.Direction == DirectionOutput && lineName == "power_status" {
			return fmt.Errorf("%w: power status GPIO line for component '%s' must be input", ErrInvalidGPIOConfiguration, componentName)
		}

		if lineConfig.Direction == DirectionInput && (lineName == "power_button" || lineName == "reset_button") {
			return fmt.Errorf("%w: control GPIO line '%s' for component '%s' must be output", ErrInvalidGPIOConfiguration, lineName, componentName)
		}
	}

	return nil
}

func (c *config) validateI2CConfig(componentName string, i2cConfig I2CConfig) error {
	if i2cConfig.DevicePath == "" {
		return fmt.Errorf("%w: I2C device path cannot be empty for component '%s'", ErrInvalidI2CConfiguration, componentName)
	}

	if !strings.HasPrefix(i2cConfig.DevicePath, "/dev/i2c") {
		return fmt.Errorf("%w: I2C device path must start with '/dev/i2c' for component '%s'", ErrInvalidI2CConfiguration, componentName)
	}

	if i2cConfig.SlaveAddress == 0 {
		return fmt.Errorf("%w: I2C slave address cannot be zero for component '%s'", ErrInvalidI2CConfiguration, componentName)
	}

	return nil
}

func (c *config) GetComponentConfig(name string) (ComponentConfig, bool) {
	config, exists := c.components[name]
	return config, exists
}

func (c *config) GetHostConfig(index int) (ComponentConfig, bool) {
	name := fmt.Sprintf("host.%d", index)
	return c.GetComponentConfig(name)
}

func (c *config) GetChassisConfig(index int) (ComponentConfig, bool) {
	name := fmt.Sprintf("chassis.%d", index)
	return c.GetComponentConfig(name)
}

func (c *config) GetBMCConfig() (ComponentConfig, bool) {
	return c.GetComponentConfig("bmc.0")
}

func (c *config) AddDefaultComponents() {
	if c.components == nil {
		c.components = make(map[string]ComponentConfig)
	}

	if c.enableHostManagement {
		for i := 0; i < c.numHosts; i++ {
			name := fmt.Sprintf("host.%d", i)
			if _, exists := c.components[name]; !exists {
				c.components[name] = newDefaultComponentConfig(name, "host", c.defaultBackend, c.gpioChip, c.i2cDevice)
			}
		}
	}

	if c.enableChassisManagement {
		for i := 0; i < c.numChassis; i++ {
			name := fmt.Sprintf("chassis.%d", i)
			if _, exists := c.components[name]; !exists {
				c.components[name] = newDefaultComponentConfig(name, "chassis", c.defaultBackend, c.gpioChip, c.i2cDevice)
			}
		}
	}

	if c.enableBMCManagement {
		name := "bmc.0"
		if _, exists := c.components[name]; !exists {
			c.components[name] = newDefaultComponentConfig(name, "bmc", c.defaultBackend, c.gpioChip, c.i2cDevice)
		}
	}
}

func newDefaultComponentConfig(name, componentType string, backend BackendType, gpioChip, i2cDevice string) ComponentConfig {
	config := ComponentConfig{
		Name:             name,
		Type:             componentType,
		Enabled:          true,
		Backend:          backend,
		OperationTimeout: DefaultOperationTimeout,
		PowerOnDelay:     DefaultPowerOnDelay,
		PowerOffDelay:    DefaultPowerOffDelay,
		ResetDelay:       DefaultResetDelay,
		ForceOffDelay:    DefaultForceOffDelay,
	}

	switch backend {
	case BackendTypeGPIO:
		config.GPIO = newDefaultGPIOConfig()
	case BackendTypeI2C:
		config.I2C = newDefaultI2CConfig(i2cDevice)
	}

	return config
}

func newDefaultGPIOConfig() GPIOConfig {
	return GPIOConfig{
		PowerButton: GPIOLineConfig{
			Direction:    DirectionOutput,
			ActiveState:  ActiveLow,
			InitialValue: 0,
			Bias:         BiasDisabled,
		},
		ResetButton: GPIOLineConfig{
			Direction:    DirectionOutput,
			ActiveState:  ActiveLow,
			InitialValue: 0,
			Bias:         BiasDisabled,
		},
		PowerStatus: GPIOLineConfig{
			Direction:   DirectionInput,
			ActiveState: ActiveHigh,
			Bias:        BiasPullDown,
		},
	}
}

func newDefaultI2CConfig(devicePath string) I2CConfig {
	return I2CConfig{
		DevicePath:    devicePath,
		SlaveAddress:  0x20,
		PowerOnReg:    0x01,
		PowerOffReg:   0x02,
		ResetReg:      0x03,
		StatusReg:     0x04,
		PowerOnValue:  0x01,
		PowerOffValue: 0x00,
		ResetValue:    0x01,
	}
}
