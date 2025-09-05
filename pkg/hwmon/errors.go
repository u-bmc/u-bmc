// SPDX-License-Identifier: BSD-3-Clause

package hwmon

import "errors"

var (
	// ErrInvalidConfig indicates that the hwmon configuration is invalid.
	ErrInvalidConfig = errors.New("invalid hwmon configuration")
	// ErrDeviceNotFound indicates that the specified hwmon device was not found.
	ErrDeviceNotFound = errors.New("hwmon device not found")
	// ErrDeviceUnavailable indicates that the hwmon device is unavailable or has been removed.
	ErrDeviceUnavailable = errors.New("hwmon device unavailable")
	// ErrSensorNotFound indicates that the specified sensor was not found on the device.
	ErrSensorNotFound = errors.New("sensor not found")
	// ErrSensorUnavailable indicates that the sensor is unavailable or has been disabled.
	ErrSensorUnavailable = errors.New("sensor unavailable")
	// ErrInvalidSensorType indicates that the specified sensor type is not supported.
	ErrInvalidSensorType = errors.New("invalid sensor type")
	// ErrInvalidSensorIndex indicates that the specified sensor index is invalid.
	ErrInvalidSensorIndex = errors.New("invalid sensor index")
	// ErrInvalidAttribute indicates that the specified sensor attribute is invalid.
	ErrInvalidAttribute = errors.New("invalid sensor attribute")
	// ErrReadFailure indicates that reading from a sensor failed.
	ErrReadFailure = errors.New("sensor read failure")
	// ErrWriteFailure indicates that writing to a sensor failed.
	ErrWriteFailure = errors.New("sensor write failure")
	// ErrReadTimeout indicates that a sensor read operation timed out.
	ErrReadTimeout = errors.New("sensor read timeout")
	// ErrWriteTimeout indicates that a sensor write operation timed out.
	ErrWriteTimeout = errors.New("sensor write timeout")
	// ErrPermissionDenied indicates that access to the sensor was denied due to insufficient permissions.
	ErrPermissionDenied = errors.New("permission denied accessing sensor")
	// ErrInvalidValue indicates that a sensor value is invalid or out of range.
	ErrInvalidValue = errors.New("invalid sensor value")
	// ErrValueParseFailure indicates that parsing a sensor value failed.
	ErrValueParseFailure = errors.New("failed to parse sensor value")
	// ErrValueOutOfRange indicates that a sensor value is outside the expected range.
	ErrValueOutOfRange = errors.New("sensor value out of range")
	// ErrReadOnlySensor indicates that an attempt was made to write to a read-only sensor.
	ErrReadOnlySensor = errors.New("sensor is read-only")
	// ErrWriteOnlySensor indicates that an attempt was made to read from a write-only sensor.
	ErrWriteOnlySensor = errors.New("sensor is write-only")
	// ErrDiscoveryFailure indicates that sensor discovery failed.
	ErrDiscoveryFailure = errors.New("sensor discovery failure")
	// ErrNilContext indicates that a nil context was provided.
	ErrNilContext = errors.New("context cannot be nil")
	// ErrInvalidPath indicates that the provided sysfs path is invalid.
	ErrInvalidPath = errors.New("invalid sysfs path")
	// ErrPathNotFound indicates that the specified sysfs path does not exist.
	ErrPathNotFound = errors.New("sysfs path not found")
	// ErrFileSystemError indicates a general filesystem error when accessing sysfs.
	ErrFileSystemError = errors.New("filesystem error accessing sysfs")
	// ErrCacheFailure indicates that a caching operation failed.
	ErrCacheFailure = errors.New("sensor cache operation failure")
	// ErrRetryExhausted indicates that all retry attempts have been exhausted.
	ErrRetryExhausted = errors.New("retry attempts exhausted")
	// ErrConcurrentAccess indicates that concurrent access to a resource was attempted inappropriately.
	ErrConcurrentAccess = errors.New("concurrent access violation")
	// ErrValidationFailed indicates that sensor validation failed.
	ErrValidationFailed = errors.New("sensor validation failed")
	// ErrUnsupportedOperation indicates that the requested operation is not supported.
	ErrUnsupportedOperation = errors.New("unsupported operation")
	// ErrDeviceNameMismatch indicates that the device name does not match expectations.
	ErrDeviceNameMismatch = errors.New("device name mismatch")
	// ErrAttributeNotSupported indicates that the sensor does not support the requested attribute.
	ErrAttributeNotSupported = errors.New("sensor attribute not supported")
	// ErrInsufficientData indicates that there is insufficient data to complete the operation.
	ErrInsufficientData = errors.New("insufficient data for operation")
	// ErrOperationCanceled indicates that the operation was canceled.
	ErrOperationCanceled = errors.New("operation canceled")
)
