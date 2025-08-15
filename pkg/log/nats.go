// SPDX-License-Identifier: BSD-3-Clause

package log

import (
	"fmt"
	"log/slog"

	"github.com/nats-io/nats-server/v2/server"
)

type NATSLogger struct {
	l *slog.Logger
}

func (l *NATSLogger) Fatalf(format string, v ...interface{}) {
	l.l.Error(fmt.Sprintf(format, v...), "NATS fatal error")
}

func (l *NATSLogger) Errorf(format string, v ...interface{}) {
	l.l.Error(fmt.Sprintf(format, v...), "NATS error")
}

func (l *NATSLogger) Warnf(format string, v ...interface{}) {
	l.l.Warn(fmt.Sprintf(format, v...), "NATS warning")
}

func (l *NATSLogger) Noticef(format string, v ...interface{}) {
	l.l.Info(fmt.Sprintf(format, v...))
}

func (l *NATSLogger) Debugf(format string, v ...interface{}) {
	l.l.Debug(fmt.Sprintf(format, v...))
}

func (l *NATSLogger) Tracef(format string, v ...interface{}) {
	l.l.Debug(fmt.Sprintf(format, v...), "NATS trace")
}

func NewNATSLogger(l *slog.Logger) server.Logger {
	return &NATSLogger{
		l: l,
	}
}
