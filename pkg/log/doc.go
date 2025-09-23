// SPDX-License-Identifier: BSD-3-Clause

// Package log provides structured logging functionality with multi-target output
// support for console and OpenTelemetry observability. The package integrates
// multiple logging libraries to provide a unified interface that outputs
// human-readable logs to the console while simultaneously sending structured
// telemetry data to OpenTelemetry for distributed tracing and monitoring.
//
// The package is built around Go's standard library slog package and provides
// adapters for various logging systems including NATS server logging, oversight
// process management logging, and QUIC connection logging. This allows for
// consistent structured logging across all components of the system.
//
// # Core Features
//
// The package provides several key features:
//
//   - Dual output: Human-readable console logs and structured OpenTelemetry data
//   - Standard library slog integration for structured logging
//   - NATS server logger adapter for consistent logging from NATS components
//   - Oversight process supervisor logger integration
//   - QUIC connection logger for network protocol debugging
//   - Automatic timestamp and debug level configuration
//
// # Basic Usage
//
// Creating and using the default logger:
//
//	logger := log.NewDefaultLogger()
//	logger.Info("BMC service starting", "version", "1.0.0", "config", "/etc/bmc/config.yaml")
//	logger.Debug("Debug information", "module", "auth", "user_count", 5)
//	logger.Error("Operation failed", "error", err, "operation", "device_scan")
//
// Using the global logger:
//
//	log.RedirectSlogger() // Redirect standard slog to use our logger
//	slog.Info("This will now use the configured logger with dual output")
//
// # Structured Logging
//
// The logger supports structured logging with key-value pairs:
//
//	func handleDeviceConnection(deviceID string, addr net.Addr) {
//		logger := log.GetGlobalLogger()
//
//		logger.Info("Device connected",
//			"device_id", deviceID,
//			"remote_addr", addr.String(),
//			"protocol", "tcp",
//			"timestamp", time.Now(),
//		)
//
//		// Perform device operations...
//
//		logger.Debug("Device authentication completed",
//			"device_id", deviceID,
//			"auth_method", "certificate",
//			"duration_ms", 150,
//		)
//	}
//
// # Error Logging with Context
//
// Enhanced error logging with contextual information:
//
//	func processDeviceCommand(deviceID string, cmd Command) error {
//		logger := log.GetGlobalLogger()
//
//		logger.Info("Processing device command",
//			"device_id", deviceID,
//			"command", cmd.Type,
//			"request_id", cmd.RequestID,
//		)
//
//		if err := executeCommand(cmd); err != nil {
//			logger.Error("Command execution failed",
//				"device_id", deviceID,
//				"command", cmd.Type,
//				"request_id", cmd.RequestID,
//				"error", err,
//				"retry_count", cmd.RetryCount,
//			)
//			return fmt.Errorf("failed to execute command %s: %w", cmd.Type, err)
//		}
//
//		logger.Info("Command completed successfully",
//			"device_id", deviceID,
//			"command", cmd.Type,
//			"request_id", cmd.RequestID,
//			"duration_ms", cmd.Duration.Milliseconds(),
//		)
//
//		return nil
//	}
//
// # NATS Server Integration
//
// Using the NATS logger adapter for consistent logging from NATS server:
//
//	func setupNATSServer() (*nats.Server, error) {
//		logger := log.GetGlobalLogger()
//		natsLogger := log.NewNATSLogger(logger)
//
//		opts := &nats.Options{
//			Host:   "127.0.0.1",
//			Port:   4222,
//			Logger: natsLogger,
//		}
//
//		server, err := nats.NewServer(opts)
//		if err != nil {
//			return nil, fmt.Errorf("failed to create NATS server: %w", err)
//		}
//
//		// NATS server logs will now be formatted consistently
//		// and sent to both console and OpenTelemetry
//		go server.Start()
//
//		return server, nil
//	}
//
// # Service Logging Pattern
//
// Recommended pattern for service initialization and lifecycle logging:
//
//	func (s *BMCService) Start(ctx context.Context) error {
//		logger := log.GetGlobalLogger()
//
//		logger.Info("BMC service starting",
//			"service", s.Name(),
//			"version", s.Version(),
//			"config_path", s.ConfigPath(),
//			"pid", os.Getpid(),
//		)
//
//		// Initialize components
//		if err := s.initializeHardware(); err != nil {
//			logger.Error("Hardware initialization failed",
//				"service", s.Name(),
//				"error", err,
//				"component", "hardware",
//			)
//			return fmt.Errorf("hardware init failed: %w", err)
//		}
//
//		logger.Info("Hardware initialized successfully",
//			"service", s.Name(),
//			"device_count", len(s.devices),
//		)
//
//		// Start service loop
//		logger.Info("BMC service ready",
//			"service", s.Name(),
//			"listen_addr", s.ListenAddr(),
//			"startup_duration_ms", time.Since(s.startTime).Milliseconds(),
//		)
//
//		return s.serve(ctx)
//	}
//
// # Request/Response Logging
//
// Logging HTTP requests and responses with correlation:
//
//	func logHTTPRequest(r *http.Request) {
//		logger := log.GetGlobalLogger()
//		requestID := r.Header.Get("X-Request-ID")
//
//		logger.Info("HTTP request received",
//			"method", r.Method,
//			"path", r.URL.Path,
//			"remote_addr", r.RemoteAddr,
//			"user_agent", r.UserAgent(),
//			"request_id", requestID,
//			"content_length", r.ContentLength,
//		)
//	}
//
//	func logHTTPResponse(status int, duration time.Duration, requestID string) {
//		logger := log.GetGlobalLogger()
//
//		level := slog.LevelInfo
//		if status >= 400 {
//			level = slog.LevelWarn
//		}
//		if status >= 500 {
//			level = slog.LevelError
//		}
//
//		logger.Log(context.Background(), level, "HTTP response sent",
//			"status", status,
//			"duration_ms", duration.Milliseconds(),
//			"request_id", requestID,
//		)
//	}
//
// # Performance and Metrics Logging
//
// Logging performance metrics and system health:
//
//	func logSystemMetrics() {
//		logger := log.GetGlobalLogger()
//
//		var m runtime.MemStats
//		runtime.ReadMemStats(&m)
//
//		logger.Debug("System metrics",
//			"goroutines", runtime.NumGoroutine(),
//			"memory_alloc_mb", m.Alloc/1024/1024,
//			"memory_sys_mb", m.Sys/1024/1024,
//			"gc_cycles", m.NumGC,
//			"cpu_cores", runtime.NumCPU(),
//		)
//	}
//
//	func logDeviceMetrics(deviceID string, metrics DeviceMetrics) {
//		logger := log.GetGlobalLogger()
//
//		logger.Info("Device metrics",
//			"device_id", deviceID,
//			"cpu_usage_percent", metrics.CPUUsage,
//			"memory_usage_percent", metrics.MemoryUsage,
//			"temperature_celsius", metrics.Temperature,
//			"fan_speed_rpm", metrics.FanSpeed,
//			"power_consumption_watts", metrics.PowerConsumption,
//		)
//	}
//
// # Error Recovery Logging
//
// Logging error recovery and fallback scenarios:
//
//	func (s *BMCService) handlePanic() {
//		if r := recover(); r != nil {
//			logger := log.GetGlobalLogger()
//
//			logger.Error("Service panic recovered",
//				"service", s.Name(),
//				"panic", r,
//				"stack", string(debug.Stack()),
//				"recovery_action", "restart",
//			)
//
//			// Attempt recovery
//			if err := s.restart(); err != nil {
//				logger.Error("Service restart failed after panic",
//					"service", s.Name(),
//					"restart_error", err,
//					"action", "manual_intervention_required",
//				)
//			} else {
//				logger.Info("Service successfully restarted after panic",
//					"service", s.Name(),
//				)
//			}
//		}
//	}
//
// # Integration with OpenTelemetry
//
// The package automatically integrates with OpenTelemetry for distributed tracing:
//
//	func processWithTracing(ctx context.Context, operation string) error {
//		logger := log.GetGlobalLogger()
//
//		// Extract trace information from context if available
//		span := trace.SpanFromContext(ctx)
//		traceID := span.SpanContext().TraceID().String()
//		spanID := span.SpanContext().SpanID().String()
//
//		logger.Info("Operation started",
//			"operation", operation,
//			"trace_id", traceID,
//			"span_id", spanID,
//		)
//
//		// The logger will automatically include trace context
//		// in OpenTelemetry output for correlation
//
//		return nil
//	}
//
// # Configuration and Best Practices
//
// Recommended initialization pattern for services:
//
//	func main() {
//		// Initialize telemetry first
//		telemetry.DefaultSetup()
//
//		// Set up global logging
//		log.RedirectSlogger()
//		logger := log.GetGlobalLogger()
//
//		logger.Info("Application starting",
//			"name", "u-bmc",
//			"version", version.BuildVersion,
//			"commit", version.BuildCommit,
//			"build_time", version.BuildTime,
//		)
//
//		// Continue with application setup...
//	}
//
// # Thread Safety
//
// All logger instances are safe for concurrent use from multiple goroutines.
// The underlying slog and zerolog implementations handle concurrent access
// appropriately.
//
// # Performance Considerations
//
// The dual-output design has minimal performance impact:
//
//   - Console output uses zerolog's efficient JSON formatting
//   - OpenTelemetry output is asynchronous and batched
//   - Debug level logs are only processed when debug logging is enabled
//   - Structured logging with key-value pairs is more efficient than string formatting
//
// For high-throughput scenarios, consider:
//
//   - Using appropriate log levels (avoid excessive debug logging in production)
//   - Batching related log entries when possible
//   - Using sampling for high-frequency events
package log
