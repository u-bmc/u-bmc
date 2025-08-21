// SPDX-License-Identifier: BSD-3-Clause

package log

import (
	"fmt"
	"log/slog"

	"github.com/nats-io/nats-server/v2/server"
)

// NATSLogger is an adapter that implements the NATS server.Logger interface
// using the standard library's slog.Logger for structured logging.
type NATSLogger struct {
	l *slog.Logger
}

// Fatalf logs a fatal error message with the given format and arguments.
// This maps to the Error level in slog with additional context indicating it's a fatal error.
func (l *NATSLogger) Fatalf(format string, v ...interface{}) {
	l.l.With("subsystem", "nats", "nats_level", "fatal").Error(fmt.Sprintf(format, v...))
}

// Errorf logs an error message with the given format and arguments.
// This maps to the Error level in slog.
func (l *NATSLogger) Errorf(format string, v ...interface{}) {
	l.l.With("subsystem", "nats", "nats_level", "error").Error(fmt.Sprintf(format, v...))
}

// Warnf logs a warning message with the given format and arguments.
// This maps to the Warn level in slog.
func (l *NATSLogger) Warnf(format string, v ...interface{}) {
	l.l.With("subsystem", "nats", "nats_level", "warn").Warn(fmt.Sprintf(format, v...))
}

// Noticef logs a notice message with the given format and arguments.
// This maps to the Info level in slog as notices are informational.
func (l *NATSLogger) Noticef(format string, v ...interface{}) {
	l.l.With("subsystem", "nats", "nats_level", "info").Info(fmt.Sprintf(format, v...))
}

// Debugf logs a debug message with the given format and arguments.
// This maps to the Debug level in slog.
func (l *NATSLogger) Debugf(format string, v ...interface{}) {
	l.l.With("subsystem", "nats", "nats_level", "debug").Debug(fmt.Sprintf(format, v...))
}

// Tracef logs a trace message with the given format and arguments.
// This maps to the Debug level in slog with additional context indicating it's a trace message.
func (l *NATSLogger) Tracef(format string, v ...interface{}) {
	l.l.With("subsystem", "nats", "nats_level", "trace").Debug(fmt.Sprintf(format, v...))
}

// NewNATSLogger creates a new NATSLogger that wraps the provided slog.Logger
// and implements the NATS server.Logger interface for use with NATS server logging.
func NewNATSLogger(l *slog.Logger) server.Logger {
	return &NATSLogger{
		l: l,
	}
}
