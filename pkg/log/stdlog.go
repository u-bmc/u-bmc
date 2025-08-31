// SPDX-License-Identifier: BSD-3-Clause

package log

import (
	"log"
	"log/slog"
)

func NewStdLoggerAt(logger *slog.Logger, level slog.Level) *log.Logger {
	return slog.NewLogLogger(logger.Handler(), level)
}

func RedirectStdLog(l *slog.Logger) {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(NewStdLoggerAt(l, slog.LevelInfo).Writer())
}
