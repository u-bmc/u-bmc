// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import "errors"

var (
	// Service-level errors
	// ErrServiceNotStarted indicates the service has not been started.
	ErrServiceNotStarted = errors.New("state manager service not started")
	// ErrServiceAlreadyStarted indicates the service is already running.
	ErrServiceAlreadyStarted = errors.New("state manager service already started")
	// ErrServiceStopped indicates the service has been stopped.
	ErrServiceStopped = errors.New("state manager service stopped")
	// ErrInvalidConfiguration indicates the service configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid service configuration")

	// Component errors
	// ErrInvalidComponentType indicates an unsupported component type was requested.
	ErrInvalidComponentType = errors.New("invalid component type")
	// ErrComponentNotFound indicates the requested component does not exist.
	ErrComponentNotFound = errors.New("component not found")
	// ErrComponentAlreadyExists indicates a component with the same ID already exists.
	ErrComponentAlreadyExists = errors.New("component already exists")
	// ErrComponentNotInitialized indicates the component has not been initialized.
	ErrComponentNotInitialized = errors.New("component not initialized")
	// ErrComponentLocked indicates the component is locked and cannot be modified.
	ErrComponentLocked = errors.New("component is locked")

	// State transition errors
	// ErrInvalidStateTransition indicates an invalid state transition was attempted.
	ErrInvalidStateTransition = errors.New("invalid state transition")
	// ErrStateTransitionTimeout indicates a state transition timed out.
	ErrStateTransitionTimeout = errors.New("state transition timeout")
	// ErrStateTransitionFailed indicates a state transition failed.
	ErrStateTransitionFailed = errors.New("state transition failed")
	// ErrInvalidTrigger indicates an invalid trigger was provided.
	ErrInvalidTrigger = errors.New("invalid trigger")
	// ErrTransitionInProgress indicates a transition is already in progress.
	ErrTransitionInProgress = errors.New("state transition already in progress")
	// ErrTransitionNotAllowed indicates the transition is not allowed in the current context.
	ErrTransitionNotAllowed = errors.New("state transition not allowed")

	// Persistence errors
	// ErrStatePersistenceFailed indicates state could not be persisted to JetStream.
	ErrStatePersistenceFailed = errors.New("failed to persist state")
	// ErrStateRecoveryFailed indicates state could not be recovered from JetStream.
	ErrStateRecoveryFailed = errors.New("failed to recover state")
	// ErrStreamCreationFailed indicates JetStream stream creation failed.
	ErrStreamCreationFailed = errors.New("failed to create JetStream stream")
	// ErrStreamNotFound indicates the JetStream stream was not found.
	ErrStreamNotFound = errors.New("JetStream stream not found")

	// Communication errors
	// ErrNATSConnectionFailed indicates connection to NATS failed.
	ErrNATSConnectionFailed = errors.New("NATS connection failed")
	// ErrJetStreamInitFailed indicates JetStream initialization failed.
	ErrJetStreamInitFailed = errors.New("JetStream initialization failed")
	// ErrMessagePublishFailed indicates publishing a message failed.
	ErrMessagePublishFailed = errors.New("failed to publish message")
	// ErrBroadcastFailed indicates broadcasting a state change failed.
	ErrBroadcastFailed = errors.New("failed to broadcast state change")

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

	// Host-specific errors
	// ErrHostPowerActionFailed indicates a host power action failed.
	ErrHostPowerActionFailed = errors.New("host power action failed")
	// ErrHostNotResponding indicates the host is not responding.
	ErrHostNotResponding = errors.New("host not responding")
	// ErrHostInvalidState indicates the host is in an invalid state.
	ErrHostInvalidState = errors.New("host in invalid state")
	// ErrHostTransitionNotSupported indicates the requested host transition is not supported.
	ErrHostTransitionNotSupported = errors.New("host transition not supported")

	// Chassis-specific errors
	// ErrChassisControlFailed indicates a chassis control operation failed.
	ErrChassisControlFailed = errors.New("chassis control failed")
	// ErrChassisNotPresent indicates the chassis is not physically present.
	ErrChassisNotPresent = errors.New("chassis not present")
	// ErrChassisPowerSupplyFault indicates a power supply fault in the chassis.
	ErrChassisPowerSupplyFault = errors.New("chassis power supply fault")
	// ErrChassisOverTemperature indicates the chassis is over temperature.
	ErrChassisOverTemperature = errors.New("chassis over temperature")
	// ErrChassisIntrusionDetected indicates chassis intrusion was detected.
	ErrChassisIntrusionDetected = errors.New("chassis intrusion detected")

	// BMC-specific errors
	// ErrBMCNotReady indicates the BMC is not ready for operations.
	ErrBMCNotReady = errors.New("BMC not ready")
	// ErrBMCResetFailed indicates BMC reset failed.
	ErrBMCResetFailed = errors.New("BMC reset failed")
	// ErrBMCFirmwareUpdateInProgress indicates a firmware update is in progress.
	ErrBMCFirmwareUpdateInProgress = errors.New("BMC firmware update in progress")
	// ErrBMCConfigurationError indicates a BMC configuration error.
	ErrBMCConfigurationError = errors.New("BMC configuration error")

	// Validation errors
	// ErrInvalidComponentID indicates an invalid component ID was provided.
	ErrInvalidComponentID = errors.New("invalid component ID")
	// ErrInvalidStateName indicates an invalid state name was provided.
	ErrInvalidStateName = errors.New("invalid state name")
	// ErrInvalidMetadata indicates invalid metadata was provided.
	ErrInvalidMetadata = errors.New("invalid metadata")

	// Concurrency errors
	// ErrConcurrentModification indicates a concurrent modification was detected.
	ErrConcurrentModification = errors.New("concurrent modification detected")
	// ErrResourceLocked indicates the requested resource is locked.
	ErrResourceLocked = errors.New("resource locked")
	// ErrDeadlockDetected indicates a potential deadlock was detected.
	ErrDeadlockDetected = errors.New("deadlock detected")
)
