// SPDX-License-Identifier: BSD-3-Clause

// Package sensormon provides a sensor monitoring service for BMC systems that reads and monitors
// hardware sensors using the hwmon subsystem and GPIO interfaces. This service focuses exclusively
// on sensor monitoring and threshold management, without thermal control functionality.
//
// # Overview
//
// The sensormon service provides NATS-based IPC endpoints for:
//   - Reading sensor values from hwmon devices
//   - Monitoring sensor thresholds and status
//   - Configuring sensor parameters
//   - Managing discrete and analog sensor readings
//   - GPIO-based sensor operations
//
// This service uses the protobuf-defined sensor types from the u-bmc API schema and provides
// a unified interface for sensor management across different hardware platforms.
//
// # Service Architecture
//
// The service follows the standard u-bmc service pattern with:
//   - NATS micro-service framework for IPC
//   - Context-aware operations with timeout support
//   - Structured logging with slog
//   - Configuration through functional options
//   - Graceful shutdown handling
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
// Creating and starting the service:
//
//	service := sensormon.New(
//		sensormon.WithServiceName("sensormon"),
//		sensormon.WithMonitoringInterval(1*time.Second),
//		sensormon.WithHwmonPath("/sys/class/hwmon"),
//		sensormon.WithEnableGPIOSensors(true),
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
// The service supports extensive configuration options:
//
//	config := sensormon.New(
//		sensormon.WithServiceName("sensormon"),
//		sensormon.WithServiceVersion("1.0.0"),
//		sensormon.WithHwmonPath("/sys/class/hwmon"),
//		sensormon.WithGPIOChipPath("/dev/gpiochip0"),
//		sensormon.WithMonitoringInterval(1*time.Second),
//		sensormon.WithThresholdCheckInterval(5*time.Second),
//		sensormon.WithEnableHwmonSensors(true),
//		sensormon.WithEnableGPIOSensors(true),
//		sensormon.WithEnableMetrics(true),
//		sensormon.WithEnableTracing(true),
//	)
//
// # Sensor Discovery
//
// The service automatically discovers sensors from multiple sources:
//
// Hwmon Sensors: Discovered from /sys/class/hwmon/hwmonN devices by scanning for:
//   - Temperature sensors (temp*_input, temp*_label)
//   - Voltage sensors (in*_input, in*_label)
//   - Fan sensors (fan*_input, pwm*)
//   - Power sensors (power*_input)
//   - Current sensors (curr*_input)
//
// GPIO Sensors: Configured sensors that use GPIO for discrete readings
//
// # Threshold Management
//
// The service monitors sensor thresholds and reports status changes:
//
//	// Analog sensors support warning and critical thresholds
//	upperThresholds := &Threshold{
//		Warning:  &wrapperspb.DoubleValue{Value: 75.0},  // 75°C warning
//		Critical: &wrapperspb.DoubleValue{Value: 85.0},  // 85°C critical
//	}
//
//	// Discrete sensors support state-based monitoring
//	discreteReading := &DiscreteSensorReading{
//		State:            SensorStatus_SENSOR_STATUS_WARNING,
//		StateDescription: "Fan speed below optimal range",
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
// State Management: Sensor readings can trigger state transitions:
//
//	// Temperature sensor triggers thermal state changes
//	if temp > criticalThreshold {
//		statemgr.Request("statemgr.thermal.critical", thermalEvent)
//	}
//
// Power Management: Power sensors inform power control decisions:
//
//	// Power consumption monitoring
//	if powerReading > powerLimit {
//		powermgr.Request("powermgr.host.power.limit", powerLimitEvent)
//	}
//
// # Performance Considerations
//
// The service is optimized for efficient sensor monitoring:
//   - Configurable monitoring intervals to balance accuracy and performance
//   - Concurrent sensor reading using goroutines
//   - Efficient hwmon file system access
//   - Optional caching of sensor metadata
//   - Context-aware operations for timeout handling
//
// # Security
//
// The service requires appropriate permissions:
//   - Read access to /sys/class/hwmon/* for sensor monitoring
//   - Write access to hwmon files for sensor configuration (if supported)
//   - GPIO access permissions for GPIO-based sensors
//   - NATS connection permissions for IPC communication
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
//   - Monitoring sensor read latencies
//   - Tracking threshold violations
//   - Alerting on sensor communication failures
//   - Performance metrics for large sensor deployments
package sensormon
