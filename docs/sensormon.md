# SensorMon Service

The `sensormon` service provides comprehensive sensor monitoring capabilities for BMC systems, utilizing both hwmon (hardware monitoring) subsystem and GPIO interfaces for sensor data collection and threshold management.

## Overview

This service has been completely refactored to:

- **Focus solely on sensor monitoring** (no thermal control - that's handled by separate services)
- **Use simplified hwmon package** that provides direct sysfs access without interpretation layers
- **Follow established u-bmc service patterns** with NATS IPC, structured logging, and configuration
- **Support protobuf-defined sensor types** from the u-bmc API schema
- **Provide threshold monitoring and violation detection**

## Architecture

### Service Structure

```
service/sensormon/
├── doc.go          # Package documentation
├── errors.go       # Error definitions
├── config.go       # Configuration options
├── sensormon.go    # Main service implementation
├── sensors.go      # Sensor list/get operations
└── monitoring.go   # Monitoring start/stop/status
```

### Updated hwmon Package

The `pkg/hwmon` package has been simplified to provide stateless, functional access to hwmon sysfs files:

```
pkg/hwmon/
├── doc.go     # Package documentation
├── errors.go  # Error definitions
└── hwmon.go   # Core hwmon functions
```

## Key Features

### Sensor Discovery

- **Automatic hwmon device discovery** from `/sys/class/hwmon/`
- **Support for multiple sensor types**: temperature, voltage, current, fan/tach, power
- **Label-based sensor identification** with fallback to generic names
- **GPIO sensor framework** (extensible for discrete sensors)

### Monitoring Capabilities

- **Configurable monitoring intervals** for continuous sensor reading
- **Concurrent sensor reads** with configurable limits
- **Threshold monitoring** with warning and critical levels
- **Status tracking** and violation detection
- **NATS broadcasting** of sensor readings and threshold events

### NATS IPC Endpoints

```
sensormon.sensors.list        # List all available sensors
sensormon.sensor.get          # Get specific sensor information
sensormon.monitoring.start    # Start continuous monitoring
sensormon.monitoring.stop     # Stop monitoring
sensormon.monitoring.status   # Get monitoring status
```

## Configuration

```go
service := sensormon.New(
    sensormon.WithServiceName("sensormon"),
    sensormon.WithHwmonPath("/sys/class/hwmon"),
    sensormon.WithMonitoringInterval(1*time.Second),
    sensormon.WithThresholdCheckInterval(5*time.Second),
    sensormon.WithEnableHwmonSensors(true),
    sensormon.WithEnableGPIOSensors(false),
    sensormon.WithEnableThresholdMonitoring(true),
    sensormon.WithMaxConcurrentReads(10),
)
```

## Hwmon Package Usage

### Basic Operations

```go
// Read temperature sensor
temp, err := hwmon.ReadInt("/sys/class/hwmon/hwmon0/temp1_input")
if err != nil {
    return err
}
tempC := float64(temp) / 1000.0 // Convert millidegrees to Celsius

// Write fan PWM
err = hwmon.WriteInt("/sys/class/hwmon/hwmon1/pwm1", 127) // 50% duty cycle

// Read sensor label
label, err := hwmon.ReadString("/sys/class/hwmon/hwmon0/temp1_label")
```

### Device Discovery

```go
// Find device by name
devicePath, err := hwmon.FindDeviceByName("k10temp")
if err != nil {
    return err
}

// List all devices
devices, err := hwmon.ListDevices()
if err != nil {
    return err
}

// List temperature sensors in device
tempAttrs, err := hwmon.ListAttributes(devicePath, `temp\d+_input`)
```

### Context Support

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

value, err := hwmon.ReadIntCtx(ctx, sensorPath)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Read operation timed out")
    }
    return err
}
```

## Sensor Types Supported

Based on protobuf schema definitions:

- **Temperature**: Celsius, Fahrenheit, Kelvin
- **Voltage**: Volts (from millivolts)
- **Current**: Amps (from milliamps)
- **Tach/Fan**: RPM
- **Power**: Watts (from microwatts)
- **Energy**: Joules
- **Pressure**: Pascals
- **Humidity**: Percent
- **Altitude**: Meters
- **Flow Rate**: Liters per minute

## Threshold Management

### Analog Sensors

```go
// Automatic threshold discovery from hwmon
// - temp1_max -> warning threshold
// - temp1_crit -> critical threshold
// - temp1_min -> lower warning threshold

// Threshold violation detection with configurable intervals
// Status updates: ENABLED -> WARNING -> CRITICAL -> ERROR
```

### Threshold Events

Threshold violations generate NATS events:

```
sensormon.events.threshold.warning   # Warning level violations
sensormon.events.threshold.critical  # Critical level violations
```

## Integration Examples

### Reading Sensors via NATS

```go
// List all sensors
req := &v1alpha1.ListSensorsRequest{}
response, err := nc.Request("sensormon.sensors.list", marshalledReq, 5*time.Second)

// Get specific sensor
req := &v1alpha1.GetSensorRequest{
    Identifier: &v1alpha1.GetSensorRequest_Name{
        Name: "CPU Temperature",
    },
}
response, err := nc.Request("sensormon.sensor.get", marshalledReq, 5*time.Second)
```

### Starting Monitoring

```go
// Start continuous monitoring
_, err := nc.Request("sensormon.monitoring.start", []byte("{}"), 5*time.Second)

// Check status
statusResp, err := nc.Request("sensormon.monitoring.status", []byte("{}"), 5*time.Second)
```

## Error Handling

### Hwmon Package Errors

```go
value, err := hwmon.ReadInt(path)
if err != nil {
    switch {
    case errors.Is(err, hwmon.ErrFileNotFound):
        log.Printf("Sensor attribute not available")
    case errors.Is(err, hwmon.ErrPermissionDenied):
        log.Printf("Insufficient permissions")
    case errors.Is(err, hwmon.ErrInvalidValue):
        log.Printf("Value could not be parsed")
    case errors.Is(err, hwmon.ErrOperationTimeout):
        log.Printf("Operation timed out")
    default:
        log.Printf("Unexpected error: %v", err)
    }
}
```

### Service Errors

All service errors are defined in `errors.go` and can be used for test matching and error wrapping.

## Performance Considerations

- **Stateless operations**: All hwmon functions are stateless and thread-safe
- **Concurrent reads**: Configurable semaphore limits concurrent sensor operations
- **Context support**: All operations support cancellation and timeouts
- **No caching**: Direct sysfs access without interpretation (caching can be added at application level)
- **Efficient discovery**: Pattern-based sensor discovery with early termination

## Security Requirements

- **Read access** to `/sys/class/hwmon/*` for sensor monitoring
- **Write access** to hwmon files for sensor configuration (if needed)
- **GPIO permissions** for GPIO-based sensors (if enabled)
- **NATS connection** permissions for IPC communication

## Differences from Previous Implementation

### Hwmon Package Changes

- **Removed complex abstractions**: No more `Sensor`, `SensorManager`, `Discoverer` types
- **Stateless functions**: All operations are function-based, not method-based
- **Direct sysfs access**: No interpretation or caching layers
- **Simplified error handling**: Fewer, more focused error types
- **Context everywhere**: All operations support context for cancellation/timeout

### Service Changes

- **NATS micro-service pattern**: Follows same pattern as `statemgr` and `powermgr`
- **Protobuf integration**: Uses generated protobuf types from schema
- **Structured monitoring**: Separate handlers for different operation types
- **Threshold automation**: Built-in threshold monitoring and violation detection
- **Focused scope**: Only sensor monitoring, no thermal control

## Future Extensions

- **GPIO sensor implementations**: Framework exists for discrete GPIO sensors
- **Sensor calibration**: Support for sensor value calibration and correction
- **Historical data**: Optional JetStream integration for sensor data persistence
- **Custom thresholds**: Runtime threshold configuration via NATS
- **Sensor groups**: Logical grouping of related sensors
- **Health monitoring**: Service health and sensor availability tracking

## Dependencies

- **NATS**: For IPC communication
- **Protobuf**: For message serialization (sensor schema)
- **slog**: For structured logging
- **OpenTelemetry**: For tracing (optional)
- **JetStream**: For data persistence (optional)

The service is designed to be lightweight, efficient, and easily extensible while maintaining compatibility with the broader u-bmc ecosystem.
