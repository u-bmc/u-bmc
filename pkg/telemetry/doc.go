// SPDX-License-Identifier: BSD-3-Clause

// Package telemetry provides OpenTelemetry integration and distributed tracing
// utilities for the u-bmc system. This package simplifies the setup and
// configuration of OpenTelemetry components including logging, tracing, and
// metrics collection, while providing utilities for context propagation
// across service boundaries.
//
// The package is designed to provide observability for distributed BMC
// systems where multiple services communicate via NATS messaging. It enables
// correlation of logs, traces, and metrics across service boundaries for
// comprehensive system monitoring and debugging.
//
// # Core Features
//
// The package provides several key capabilities:
//
//   - Default OpenTelemetry setup with no-op providers for development
//   - Context propagation utilities for distributed tracing
//   - Integration with NATS micro services for trace context extraction
//   - Standardized telemetry configuration for consistent observability
//   - Support for both development and production telemetry backends
//
// # Basic Setup
//
// Initialize OpenTelemetry with default configuration:
//
//	func main() {
//		// Initialize telemetry before any other components
//		telemetry.DefaultSetup()
//
//		// Continue with application initialization
//		logger := log.GetGlobalLogger()
//		logger.Info("Application starting with telemetry enabled")
//
//		// Your application code here
//	}
//
// # Distributed Tracing with NATS
//
// Extract and propagate trace context in NATS micro services:
//
//	func setupMicroService() error {
//		nc, err := nats.Connect(nats.DefaultURL)
//		if err != nil {
//			return err
//		}
//
//		config := micro.Config{
//			Name:    "device-service",
//			Version: "1.0.0",
//		}
//
//		svc, err := micro.AddService(nc, config)
//		if err != nil {
//			return err
//		}
//
//		// Add endpoint with trace context extraction
//		return svc.AddEndpoint("device.info", micro.HandlerFunc(func(req micro.Request) {
//			// Extract distributed trace context from request headers
//			ctx := telemetry.GetCtxFromReq(req)
//
//			// Use context for operations - traces will be correlated
//			deviceInfo, err := getDeviceInfo(ctx, req.Data())
//			if err != nil {
//				req.Error("500", "Internal Server Error", nil)
//				return
//			}
//
//			req.Respond(deviceInfo)
//		}))
//	}
//
// # Manual Context Propagation
//
// For scenarios where manual context propagation is needed:
//
//	func processRequest(ctx context.Context, request []byte) error {
//		// Create a new span for this operation
//		tracer := otel.Tracer("bmc-service")
//		ctx, span := tracer.Start(ctx, "process_request")
//		defer span.End()
//
//		// Add attributes to the span
//		span.SetAttributes(
//			attribute.String("request.size", fmt.Sprintf("%d", len(request))),
//			attribute.String("service.component", "request_processor"),
//		)
//
//		// Process the request with trace context
//		result, err := doProcessing(ctx, request)
//		if err != nil {
//			span.RecordError(err)
//			span.SetStatus(codes.Error, err.Error())
//			return err
//		}
//
//		span.SetAttributes(
//			attribute.String("result.status", "success"),
//			attribute.Int("result.size", len(result)),
//		)
//
//		return nil
//	}
//
// # Service-to-Service Communication
//
// Propagating trace context between services via NATS:
//
//	func callRemoteService(ctx context.Context, nc *nats.Conn, data []byte) ([]byte, error) {
//		// Create a new span for the remote call
//		tracer := otel.Tracer("bmc-client")
//		ctx, span := tracer.Start(ctx, "remote_service_call")
//		defer span.End()
//
//		// Create request with trace context in headers
//		msg := &nats.Msg{
//			Subject: "remote.service.endpoint",
//			Data:    data,
//			Header:  make(nats.Header),
//		}
//
//		// Inject trace context into message headers
//		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(msg.Header))
//
//		// Send request and wait for response
//		response, err := nc.RequestMsg(msg, 30*time.Second)
//		if err != nil {
//			span.RecordError(err)
//			span.SetStatus(codes.Error, "Remote call failed")
//			return nil, fmt.Errorf("remote service call failed: %w", err)
//		}
//
//		span.SetAttributes(
//			attribute.String("response.status", "success"),
//			attribute.Int("response.size", len(response.Data)),
//		)
//
//		return response.Data, nil
//	}
//
// # BMC Device Monitoring
//
// Using telemetry for hardware monitoring and alerting:
//
//	func monitorDeviceHealth(ctx context.Context, deviceID string) error {
//		tracer := otel.Tracer("bmc-monitor")
//		ctx, span := tracer.Start(ctx, "device_health_check")
//		defer span.End()
//
//		span.SetAttributes(
//			attribute.String("device.id", deviceID),
//			attribute.String("monitor.type", "health_check"),
//		)
//
//		// Get device metrics
//		metrics, err := getDeviceMetrics(ctx, deviceID)
//		if err != nil {
//			span.RecordError(err)
//			span.SetStatus(codes.Error, "Failed to get device metrics")
//			return err
//		}
//
//		// Add metrics as span attributes
//		span.SetAttributes(
//			attribute.Float64("device.temperature", metrics.Temperature),
//			attribute.Float64("device.cpu_usage", metrics.CPUUsage),
//			attribute.Float64("device.memory_usage", metrics.MemoryUsage),
//			attribute.Int64("device.fan_speed", metrics.FanSpeed),
//		)
//
//		// Check for alerts
//		if metrics.Temperature > 80.0 {
//			span.AddEvent("temperature_alert", trace.WithAttributes(
//				attribute.Float64("threshold", 80.0),
//				attribute.Float64("current", metrics.Temperature),
//			))
//
//			// Send alert with trace context
//			return sendAlert(ctx, deviceID, "high_temperature", metrics)
//		}
//
//		span.SetStatus(codes.Ok, "Device healthy")
//		return nil
//	}
//
// # HTTP Server Integration
//
// Adding telemetry to HTTP servers:
//
//	func setupHTTPServer() error {
//		mux := http.NewServeMux()
//
//		// Add telemetry middleware
//		mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
//			// Extract trace context from HTTP headers
//			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
//
//			tracer := otel.Tracer("bmc-api")
//			ctx, span := tracer.Start(ctx, "get_devices")
//			defer span.End()
//
//			span.SetAttributes(
//				attribute.String("http.method", r.Method),
//				attribute.String("http.url", r.URL.String()),
//				attribute.String("http.user_agent", r.UserAgent()),
//			)
//
//			devices, err := getDevices(ctx)
//			if err != nil {
//				span.RecordError(err)
//				span.SetStatus(codes.Error, "Failed to get devices")
//				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//				return
//			}
//
//			// Inject trace context into response headers for client correlation
//			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))
//
//			w.Header().Set("Content-Type", "application/json")
//			json.NewEncoder(w).Encode(devices)
//
//			span.SetAttributes(
//				attribute.Int("http.status_code", http.StatusOK),
//				attribute.Int("devices.count", len(devices)),
//			)
//		})
//
//		return http.ListenAndServe(":8080", mux)
//	}
//
// # Error Tracking and Debugging
//
// Using telemetry for error tracking across services:
//
//	func handleDeviceOperation(ctx context.Context, deviceID string, operation string) error {
//		tracer := otel.Tracer("bmc-operations")
//		ctx, span := tracer.Start(ctx, fmt.Sprintf("device_operation_%s", operation))
//		defer span.End()
//
//		span.SetAttributes(
//			attribute.String("device.id", deviceID),
//			attribute.String("operation.type", operation),
//		)
//
//		// Add breadcrumb for debugging
//		span.AddEvent("operation_started", trace.WithAttributes(
//			attribute.String("timestamp", time.Now().Format(time.RFC3339)),
//		))
//
//		// Simulate operation with potential errors
//		if err := executeDeviceOperation(ctx, deviceID, operation); err != nil {
//			// Record detailed error information
//			span.RecordError(err)
//			span.SetStatus(codes.Error, "Operation failed")
//			span.SetAttributes(
//				attribute.String("error.type", fmt.Sprintf("%T", err)),
//				attribute.String("error.message", err.Error()),
//			)
//
//			// Add context for debugging
//			span.AddEvent("operation_failed", trace.WithAttributes(
//				attribute.String("failure_reason", err.Error()),
//				attribute.String("recovery_action", "retry_with_backoff"),
//			))
//
//			return fmt.Errorf("device operation %s failed for %s: %w", operation, deviceID, err)
//		}
//
//		span.AddEvent("operation_completed")
//		span.SetStatus(codes.Ok, "Operation successful")
//		return nil
//	}
//
// # Configuration for Different Environments
//
// Setting up telemetry for different deployment environments:
//
//	func setupTelemetryForEnvironment(env string) error {
//		switch env {
//		case "development":
//			// Use default no-op setup for development
//			telemetry.DefaultSetup()
//			log.Println("Telemetry: Development mode (no-op providers)")
//
//		case "staging":
//			// Setup with console exporters for staging
//			if err := setupConsoleExporters(); err != nil {
//				return fmt.Errorf("failed to setup console exporters: %w", err)
//			}
//			log.Println("Telemetry: Staging mode (console exporters)")
//
//		case "production":
//			// Setup with OTLP exporters for production
//			if err := setupOTLPExporters(); err != nil {
//				return fmt.Errorf("failed to setup OTLP exporters: %w", err)
//			}
//			log.Println("Telemetry: Production mode (OTLP exporters)")
//
//		default:
//			// Default to no-op for unknown environments
//			telemetry.DefaultSetup()
//			log.Printf("Telemetry: Unknown environment '%s', using no-op providers", env)
//		}
//
//		return nil
//	}
//
// # Performance Monitoring
//
// Using telemetry for performance monitoring:
//
//	func monitorPerformance(ctx context.Context, operation func() error) error {
//		tracer := otel.Tracer("bmc-performance")
//		ctx, span := tracer.Start(ctx, "performance_monitor")
//		defer span.End()
//
//		startTime := time.Now()
//		startMemory := getMemoryUsage()
//
//		// Execute operation
//		err := operation()
//
//		duration := time.Since(startTime)
//		endMemory := getMemoryUsage()
//		memoryDelta := endMemory - startMemory
//
//		// Record performance metrics
//		span.SetAttributes(
//			attribute.Int64("performance.duration_ns", duration.Nanoseconds()),
//			attribute.Int64("performance.memory_delta_bytes", memoryDelta),
//			attribute.Float64("performance.duration_ms", float64(duration.Nanoseconds())/1e6),
//		)
//
//		// Add performance events
//		if duration > 100*time.Millisecond {
//			span.AddEvent("slow_operation", trace.WithAttributes(
//				attribute.String("threshold", "100ms"),
//				attribute.String("actual", duration.String()),
//			))
//		}
//
//		if memoryDelta > 10*1024*1024 { // 10MB
//			span.AddEvent("high_memory_usage", trace.WithAttributes(
//				attribute.String("threshold", "10MB"),
//				attribute.String("actual", fmt.Sprintf("%.2fMB", float64(memoryDelta)/1024/1024)),
//			))
//		}
//
//		if err != nil {
//			span.RecordError(err)
//			span.SetStatus(codes.Error, "Operation failed")
//		} else {
//			span.SetStatus(codes.Ok, "Operation successful")
//		}
//
//		return err
//	}
//
// # Best Practices
//
// When using this package:
//
//   - Initialize telemetry early in the application lifecycle
//   - Use meaningful span names that describe the operation
//   - Add relevant attributes to spans for filtering and grouping
//   - Propagate context through all service calls
//   - Record errors with appropriate status codes
//   - Use events for significant milestones or debugging information
//   - Configure appropriate sampling rates for production workloads
//   - Monitor telemetry overhead and adjust configuration as needed
//
// # Thread Safety
//
// All functions in this package are safe for concurrent use. The underlying
// OpenTelemetry SDK handles concurrent access to tracers, spans, and
// propagators appropriately.
//
// # Resource Usage
//
// The telemetry system has minimal overhead when using no-op providers:
//
//   - No-op tracers have negligible performance impact
//   - Context propagation uses efficient header manipulation
//   - Span creation and attribute setting are optimized for no-op case
//   - Memory usage is minimal with no active telemetry backends
//
// For production deployments with active telemetry:
//
//   - Configure appropriate batch sizes for exporters
//   - Use sampling to reduce telemetry volume
//   - Monitor resource usage and adjust configuration
//   - Consider using resource detection for automatic tagging
package telemetry
