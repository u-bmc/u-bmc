// SPDX-License-Identifier: BSD-3-Clause

//nolint:gosec
package i2c

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// Linux I2C subsystem constants.
const (
	// ioctl commands.
	i2cSlave      = 0x0703 // Use this slave address
	i2cSlaveForce = 0x0706 // Use this slave address, even if busy
	i2cTenBit     = 0x0704 // Use 10-bit addresses
	i2cFuncs      = 0x0705 // Get the adapter functionality mask
	i2cRdwr       = 0x0707 // Combined R/W transfer (one STOP only)
	i2cPEC        = 0x0708 // != 0 to use PEC with SMBus
	i2cSMBus      = 0x0720 // SMBus transfer
	i2cTimeout    = 0x0702 // Set timeout in units of 10 ms
	i2cRetries    = 0x0701 // Set number of retries

	// SMBus transaction types.
	i2cSMBusWrite = 0
	i2cSMBusRead  = 1

	// SMBus data sizes.
	i2cSMBusQuick         = 0
	i2cSMBusByte          = 1
	i2cSMBusByteData      = 2
	i2cSMBusWordData      = 3
	i2cSMBusProcCall      = 4
	i2cSMBusBlockData     = 5
	i2cSMBusI2CBlockData  = 8
	i2cSMBusBlockProcCall = 7
	i2cSMBusBlockMax      = 32

	// I2C functionality flags.
	i2cFuncI2C                    = 0x00000001
	i2cFuncTenBitAddr             = 0x00000002
	i2cFuncProtocolMangling       = 0x00000004
	i2cFuncSMBusPEC               = 0x00000008
	i2cFuncNoStart                = 0x00000010
	i2cFuncSlaveAck               = 0x00000020
	i2cFuncSMBusBlockProcCall     = 0x00008000
	i2cFuncSMBusQuick             = 0x00010000
	i2cFuncSMBusReadByte          = 0x00020000
	i2cFuncSMBusWriteByte         = 0x00040000
	i2cFuncSMBusReadByteData      = 0x00080000
	i2cFuncSMBusWriteByteData     = 0x00100000
	i2cFuncSMBusReadWordData      = 0x00200000
	i2cFuncSMBusWriteWordData     = 0x00400000
	i2cFuncSMBusReadBlockData     = 0x01000000
	i2cFuncSMBusWriteBlockData    = 0x02000000
	i2cFuncSMBusReadI2CBlockData  = 0x04000000
	i2cFuncSMBusWriteI2CBlockData = 0x08000000
)

// Conn represents a connection to an I2C device.
type Conn struct {
	file         *os.File
	config       *Config
	capabilities uint32
	currentAddr  uint16
	addrSet      bool
}

// i2cMsg represents an I2C message for combined transactions.
type i2cMsg struct {
	addr  uint16
	flags uint16
	len   uint16
	buf   uintptr
}

// i2cRdwrIoctlData represents the data structure for I2C_RDWR ioctl.
type i2cRdwrIoctlData struct {
	msgs  uintptr
	nmsgs uint32
}

// i2cSMBusData represents SMBus data for ioctl operations.
type i2cSMBusData struct {
	byte0 uint8
	word  uint16
	block [i2cSMBusBlockMax + 2]uint8
}

// i2cSMBusIoctlData represents the data structure for SMBus ioctl.
type i2cSMBusIoctlData struct {
	readWrite uint8
	command   uint8
	size      uint32
	data      uintptr
}

// Open opens a connection to an I2C device using the provided configuration.
func Open(cfg *Config) (*Conn, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}

	devicePath := cfg.GetDevicePath()
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrBusNotFound, devicePath)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("%w: %s", ErrBusAccessDenied, devicePath)
		}
		return nil, fmt.Errorf("%w: %s: %w", ErrBusOpenFailed, devicePath, err)
	}

	conn := &Conn{
		file:   file,
		config: cfg,
	}

	// Get adapter capabilities
	if err := conn.getCapabilities(); err != nil {
		_ = conn.file.Close()
		return nil, fmt.Errorf("failed to get adapter capabilities: %w", err)
	}

	// Configure the connection
	if err := conn.configure(); err != nil {
		_ = conn.file.Close()
		return nil, fmt.Errorf("failed to configure connection: %w", err)
	}

	// Set device address
	if err := conn.SetAddress(cfg.Address); err != nil {
		_ = conn.file.Close()
		return nil, fmt.Errorf("failed to set device address: %w", err)
	}

	return conn, nil
}

// OpenBus opens a connection to an I2C bus without setting a device address.
// This is useful for scanning the bus or when you need to communicate with multiple devices.
func OpenBus(bus int) (*Conn, error) {
	cfg := NewConfig(WithBus(bus))

	devicePath := cfg.GetDevicePath()
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrBusNotFound, devicePath)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("%w: %s", ErrBusAccessDenied, devicePath)
		}
		return nil, fmt.Errorf("%w: %s: %w", ErrBusOpenFailed, devicePath, err)
	}

	conn := &Conn{
		file:   file,
		config: cfg,
	}

	// Get adapter capabilities
	if err := conn.getCapabilities(); err != nil {
		_ = conn.file.Close()
		return nil, fmt.Errorf("failed to get adapter capabilities: %w", err)
	}

	// Configure the connection (without setting address)
	if err := conn.configure(); err != nil {
		_ = conn.file.Close()
		return nil, fmt.Errorf("failed to configure connection: %w", err)
	}

	return conn, nil
}

// Close closes the connection to the I2C device.
func (c *Conn) Close() error {
	if c.file == nil {
		return nil
	}

	err := c.file.Close()
	c.file = nil
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBusCloseFailed, err)
	}
	return nil
}

// SetAddress sets the I2C device address for subsequent operations.
func (c *Conn) SetAddress(addr uint16) error {
	if !c.config.IsValidAddress(addr) {
		return fmt.Errorf("%w: 0x%02x", ErrInvalidAddress, addr)
	}

	// Skip if address is already set
	if c.addrSet && c.currentAddr == addr {
		return nil
	}

	var ioctlCmd uintptr = i2cSlave
	if c.config.ForceAddress {
		ioctlCmd = i2cSlaveForce
	}

	if err := c.ioctl(ioctlCmd, uintptr(addr)); err != nil {
		return fmt.Errorf("%w: failed to set address 0x%02x: %w", ErrDeviceNotResponding, addr, err)
	}

	c.currentAddr = addr
	c.addrSet = true
	return nil
}

// Read reads data from the I2C device into the provided buffer.
func (c *Conn) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	var errs []error //nolint:prealloc
	for attempt := range c.config.Retries {
		if attempt > 0 {
			time.Sleep(10 * time.Millisecond) // Brief delay between retries
		}
		n, err := c.file.Read(buf)
		if err == nil {
			return n, nil
		}
		errs = append(errs, err)
	}

	return 0, fmt.Errorf("%w: %w", ErrReadFailed, errors.Join(errs...))
}

// Write writes data to the I2C device.
func (c *Conn) Write(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	var errs []error //nolint:prealloc
	for attempt := 0; attempt <= c.config.Retries; attempt++ {
		if attempt > 0 {
			time.Sleep(10 * time.Millisecond) // Brief delay between retries
		}
		n, err := c.file.Write(buf)
		if err == nil {
			return n, nil
		}
		errs = append(errs, err)
	}

	return 0, fmt.Errorf("%w: %w", ErrWriteFailed, errors.Join(errs...))
}

// WriteByte writes a single byte to the I2C device.
func (c *Conn) WriteByte(b byte) error {
	_, err := c.Write([]byte{b})
	return err
}

// ReadByte reads a single byte from the I2C device.
func (c *Conn) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	if _, err := c.Read(buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

// Transaction performs a combined write-then-read transaction.
// This is useful for register-based devices where you write a register address
// and then read the data without releasing the bus.
func (c *Conn) Transaction(writeData []byte, readBuf []byte) error {
	// Check if at least one operation is requested
	if len(writeData) == 0 && len(readBuf) == 0 {
		return fmt.Errorf("%w: no data to read or write", ErrInvalidDataLength)
	}

	if !c.supportsI2C() {
		return ErrSMBusNotSupported
	}

	msgs := make([]i2cMsg, 0, 2)

	// Add write message if writeData is provided
	if len(writeData) > 0 {
		msgs = append(msgs, i2cMsg{
			addr:  c.currentAddr,
			flags: 0, // Write
			len:   uint16(len(writeData)),
			buf:   uintptr(unsafe.Pointer(&writeData[0])),
		})
	}

	// Add read message if readBuf is provided
	if len(readBuf) > 0 {
		msgs = append(msgs, i2cMsg{
			addr:  c.currentAddr,
			flags: 1, // Read
			len:   uint16(len(readBuf)),
			buf:   uintptr(unsafe.Pointer(&readBuf[0])),
		})
	}

	data := i2cRdwrIoctlData{
		msgs:  uintptr(unsafe.Pointer(&msgs[0])),
		nmsgs: uint32(len(msgs)),
	}

	var lastErr error
	for attempt := range c.config.Retries {
		if attempt > 0 {
			time.Sleep(10 * time.Millisecond)
		}

		if err := c.ioctl(i2cRdwr, uintptr(unsafe.Pointer(&data))); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("%w: %w", ErrTransactionFailed, lastErr)
}

// GetCapabilities returns the I2C adapter capabilities.
func (c *Conn) GetCapabilities() uint32 {
	return c.capabilities
}

// SupportsProtocol checks if the adapter supports the specified protocol.
func (c *Conn) SupportsProtocol(protocol Protocol) bool {
	switch protocol {
	case ProtocolI2C:
		return c.supportsI2C()
	case ProtocolSMBus:
		return c.supportsSMBus()
	case ProtocolPMBus:
		return c.supportsSMBus() // PMBus is built on SMBus
	case ProtocolI3C:
		// I3C support would require additional kernel support
		return false
	default:
		return false
	}
}

// Config returns a copy of the connection configuration.
func (c *Conn) Config() Config {
	return *c.config
}

// IsConnected returns true if the connection is still valid.
func (c *Conn) IsConnected() bool {
	return c.file != nil
}

// configure sets up the I2C connection based on the configuration.
func (c *Conn) configure() error {
	// Set timeout (in units of 10ms)
	timeoutUnits := int(c.config.Timeout.Milliseconds() / 10)
	timeoutUnits = max(timeoutUnits, 1) // Minimum 1 unit (10ms)
	if err := c.ioctl(i2cTimeout, uintptr(timeoutUnits)); err != nil {
		return fmt.Errorf("failed to set timeout: %w", err)
	}

	// Set retry count
	if err := c.ioctl(i2cRetries, uintptr(c.config.Retries)); err != nil {
		return fmt.Errorf("failed to set retries: %w", err)
	}

	// Enable 10-bit addressing if configured
	if c.config.AddressMode == AddressMode10Bit || c.config.Use10BitAddress {
		if !c.supports10BitAddr() {
			return fmt.Errorf("%w: 10-bit addressing not supported by adapter", ErrOperationFailed)
		}
		if err := c.ioctl(i2cTenBit, 1); err != nil {
			return fmt.Errorf("failed to enable 10-bit addressing: %w", err)
		}
	}

	// Enable PEC for SMBus if configured
	if c.config.PEC && (c.config.Protocol == ProtocolSMBus || c.config.Protocol == ProtocolPMBus) {
		if !c.supportsPEC() {
			return ErrPECNotSupported
		}
		if err := c.ioctl(i2cPEC, 1); err != nil {
			return fmt.Errorf("failed to enable PEC: %w", err)
		}
	}

	return nil
}

// getCapabilities retrieves the I2C adapter capabilities.
func (c *Conn) getCapabilities() error {
	var funcs uint32
	if err := c.ioctl(i2cFuncs, uintptr(unsafe.Pointer(&funcs))); err != nil {
		return fmt.Errorf("failed to get adapter capabilities: %w", err)
	}
	c.capabilities = funcs
	return nil
}

// supportsI2C checks if the adapter supports basic I2C operations.
func (c *Conn) supportsI2C() bool {
	return c.capabilities&i2cFuncI2C != 0
}

// supportsSMBus checks if the adapter supports SMBus operations.
func (c *Conn) supportsSMBus() bool {
	required := uint32(i2cFuncSMBusReadByte | i2cFuncSMBusWriteByte |
		i2cFuncSMBusReadByteData | i2cFuncSMBusWriteByteData)
	return c.capabilities&required == required
}

// supports10BitAddr checks if the adapter supports 10-bit addressing.
func (c *Conn) supports10BitAddr() bool {
	return c.capabilities&i2cFuncTenBitAddr != 0
}

// supportsPEC checks if the adapter supports Packet Error Checking.
func (c *Conn) supportsPEC() bool {
	return c.capabilities&i2cFuncSMBusPEC != 0
}

// ioctl performs an ioctl system call on the I2C device file.
func (c *Conn) ioctl(cmd, arg uintptr) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, c.file.Fd(), cmd, arg); errno != 0 {
		return errno
	}
	return nil
}

// smbusAccess performs an SMBus access operation.
func (c *Conn) smbusAccess(readWrite uint8, command uint8, size uint32, data *i2cSMBusData) error {
	args := i2cSMBusIoctlData{
		readWrite: readWrite,
		command:   command,
		size:      size,
		data:      uintptr(unsafe.Pointer(data)),
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.Retries; attempt++ {
		if attempt > 0 {
			time.Sleep(10 * time.Millisecond)
		}

		if err := c.ioctl(i2cSMBus, uintptr(unsafe.Pointer(&args))); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("%w: %w", ErrOperationFailed, lastErr)
}
