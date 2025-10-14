// SPDX-License-Identifier: BSD-3-Clause

// Package sensormon provides a sensor monitoring service for BMC systems that reads and monitors
// hardware sensors using the hwmon subsystem, GPIO interfaces, and mock backends. This service
// focuses exclusively on sensor monitoring and threshold management, without thermal control
// functionality. It supports user-defined sensor configurations with custom callbacks and
// comprehensive mock backends for testing.
//
// # Overview
//
// The sensormon service provides NATS-based IPC endpoints for:
//   - Reading sensor values from hwmon devices, GPIO, and mock backends
//   - Monitoring sensor thresholds and status with custom callbacks
//   - Configuring sensor parameters through user-defined sensor definitions
//   - Managing discrete and analog sensor readings
//   - GPIO-based sensor operations
//   - Mock sensor operations for testing and development
//
// This service uses the protobuf-defined sensor types from the u-bmc API schema and provides
// a unified interface for sensor management across different hardware platforms with extensive
// customization capabilities.
//
// # Service Architecture
//
// The service follows the standard u-bmc service pattern with:
//   - NATS micro-service framework for IPC
//   - Context-aware operations with timeout support
//   - Structured logging with slog
//   - Configuration through functional options
//   - Graceful shutdown handling
//   - Observability handled by global telemetry service
//
// # Sensor Types Supported
//
// The service supports all sensor types defined in the protobuf schema:
//   - Temperature sensors (Celsius, Fahrenheit, Kelvin)
//   - Voltage sensors (Volts)
//   - Current sensors (Amps)
//   - Tachometer/Fan sensors (RPM)
//   - Power sensors (Watts)
//   - Energy sensors (Joules)
//   - Pressure sensors (Pascals)
//   - Humidity sensors (Percent)
//   - Altitude sensors (Meters)
//   - Flow rate sensors (Liters per minute)
//
// All sensor types support both hardware backends (hwmon, GPIO) and mock backends for testing.
//
// # IPC Endpoints
//
// The service exposes the following NATS endpoints:
//
// Sensor Listing:
//   - sensormon.sensors.list - List all available sensors
//
// Sensor Operations:
//   - sensormon.sensor.get - Get sensor information and current reading
//   - sensormon.sensor.read - Read current sensor value
//   - sensormon.sensor.configure - Configure sensor parameters
//
// Monitoring Operations:
//   - sensormon.monitoring.start - Start continuous monitoring
//   - sensormon.monitoring.stop - Stop monitoring
//   - sensormon.monitoring.status - Get monitoring status
//
// # Usage Examples
//
// Creating and starting the service with enhanced configuration:
//
//	// Define custom sensors
//	sensors := []sensormon.SensorDefinition{
//		sensormon.NewTemperatureSensor("cpu_temp", "CPU Temperature", sensormon.BackendTypeHwmon),
//		sensormon.NewFanSensor("fan1", "System Fan 1", sensormon.BackendTypeMock),
//	}
//
//	// Configure callbacks
//	callbacks := sensormon.SensorCallbacks{
//		OnThresholdCritical: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
//			// Handle critical temperature
//			return nil
//		},
//	}
//
//	service := sensormon.New(
//		sensormon.WithServiceName("sensormon"),
//		sensormon.WithMonitoringInterval(1*time.Second),
//		sensormon.WithHwmonPath("/sys/class/hwmon"),
//		sensormon.WithGPIOSensors(true),
//		sensormon.WithSensorDefinitions(sensors...),
//		sensormon.WithCallbacks(callbacks),
//	)
//
//	ctx := context.Background()
//	err := service.Run(ctx, ipcConn)
//
// Reading a sensor via NATS:
//
//	req := &GetSensorRequest{
//		Identifier: &GetSensorRequest_Name{
//			Name: "CPU Temperature",
//		},
//	}
//
//	response, err := nc.Request("sensormon.sensor.get", req, 5*time.Second)
//
// Listing all sensors:
//
//	req := &ListSensorsRequest{}
//	response, err := nc.Request("sensormon.sensors.list", req, 5*time.Second)
//
// # Configuration
//
// The service supports extensive configuration options including user-defined sensors:
//
//	// Define hwmon sensor with specific configuration
//	hwmonSensor := sensormon.SensorDefinition{
//		ID:      "motherboard_temp",
//		Name:    "Motherboard Temperature",
//		Context: v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE,
//		Unit:    v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS,
//		Backend: sensormon.BackendTypeHwmon,
//		UpperThresholds: &sensormon.Threshold{
//			Warning:  &[]float64{65.0}[0],
//			Critical: &[]float64{75.0}[0],
//		},
//		HwmonConfig: &sensormon.HwmonSensorConfig{
//			MatchPattern:  "nct6775",
//			AttributeName: "temp1_input",
//			ScaleFactor:   1000,
//		},
//		Enabled: true,
//	}
//
//	// Define mock sensor for testing
//	mockSensor := sensormon.SensorDefinition{
//		ID:      "test_fan",
//		Name:    "Test Fan",
//		Context: v1alpha1.SensorContext_SENSOR_CONTEXT_TACH,
//		Unit:    v1alpha1.SensorUnit_SENSOR_UNIT_RPM,
//		Backend: sensormon.BackendTypeMock,
//		MockConfig: &sensormon.MockSensorConfig{
//			Behavior:  sensormon.MockBehaviorSine,
//			BaseValue: 1200.0,
//			Variance:  200.0,
//			Period:    30 * time.Second,
//		},
//		Enabled: true,
//	}
//
//	config := sensormon.New(
//		sensormon.WithServiceName("sensormon"),
//		sensormon.WithServiceVersion("1.0.0"),
//		sensormon.WithHwmonPath("/sys/class/hwmon"),
//		sensormon.WithGPIOChipPath("/dev/gpiochip0"),
//		sensormon.WithMonitoringInterval(1*time.Second),
//		sensormon.WithThresholdCheckInterval(5*time.Second),
//		sensormon.WithHwmonSensors(true),
//		sensormon.WithGPIOSensors(true),
//		sensormon.WithMockSensors(true),
//		sensormon.WithSensorDefinitions(hwmonSensor, mockSensor),
//	)
//
// # Sensor Discovery
//
// The service discovers and manages sensors from multiple sources:
//
// User-Defined Sensors: Explicitly configured sensors with custom parameters:
//   - Detailed threshold configuration with warning and critical levels
//   - Custom location information and metadata
//   - Backend-specific configuration (hwmon patterns, GPIO lines, mock behavior)
//   - Event callbacks for custom handling
//
// Auto-Discovered Hwmon Sensors: Discovered from /sys/class/hwmon/hwmonN devices by scanning for:
//   - Temperature sensors (temp*_input, temp*_label)
//   - Voltage sensors (in*_input, in*_label)
//   - Fan sensors (fan*_input, pwm*)
//   - Power sensors (power*_input)
//   - Current sensors (curr*_input)
//
// GPIO Sensors: Configured sensors that use GPIO for discrete readings
//
// Mock Sensors: Simulated sensors for testing with configurable behaviors:
//   - Fixed values for consistent testing
//   - Randomized values with configurable variance
//   - Sine wave patterns for dynamic testing
//   - Step functions for threshold testing
//
// # Threshold Management
//
// The service monitors sensor thresholds and executes custom callbacks:
//
//	// Define thresholds in sensor configuration
//	sensor := sensormon.SensorDefinition{
//		ID: "cpu_temp",
//		UpperThresholds: &sensormon.Threshold{
//			Warning:  &[]float64{75.0}[0],  // 75°C warning
//			Critical: &[]float64{85.0}[0],  // 85°C critical
//		},
//		LowerThresholds: &sensormon.Threshold{
//			Warning:  &[]float64{500.0}[0], // 500 RPM warning for fans
//			Critical: &[]float64{200.0}[0], // 200 RPM critical for fans
//		},
//	}
//
//	// Configure callbacks for threshold events
//	callbacks := sensormon.SensorCallbacks{
//		OnThresholdCritical: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
//			// Custom handling for critical thresholds
//			log.Printf("Critical threshold exceeded for sensor %s", sensorID)
//			return triggerEmergencyCooling()
//		},
//		OnSensorError: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
//			// Custom handling for sensor errors
//			return handleSensorFailure(sensorID)
//		},
//	}
//
// # Error Handling
//
// The service provides structured error responses through NATS:
//
//	response, err := nc.Request("sensormon.sensor.get", req, timeout)
//	if err != nil {
//		// Handle NATS communication error
//		return err
//	}
//
//	var sensorResponse GetSensorResponse
//	err = proto.Unmarshal(response.Data, &sensorResponse)
//	if err != nil {
//		// Handle protobuf unmarshaling error
//		return err
//	}
//
//	if len(sensorResponse.Sensors) == 0 {
//		// Sensor not found
//		return errors.New("sensor not found")
//	}
//
// # Integration with Other Services
//
// The sensormon service is designed to integrate with other u-bmc services:
//
// State Management: Sensor readings can trigger state transitions through callbacks:
//
//	callbacks := sensormon.SensorCallbacks{
//		OnThresholdCritical: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
//			// Trigger thermal state change
//			return statemgr.Request("statemgr.thermal.critical", thermalEvent)
//		},
//	}
//
// Power Management: Power sensors inform power control decisions through callbacks:
//
//	callbacks := sensormon.SensorCallbacks{
//		OnSensorRead: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
//			if sensorID == "system_power" && data.(float64) > powerLimit {
//				return powermgr.Request("powermgr.host.power.limit", powerLimitEvent)
//			}
//			return nil
//		},
//	}
//
// Testing Integration: Mock sensors enable comprehensive testing:
//
//	// Configure mock sensors that simulate real hardware behavior
//	mockConfig := sensormon.NewMockTemperatureSensor(45.0) // Base temp 45°C
//	mockConfig.Behavior = sensormon.MockBehaviorSine
//	mockConfig.Period = 60 * time.Second // 1-minute temperature cycle
//
// # Performance Considerations
//
// The service is optimized for efficient sensor monitoring:
//   - Configurable monitoring intervals to balance accuracy and performance
//   - Concurrent sensor reading using goroutines with configurable limits
//   - Efficient hwmon file system access with pattern matching and discovery
//   - Backend-specific optimizations (hwmon caching, GPIO debouncing, mock efficiency)
//   - Context-aware operations for timeout handling
//   - Event-driven callbacks to minimize overhead
//   - Mock backends with zero system resource usage for testing
//
// # Security
//
// The service requires appropriate permissions:
//   - Read access to /sys/class/hwmon/* for hwmon sensor monitoring
//   - Write access to hwmon files for sensor configuration (if supported)
//   - GPIO access permissions for GPIO-based sensors (/dev/gpiochip*)
//   - NATS connection permissions for IPC communication
//   - No special permissions required for mock backends (testing only)
//
// # Monitoring and Observability
//
// The service provides comprehensive monitoring capabilities:
//   - OpenTelemetry tracing for request tracking
//   - Structured logging with contextual information
//   - Metrics collection for sensor read operations
//   - Health status reporting through NATS endpoints
//
// For production deployments, consider:
//   - Monitoring sensor read latencies and callback execution times
//   - Tracking threshold violations and callback success rates
//   - Alerting on sensor communication failures and callback errors
//   - Performance metrics for large sensor deployments
//   - Using mock backends for integration testing and CI/CD pipelines
//   - Implementing comprehensive sensor definitions with proper thresholds
//   - Setting up event callbacks for automated responses to sensor events
package sensormon
