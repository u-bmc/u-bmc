// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import "errors"

var (
	// ErrServiceNotConfigured is returned when the telemetry service is not properly configured.
	ErrServiceNotConfigured = errors.New("service not configured")

	// ErrInvalidServiceName is returned when an invalid service name is provided.
	ErrInvalidServiceName = errors.New("invalid service name")

	// ErrProviderInitializationFailed is returned when the telemetry provider fails to initialize.
	ErrProviderInitializationFailed = errors.New("provider initialization failed")

	// ErrServiceAlreadyRunning is returned when attempting to start a service that is already running.
	ErrServiceAlreadyRunning = errors.New("service already running")

	// ErrServiceNotRunning is returned when attempting to stop a service that is not running.
	ErrServiceNotRunning = errors.New("service not running")

	// ErrExportConfigurationInvalid is returned when the export configuration is invalid.
	ErrExportConfigurationInvalid = errors.New("export configuration invalid")

	// ErrCollectorSetupFailed is returned when the telemetry collector setup fails.
	ErrCollectorSetupFailed = errors.New("collector setup failed")

	// ErrAggregatorSetupFailed is returned when the telemetry aggregator setup fails.
	ErrAggregatorSetupFailed = errors.New("aggregator setup failed")

	// ErrMetricsCollectionFailed is returned when metrics collection fails.
	ErrMetricsCollectionFailed = errors.New("metrics collection failed")

	// ErrTracesCollectionFailed is returned when traces collection fails.
	ErrTracesCollectionFailed = errors.New("traces collection failed")

	// ErrLogsCollectionFailed is returned when logs collection fails.
	ErrLogsCollectionFailed = errors.New("logs collection failed")

	// ErrFilterSetupFailed is returned when telemetry filter setup fails.
	ErrFilterSetupFailed = errors.New("filter setup failed")

	// ErrShutdownTimeout is returned when the service shutdown times out.
	ErrShutdownTimeout = errors.New("shutdown timeout")

	// ErrContextCancelled is returned when the service context is cancelled.
	ErrContextCancelled = errors.New("context cancelled")

	// ErrIPCConnectionFailed is returned when IPC connection fails.
	ErrIPCConnectionFailed = errors.New("IPC connection failed")
)
