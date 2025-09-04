// SPDX-License-Identifier: BSD-3-Clause

//nolint:gosec
package i2c

import (
	"fmt"
)

// SMBus transaction types - these extend the constants from conn.go.
const (
	// SMBus command constants.
	smbusQuickWrite = 0
	smbusQuickRead  = 1
)

// QuickCommand sends an SMBus Quick Command.
// This is typically used to turn a device on/off or test if a device is present.
// The bit parameter determines if this is a quick write (0) or quick read (1).
func (c *Conn) QuickCommand(bit uint8) error {
	if !c.supportsSMBus() {
		return ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	var readWrite uint8 = i2cSMBusWrite
	if bit != 0 {
		readWrite = i2cSMBusRead
	}

	if err := c.smbusAccess(readWrite, 0, i2cSMBusQuick, nil); err != nil {
		return fmt.Errorf("%w: %w", ErrSMBusQuickFailed, err)
	}

	return nil
}

// SendByte sends a single byte using SMBus Send Byte protocol.
func (c *Conn) SendByte(value uint8) error {
	if !c.supportsSMBus() {
		return ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	data := &i2cSMBusData{}
	data.byte0 = value

	if err := c.smbusAccess(i2cSMBusWrite, value, i2cSMBusByte, data); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteFailed, err)
	}

	return nil
}

// ReceiveByte receives a single byte using SMBus Receive Byte protocol.
func (c *Conn) ReceiveByte() (uint8, error) {
	if !c.supportsSMBus() {
		return 0, ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	data := &i2cSMBusData{}

	if err := c.smbusAccess(i2cSMBusRead, 0, i2cSMBusByte, data); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrReadFailed, err)
	}

	return data.byte0, nil
}

// WriteByteData writes a single byte to a specific register using SMBus Write Byte Data.
func (c *Conn) WriteByteData(register, value uint8) error {
	if !c.supportsSMBus() {
		return ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	data := &i2cSMBusData{}
	data.byte0 = value

	if err := c.smbusAccess(i2cSMBusWrite, register, i2cSMBusByteData, data); err != nil {
		return fmt.Errorf("%w: register 0x%02x: %w", ErrWriteFailed, register, err)
	}

	return nil
}

// ReadByteData reads a single byte from a specific register using SMBus Read Byte Data.
func (c *Conn) ReadByteData(register uint8) (uint8, error) {
	if !c.supportsSMBus() {
		return 0, ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	data := &i2cSMBusData{}

	if err := c.smbusAccess(i2cSMBusRead, register, i2cSMBusByteData, data); err != nil {
		return 0, fmt.Errorf("%w: register 0x%02x: %w", ErrReadFailed, register, err)
	}

	return data.byte0, nil
}

// WriteWordData writes a 16-bit word to a specific register using SMBus Write Word Data.
// The word is sent in little-endian format (LSB first).
func (c *Conn) WriteWordData(register uint8, value uint16) error {
	if !c.supportsSMBus() {
		return ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	if c.capabilities&i2cFuncSMBusWriteWordData == 0 {
		return ErrSMBusUnsupportedCommand
	}

	data := &i2cSMBusData{}
	data.word = value

	if err := c.smbusAccess(i2cSMBusWrite, register, i2cSMBusWordData, data); err != nil {
		return fmt.Errorf("%w: register 0x%02x: %w", ErrWriteFailed, register, err)
	}

	return nil
}

// ReadWordData reads a 16-bit word from a specific register using SMBus Read Word Data.
// The word is received in little-endian format (LSB first).
func (c *Conn) ReadWordData(register uint8) (uint16, error) {
	if !c.supportsSMBus() {
		return 0, ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	if c.capabilities&i2cFuncSMBusReadWordData == 0 {
		return 0, ErrSMBusUnsupportedCommand
	}

	data := &i2cSMBusData{}

	if err := c.smbusAccess(i2cSMBusRead, register, i2cSMBusWordData, data); err != nil {
		return 0, fmt.Errorf("%w: register 0x%02x: %w", ErrReadFailed, register, err)
	}

	return data.word, nil
}

// ProcessCall performs an SMBus Process Call operation.
// This writes a 16-bit word to a register and reads back a 16-bit response.
func (c *Conn) ProcessCall(register uint8, value uint16) (uint16, error) {
	if !c.supportsSMBus() {
		return 0, ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	data := &i2cSMBusData{}
	data.word = value

	if err := c.smbusAccess(i2cSMBusWrite, register, i2cSMBusProcCall, data); err != nil {
		return 0, fmt.Errorf("%w: register 0x%02x: %w", ErrOperationFailed, register, err)
	}

	return data.word, nil
}

// WriteBlockData writes a block of data to a specific register using SMBus Write Block Data.
// The first byte of the transmission contains the number of data bytes to follow.
func (c *Conn) WriteBlockData(register uint8, values []byte) error {
	if !c.supportsSMBus() {
		return ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	if len(values) == 0 {
		return ErrInvalidDataLength
	}

	if len(values) > i2cSMBusBlockMax {
		return fmt.Errorf("%w: maximum %d bytes allowed, got %d",
			ErrBufferTooLarge, i2cSMBusBlockMax, len(values))
	}

	if c.capabilities&i2cFuncSMBusWriteBlockData == 0 {
		return ErrSMBusUnsupportedCommand
	}

	data := &i2cSMBusData{}
	data.block[0] = uint8(len(values))
	copy(data.block[1:], values)

	if err := c.smbusAccess(i2cSMBusWrite, register, i2cSMBusBlockData, data); err != nil {
		return fmt.Errorf("%w: register 0x%02x: %w", ErrWriteFailed, register, err)
	}

	return nil
}

// ReadBlockData reads a block of data from a specific register using SMBus Read Block Data.
// The first byte received contains the number of data bytes to follow.
// The buffer must be large enough to hold the maximum possible block size.
func (c *Conn) ReadBlockData(register uint8, buffer []byte) (int, error) {
	if !c.supportsSMBus() {
		return 0, ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	if len(buffer) == 0 {
		return 0, ErrInvalidDataLength
	}

	if c.capabilities&i2cFuncSMBusReadBlockData == 0 {
		return 0, ErrSMBusUnsupportedCommand
	}

	data := &i2cSMBusData{}

	if err := c.smbusAccess(i2cSMBusRead, register, i2cSMBusBlockData, data); err != nil {
		return 0, fmt.Errorf("%w: register 0x%02x: %w", ErrReadFailed, register, err)
	}

	blockLength := int(data.block[0])
	if blockLength > len(buffer) {
		return 0, fmt.Errorf("%w: need %d bytes, buffer has %d",
			ErrBufferTooSmall, blockLength, len(buffer))
	}

	copy(buffer, data.block[1:blockLength+1])
	return blockLength, nil
}

// WriteI2CBlockData writes a block of data using SMBus I2C Block Write.
// Unlike WriteBlockData, this doesn't include a length byte.
func (c *Conn) WriteI2CBlockData(register uint8, values []byte) error {
	if !c.supportsSMBus() {
		return ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	if len(values) == 0 {
		return ErrInvalidDataLength
	}

	if len(values) > i2cSMBusBlockMax {
		return fmt.Errorf("%w: maximum %d bytes allowed, got %d",
			ErrBufferTooLarge, i2cSMBusBlockMax, len(values))
	}

	if c.capabilities&i2cFuncSMBusWriteI2CBlockData == 0 {
		return ErrSMBusUnsupportedCommand
	}

	data := &i2cSMBusData{}
	data.block[0] = uint8(len(values))
	copy(data.block[1:], values)

	if err := c.smbusAccess(i2cSMBusWrite, register, i2cSMBusI2CBlockData, data); err != nil {
		return fmt.Errorf("%w: register 0x%02x: %w", ErrWriteFailed, register, err)
	}

	return nil
}

// ReadI2CBlockData reads a fixed number of bytes using SMBus I2C Block Read.
// Unlike ReadBlockData, this doesn't expect a length byte from the device.
func (c *Conn) ReadI2CBlockData(register uint8, length uint8, buffer []byte) error {
	if !c.supportsSMBus() {
		return ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	if length == 0 || length > i2cSMBusBlockMax {
		return fmt.Errorf("%w: length must be 1-%d, got %d",
			ErrInvalidDataLength, i2cSMBusBlockMax, length)
	}

	if len(buffer) < int(length) {
		return fmt.Errorf("%w: need %d bytes, buffer has %d",
			ErrBufferTooSmall, length, len(buffer))
	}

	if c.capabilities&i2cFuncSMBusReadI2CBlockData == 0 {
		return ErrSMBusUnsupportedCommand
	}

	data := &i2cSMBusData{}
	data.block[0] = length

	if err := c.smbusAccess(i2cSMBusRead, register, i2cSMBusI2CBlockData, data); err != nil {
		return fmt.Errorf("%w: register 0x%02x: %w", ErrReadFailed, register, err)
	}

	copy(buffer, data.block[1:length+1])
	return nil
}

// BlockProcessCall performs an SMBus Block Process Call operation.
// This writes a block of data and reads back a block response.
func (c *Conn) BlockProcessCall(register uint8, writeValues []byte, readBuffer []byte) (int, error) {
	if !c.supportsSMBus() {
		return 0, ErrSMBusNotSupported
	}

	if c.config.Protocol != ProtocolSMBus && c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	if len(writeValues) == 0 || len(writeValues) > i2cSMBusBlockMax {
		return 0, fmt.Errorf("%w: write length must be 1-%d, got %d",
			ErrInvalidDataLength, i2cSMBusBlockMax, len(writeValues))
	}

	if c.capabilities&i2cFuncSMBusBlockProcCall == 0 {
		return 0, ErrSMBusUnsupportedCommand
	}

	data := &i2cSMBusData{}
	data.block[0] = uint8(len(writeValues))
	copy(data.block[1:], writeValues)

	if err := c.smbusAccess(i2cSMBusWrite, register, i2cSMBusBlockProcCall, data); err != nil {
		return 0, fmt.Errorf("%w: register 0x%02x: %w", ErrOperationFailed, register, err)
	}

	blockLength := int(data.block[0])
	if blockLength > len(readBuffer) {
		return 0, fmt.Errorf("%w: need %d bytes, buffer has %d",
			ErrBufferTooSmall, blockLength, len(readBuffer))
	}

	copy(readBuffer, data.block[1:blockLength+1])
	return blockLength, nil
}

// ScanBus scans the I2C bus for responding devices.
// It returns a slice of addresses that responded to the Quick Command.
func (c *Conn) ScanBus() ([]uint8, error) {
	if !c.supportsSMBus() {
		return nil, ErrSMBusNotSupported
	}

	var devices []uint8

	// Scan the valid address range for 7-bit addressing
	for addr := uint8(0x08); addr <= 0x77; addr++ {
		// Temporarily set the address
		oldAddr := c.currentAddr
		oldAddrSet := c.addrSet

		if err := c.SetAddress(uint16(addr)); err != nil {
			continue
		}

		// Try Quick Command
		if err := c.QuickCommand(smbusQuickRead); err == nil {
			devices = append(devices, addr)
		}

		// Restore previous address if it was set
		if oldAddrSet {
			// Best effort to restore - log but don't fail the scan
			if err := c.SetAddress(oldAddr); err != nil {
				// Can't return error here as it would lose scan results
				// Consider logging this error if a logger is available
				_ = err // Explicitly ignore after consideration
			}
		} else {
			c.addrSet = false
		}
	}

	return devices, nil
}

// GetSMBusCapabilities returns a human-readable list of supported SMBus operations.
func (c *Conn) GetSMBusCapabilities() []string {
	var capabilities []string

	if c.capabilities&i2cFuncSMBusQuick != 0 {
		capabilities = append(capabilities, "Quick Command")
	}
	if c.capabilities&i2cFuncSMBusReadByte != 0 {
		capabilities = append(capabilities, "Receive Byte")
	}
	if c.capabilities&i2cFuncSMBusWriteByte != 0 {
		capabilities = append(capabilities, "Send Byte")
	}
	if c.capabilities&i2cFuncSMBusReadByteData != 0 {
		capabilities = append(capabilities, "Read Byte Data")
	}
	if c.capabilities&i2cFuncSMBusWriteByteData != 0 {
		capabilities = append(capabilities, "Write Byte Data")
	}
	if c.capabilities&i2cFuncSMBusReadWordData != 0 {
		capabilities = append(capabilities, "Read Word Data")
	}
	if c.capabilities&i2cFuncSMBusWriteWordData != 0 {
		capabilities = append(capabilities, "Write Word Data")
	}
	if c.capabilities&i2cFuncSMBusReadBlockData != 0 {
		capabilities = append(capabilities, "Read Block Data")
	}
	if c.capabilities&i2cFuncSMBusWriteBlockData != 0 {
		capabilities = append(capabilities, "Write Block Data")
	}
	if c.capabilities&i2cFuncSMBusReadI2CBlockData != 0 {
		capabilities = append(capabilities, "Read I2C Block Data")
	}
	if c.capabilities&i2cFuncSMBusWriteI2CBlockData != 0 {
		capabilities = append(capabilities, "Write I2C Block Data")
	}
	if c.capabilities&i2cFuncSMBusBlockProcCall != 0 {
		capabilities = append(capabilities, "Block Process Call")
	}
	if c.capabilities&i2cFuncSMBusPEC != 0 {
		capabilities = append(capabilities, "Packet Error Checking (PEC)")
	}

	return capabilities
}

// ValidateConnection performs a basic connectivity test to the device.
// It attempts a Quick Command to verify the device is present and responding.
func (c *Conn) ValidateConnection() error {
	if !c.addrSet {
		return fmt.Errorf("%w: device address not set", ErrInvalidConfig)
	}

	// Try Quick Command as a basic connectivity test
	if err := c.QuickCommand(smbusQuickRead); err != nil {
		return fmt.Errorf("%w: %w", ErrDeviceNotResponding, err)
	}

	return nil
}
