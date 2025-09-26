// SPDX-License-Identifier: BSD-3-Clause

package i2c

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// I2C ioctl commands.
const (
	I2C_SLAVE = 0x0703
	I2C_SMBUS = 0x0720
)

// SMBus transaction types.
const (
	I2C_SMBUS_WRITE = 0
	I2C_SMBUS_READ  = 1
)

// SMBus protocols.
const (
	I2C_SMBUS_QUICK            = 0
	I2C_SMBUS_BYTE             = 1
	I2C_SMBUS_BYTE_DATA        = 2
	I2C_SMBUS_WORD_DATA        = 3
	I2C_SMBUS_PROC_CALL        = 4
	I2C_SMBUS_BLOCK_DATA       = 5
	I2C_SMBUS_I2C_BLOCK_BROKEN = 6
	I2C_SMBUS_BLOCK_PROC_CALL  = 7
	I2C_SMBUS_I2C_BLOCK_DATA   = 8
)

// smbusIoctlData represents the data structure for SMBus ioctl operations.
type smbusIoctlData struct {
	readWrite uint8
	command   uint8
	size      uint32
	data      uintptr
}

// WriteRegister writes a single byte value to the specified register of an I2C device.
//
// Parameters:
//   - devicePath: Path to the I2C device (e.g., "/dev/i2c-0")
//   - slaveAddr: I2C slave address of the target device
//   - register: Register address to write to
//   - value: Byte value to write
//
// Returns an error if the operation fails.
func WriteRegister(devicePath string, slaveAddr uint8, register uint8, value uint8) error {
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open I2C device %s: %w", devicePath, err)
	}
	defer file.Close()

	// Set slave address
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SLAVE, uintptr(slaveAddr))
	if errno != 0 {
		return fmt.Errorf("failed to set I2C slave address 0x%02x: %w", slaveAddr, errno)
	}

	// Prepare SMBus data
	data := smbusIoctlData{
		readWrite: I2C_SMBUS_WRITE,
		command:   register,
		size:      I2C_SMBUS_BYTE_DATA,
		data:      uintptr(unsafe.Pointer(&value)),
	}

	// Perform SMBus write
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&data)))
	if errno != 0 {
		return fmt.Errorf("failed to write I2C register 0x%02x: %w", register, errno)
	}

	return nil
}

// ReadRegister reads a single byte value from the specified register of an I2C device.
//
// Parameters:
//   - devicePath: Path to the I2C device (e.g., "/dev/i2c-0")
//   - slaveAddr: I2C slave address of the target device
//   - register: Register address to read from
//
// Returns the byte value read from the register and any error encountered.
func ReadRegister(devicePath string, slaveAddr uint8, register uint8) (uint8, error) {
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to open I2C device %s: %w", devicePath, err)
	}
	defer file.Close()

	// Set slave address
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SLAVE, uintptr(slaveAddr))
	if errno != 0 {
		return 0, fmt.Errorf("failed to set I2C slave address 0x%02x: %w", slaveAddr, errno)
	}

	var value uint8

	// Prepare SMBus data
	data := smbusIoctlData{
		readWrite: I2C_SMBUS_READ,
		command:   register,
		size:      I2C_SMBUS_BYTE_DATA,
		data:      uintptr(unsafe.Pointer(&value)),
	}

	// Perform SMBus read
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&data)))
	if errno != 0 {
		return 0, fmt.Errorf("failed to read I2C register 0x%02x: %w", register, errno)
	}

	return value, nil
}

// WriteByte writes a single byte to an I2C device without specifying a register.
//
// Parameters:
//   - devicePath: Path to the I2C device (e.g., "/dev/i2c-0")
//   - slaveAddr: I2C slave address of the target device
//   - value: Byte value to write
//
// Returns an error if the operation fails.
func WriteByte(devicePath string, slaveAddr uint8, value uint8) error {
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open I2C device %s: %w", devicePath, err)
	}
	defer file.Close()

	// Set slave address
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SLAVE, uintptr(slaveAddr))
	if errno != 0 {
		return fmt.Errorf("failed to set I2C slave address 0x%02x: %w", slaveAddr, errno)
	}

	// Write single byte
	_, err = file.Write([]byte{value})
	if err != nil {
		return fmt.Errorf("failed to write byte 0x%02x to I2C device: %w", value, err)
	}

	return nil
}

// ReadByte reads a single byte from an I2C device without specifying a register.
//
// Parameters:
//   - devicePath: Path to the I2C device (e.g., "/dev/i2c-0")
//   - slaveAddr: I2C slave address of the target device
//
// Returns the byte value read from the device and any error encountered.
func ReadByte(devicePath string, slaveAddr uint8) (uint8, error) {
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to open I2C device %s: %w", devicePath, err)
	}
	defer file.Close()

	// Set slave address
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SLAVE, uintptr(slaveAddr))
	if errno != 0 {
		return 0, fmt.Errorf("failed to set I2C slave address 0x%02x: %w", slaveAddr, errno)
	}

	// Read single byte
	buf := make([]byte, 1)
	_, err = file.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("failed to read byte from I2C device: %w", err)
	}

	return buf[0], nil
}

// WriteBlock writes a block of data to the specified register of an I2C device.
//
// Parameters:
//   - devicePath: Path to the I2C device (e.g., "/dev/i2c-0")
//   - slaveAddr: I2C slave address of the target device
//   - register: Register address to write to
//   - data: Block of data to write
//
// Returns an error if the operation fails.
func WriteBlock(devicePath string, slaveAddr uint8, register uint8, data []byte) error {
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open I2C device %s: %w", devicePath, err)
	}
	defer file.Close()

	// Set slave address
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SLAVE, uintptr(slaveAddr))
	if errno != 0 {
		return fmt.Errorf("failed to set I2C slave address 0x%02x: %w", slaveAddr, errno)
	}

	// Prepare write buffer with register address followed by data
	writeData := make([]byte, len(data)+1)
	writeData[0] = register
	copy(writeData[1:], data)

	// Write data
	_, err = file.Write(writeData)
	if err != nil {
		return fmt.Errorf("failed to write block to I2C register 0x%02x: %w", register, err)
	}

	return nil
}

// ReadBlock reads a block of data from the specified register of an I2C device.
//
// Parameters:
//   - devicePath: Path to the I2C device (e.g., "/dev/i2c-0")
//   - slaveAddr: I2C slave address of the target device
//   - register: Register address to read from
//   - length: Number of bytes to read
//
// Returns the block of data read and any error encountered.
func ReadBlock(devicePath string, slaveAddr uint8, register uint8, length int) ([]byte, error) {
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C device %s: %w", devicePath, err)
	}
	defer file.Close()

	// Set slave address
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SLAVE, uintptr(slaveAddr))
	if errno != 0 {
		return nil, fmt.Errorf("failed to set I2C slave address 0x%02x: %w", slaveAddr, errno)
	}

	// Write register address
	_, err = file.Write([]byte{register})
	if err != nil {
		return nil, fmt.Errorf("failed to write register address 0x%02x: %w", register, err)
	}

	// Read data block
	data := make([]byte, length)
	_, err = file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read block from I2C register 0x%02x: %w", register, err)
	}

	return data, nil
}

// DeviceExists checks if an I2C device responds at the given address.
//
// Parameters:
//   - devicePath: Path to the I2C device (e.g., "/dev/i2c-0")
//   - slaveAddr: I2C slave address to probe
//
// Returns true if the device responds, false otherwise.
func DeviceExists(devicePath string, slaveAddr uint8) bool {
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return false
	}
	defer file.Close()

	// Set slave address
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SLAVE, uintptr(slaveAddr))
	if errno != 0 {
		return false
	}

	// Try to perform a quick operation (write 0 bits)
	data := smbusIoctlData{
		readWrite: I2C_SMBUS_WRITE,
		command:   0,
		size:      I2C_SMBUS_QUICK,
		data:      0,
	}

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&data)))
	return errno == 0
}
