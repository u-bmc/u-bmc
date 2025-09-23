// SPDX-License-Identifier: BSD-3-Clause

package ipc

import "errors"

var (
	// Connection errors
	// ErrConnectionFailed indicates that a connection could not be established.
	ErrConnectionFailed = errors.New("failed to establish IPC connection")
	// ErrConnectionUnavailable indicates that no connection is available.
	ErrConnectionUnavailable = errors.New("IPC connection not available")
	// ErrConnectionClosed indicates that the connection has been closed.
	ErrConnectionClosed = errors.New("IPC connection closed")
	// ErrConnectionTimeout indicates that a connection operation timed out.
	ErrConnectionTimeout = errors.New("IPC connection timeout")

	// Provider errors
	// ErrProviderNotReady indicates that the connection provider is not ready.
	ErrProviderNotReady = errors.New("IPC connection provider not ready")
	// ErrProviderUnavailable indicates that the connection provider is unavailable.
	ErrProviderUnavailable = errors.New("IPC connection provider unavailable")
	// ErrProviderInitFailed indicates that provider initialization failed.
	ErrProviderInitFailed = errors.New("IPC connection provider initialization failed")

	// Communication errors
	// ErrRequestFailed indicates that a request operation failed.
	ErrRequestFailed = errors.New("IPC request failed")
	// ErrResponseFailed indicates that a response operation failed.
	ErrResponseFailed = errors.New("IPC response failed")
	// ErrMessageCorrupted indicates that a message was corrupted during transmission.
	ErrMessageCorrupted = errors.New("IPC message corrupted")
	// ErrMessageTooLarge indicates that a message exceeds size limits.
	ErrMessageTooLarge = errors.New("IPC message too large")

	// Service errors
	// ErrServiceNotFound indicates that a requested service was not found.
	ErrServiceNotFound = errors.New("IPC service not found")
	// ErrServiceUnavailable indicates that a service is temporarily unavailable.
	ErrServiceUnavailable = errors.New("IPC service unavailable")
	// ErrServiceBusy indicates that a service is currently busy.
	ErrServiceBusy = errors.New("IPC service busy")
	// ErrServiceShutdown indicates that a service is shutting down.
	ErrServiceShutdown = errors.New("IPC service shutting down")

	// Authentication and authorization errors
	// ErrUnauthorized indicates that the request lacks proper authorization.
	ErrUnauthorized = errors.New("IPC request unauthorized")
	// ErrPermissionDenied indicates that permission was denied for the operation.
	ErrPermissionDenied = errors.New("IPC permission denied")
	// ErrAuthenticationFailed indicates that authentication failed.
	ErrAuthenticationFailed = errors.New("IPC authentication failed")

	// Protocol errors
	// ErrInvalidRequest indicates that the request format is invalid.
	ErrInvalidRequest = errors.New("invalid IPC request format")
	// ErrInvalidResponse indicates that the response format is invalid.
	ErrInvalidResponse = errors.New("invalid IPC response format")
	// ErrProtocolViolation indicates that the communication protocol was violated.
	ErrProtocolViolation = errors.New("IPC protocol violation")
	// ErrUnsupportedOperation indicates that the operation is not supported.
	ErrUnsupportedOperation = errors.New("IPC operation not supported")

	// Resource errors
	// ErrResourceExhausted indicates that system resources are exhausted.
	ErrResourceExhausted = errors.New("IPC resources exhausted")
	// ErrMemoryLimitExceeded indicates that memory limits were exceeded.
	ErrMemoryLimitExceeded = errors.New("IPC memory limit exceeded")
	// ErrConnectionLimitExceeded indicates that connection limits were exceeded.
	ErrConnectionLimitExceeded = errors.New("IPC connection limit exceeded")

	// Configuration errors
	// ErrInvalidConfiguration indicates that the IPC configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid IPC configuration")
	// ErrMissingConfiguration indicates that required configuration is missing.
	ErrMissingConfiguration = errors.New("missing IPC configuration")
	// ErrConfigurationConflict indicates conflicting configuration values.
	ErrConfigurationConflict = errors.New("IPC configuration conflict")

	// Network errors
	// ErrNetworkError indicates a network-related error occurred.
	ErrNetworkError = errors.New("IPC network error")
	// ErrAddressInUse indicates that the address is already in use.
	ErrAddressInUse = errors.New("IPC address already in use")
	// ErrAddressNotAvailable indicates that the address is not available.
	ErrAddressNotAvailable = errors.New("IPC address not available")

	// Serialization errors
	// ErrSerializationFailed indicates that data serialization failed.
	ErrSerializationFailed = errors.New("IPC data serialization failed")
	// ErrDeserializationFailed indicates that data deserialization failed.
	ErrDeserializationFailed = errors.New("IPC data deserialization failed")
	// ErrEncodingError indicates an encoding error occurred.
	ErrEncodingError = errors.New("IPC encoding error")
	// ErrDecodingError indicates a decoding error occurred.
	ErrDecodingError = errors.New("IPC decoding error")
)
