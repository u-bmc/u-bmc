// SPDX-License-Identifier: BSD-3-Clause

package log

import "errors"

var (
	// ErrLoggerInitialization indicates a failure during logger initialization.
	ErrLoggerInitialization = errors.New("failed to initialize logger")
	// ErrLoggerConfiguration indicates an invalid logger configuration.
	ErrLoggerConfiguration = errors.New("invalid logger configuration")
	// ErrHandlerCreation indicates a failure to create a log handler.
	ErrHandlerCreation = errors.New("failed to create log handler")
	// ErrOutputTarget indicates a failure with a log output target.
	ErrOutputTarget = errors.New("log output target error")
	// ErrTelemetryProvider indicates a failure with the OpenTelemetry provider.
	ErrTelemetryProvider = errors.New("OpenTelemetry provider error")
	// ErrNATSLogger indicates a failure in the NATS logger adapter.
	ErrNATSLogger = errors.New("NATS logger adapter error")
	// ErrOversightLogger indicates a failure in the oversight logger adapter.
	ErrOversightLogger = errors.New("oversight logger adapter error")
	// ErrQUICLogger indicates a failure in the QUIC logger adapter.
	ErrQUICLogger = errors.New("QUIC logger adapter error")
	// ErrLogLevel indicates an invalid log level configuration.
	ErrLogLevel = errors.New("invalid log level")
	// ErrLogFormat indicates an invalid log format configuration.
	ErrLogFormat = errors.New("invalid log format")
	// ErrConsoleWriter indicates a failure with the console writer.
	ErrConsoleWriter = errors.New("console writer error")
	// ErrStructuredLogging indicates a failure in structured logging operations.
	ErrStructuredLogging = errors.New("structured logging error")
)
