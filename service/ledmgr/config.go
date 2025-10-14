// SPDX-License-Identifier: BSD-3-Clause

package ledmgr

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultServiceName        = "ledmgr"
	DefaultServiceDescription = "LED management service for BMC components"
	DefaultServiceVersion     = "1.0.0"
	DefaultGPIOChip           = "/dev/gpiochip0"
	DefaultI2CDevice          = "/dev/i2c-0"
	DefaultOperationTimeout   = 5 * time.Second
	DefaultBlinkInterval      = 500 * time.Millisecond
)

type BackendType string

const (
	BackendTypeGPIO BackendType = "gpio"
	BackendTypeI2C  BackendType = "i2c"
)

type LEDType string

const (
	LEDTypePower    LEDType = "power"
	LEDTypeStatus   LEDType = "status"
	LEDTypeError    LEDType = "error"
	LEDTypeIdentify LEDType = "identify"
)

type LEDState string

const (
	LEDStateOff       LEDState = "off"
	LEDStateOn        LEDState = "on"
	LEDStateBlink     LEDState = "blink"
	LEDStateFastBlink LEDState = "fast_blink"
)

type GPIOActiveState int

const (
	ActiveHigh GPIOActiveState = iota
	ActiveLow
)

type LEDGPIOConfig struct {
	Line        string
	ActiveState GPIOActiveState
}

type LEDI2CConfig struct {
	DevicePath   string
	SlaveAddress uint8
	Register     uint8
	OnValue      uint8
	OffValue     uint8
	BlinkValue   uint8
}

type LEDConfig struct {
	Type    LEDType
	Enabled bool
	Backend BackendType
	GPIO    LEDGPIOConfig
	I2C     LEDI2CConfig
}

type ComponentConfig struct {
	Name             string
	Type             string
	Enabled          bool
	LEDs             map[LEDType]LEDConfig
	OperationTimeout time.Duration
	BlinkInterval    time.Duration
}

type config struct {
	serviceName             string
	serviceDescription      string
	serviceVersion          string
	gpioChip                string
	i2cDevice               string
	defaultBackend          BackendType
	components              map[string]ComponentConfig
	enableHostManagement    bool
	enableChassisManagement bool
	enableBMCManagement     bool
	numHosts                int
	numChassis              int
	defaultOperationTimeout time.Duration
	defaultBlinkInterval    time.Duration
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

type defaultBlinkIntervalOption struct {
	interval time.Duration
}

func (o *defaultBlinkIntervalOption) apply(c *config) {
	c.defaultBlinkInterval = o.interval
}

func WithDefaultBlinkInterval(interval time.Duration) Option {
	return &defaultBlinkIntervalOption{interval: interval}
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

	if c.defaultBlinkInterval <= 0 {
		return fmt.Errorf("%w: default blink interval must be positive", ErrInvalidConfiguration)
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

	if component.OperationTimeout <= 0 {
		return fmt.Errorf("%w: operation timeout must be positive for component '%s'", ErrInvalidConfiguration, name)
	}

	if component.BlinkInterval <= 0 {
		return fmt.Errorf("%w: blink interval must be positive for component '%s'", ErrInvalidConfiguration, name)
	}

	for ledType, ledConfig := range component.LEDs {
		if err := c.validateLEDConfig(name, string(ledType), ledConfig); err != nil {
			return err
		}
	}

	return nil
}

func (c *config) validateLEDConfig(componentName, ledTypeName string, ledConfig LEDConfig) error {
	if ledConfig.Backend != BackendTypeGPIO && ledConfig.Backend != BackendTypeI2C {
		return fmt.Errorf("%w: invalid backend type '%s' for LED '%s' of component '%s'", ErrInvalidConfiguration, ledConfig.Backend, ledTypeName, componentName)
	}

	if ledConfig.Backend == BackendTypeGPIO {
		if ledConfig.GPIO.Line == "" {
			return fmt.Errorf("%w: GPIO line cannot be empty for LED '%s' of component '%s'", ErrInvalidGPIOConfiguration, ledTypeName, componentName)
		}
	}

	if ledConfig.Backend == BackendTypeI2C {
		if ledConfig.I2C.DevicePath == "" {
			return fmt.Errorf("%w: I2C device path cannot be empty for LED '%s' of component '%s'", ErrInvalidI2CConfiguration, ledTypeName, componentName)
		}

		if !strings.HasPrefix(ledConfig.I2C.DevicePath, "/dev/i2c") {
			return fmt.Errorf("%w: I2C device path must start with '/dev/i2c' for LED '%s' of component '%s'", ErrInvalidI2CConfiguration, ledTypeName, componentName)
		}

		if ledConfig.I2C.SlaveAddress == 0 {
			return fmt.Errorf("%w: I2C slave address cannot be zero for LED '%s' of component '%s'", ErrInvalidI2CConfiguration, ledTypeName, componentName)
		}
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
				c.components[name] = newDefaultComponentConfig(name, "host", c.defaultBackend, c.gpioChip, c.i2cDevice, c.defaultOperationTimeout, c.defaultBlinkInterval)
			}
		}
	}

	if c.enableChassisManagement {
		for i := 0; i < c.numChassis; i++ {
			name := fmt.Sprintf("chassis.%d", i)
			if _, exists := c.components[name]; !exists {
				c.components[name] = newDefaultComponentConfig(name, "chassis", c.defaultBackend, c.gpioChip, c.i2cDevice, c.defaultOperationTimeout, c.defaultBlinkInterval)
			}
		}
	}

	if c.enableBMCManagement {
		name := "bmc.0"
		if _, exists := c.components[name]; !exists {
			c.components[name] = newDefaultComponentConfig(name, "bmc", c.defaultBackend, c.gpioChip, c.i2cDevice, c.defaultOperationTimeout, c.defaultBlinkInterval)
		}
	}
}

func newDefaultComponentConfig(name, componentType string, backend BackendType, gpioChip, i2cDevice string, operationTimeout, blinkInterval time.Duration) ComponentConfig {
	config := ComponentConfig{
		Name:             name,
		Type:             componentType,
		Enabled:          true,
		LEDs:             make(map[LEDType]LEDConfig),
		OperationTimeout: operationTimeout,
		BlinkInterval:    blinkInterval,
	}

	ledTypes := []LEDType{LEDTypePower, LEDTypeStatus, LEDTypeError}
	if componentType == "host" || componentType == "chassis" {
		ledTypes = append(ledTypes, LEDTypeIdentify)
	}

	for i, ledType := range ledTypes {
		ledConfig := LEDConfig{
			Type:    ledType,
			Enabled: true,
			Backend: backend,
		}

		switch backend {
		case BackendTypeGPIO:
			ledConfig.GPIO = LEDGPIOConfig{
				Line:        fmt.Sprintf("%s-%s-led-%d", name, string(ledType), i),
				ActiveState: ActiveHigh,
			}
		case BackendTypeI2C:
			ledConfig.I2C = LEDI2CConfig{
				DevicePath:   i2cDevice,
				SlaveAddress: 0x20,
				Register:     uint8(0x10 + i), //nolint:gosec
				OnValue:      0xFF,
				OffValue:     0x00,
				BlinkValue:   0x55,
			}
		}

		config.LEDs[ledType] = ledConfig
	}

	return config
}
