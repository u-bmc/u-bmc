// SPDX-License-Identifier: BSD-3-Clause

// Package telemetry provides reusable abstractions around the OpenTelemetry SDK for Go.
// It offers simplified configuration and setup for metrics, traces, and logs collection
// with support for various exporters including OTLP HTTP, OTLP gRPC, and no-op providers.
//
// The package supports four main operation modes:
//   - NoOp: Discards all telemetry data with minimal overhead
//   - OTLP HTTP: Exports telemetry data via OTLP over HTTP
//   - OTLP gRPC: Exports telemetry data via OTLP over gRPC
//   - Dual: Exports telemetry data via both HTTP and gRPC
//
// Example usage:
//
//	// Initialize with OTLP HTTP exporter
//	provider, err := telemetry.NewProvider(telemetry.WithOTLPHTTP("http://localhost:4318"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer provider.Shutdown(context.Background())
//
//	// Use the provider for telemetry collection
//	tracer := provider.Tracer("my-service")
package telemetry
