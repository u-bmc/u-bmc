// SPDX-License-Identifier: BSD-3-Clause

// Package telemetry provides a service that acts as a telemetry collector and aggregator.
// It implements the u-bmc service interface and provides OpenTelemetry data collection
// capabilities with support for various export formats including OTLP HTTP, OTLP gRPC,
// and no-op operation modes.
//
// The telemetry service can be configured to:
//   - Collect metrics, traces, and logs from other u-bmc services
//   - Apply filtering and aggregation as supported by the OTLP specification
//   - Export telemetry data to configured OTLP-compatible endpoints
//   - Operate in no-op mode for minimal overhead when telemetry is not needed
//
// The service supports four main operation modes:
//   - NoOp: Discards all telemetry data with minimal overhead
//   - OTLP HTTP: Exports telemetry data via OTLP over HTTP
//   - OTLP gRPC: Exports telemetry data via OTLP over gRPC
//   - Dual: Exports telemetry data via both HTTP and gRPC protocols
//
// Configuration is handled through functional options that allow fine-tuning
// of export behavior, sampling rates, batch sizes, and endpoint settings.
package telemetry
