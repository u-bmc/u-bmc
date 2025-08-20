// SPDX-License-Identifier: BSD-3-Clause

package log

import (
	"log/slog"

	"github.com/rs/zerolog"
	slogmulti "github.com/samber/slog-multi"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
)

// NewDefaultLogger creates a new structured logger that outputs to both console and OpenTelemetry.
// The logger uses zerolog for console output with timestamps and debug level logging,
// and sends telemetry data to the global OpenTelemetry logger provider.
// This is the recommended way to create a new logger instance for application use.
func NewDefaultLogger() *slog.Logger {
	zeroLogger := zerolog.
		New(zerolog.NewConsoleWriter()).
		With().
		Timestamp().
		Logger()

	provider := global.GetLoggerProvider()

	otelHandler := otelslog.NewHandler("u-bmc", otelslog.WithLoggerProvider(provider))
	return slog.New(slogmulti.Fanout(
		slogzerolog.Option{Level: slog.LevelDebug, Logger: &zeroLogger}.NewZerologHandler(),
		otelHandler,
	))
}

// GetGlobalLogger returns a structured logger configured for global application use.
// Like NewDefaultLogger, it outputs to both console and OpenTelemetry with debug level logging.
// The logger uses zerolog for human-readable console output with timestamps,
// while simultaneously sending structured log data to OpenTelemetry for observability.
// Use this function when you need a logger instance that matches the global logging configuration.
func GetGlobalLogger() *slog.Logger {
	zeroLogger := zerolog.
		New(zerolog.NewConsoleWriter()).
		With().
		Timestamp().
		Logger()

	provider := global.GetLoggerProvider()

	otelHandler := otelslog.NewHandler("u-bmc", otelslog.WithLoggerProvider(provider))
	return slog.New(slogmulti.Fanout(
		slogzerolog.Option{Level: slog.LevelDebug, Logger: &zeroLogger}.NewZerologHandler(),
		otelHandler,
	))
}
