// SPDX-License-Identifier: BSD-3-Clause

package i2c

import (
	"fmt"
	"time"
)

// Protocol represents the communication protocol to use.
type Protocol int

const (
	// ProtocolI2C uses standard I2C communication.
	ProtocolI2C Protocol = iota
	// ProtocolI3C uses I3C communication with backward I2C compatibility.
	ProtocolI3C
	// ProtocolSMBus uses SMBus communication (subset of I2C with additional specifications).
	ProtocolSMBus
	// ProtocolPMBus uses PMBus communication (extension of SMBus for power management).
	ProtocolPMBus
)

// PMBusFormat represents the data format used for PMBus communications.
type PMBusFormat int

const (
	// PMBusFormatLinear uses LINEAR11 or LINEAR16 format.
	PMBusFormatLinear PMBusFormat = iota
	// PMBusFormatDirect uses DIRECT format with coefficients.
	PMBusFormatDirect
)

// AddressMode represents the I2C addressing mode.
type AddressMode int

const (
	// AddressMode7Bit uses 7-bit I2C addressing (0x00-0x7F).
	AddressMode7Bit AddressMode = iota
	// AddressMode10Bit uses 10-bit I2C addressing (0x000-0x3FF).
	AddressMode10Bit
)

// Config holds the configuration for I2C/I3C/SMBus/PMBus communication.
type Config struct {
	// Bus is the I2C bus number (corresponds to /dev/i2c-N).
	Bus int
	// Address is the device address on the I2C bus.
	Address uint16
	// Protocol specifies which communication protocol to use.
	Protocol Protocol
	// AddressMode specifies whether to use 7-bit or 10-bit addressing.
	AddressMode AddressMode
	// ForceAddress uses I2C_SLAVE_FORCE instead of I2C_SLAVE when setting device address.
	// This bypasses the kernel's busy device check.
	ForceAddress bool
	// Timeout is the maximum time to wait for I2C operations.
	Timeout time.Duration
	// Retries is the number of times to retry failed operations.
	Retries int
	// PEC enables Packet Error Checking for SMBus operations.
	PEC bool
	// PMBusFormat specifies the data format for PMBus operations.
	PMBusFormat PMBusFormat
	// PMBusCoefficients holds the coefficients for DIRECT format PMBus operations.
	PMBusCoefficients *PMBusCoefficients
	// ClockFrequency is the desired I2C clock frequency in Hz.
	// Set to 0 to use adapter default.
	ClockFrequency uint32
	// Use10BitAddress forces 10-bit addressing mode.
	Use10BitAddress bool
}

// PMBusCoefficients holds the coefficients for DIRECT format PMBus calculations.
type PMBusCoefficients struct {
	// M is the slope coefficient.
	M int16
	// B is the offset coefficient.
	B int16
	// R is the exponent coefficient.
	R int8
}

// Option represents a configuration option for I2C communication.
type Option interface {
	apply(*Config)
}

type busOption struct {
	bus int
}

func (o *busOption) apply(c *Config) {
	c.Bus = o.bus
}

// WithBus sets the I2C bus number to use.
// The bus number corresponds to /dev/i2c-N where N is the bus number.
func WithBus(bus int) Option {
	return &busOption{
		bus: bus,
	}
}

type addressOption struct {
	address uint16
}

func (o *addressOption) apply(c *Config) {
	c.Address = o.address
}

// WithAddress sets the I2C device address.
// For 7-bit addressing, use values 0x00-0x7F.
// For 10-bit addressing, use values 0x000-0x3FF.
func WithAddress(address uint16) Option {
	return &addressOption{
		address: address,
	}
}

type protocolOption struct {
	protocol Protocol
}

func (o *protocolOption) apply(c *Config) {
	c.Protocol = o.protocol
}

// WithProtocol sets the communication protocol to use.
func WithProtocol(protocol Protocol) Option {
	return &protocolOption{
		protocol: protocol,
	}
}

type addressModeOption struct {
	mode AddressMode
}

func (o *addressModeOption) apply(c *Config) {
	c.AddressMode = o.mode
}

// WithAddressMode sets the I2C addressing mode (7-bit or 10-bit).
func WithAddressMode(mode AddressMode) Option {
	return &addressModeOption{
		mode: mode,
	}
}

type forceAddressOption struct {
	force bool
}

func (o *forceAddressOption) apply(c *Config) {
	c.ForceAddress = o.force
}

// WithForceAddress enables or disables forced address mode.
// When enabled, uses I2C_SLAVE_FORCE instead of I2C_SLAVE, bypassing
// the kernel's check for busy devices.
func WithForceAddress(force bool) Option {
	return &forceAddressOption{
		force: force,
	}
}

type timeoutOption struct {
	timeout time.Duration
}

func (o *timeoutOption) apply(c *Config) {
	c.Timeout = o.timeout
}

// WithTimeout sets the maximum time to wait for I2C operations.
func WithTimeout(timeout time.Duration) Option {
	return &timeoutOption{
		timeout: timeout,
	}
}

type retriesOption struct {
	retries int
}

func (o *retriesOption) apply(c *Config) {
	c.Retries = o.retries
}

// WithRetries sets the number of times to retry failed operations.
func WithRetries(retries int) Option {
	return &retriesOption{
		retries: retries,
	}
}

type pecOption struct {
	pec bool
}

func (o *pecOption) apply(c *Config) {
	c.PEC = o.pec
}

// WithPEC enables or disables Packet Error Checking for SMBus operations.
func WithPEC(pec bool) Option {
	return &pecOption{
		pec: pec,
	}
}

type pmbusFormatOption struct {
	format PMBusFormat
}

func (o *pmbusFormatOption) apply(c *Config) {
	c.PMBusFormat = o.format
}

// WithPMBusFormat sets the data format for PMBus operations.
func WithPMBusFormat(format PMBusFormat) Option {
	return &pmbusFormatOption{
		format: format,
	}
}

type pmbusCoefficientsOption struct {
	coefficients *PMBusCoefficients
}

func (o *pmbusCoefficientsOption) apply(c *Config) {
	c.PMBusCoefficients = o.coefficients
}

// WithPMBusCoefficients sets the coefficients for DIRECT format PMBus operations.
func WithPMBusCoefficients(m, b int16, r int8) Option {
	return &pmbusCoefficientsOption{
		coefficients: &PMBusCoefficients{
			M: m,
			B: b,
			R: r,
		},
	}
}

type clockFrequencyOption struct {
	frequency uint32
}

func (o *clockFrequencyOption) apply(c *Config) {
	c.ClockFrequency = o.frequency
}

// WithClockFrequency sets the desired I2C clock frequency in Hz.
// Set to 0 to use the adapter's default frequency.
// Common values: 100000 (100kHz), 400000 (400kHz), 1000000 (1MHz).
func WithClockFrequency(frequency uint32) Option {
	return &clockFrequencyOption{
		frequency: frequency,
	}
}

type use10BitAddressOption struct {
	use10Bit bool
}

func (o *use10BitAddressOption) apply(c *Config) {
	c.Use10BitAddress = o.use10Bit
}

// WithUse10BitAddress forces 10-bit addressing mode.
// This is an alternative to WithAddressMode for backward compatibility.
func WithUse10BitAddress(use10Bit bool) Option {
	return &use10BitAddressOption{
		use10Bit: use10Bit,
	}
}

// NewConfig creates a new Config with default values and applies the provided options.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Bus:               0,
		Address:           0x00,
		Protocol:          ProtocolI2C,
		AddressMode:       AddressMode7Bit,
		ForceAddress:      false,
		Timeout:           1 * time.Second,
		Retries:           3,
		PEC:               false,
		PMBusFormat:       PMBusFormatLinear,
		PMBusCoefficients: nil,
		ClockFrequency:    0, // Use adapter default
		Use10BitAddress:   false,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	// Validate bus number
	if c.Bus < 0 {
		return fmt.Errorf("%w: bus number cannot be negative", ErrInvalidBusNumber)
	}

	// Validate address based on addressing mode
	if c.AddressMode == AddressMode7Bit || !c.Use10BitAddress {
		if c.Address > 0x7F {
			return fmt.Errorf("%w: address 0x%02x exceeds 7-bit limit (0x7F)", ErrInvalidAddress, c.Address)
		}
		// Check reserved addresses for 7-bit mode
		if c.Address < 0x08 || c.Address > 0x77 {
			return fmt.Errorf("%w: address 0x%02x is in reserved range", ErrInvalidAddress, c.Address)
		}
	} else {
		if c.Address > 0x3FF {
			return fmt.Errorf("%w: address 0x%03x exceeds 10-bit limit (0x3FF)", ErrInvalidAddress, c.Address)
		}
	}

	// Validate timeout
	if c.Timeout <= 0 {
		return fmt.Errorf("%w: timeout must be positive", ErrInvalidTimeout)
	}

	// Validate retries
	if c.Retries < 0 {
		return fmt.Errorf("%w: retries cannot be negative", ErrInvalidRetryCount)
	}

	// Validate protocol-specific options
	switch c.Protocol {
	case ProtocolSMBus:
		// SMBus-specific validations
		if c.PEC && c.Address == 0x00 {
			return fmt.Errorf("%w: PEC not supported for general call address", ErrInvalidConfig)
		}
	case ProtocolPMBus:
		// PMBus-specific validations
		if c.PMBusFormat == PMBusFormatDirect && c.PMBusCoefficients == nil {
			return fmt.Errorf("%w: coefficients required for DIRECT format", ErrInvalidConfig)
		}
		if c.PMBusCoefficients != nil {
			if c.PMBusCoefficients.R < -15 || c.PMBusCoefficients.R > 15 {
				return fmt.Errorf("%w: R coefficient must be between -15 and 15", ErrPMBusCoefficientsInvalid)
			}
		}
	case ProtocolI3C:
		// I3C-specific validations
		if c.AddressMode == AddressMode10Bit {
			return fmt.Errorf("%w: I3C does not support 10-bit addressing", ErrInvalidConfig)
		}
	case ProtocolI2C:
		// Standard I2C validations
		// No additional validations needed
	default:
		return fmt.Errorf("%w: unsupported protocol", ErrInvalidProtocol)
	}

	// Validate clock frequency
	if c.ClockFrequency > 0 {
		// Check reasonable frequency limits
		if c.ClockFrequency < 1000 { // 1kHz minimum
			return fmt.Errorf("%w: clock frequency too low (minimum 1kHz)", ErrInvalidConfig)
		}
		if c.ClockFrequency > 5000000 { // 5MHz maximum
			return fmt.Errorf("%w: clock frequency too high (maximum 5MHz)", ErrInvalidConfig)
		}
	}

	return nil
}

// IsValidAddress checks if an address is valid for the current configuration.
func (c *Config) IsValidAddress(addr uint16) bool {
	if c.AddressMode == AddressMode7Bit || !c.Use10BitAddress {
		// Check if it's within 7-bit range
		if addr > 0x7F {
			return false
		}
		// Check if it's not in reserved range (but allow for special cases)
		return (addr >= 0x08 && addr <= 0x77) ||
			c.ForceAddress // Allow reserved addresses with ForceAddress
	}
	return addr <= 0x3FF
}

// SupportsProtocol checks if the configuration supports a specific protocol.
func (c *Config) SupportsProtocol(protocol Protocol) bool {
	switch protocol {
	case ProtocolI2C:
		return true // Always supported
	case ProtocolI3C:
		return c.AddressMode == AddressMode7Bit
	case ProtocolSMBus:
		return true // Generally supported if I2C is supported
	case ProtocolPMBus:
		return true // PMBus is built on SMBus
	default:
		return false
	}
}

// GetDevicePath returns the device path for the configured bus.
func (c *Config) GetDevicePath() string {
	return fmt.Sprintf("/dev/i2c-%d", c.Bus)
}

// String returns a string representation of the configuration.
func (c *Config) String() string {
	protocol := "I2C"
	switch c.Protocol {
	case ProtocolI3C:
		protocol = "I3C"
	case ProtocolSMBus:
		protocol = "SMBus"
	case ProtocolPMBus:
		protocol = "PMBus"
	}

	addressMode := "7-bit"
	if c.AddressMode == AddressMode10Bit || c.Use10BitAddress {
		addressMode = "10-bit"
	}

	return fmt.Sprintf("%s bus=%d addr=0x%02x mode=%s timeout=%v retries=%d",
		protocol, c.Bus, c.Address, addressMode, c.Timeout, c.Retries)
}
