// SPDX-License-Identifier: BSD-3-Clause

package thermalmgr

import "errors"

var (
	// ErrServiceAlreadyStarted indicates that the thermal manager service is already running.
	ErrServiceAlreadyStarted = errors.New("thermal manager service already started")
	// ErrInvalidConfiguration indicates that the thermal manager configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid thermal manager configuration")
	// ErrNATSConnectionFailed indicates that the NATS connection failed.
	ErrNATSConnectionFailed = errors.New("NATS connection failed")
	// ErrJetStreamInitFailed indicates that JetStream initialization failed.
	ErrJetStreamInitFailed = errors.New("JetStream initialization failed")
	// ErrMicroServiceCreationFailed indicates that micro service creation failed.
	ErrMicroServiceCreationFailed = errors.New("micro service creation failed")
	// ErrEndpointRegistrationFailed indicates that endpoint registration failed.
	ErrEndpointRegistrationFailed = errors.New("endpoint registration failed")
	// ErrStreamCreationFailed indicates that JetStream stream creation failed.
	ErrStreamCreationFailed = errors.New("stream creation failed")
	// ErrThermalZoneInitFailed indicates that thermal zone initialization failed.
	ErrThermalZoneInitFailed = errors.New("thermal zone initialization failed")
	// ErrCoolingDeviceInitFailed indicates that cooling device initialization failed.
	ErrCoolingDeviceInitFailed = errors.New("cooling device initialization failed")
	// ErrThermalControlNotRunning indicates that thermal control is not currently running.
	ErrThermalControlNotRunning = errors.New("thermal control not running")
	// ErrThermalControlAlreadyRunning indicates that thermal control is already running.
	ErrThermalControlAlreadyRunning = errors.New("thermal control already running")
	// ErrEmergencyThermalCondition indicates that an emergency thermal condition has been detected.
	ErrEmergencyThermalCondition = errors.New("emergency thermal condition detected")
	// ErrThermalZoneNotConfigured indicates that the requested thermal zone is not configured.
	ErrThermalZoneNotConfigured = errors.New("thermal zone not configured")
	// ErrCoolingDeviceNotConfigured indicates that the requested cooling device is not configured.
	ErrCoolingDeviceNotConfigured = errors.New("cooling device not configured")
	// ErrInvalidThermalRequest indicates that the thermal management request is invalid.
	ErrInvalidThermalRequest = errors.New("invalid thermal request")
	// ErrThermalOperationFailed indicates that a thermal operation failed.
	ErrThermalOperationFailed = errors.New("thermal operation failed")
	// ErrSensorCommunicationFailed indicates that communication with sensors failed.
	ErrSensorCommunicationFailed = errors.New("sensor communication failed")
	// ErrPowerMgrCommunicationFailed indicates that communication with power manager failed.
	ErrPowerMgrCommunicationFailed = errors.New("power manager communication failed")
	// ErrThermalDiscoveryFailed indicates that thermal device discovery failed.
	ErrThermalDiscoveryFailed = errors.New("thermal device discovery failed")
)
