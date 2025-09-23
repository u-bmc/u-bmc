// SPDX-License-Identifier: BSD-3-Clause

package log

import (
	"log"
	"log/slog"
)

// NewStdLoggerAt creates a new standard library log.Logger that wraps the provided
// slog.Logger and logs all messages at the specified level. This is useful for
// integrating third-party libraries that expect a standard log.Logger interface.
func NewStdLoggerAt(logger *slog.Logger, level slog.Level) *log.Logger {
	return slog.NewLogLogger(logger.Handler(), level)
}

// RedirectStdLog configures the standard library log package to output through
// the provided slog.Logger at Info level. This redirects all standard log output
// to use structured logging, ensuring consistent log formatting across the application.
func RedirectStdLog(l *slog.Logger) {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(NewStdLoggerAt(l, slog.LevelInfo).Writer())
}
