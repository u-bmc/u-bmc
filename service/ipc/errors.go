// SPDX-License-Identifier: BSD-3-Clause

package ipc

import "errors"

var (
	// Service-level errors
	// ErrServiceNotStarted indicates the IPC service has not been started.
	ErrServiceNotStarted = errors.New("IPC service not started")
	// ErrServiceAlreadyStarted indicates the IPC service is already running.
	ErrServiceAlreadyStarted = errors.New("IPC service already started")
	// ErrServiceStopped indicates the IPC service has been stopped.
	ErrServiceStopped = errors.New("IPC service stopped")
	// ErrInvalidConfiguration indicates the service configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid IPC service configuration")

	// Server errors
	// ErrServerCreationFailed indicates NATS server creation failed.
	ErrServerCreationFailed = errors.New("failed to create NATS server")
	// ErrServerStartFailed indicates NATS server startup failed.
	ErrServerStartFailed = errors.New("failed to start NATS server")
	// ErrServerNotReady indicates the NATS server is not ready for connections.
	ErrServerNotReady = errors.New("NATS server not ready for connections")
	// ErrServerShutdownFailed indicates NATS server shutdown failed.
	ErrServerShutdownFailed = errors.New("failed to shutdown NATS server")
	// ErrServerTimeout indicates a server operation timed out.
	ErrServerTimeout = errors.New("NATS server operation timeout")

	// Connection errors
	// ErrConnectionFailed indicates connection creation failed.
	ErrConnectionFailed = errors.New("failed to create connection")
	// ErrInProcessConnFailed indicates in-process connection creation failed.
	ErrInProcessConnFailed = errors.New("failed to create in-process connection")
	// ErrConnectionNotAvailable indicates no connection is available.
	ErrConnectionNotAvailable = errors.New("connection not available")
	// ErrConnectionTimeout indicates a connection operation timed out.
	ErrConnectionTimeout = errors.New("connection timeout")

	// JetStream errors
	// ErrJetStreamNotEnabled indicates JetStream is not enabled on the server.
	ErrJetStreamNotEnabled = errors.New("JetStream not enabled")
	// ErrJetStreamInitFailed indicates JetStream initialization failed.
	ErrJetStreamInitFailed = errors.New("JetStream initialization failed")
	// ErrStreamCreationFailed indicates JetStream stream creation failed.
	ErrStreamCreationFailed = errors.New("failed to create JetStream stream")
	// ErrConsumerCreationFailed indicates JetStream consumer creation failed.
	ErrConsumerCreationFailed = errors.New("failed to create JetStream consumer")

	// Storage errors
	// ErrStorageDirInvalid indicates the storage directory is invalid.
	ErrStorageDirInvalid = errors.New("invalid storage directory")
	// ErrStorageDirNotWritable indicates the storage directory is not writable.
	ErrStorageDirNotWritable = errors.New("storage directory not writable")
	// ErrStorageSpaceInsufficient indicates insufficient storage space.
	ErrStorageSpaceInsufficient = errors.New("insufficient storage space")
	// ErrStorageCorrupted indicates storage corruption was detected.
	ErrStorageCorrupted = errors.New("storage corruption detected")

	// Configuration errors
	// ErrInvalidServerName indicates an invalid server name was provided.
	ErrInvalidServerName = errors.New("invalid server name")
	// ErrInvalidPort indicates an invalid port was specified.
	ErrInvalidPort = errors.New("invalid port")
	// ErrInvalidHost indicates an invalid host was specified.
	ErrInvalidHost = errors.New("invalid host")
	// ErrInvalidTimeout indicates an invalid timeout value was provided.
	ErrInvalidTimeout = errors.New("invalid timeout value")

	// Lifecycle errors
	// ErrShutdownTimeout indicates shutdown operation timed out.
	ErrShutdownTimeout = errors.New("shutdown timeout")
	// ErrStartupTimeout indicates startup operation timed out.
	ErrStartupTimeout = errors.New("startup timeout")
	// ErrGracefulShutdownFailed indicates graceful shutdown failed.
	ErrGracefulShutdownFailed = errors.New("graceful shutdown failed")

	// Resource errors
	// ErrResourceExhausted indicates system resources are exhausted.
	ErrResourceExhausted = errors.New("system resources exhausted")
	// ErrMemoryLimitExceeded indicates memory limit was exceeded.
	ErrMemoryLimitExceeded = errors.New("memory limit exceeded")
	// ErrFileDescriptorLimitExceeded indicates file descriptor limit was exceeded.
	ErrFileDescriptorLimitExceeded = errors.New("file descriptor limit exceeded")

	// Security errors
	// ErrUnauthorizedAccess indicates unauthorized access attempt.
	ErrUnauthorizedAccess = errors.New("unauthorized access")
	// ErrPermissionDenied indicates permission was denied.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrTLSConfigurationFailed indicates TLS configuration failed.
	ErrTLSConfigurationFailed = errors.New("TLS configuration failed")

	// Monitoring errors
	// ErrMetricsCollectionFailed indicates metrics collection failed.
	ErrMetricsCollectionFailed = errors.New("metrics collection failed")
	// ErrHealthCheckFailed indicates health check failed.
	ErrHealthCheckFailed = errors.New("health check failed")
	// ErrTracingSetupFailed indicates tracing setup failed.
	ErrTracingSetupFailed = errors.New("tracing setup failed")
)
