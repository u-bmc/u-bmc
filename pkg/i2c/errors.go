// SPDX-License-Identifier: BSD-3-Clause

package i2c

import "errors"

var (
	// Bus and device access errors.

	// ErrBusNotFound indicates that the specified I2C bus device file does not exist.
	ErrBusNotFound = errors.New("I2C bus device not found")
	// ErrBusAccessDenied indicates insufficient permissions to access the I2C bus device.
	ErrBusAccessDenied = errors.New("access denied to I2C bus device")
	// ErrBusOpenFailed indicates a failure to open the I2C bus device file.
	ErrBusOpenFailed = errors.New("failed to open I2C bus device")
	// ErrBusCloseFailed indicates a failure to close the I2C bus device file.
	ErrBusCloseFailed = errors.New("failed to close I2C bus device")

	// Device communication errors.

	// ErrDeviceNotResponding indicates that the I2C device did not acknowledge communication attempts.
	ErrDeviceNotResponding = errors.New("I2C device not responding")
	// ErrAddressNACK indicates that the device address was not acknowledged.
	ErrAddressNACK = errors.New("I2C device address not acknowledged")
	// ErrDataNACK indicates that data transmission was not acknowledged by the device.
	ErrDataNACK = errors.New("I2C data transmission not acknowledged")
	// ErrArbitrationLost indicates that bus arbitration was lost during communication.
	ErrArbitrationLost = errors.New("I2C bus arbitration lost")
	// ErrBusError indicates a general I2C bus error occurred.
	ErrBusError = errors.New("I2C bus error")

	// Protocol-specific errors.

	// ErrProtocolViolation indicates a violation of the selected protocol specifications.
	ErrProtocolViolation = errors.New("protocol violation")
	// ErrSMBusNotSupported indicates that SMBus operations are not supported by the adapter.
	ErrSMBusNotSupported = errors.New("SMBus operations not supported by adapter")
	// ErrPECNotSupported indicates that Packet Error Checking is not supported by the adapter.
	ErrPECNotSupported = errors.New("packet Error Checking not supported by adapter")
	// ErrChecksumMismatch indicates that the Packet Error Checking (PEC) validation failed.
	ErrChecksumMismatch = errors.New("packet error checking (PEC) checksum mismatch")
	// ErrI3CNotSupported indicates that I3C operations are not supported by the adapter.
	ErrI3CNotSupported = errors.New("I3C operations not supported by adapter")

	// Data and parameter validation errors.

	// ErrInvalidBusNumber indicates that the specified bus number is invalid.
	ErrInvalidBusNumber = errors.New("invalid I2C bus number")
	// ErrInvalidAddress indicates that the specified device address is invalid.
	ErrInvalidAddress = errors.New("invalid I2C device address")
	// ErrInvalidRegister indicates that the specified register address is invalid.
	ErrInvalidRegister = errors.New("invalid register address")
	// ErrInvalidDataLength indicates that the data length is invalid for the operation.
	ErrInvalidDataLength = errors.New("invalid data length for operation")
	// ErrBufferTooSmall indicates that the provided buffer is too small for the operation.
	ErrBufferTooSmall = errors.New("buffer too small for operation")
	// ErrBufferTooLarge indicates that the provided buffer exceeds maximum size limits.
	ErrBufferTooLarge = errors.New("buffer too large for operation")
	// ErrInvalidData indicates that the provided data is invalid or out of range.
	ErrInvalidData = errors.New("invalid data for operation")

	// Configuration errors.

	// ErrInvalidConfig indicates that the provided configuration is invalid.
	ErrInvalidConfig = errors.New("invalid I2C configuration")
	// ErrInvalidProtocol indicates that the specified protocol is invalid or unsupported.
	ErrInvalidProtocol = errors.New("invalid or unsupported protocol")
	// ErrInvalidTimeout indicates that the specified timeout value is invalid.
	ErrInvalidTimeout = errors.New("invalid timeout value")
	// ErrInvalidRetryCount indicates that the specified retry count is invalid.
	ErrInvalidRetryCount = errors.New("invalid retry count")

	// Operation errors.

	// ErrTimeout indicates that an I2C operation timed out.
	ErrTimeout = errors.New("I2C operation timeout")
	// ErrOperationFailed indicates that an I2C operation failed for an unspecified reason.
	ErrOperationFailed = errors.New("I2C operation failed")
	// ErrReadFailed indicates that a read operation failed.
	ErrReadFailed = errors.New("I2C read operation failed")
	// ErrWriteFailed indicates that a write operation failed.
	ErrWriteFailed = errors.New("I2C write operation failed")
	// ErrTransactionFailed indicates that a combined I2C transaction failed.
	ErrTransactionFailed = errors.New("I2C transaction failed")

	// SMBus specific errors.

	// ErrSMBusQuickFailed indicates that an SMBus Quick Command failed.
	ErrSMBusQuickFailed = errors.New("SMBus Quick Command failed")
	// ErrSMBusBlockSizeMismatch indicates that the block size doesn't match expected value.
	ErrSMBusBlockSizeMismatch = errors.New("SMBus block size mismatch")
	// ErrSMBusUnsupportedCommand indicates that the SMBus command is not supported.
	ErrSMBusUnsupportedCommand = errors.New("SMBus command not supported")

	// PMBus specific errors.

	// ErrPMBusInvalidCommand indicates that the PMBus command is invalid.
	ErrPMBusInvalidCommand = errors.New("invalid PMBus command")
	// ErrPMBusDataFormatError indicates an error in PMBus data format conversion.
	ErrPMBusDataFormatError = errors.New("PMBus data format conversion error")
	// ErrPMBusLinearFormatError indicates an error in PMBus LINEAR format handling.
	ErrPMBusLinearFormatError = errors.New("PMBus LINEAR format error")
	// ErrPMBusDirectFormatError indicates an error in PMBus DIRECT format handling.
	ErrPMBusDirectFormatError = errors.New("PMBus DIRECT format error")
	// ErrPMBusCoefficientsInvalid indicates that PMBus coefficients are invalid.
	ErrPMBusCoefficientsInvalid = errors.New("PMBus coefficients invalid")

	// I3C specific errors.

	// ErrI3CDynamicAddressFailed indicates that I3C dynamic address assignment failed.
	ErrI3CDynamicAddressFailed = errors.New("I3C dynamic address assignment failed")
	// ErrI3CCCCFailed indicates that an I3C Common Command Code operation failed.
	ErrI3CCCCFailed = errors.New("I3C Common Command Code operation failed")
	// ErrI3CHotJoinFailed indicates that I3C Hot-Join operation failed.
	ErrI3CHotJoinFailed = errors.New("I3C Hot-Join operation failed")
	// ErrI3CIBIFailed indicates that an I3C In-Band Interrupt failed.
	ErrI3CIBIFailed = errors.New("I3C In-Band Interrupt operation failed")

	// System and hardware errors.

	// ErrAdapterNotFound indicates that no I2C adapter was found for the specified bus.
	ErrAdapterNotFound = errors.New("I2C adapter not found")
	// ErrAdapterBusy indicates that the I2C adapter is busy and cannot perform the operation.
	ErrAdapterBusy = errors.New("I2C adapter busy")
	// ErrHardwareFault indicates a hardware-level fault in the I2C subsystem.
	ErrHardwareFault = errors.New("I2C hardware fault")
	// ErrKernelDriverError indicates an error in the kernel I2C driver.
	ErrKernelDriverError = errors.New("kernel I2C driver error")

	// Advanced feature errors.

	// ErrClockStretchingTimeout indicates that clock stretching exceeded timeout limits.
	ErrClockStretchingTimeout = errors.New("I2C clock stretching timeout")
	// ErrMultimasterConflict indicates a conflict in multi-master I2C configuration.
	ErrMultimasterConflict = errors.New("I2C multi-master conflict")
	// ErrBusRecoveryFailed indicates that I2C bus recovery failed.
	ErrBusRecoveryFailed = errors.New("I2C bus recovery failed")
)
