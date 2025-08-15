// SPDX-License-Identifier: BSD-3-Clause

package log

import (
	"log/slog"

	"cirello.io/oversight/v2"
)

func NewOversightLogger(l *slog.Logger) oversight.Logger {
	return func(args ...any) {
		l.Debug("oversight", args...)
	}
}
