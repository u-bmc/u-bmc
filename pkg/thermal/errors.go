// SPDX-License-Identifier: BSD-3-Clause

package thermal

import "errors"

var (
	// ErrPIDNotInitialized indicates that the PID controller has not been initialized.
	ErrPIDNotInitialized = errors.New("PID controller not initialized")
	// ErrInvalidTemperature indicates that the temperature value is invalid or out of range.
	ErrInvalidTemperature = errors.New("invalid temperature value")
	// ErrInvalidPIDConfig indicates that the PID configuration parameters are invalid.
	ErrInvalidPIDConfig = errors.New("invalid PID configuration")
	// ErrCoolingDeviceUnavailable indicates that the cooling device is not available or accessible.
	ErrCoolingDeviceUnavailable = errors.New("cooling device unavailable")
	// ErrInvalidCoolingPower indicates that the cooling power value is out of valid range.
	ErrInvalidCoolingPower = errors.New("invalid cooling power value")
	// ErrThermalZoneNotFound indicates that the specified thermal zone was not found.
	ErrThermalZoneNotFound = errors.New("thermal zone not found")
	// ErrThermalZoneAlreadyExists indicates that a thermal zone with the same name already exists.
	ErrThermalZoneAlreadyExists = errors.New("thermal zone already exists")
	// ErrInvalidZoneConfiguration indicates that the thermal zone configuration is invalid.
	ErrInvalidZoneConfiguration = errors.New("invalid thermal zone configuration")
	// ErrSensorReadFailure indicates that reading from a temperature sensor failed.
	ErrSensorReadFailure = errors.New("sensor read failure")
	// ErrCoolingControlFailure indicates that controlling a cooling device failed.
	ErrCoolingControlFailure = errors.New("cooling control failure")
	// ErrCriticalTemperature indicates that a critical temperature threshold has been exceeded.
	ErrCriticalTemperature = errors.New("critical temperature exceeded")
	// ErrEmergencyShutdownRequired indicates that an emergency shutdown is required due to thermal conditions.
	ErrEmergencyShutdownRequired = errors.New("emergency shutdown required")
	// ErrThermalProfileNotFound indicates that the specified thermal profile was not found.
	ErrThermalProfileNotFound = errors.New("thermal profile not found")
	// ErrInvalidSampleTime indicates that the PID sample time is invalid.
	ErrInvalidSampleTime = errors.New("invalid PID sample time")
	// ErrOutputLimitsInvalid indicates that the PID output limits are invalid.
	ErrOutputLimitsInvalid = errors.New("invalid PID output limits")
	// ErrDeviceTypeUnsupported indicates that the cooling device type is not supported.
	ErrDeviceTypeUnsupported = errors.New("unsupported cooling device type")
	// ErrHwmonPathInvalid indicates that the hwmon path for a device is invalid.
	ErrHwmonPathInvalid = errors.New("invalid hwmon path")
	// ErrThermalOperationTimeout indicates that a thermal operation timed out.
	ErrThermalOperationTimeout = errors.New("thermal operation timeout")
)
