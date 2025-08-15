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
