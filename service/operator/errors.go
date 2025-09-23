// SPDX-License-Identifier: BSD-3-Clause

package operator

import "errors"

var (
	// Configuration errors
	// ErrNameEmpty indicates that the operator name cannot be empty.
	ErrNameEmpty = errors.New("operator name cannot be empty")
	// ErrInvalidConfiguration indicates that the operator configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid operator configuration")
	// ErrMissingConfiguration indicates that required configuration is missing.
	ErrMissingConfiguration = errors.New("missing operator configuration")
	// ErrConfigurationConflict indicates conflicting configuration values.
	ErrConfigurationConflict = errors.New("operator configuration conflict")

	// IPC errors
	// ErrIPCNil indicates that IPC service is not configured.
	ErrIPCNil = errors.New("IPC service not configured: provide either ipcConn or WithIPC option")
	// ErrIPCConnectionFailed indicates that IPC connection establishment failed.
	ErrIPCConnectionFailed = errors.New("failed to establish IPC connection")
	// ErrIPCServiceFailed indicates that the IPC service failed to start.
	ErrIPCServiceFailed = errors.New("IPC service startup failed")

	// Service management errors
	// ErrServiceNotFound indicates that a requested service was not found.
	ErrServiceNotFound = errors.New("service not found")
	// ErrServiceStartupFailed indicates that a service failed to start.
	ErrServiceStartupFailed = errors.New("service startup failed")
	// ErrServiceShutdownFailed indicates that a service failed to shutdown gracefully.
	ErrServiceShutdownFailed = errors.New("service shutdown failed")
	// ErrServiceTimeout indicates that a service operation timed out.
	ErrServiceTimeout = errors.New("service operation timeout")
	// ErrServiceDependencyFailed indicates that a service dependency failed.
	ErrServiceDependencyFailed = errors.New("service dependency failed")

	// Process management errors
	// ErrAddProcess indicates that adding a process to supervision failed.
	ErrAddProcess = errors.New("failed to add process to supervision tree")
	// ErrAddExtraService indicates that adding an extra service failed.
	ErrAddExtraService = errors.New("failed to add extra service to supervision tree")
	// ErrProcessCrashed indicates that a supervised process crashed.
	ErrProcessCrashed = errors.New("supervised process crashed")
	// ErrProcessKilled indicates that a supervised process was killed.
	ErrProcessKilled = errors.New("supervised process was killed")

	// System initialization errors
	// ErrSetupMounts indicates that filesystem mount setup failed.
	ErrSetupMounts = errors.New("failed to setup filesystem mounts")
	// ErrIDGeneration indicates that persistent ID generation failed.
	ErrIDGeneration = errors.New("failed to generate persistent ID")
	// ErrSystemInitFailed indicates that system initialization failed.
	ErrSystemInitFailed = errors.New("system initialization failed")
	// ErrTelemetrySetupFailed indicates that telemetry setup failed.
	ErrTelemetrySetupFailed = errors.New("telemetry setup failed")

	// Supervision errors
	// ErrSupervisionTreeFailed indicates that the supervision tree failed.
	ErrSupervisionTreeFailed = errors.New("supervision tree failed")
	// ErrRestartPolicyViolated indicates that restart policy was violated.
	ErrRestartPolicyViolated = errors.New("restart policy violated")
	// ErrMaxRestartsExceeded indicates that maximum restart count was exceeded.
	ErrMaxRestartsExceeded = errors.New("maximum restart count exceeded")
	// ErrSupervisionTimeout indicates that supervision operation timed out.
	ErrSupervisionTimeout = errors.New("supervision operation timeout")

	// Runtime errors
	// ErrPanicked indicates that the operator panicked during execution.
	ErrPanicked = errors.New("operator panicked")
	// ErrDeadlock indicates that a deadlock was detected.
	ErrDeadlock = errors.New("deadlock detected")
	// ErrResourceExhausted indicates that system resources are exhausted.
	ErrResourceExhausted = errors.New("system resources exhausted")
	// ErrMemoryLimitExceeded indicates that memory limits were exceeded.
	ErrMemoryLimitExceeded = errors.New("memory limit exceeded")

	// Lifecycle errors
	// ErrStartupTimeout indicates that startup operation timed out.
	ErrStartupTimeout = errors.New("operator startup timeout")
	// ErrShutdownTimeout indicates that shutdown operation timed out.
	ErrShutdownTimeout = errors.New("operator shutdown timeout")
	// ErrGracefulShutdownFailed indicates that graceful shutdown failed.
	ErrGracefulShutdownFailed = errors.New("graceful shutdown failed")
	// ErrForcedShutdown indicates that forced shutdown was required.
	ErrForcedShutdown = errors.New("forced shutdown required")

	// Health and monitoring errors
	// ErrHealthCheckFailed indicates that health check failed.
	ErrHealthCheckFailed = errors.New("operator health check failed")
	// ErrMetricsCollectionFailed indicates that metrics collection failed.
	ErrMetricsCollectionFailed = errors.New("metrics collection failed")
	// ErrTracingSetupFailed indicates that tracing setup failed.
	ErrTracingSetupFailed = errors.New("tracing setup failed")
	// ErrLoggingSetupFailed indicates that logging setup failed.
	ErrLoggingSetupFailed = errors.New("logging setup failed")

	// Security and permissions errors
	// ErrPermissionDenied indicates that permission was denied.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrUnauthorizedAccess indicates unauthorized access attempt.
	ErrUnauthorizedAccess = errors.New("unauthorized access attempt")
	// ErrSecurityViolation indicates that a security violation occurred.
	ErrSecurityViolation = errors.New("security violation detected")

	// Storage and filesystem errors
	// ErrStorageNotAvailable indicates that storage is not available.
	ErrStorageNotAvailable = errors.New("storage not available")
	// ErrFilesystemCorrupted indicates that filesystem corruption was detected.
	ErrFilesystemCorrupted = errors.New("filesystem corruption detected")
	// ErrDiskSpaceExhausted indicates that disk space is exhausted.
	ErrDiskSpaceExhausted = errors.New("disk space exhausted")
	// ErrMountPointUnavailable indicates that a mount point is unavailable.
	ErrMountPointUnavailable = errors.New("mount point unavailable")

	// Network and communication errors
	// ErrNetworkUnavailable indicates that network is unavailable.
	ErrNetworkUnavailable = errors.New("network unavailable")
	// ErrCommunicationFailed indicates that inter-service communication failed.
	ErrCommunicationFailed = errors.New("inter-service communication failed")
	// ErrConnectionLost indicates that a connection was lost.
	ErrConnectionLost = errors.New("connection lost")
	// ErrProtocolError indicates that a protocol error occurred.
	ErrProtocolError = errors.New("protocol error")
)
