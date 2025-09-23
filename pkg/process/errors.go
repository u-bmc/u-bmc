// SPDX-License-Identifier: BSD-3-Clause

package process

import "errors"

var (
	// ErrServicePanic indicates a service panicked during execution.
	ErrServicePanic = errors.New("service panicked during execution")
	// ErrServiceInitialization indicates a failure during service initialization.
	ErrServiceInitialization = errors.New("service initialization failed")
	// ErrServiceShutdown indicates a failure during service shutdown.
	ErrServiceShutdown = errors.New("service shutdown failed")
	// ErrIPCConnection indicates a failure to establish or use the IPC connection.
	ErrIPCConnection = errors.New("IPC connection error")
	// ErrServiceTimeout indicates a service operation timed out.
	ErrServiceTimeout = errors.New("service operation timed out")
	// ErrServiceNotRunning indicates an operation was attempted on a non-running service.
	ErrServiceNotRunning = errors.New("service is not running")
	// ErrServiceAlreadyRunning indicates an attempt to start an already running service.
	ErrServiceAlreadyRunning = errors.New("service is already running")
	// ErrInvalidService indicates an invalid or nil service was provided.
	ErrInvalidService = errors.New("invalid service provided")
	// ErrContextCanceled indicates the service context was canceled.
	ErrContextCanceled = errors.New("service context was canceled")
	// ErrOversightConfiguration indicates an error in oversight tree configuration.
	ErrOversightConfiguration = errors.New("oversight configuration error")
	// ErrChildProcessCreation indicates a failure to create a child process.
	ErrChildProcessCreation = errors.New("failed to create child process")
	// ErrServiceRecovery indicates a failure during service recovery after panic.
	ErrServiceRecovery = errors.New("service recovery failed")
)
