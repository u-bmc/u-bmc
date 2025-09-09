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
	// ErrPowerSequenceViolation indicates a power operation would violate sequencing rules.
	ErrPowerSequenceViolation = errors.New("power sequence violation")

	// GPIO and hardware errors
	// ErrGPIOOperationFailed indicates a GPIO operation failed.
	ErrGPIOOperationFailed = errors.New("GPIO operation failed")
	// ErrGPIONotConfigured indicates GPIO is not configured for the component.
	ErrGPIONotConfigured = errors.New("GPIO not configured")
	// ErrGPIOLineNotFound indicates the specified GPIO line was not found.
	ErrGPIOLineNotFound = errors.New("GPIO line not found")
	// ErrGPIOPermissionDenied indicates insufficient permissions for GPIO access.
	ErrGPIOPermissionDenied = errors.New("GPIO permission denied")
	// ErrHardwareNotResponding indicates the hardware is not responding.
	ErrHardwareNotResponding = errors.New("hardware not responding")
	// ErrHardwareFailure indicates a hardware failure was detected.
	ErrHardwareFailure = errors.New("hardware failure")

	// Safety and protection errors
	// ErrSafetyInterlock indicates a safety interlock prevented the operation.
	ErrSafetyInterlock = errors.New("safety interlock active")
	// ErrThermalProtection indicates thermal protection prevented the operation.
	ErrThermalProtection = errors.New("thermal protection active")
	// ErrOvercurrentProtection indicates overcurrent protection prevented the operation.
	ErrOvercurrentProtection = errors.New("overcurrent protection active")
	// ErrPowerSupplyOverload indicates power supply overload prevented the operation.
	ErrPowerSupplyOverload = errors.New("power supply overload")
	// ErrEmergencyShutdown indicates an emergency shutdown was triggered.
	ErrEmergencyShutdown = errors.New("emergency shutdown triggered")

	// Power monitoring errors
	// ErrPowerReadFailed indicates power reading failed.
	ErrPowerReadFailed = errors.New("power reading failed")
	// ErrPowerSensorNotFound indicates power sensor was not found.
	ErrPowerSensorNotFound = errors.New("power sensor not found")
	// ErrPowerDataUnavailable indicates power data is not available.
	ErrPowerDataUnavailable = errors.New("power data unavailable")
	// ErrPowerMonitoringDisabled indicates power monitoring is disabled.
	ErrPowerMonitoringDisabled = errors.New("power monitoring disabled")

	// Power capping errors
	// ErrPowerCapExceeded indicates the operation would exceed power cap limits.
	ErrPowerCapExceeded = errors.New("power cap exceeded")
	// ErrInvalidPowerCap indicates an invalid power cap value was specified.
	ErrInvalidPowerCap = errors.New("invalid power cap")
	// ErrPowerCapNotSupported indicates power capping is not supported.
	ErrPowerCapNotSupported = errors.New("power capping not supported")
	// ErrPowerCapEnforcementFailed indicates power cap enforcement failed.
	ErrPowerCapEnforcementFailed = errors.New("power cap enforcement failed")

	// Communication errors
	// ErrNATSConnectionFailed indicates connection to NATS failed.
	ErrNATSConnectionFailed = errors.New("NATS connection failed")
	// ErrMessagePublishFailed indicates publishing a message failed.
	ErrMessagePublishFailed = errors.New("failed to publish message")
	// ErrBroadcastFailed indicates broadcasting a power event failed.
	ErrBroadcastFailed = errors.New("failed to broadcast power event")

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
	// ErrChassisPowerSupplyFault indicates a power supply fault in the chassis.
	ErrChassisPowerSupplyFault = errors.New("chassis power supply fault")
	// ErrChassisOverTemperature indicates the chassis is over temperature.
	ErrChassisOverTemperature = errors.New("chassis over temperature")

	// BMC-specific power errors
	// ErrBMCResetFailed indicates BMC reset operation failed.
	ErrBMCResetFailed = errors.New("BMC reset failed")
	// ErrBMCPowerOperationDenied indicates BMC denied the power operation.
	ErrBMCPowerOperationDenied = errors.New("BMC power operation denied")
	// ErrBMCNotReady indicates the BMC is not ready for power operations.
	ErrBMCNotReady = errors.New("BMC not ready")

	// Backend errors
	// ErrBackendNotConfigured indicates no backend is configured for the operation.
	ErrBackendNotConfigured = errors.New("backend not configured")
	// ErrBackendInitializationFailed indicates backend initialization failed.
	ErrBackendInitializationFailed = errors.New("backend initialization failed")
	// ErrBackendOperationFailed indicates a backend operation failed.
	ErrBackendOperationFailed = errors.New("backend operation failed")
	// ErrBackendNotSupported indicates the backend does not support the operation.
	ErrBackendNotSupported = errors.New("backend not supported")

	// Validation errors
	// ErrInvalidComponentID indicates an invalid component ID was provided.
	ErrInvalidComponentID = errors.New("invalid component ID")
	// ErrInvalidPowerValue indicates an invalid power value was provided.
	ErrInvalidPowerValue = errors.New("invalid power value")
	// ErrInvalidDuration indicates an invalid duration was provided.
	ErrInvalidDuration = errors.New("invalid duration")
	// ErrInvalidGPIOConfiguration indicates invalid GPIO configuration.
	ErrInvalidGPIOConfiguration = errors.New("invalid GPIO configuration")

	// Concurrency errors
	// ErrConcurrentAccess indicates concurrent access to a power resource.
	ErrConcurrentAccess = errors.New("concurrent access to power resource")
	// ErrResourceLocked indicates the power resource is locked.
	ErrResourceLocked = errors.New("power resource locked")
	// ErrDeadlockDetected indicates a potential deadlock was detected.
	ErrDeadlockDetected = errors.New("power operation deadlock detected")

	// Context errors
	// ErrOperationCanceled indicates the operation was canceled.
	ErrOperationCanceled = errors.New("power operation canceled")
	// ErrContextCanceled indicates the operation context was canceled.
	ErrContextCanceled = errors.New("power operation context canceled")
	// ErrDeadlineExceeded indicates the operation deadline was exceeded.
	ErrDeadlineExceeded = errors.New("power operation deadline exceeded")
)
