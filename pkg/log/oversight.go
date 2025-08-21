// SPDX-License-Identifier: BSD-3-Clause

package log

import (
	"fmt"
	"log/slog"

	"cirello.io/oversight/v2"
)

// NewOversightLogger creates an oversight.Logger that wraps the provided slog.Logger.
// The returned logger will log oversight messages at the Debug level with the prefix "oversight".
// This is useful for integrating oversight supervision tree logging with structured logging.
func NewOversightLogger(l *slog.Logger) oversight.Logger {
	return func(args ...any) {
		l.Debug("oversight", "msg", fmt.Sprint(args...))
	}
}
