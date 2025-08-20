// SPDX-License-Identifier: BSD-3-Clause

package log

import "log/slog"

// QLogger is a wrapper around slog.Logger that implements io.WriteCloser
// for compatibility with libraries like QUIC that expect a writer interface.
type QLogger struct {
	l *slog.Logger
}

// Write implements io.Writer by logging the provided bytes as an Info level message.
// It converts the byte slice to a string and logs it using the underlying slog.Logger.
// Always returns the length of the input and a nil error.
func (l *QLogger) Write(b []byte) (n int, err error) {
	l.l.Info(string(b))

	return len(b), nil
}

// Close implements io.Closer. This implementation has no resources to clean up,
// so it always returns nil.
func (l *QLogger) Close() error {
	// No resources to close in this implementation
	return nil
}

// NewQLogger creates a new QLogger that wraps the provided slog.Logger.
// The returned QLogger can be used anywhere an io.WriteCloser is expected.
func NewQLogger(l *slog.Logger) *QLogger {
	return &QLogger{
		l: l,
	}
}
