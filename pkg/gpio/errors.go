// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package gpio

import "errors"

var (
	// ErrChipNotFound indicates that the specified GPIO chip could not be found.
	ErrChipNotFound = errors.New("GPIO chip not found")
	// ErrChipAccessDenied indicates insufficient permissions to access the GPIO chip.
	ErrChipAccessDenied = errors.New("access denied to GPIO chip")
	// ErrChipAlreadyOpen indicates that the GPIO chip is already open.
	ErrChipAlreadyOpen = errors.New("GPIO chip already open")
	// ErrChipClosed indicates that an operation was attempted on a closed GPIO chip.
	ErrChipClosed = errors.New("GPIO chip is closed")
	// ErrLineNotFound indicates that the specified GPIO line could not be found.
	ErrLineNotFound = errors.New("GPIO line not found")
	// ErrLineAlreadyRequested indicates that the GPIO line is already in use.
	ErrLineAlreadyRequested = errors.New("GPIO line already requested")
	// ErrLineNotRequested indicates that an operation was attempted on a line that hasn't been requested.
	ErrLineNotRequested = errors.New("GPIO line not requested")
	// ErrLineClosed indicates that an operation was attempted on a closed GPIO line.
	ErrLineClosed = errors.New("GPIO line is closed")
	// ErrInvalidLineNumber indicates that the provided line number is invalid for the chip.
	ErrInvalidLineNumber = errors.New("invalid GPIO line number")
	// ErrInvalidValue indicates that an invalid value was provided for a GPIO operation.
	ErrInvalidValue = errors.New("invalid GPIO value")
	// ErrInvalidDirection indicates that an invalid direction was specified for the GPIO line.
	ErrInvalidDirection = errors.New("invalid GPIO direction")
	// ErrInvalidBias indicates that an invalid bias setting was specified for the GPIO line.
	ErrInvalidBias = errors.New("invalid GPIO bias setting")
	// ErrInvalidEdge indicates that an invalid edge detection setting was specified.
	ErrInvalidEdge = errors.New("invalid GPIO edge detection setting")
	// ErrInvalidDrive indicates that an invalid drive setting was specified for the GPIO line.
	ErrInvalidDrive = errors.New("invalid GPIO drive setting")
	// ErrReadOperation indicates a failure during a GPIO read operation.
	ErrReadOperation = errors.New("GPIO read operation failed")
	// ErrWriteOperation indicates a failure during a GPIO write operation.
	ErrWriteOperation = errors.New("GPIO write operation failed")
	// ErrToggleOperation indicates a failure during a GPIO toggle operation.
	ErrToggleOperation = errors.New("GPIO toggle operation failed")
	// ErrBulkOperation indicates a failure during a bulk GPIO operation.
	ErrBulkOperation = errors.New("GPIO bulk operation failed")
	// ErrEventMonitoring indicates a failure in GPIO event monitoring.
	ErrEventMonitoring = errors.New("GPIO event monitoring failed")
	// ErrEventTimeout indicates that a GPIO event monitoring operation timed out.
	ErrEventTimeout = errors.New("GPIO event monitoring timeout")
	// ErrConfigurationConflict indicates conflicting configuration options.
	ErrConfigurationConflict = errors.New("conflicting GPIO configuration options")
	// ErrInvalidConfiguration indicates that the provided GPIO configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid GPIO configuration")
	// ErrUnsupportedOperation indicates that the requested operation is not supported.
	ErrUnsupportedOperation = errors.New("unsupported GPIO operation")
	// ErrPermissionDenied indicates insufficient permissions for the GPIO operation.
	ErrPermissionDenied = errors.New("permission denied for GPIO operation")
	// ErrResourceBusy indicates that the GPIO resource is currently busy.
	ErrResourceBusy = errors.New("GPIO resource is busy")
	// ErrHardwareNotSupported indicates that the hardware does not support the requested operation.
	ErrHardwareNotSupported = errors.New("hardware does not support requested GPIO operation")
	// ErrManagerClosed indicates that an operation was attempted on a closed GPIO manager.
	ErrManagerClosed = errors.New("GPIO manager is closed")
	// ErrManagerNotInitialized indicates that the GPIO manager has not been properly initialized.
	ErrManagerNotInitialized = errors.New("GPIO manager not initialized")
	// ErrInvalidChipPath indicates that the provided GPIO chip path is invalid.
	ErrInvalidChipPath = errors.New("invalid GPIO chip path")
	// ErrInvalidLineName indicates that the provided GPIO line name is invalid.
	ErrInvalidLineName = errors.New("invalid GPIO line name")
	// ErrInvalidTimeout indicates that an invalid timeout value was provided.
	ErrInvalidTimeout = errors.New("invalid timeout value")
	// ErrOperationCanceled indicates that a GPIO operation was canceled.
	ErrOperationCanceled = errors.New("GPIO operation canceled")
	// ErrContextCanceled indicates that the operation context was canceled.
	ErrContextCanceled = errors.New("GPIO operation context canceled")
	// ErrDeadlineExceeded indicates that the operation deadline was exceeded.
	ErrDeadlineExceeded = errors.New("GPIO operation deadline exceeded")
)
