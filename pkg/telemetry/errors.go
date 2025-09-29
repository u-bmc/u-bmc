// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import "errors"

var (
	// ErrInvalidExporterType is returned when an unsupported exporter type is specified.
	ErrInvalidExporterType = errors.New("invalid exporter type")

	// ErrMissingEndpoint is returned when an endpoint is required but not provided.
	ErrMissingEndpoint = errors.New("missing endpoint")

	// ErrProviderNotInitialized is returned when attempting to use a provider that hasn't been initialized.
	ErrProviderNotInitialized = errors.New("provider not initialized")

	// ErrInvalidConfiguration is returned when the telemetry configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrShutdownFailed is returned when the provider fails to shutdown cleanly.
	ErrShutdownFailed = errors.New("shutdown failed")

	// ErrExporterSetupFailed is returned when an exporter fails to initialize.
	ErrExporterSetupFailed = errors.New("exporter setup failed")

	// ErrInvalidHeaders is returned when invalid headers are provided for OTLP exporters.
	ErrInvalidHeaders = errors.New("invalid headers")

	// ErrConnectionFailed is returned when connection to OTLP endpoint fails.
	ErrConnectionFailed = errors.New("connection failed")
)
