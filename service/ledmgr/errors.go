// SPDX-License-Identifier: BSD-3-Clause

package ledmgr

import "errors"

var (
	// Service-level errors
	// ErrServiceNotStarted indicates the LED manager service has not been started.
	ErrServiceNotStarted = errors.New("LED manager service not started")
	// ErrServiceAlreadyStarted indicates the LED manager service is already running.
	ErrServiceAlreadyStarted = errors.New("LED manager service already started")
	// ErrServiceStopped indicates the LED manager service has been stopped.
	ErrServiceStopped = errors.New("LED manager service stopped")
	// ErrInvalidConfiguration indicates the service configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid LED manager configuration")

	// Component errors
	// ErrComponentNotFound indicates the requested component does not exist.
	ErrComponentNotFound = errors.New("component not found")
	// ErrComponentNotConfigured indicates the component has not been configured for LED management.
	ErrComponentNotConfigured = errors.New("component not configured")
	// ErrComponentDisabled indicates the component is disabled and cannot be controlled.
	ErrComponentDisabled = errors.New("component disabled")
	// ErrInvalidComponentType indicates an unsupported component type was requested.
	ErrInvalidComponentType = errors.New("invalid component type")
	// ErrComponentBusy indicates the component is busy with another operation.
	ErrComponentBusy = errors.New("component busy")

	// LED operation errors
	// ErrLEDOperationFailed indicates an LED operation failed to complete.
	ErrLEDOperationFailed = errors.New("LED operation failed")
	// ErrLEDOperationTimeout indicates an LED operation timed out.
	ErrLEDOperationTimeout = errors.New("LED operation timeout")
	// ErrLEDOperationInProgress indicates an LED operation is already in progress.
	ErrLEDOperationInProgress = errors.New("LED operation already in progress")
	// ErrLEDOperationNotSupported indicates the requested LED operation is not supported.
	ErrLEDOperationNotSupported = errors.New("LED operation not supported")
	// ErrInvalidLEDAction indicates an invalid LED action was requested.
	ErrInvalidLEDAction = errors.New("invalid LED action")
	// ErrInvalidLEDState indicates an invalid LED state was requested.
	ErrInvalidLEDState = errors.New("invalid LED state")

	// Backend errors
	// ErrBackendNotConfigured indicates no backend is configured for the operation.
	ErrBackendNotConfigured = errors.New("backend not configured")
	// ErrBackendInitializationFailed indicates backend initialization failed.
	ErrBackendInitializationFailed = errors.New("backend initialization failed")
	// ErrBackendOperationFailed indicates a backend operation failed.
	ErrBackendOperationFailed = errors.New("backend operation failed")
	// ErrBackendNotSupported indicates the backend does not support the operation.
	ErrBackendNotSupported = errors.New("backend not supported")

	// GPIO errors
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
	// ErrLEDHardwareFailure indicates LED hardware failure was detected.
	ErrLEDHardwareFailure = errors.New("LED hardware failure")

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

	// LED-specific errors
	// ErrPowerLEDFailed indicates power LED operation failed.
	ErrPowerLEDFailed = errors.New("power LED operation failed")
	// ErrStatusLEDFailed indicates status LED operation failed.
	ErrStatusLEDFailed = errors.New("status LED operation failed")
	// ErrErrorLEDFailed indicates error LED operation failed.
	ErrErrorLEDFailed = errors.New("error LED operation failed")
	// ErrIdentifyLEDFailed indicates identify LED operation failed.
	ErrIdentifyLEDFailed = errors.New("identify LED operation failed")

	// Validation errors
	// ErrInvalidComponentID indicates an invalid component ID was provided.
	ErrInvalidComponentID = errors.New("invalid component ID")
	// ErrInvalidLEDType indicates an invalid LED type was provided.
	ErrInvalidLEDType = errors.New("invalid LED type")
	// ErrInvalidBrightness indicates an invalid brightness value was provided.
	ErrInvalidBrightness = errors.New("invalid brightness value")
	// ErrInvalidBlinkPattern indicates an invalid blink pattern was provided.
	ErrInvalidBlinkPattern = errors.New("invalid blink pattern")

	// Concurrency errors
	// ErrConcurrentAccess indicates concurrent access to an LED resource.
	ErrConcurrentAccess = errors.New("concurrent access to LED resource")
	// ErrResourceLocked indicates the LED resource is locked.
	ErrResourceLocked = errors.New("LED resource locked")

	// Context errors
	// ErrOperationCanceled indicates the operation was canceled.
	ErrOperationCanceled = errors.New("LED operation canceled")
	// ErrContextCanceled indicates the operation context was canceled.
	ErrContextCanceled = errors.New("LED operation context canceled")
	// ErrDeadlineExceeded indicates the operation deadline was exceeded.
	ErrDeadlineExceeded = errors.New("LED operation deadline exceeded")
)
