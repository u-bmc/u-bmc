// SPDX-License-Identifier: BSD-3-Clause

package log

import "log/slog"

type QLogger struct {
	l *slog.Logger
}

func (l *QLogger) Write(b []byte) (n int, err error) {
	l.l.Info(string(b))

	return len(b), nil
}

func (l *QLogger) Close() error {
	// No resources to close in this implementation
	return nil
}

func NewQLogger(l *slog.Logger) *QLogger {
	return &QLogger{
		l: l,
	}
}
