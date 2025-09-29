// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import "errors"

var (
	// ErrServiceAlreadyStarted indicates that the sensor monitoring service is already running.
	ErrServiceAlreadyStarted = errors.New("sensor monitoring service already started")
	// ErrServiceNotStarted indicates that the sensor monitoring service is not running.
	ErrServiceNotStarted = errors.New("sensor monitoring service not started")
	// ErrInvalidConfiguration indicates that the service configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid sensormon configuration")
	// ErrNATSConnectionFailed indicates that the NATS connection could not be established.
	ErrNATSConnectionFailed = errors.New("failed to connect to NATS")
	// ErrJetStreamInitFailed indicates that JetStream initialization failed.
	ErrJetStreamInitFailed = errors.New("failed to initialize JetStream")
	// ErrMicroServiceCreationFailed indicates that micro service creation failed.
	ErrMicroServiceCreationFailed = errors.New("failed to create micro service")
	// ErrEndpointRegistrationFailed indicates that endpoint registration failed.
	ErrEndpointRegistrationFailed = errors.New("failed to register endpoint")
	// ErrSensorNotFound indicates that the specified sensor was not found.
	ErrSensorNotFound = errors.New("sensor not found")
	// ErrSensorReadFailed indicates that reading from a sensor failed.
	ErrSensorReadFailed = errors.New("sensor read failed")
	// ErrSensorConfigurationFailed indicates that sensor configuration failed.
	ErrSensorConfigurationFailed = errors.New("sensor configuration failed")
	// ErrInvalidSensorRequest indicates that the sensor request is invalid.
	ErrInvalidSensorRequest = errors.New("invalid sensor request")
	// ErrMonitoringNotStarted indicates that sensor monitoring is not active.
	ErrMonitoringNotStarted = errors.New("sensor monitoring not started")
	// ErrMonitoringAlreadyStarted indicates that sensor monitoring is already active.
	ErrMonitoringAlreadyStarted = errors.New("sensor monitoring already started")
	// ErrThresholdViolation indicates that a sensor threshold has been exceeded.
	ErrThresholdViolation = errors.New("sensor threshold violation")
	// ErrSensorDiscoveryFailed indicates that sensor discovery failed.
	ErrSensorDiscoveryFailed = errors.New("sensor discovery failed")
	// ErrHwmonAccessFailed indicates that hwmon access failed.
	ErrHwmonAccessFailed = errors.New("hwmon access failed")
	// ErrGPIOAccessFailed indicates that GPIO access failed.
	ErrGPIOAccessFailed = errors.New("GPIO access failed")
	// ErrSensorTypeUnsupported indicates that the sensor type is not supported.
	ErrSensorTypeUnsupported = errors.New("sensor type unsupported")
	// ErrInvalidFieldMask indicates that the provided field mask is invalid.
	ErrInvalidFieldMask = errors.New("invalid field mask")
	// ErrOperationTimeout indicates that a sensor operation timed out.
	ErrOperationTimeout = errors.New("sensor operation timeout")
	// ErrRequestMarshalingFailed indicates that request marshaling failed.
	ErrRequestMarshalingFailed = errors.New("request marshaling failed")
	// ErrResponseMarshalingFailed indicates that response marshaling failed.
	ErrResponseMarshalingFailed = errors.New("response marshaling failed")
	// ErrContextCanceled indicates that the operation was canceled.
	ErrContextCanceled = errors.New("operation canceled")
)
