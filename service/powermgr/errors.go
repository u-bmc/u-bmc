// SPDX-License-Identifier: BSD-3-Clause

package powermgr

import "errors"

var (
	// Service-level errors
	// ErrServiceNotStarted indicates the power manager service has not been started.
	ErrServiceNotStarted = errors.New("power manager service not started")
	// ErrServiceAlreadyStarted indicates the power manager service is already running.
	ErrServiceAlreadyStarted = errors.New("power manager service already started")
	// ErrServiceStopped indicates the power manager service has been stopped.
	ErrServiceStopped = errors.New("power manager service stopped")
	// ErrInvalidConfiguration indicates the service configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid power manager configuration")

	// Component errors
	// ErrComponentNotFound indicates the requested component does not exist.
	ErrComponentNotFound = errors.New("component not found")
	// ErrComponentNotConfigured indicates the component has not been configured for power management.
	ErrComponentNotConfigured = errors.New("component not configured")
	// ErrComponentDisabled indicates the component is disabled and cannot be controlled.
	ErrComponentDisabled = errors.New("component disabled")
	// ErrInvalidComponentType indicates an unsupported component type was requested.
	ErrInvalidComponentType = errors.New("invalid component type")
	// ErrComponentBusy indicates the component is busy with another operation.
	ErrComponentBusy = errors.New("component busy")

	// Power operation errors
	// ErrPowerOperationFailed indicates a power operation failed to complete.
	ErrPowerOperationFailed = errors.New("power operation failed")
	// ErrPowerOperationTimeout indicates a power operation timed out.
	ErrPowerOperationTimeout = errors.New("power operation timeout")
	// ErrPowerOperationInProgress indicates a power operation is already in progress.
	ErrPowerOperationInProgress = errors.New("power operation already in progress")
	// ErrPowerOperationNotSupported indicates the requested power operation is not supported.
	ErrPowerOperationNotSupported = errors.New("power operation not supported")
	// ErrInvalidPowerAction indicates an invalid power action was requested.
	ErrInvalidPowerAction = errors.New("invalid power action")

	// Backend errors
	// ErrBackendNotConfigured indicates no backend is configured for the operation.
	ErrBackendNotConfigured = errors.New("backend not configured")
	// ErrBackendInitializationFailed indicates backend initialization failed.
	ErrBackendInitializationFailed = errors.New("backend initialization failed")
	// ErrBackendOperationFailed indicates a backend operation failed.
	ErrBackendOperationFailed = errors.New("backend operation failed")
	// ErrBackendNotSupported indicates the backend does not support the operation.
	ErrBackendNotSupported = errors.New("backend not supported")
	// ErrCallbackNotSet indicates the required callback function is not set.
	ErrCallbackNotSet = errors.New("callback function not set")
	// ErrCallbackFailed indicates a callback function failed.
	ErrCallbackFailed = errors.New("callback function failed")

	// GPIO and hardware errors
	// ErrGPIOOperationFailed indicates a GPIO operation failed.
	ErrGPIOOperationFailed = errors.New("GPIO operation failed")
	// ErrGPIONotConfigured indicates GPIO is not configured for the component.
	ErrGPIONotConfigured = errors.New("GPIO not configured")
	// ErrGPIOLineNotFound indicates the specified GPIO line was not found.
	ErrGPIOLineNotFound = errors.New("GPIO line not found")
	// ErrGPIOPermissionDenied indicates insufficient permissions for GPIO access.
	ErrGPIOPermissionDenied = errors.New("GPIO permission denied")
	// ErrInvalidGPIOConfiguration indicates invalid GPIO configuration.
	ErrInvalidGPIOConfiguration = errors.New("invalid GPIO configuration")

	// I2C errors
	// ErrI2COperationFailed indicates an I2C operation failed.
	ErrI2COperationFailed = errors.New("I2C operation failed")
	// ErrI2CNotConfigured indicates I2C is not configured for the component.
	ErrI2CNotConfigured = errors.New("I2C not configured")
	// ErrI2CDeviceNotFound indicates the specified I2C device was not found.
	ErrI2CDeviceNotFound = errors.New("I2C device not found")
	// ErrI2CPermissionDenied indicates insufficient permissions for I2C access.
	ErrI2CPermissionDenied = errors.New("I2C permission denied")
	// ErrI2CSlaveNotResponding indicates the I2C slave device is not responding.
	ErrI2CSlaveNotResponding = errors.New("I2C slave not responding")
	// ErrInvalidI2CConfiguration indicates invalid I2C configuration.
	ErrInvalidI2CConfiguration = errors.New("invalid I2C configuration")
	// ErrI2CAddressInUse indicates the I2C address is already in use.
	ErrI2CAddressInUse = errors.New("I2C address in use")
	// ErrI2CTransmissionFailed indicates I2C data transmission failed.
	ErrI2CTransmissionFailed = errors.New("I2C transmission failed")

	// Hardware errors
	// ErrHardwareNotResponding indicates the hardware is not responding.
	ErrHardwareNotResponding = errors.New("hardware not responding")
	// ErrHardwareFailure indicates a hardware failure was detected.
	ErrHardwareFailure = errors.New("hardware failure")

	// Communication errors
	// ErrNATSConnectionFailed indicates connection to NATS failed.
	ErrNATSConnectionFailed = errors.New("NATS connection failed")
	// ErrMessagePublishFailed indicates publishing a message failed.
	ErrMessagePublishFailed = errors.New("failed to publish message")

	// Request/Response errors
	// ErrInvalidRequest indicates the request format is invalid.
	ErrInvalidRequest = errors.New("invalid request format")
	// ErrMissingRequiredField indicates a required field is missing in the request.
	ErrMissingRequiredField = errors.New("missing required field")
	// ErrMarshalingFailed indicates protobuf marshaling failed.
	ErrMarshalingFailed = errors.New("marshaling failed")
	// ErrUnmarshalingFailed indicates protobuf unmarshaling failed.
	ErrUnmarshalingFailed = errors.New("unmarshaling failed")
	// ErrResponseTimeout indicates a response timeout occurred.
	ErrResponseTimeout = errors.New("response timeout")

	// Host-specific power errors
	// ErrHostPowerOnFailed indicates host power on operation failed.
	ErrHostPowerOnFailed = errors.New("host power on failed")
	// ErrHostPowerOffFailed indicates host power off operation failed.
	ErrHostPowerOffFailed = errors.New("host power off failed")
	// ErrHostResetFailed indicates host reset operation failed.
	ErrHostResetFailed = errors.New("host reset failed")
	// ErrHostNotPowered indicates the host is not powered when operation requires it.
	ErrHostNotPowered = errors.New("host not powered")
	// ErrHostAlreadyPowered indicates the host is already powered when operation conflicts.
	ErrHostAlreadyPowered = errors.New("host already powered")

	// Chassis-specific power errors
	// ErrChassisPowerOnFailed indicates chassis power on operation failed.
	ErrChassisPowerOnFailed = errors.New("chassis power on failed")
	// ErrChassisPowerOffFailed indicates chassis power off operation failed.
	ErrChassisPowerOffFailed = errors.New("chassis power off failed")
	// ErrChassisNotPresent indicates the chassis is not physically present.
	ErrChassisNotPresent = errors.New("chassis not present")

	// BMC-specific power errors
	// ErrBMCResetFailed indicates BMC reset operation failed.
	ErrBMCResetFailed = errors.New("BMC reset failed")
	// ErrBMCPowerOperationDenied indicates BMC denied the power operation.
	ErrBMCPowerOperationDenied = errors.New("BMC power operation denied")
	// ErrBMCNotReady indicates the BMC is not ready for power operations.
	ErrBMCNotReady = errors.New("BMC not ready")

	// Validation errors
	// ErrInvalidComponentID indicates an invalid component ID was provided.
	ErrInvalidComponentID = errors.New("invalid component ID")

	// Concurrency errors
	// ErrConcurrentAccess indicates concurrent access to a power resource.
	ErrConcurrentAccess = errors.New("concurrent access to power resource")
	// ErrResourceLocked indicates the power resource is locked.
	ErrResourceLocked = errors.New("power resource locked")

	// Context errors
	// ErrOperationCanceled indicates the operation was canceled.
	ErrOperationCanceled = errors.New("power operation canceled")
	// ErrContextCanceled indicates the operation context was canceled.
	ErrContextCanceled = errors.New("power operation context canceled")
	// ErrDeadlineExceeded indicates the operation deadline was exceeded.
	ErrDeadlineExceeded = errors.New("power operation deadline exceeded")
)
