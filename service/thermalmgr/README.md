# Thermal Manager Service

The thermal manager service (`thermalmgr`) provides comprehensive thermal management for BMC systems through PID-based temperature control and cooling device management. It integrates with other BMC services to maintain system temperatures within safe operating limits.

## Overview

The thermal management system consists of:

- **Thermal Package** (`pkg/thermal`) - Core thermal management functionality
- **Thermal Manager Service** (`service/thermalmgr`) - NATS-based thermal management service
- **Sensor Integration** - Temperature monitoring via `sensormon`
- **Power Integration** - Emergency shutdown coordination via `powermgr`

## Key Features

- **PID Control Loops**: Software-based temperature regulation using the `pid-go` library
- **Thermal Zones**: Logical groupings of sensors and cooling devices
- **Cooling Device Control**: Fan, pump, and thermal device management via hwmon
- **Emergency Response**: Critical temperature handling and shutdown coordination
- **NATS Integration**: Service-to-service communication and event handling
- **Hardware Discovery**: Automatic detection of cooling devices from hwmon

## Architecture

### Thermal Zones

Thermal zones represent logical groupings of temperature sensors and cooling devices that work together to maintain target temperatures:

```go
type ThermalZone struct {
    Name                string
    SensorPaths         []string
    CoolingDevices      []*CoolingDevice
    TargetTemperature   float64
    WarningTemperature  float64
    CriticalTemperature float64
    PIDConfig           PIDConfig
}
```

### Cooling Devices

Cooling devices represent controllable thermal management hardware:

```go
type CoolingDevice struct {
    Name         string
    Type         v1alpha1.CoolingDeviceType
    HwmonPath    string
    MinPower     float64
    MaxPower     float64
    CurrentPower float64
    Status       v1alpha1.CoolingDeviceStatus
}
```

### PID Control

Each thermal zone uses a PID controller for precise temperature regulation:

```go
type PIDConfig struct {
    Kp         float64       // Proportional gain
    Ki         float64       // Integral gain
    Kd         float64       // Derivative gain
    SampleTime time.Duration // Control loop interval
    OutputMin  float64       // Minimum output (0%)
    OutputMax  float64       // Maximum output (100%)
}
```

## Service Integration

### Sensormon Integration

The thermal manager integrates with `sensormon` for temperature monitoring:

- **Temperature Updates**: Regular temperature readings from sensors
- **Threshold Alerts**: Critical and warning temperature notifications
- **Emergency Alerts**: Immediate notifications for thermal emergencies
- **Sensor Configuration**: Dynamic sensor-to-zone assignment

### Powermgr Integration

Emergency thermal conditions trigger power management actions:

- **Emergency Shutdown**: Component shutdown when cooling is insufficient
- **Power Throttling**: Power reduction for thermal management
- **Immediate Shutdown**: Force shutdown for critical conditions

## NATS Endpoints

### Thermal Zone Management

```
thermalmgr.zones.list          # List all thermal zones
thermalmgr.zone.get            # Get thermal zone details
thermalmgr.zone.set            # Update thermal zone configuration
```

### Cooling Device Management

```
thermalmgr.devices.list        # List all cooling devices
thermalmgr.device.get          # Get cooling device details
thermalmgr.device.set          # Control cooling device power
```

### Thermal Control

```
thermalmgr.control.start       # Start thermal control loops
thermalmgr.control.stop        # Stop thermal control loops
thermalmgr.control.status      # Get thermal control status
thermalmgr.control.emergency   # Handle emergency conditions
```

## Configuration

### Basic Configuration

```go
config := thermalmgr.NewConfig(
    thermalmgr.WithServiceName("thermalmgr"),
    thermalmgr.WithThermalControlInterval(time.Second),
    thermalmgr.WithDefaultPIDConfig(1.0, 0.1, 0.05),
    thermalmgr.WithTemperatureThresholds(75.0, 85.0, 95.0),
    thermalmgr.WithHwmonPath("/sys/class/hwmon"),
    thermalmgr.WithDiscovery(true),
)
```

### PID Tuning Examples

**Aggressive Cooling Profile:**
```go
aggressiveConfig := thermal.PIDConfig{
    Kp:         3.0,  // High proportional gain
    Ki:         1.0,  // Moderate integral gain
    Kd:         0.5,  // Higher derivative gain
    SampleTime: 500 * time.Millisecond,
    OutputMin:  0.0,
    OutputMax:  100.0,
}
```

**Quiet Cooling Profile:**
```go
quietConfig := thermal.PIDConfig{
    Kp:         1.0,  // Lower proportional gain
    Ki:         0.2,  // Lower integral gain
    Kd:         0.05, // Lower derivative gain
    SampleTime: 2 * time.Second,
    OutputMin:  0.0,
    OutputMax:  60.0, // Limit maximum cooling
}
```

## Usage Examples

### Creating a Thermal Zone

```go
zone := &thermal.ThermalZone{
    Name:              "cpu_zone",
    SensorPaths:       []string{"/sys/class/hwmon/hwmon0/temp1_input"},
    TargetTemperature: 65.0,
    PIDConfig: thermal.PIDConfig{
        Kp:         2.0,
        Ki:         0.5,
        Kd:         0.1,
        SampleTime: time.Second,
        OutputMin:  0.0,
        OutputMax:  100.0,
    },
}

err := thermal.InitializeThermalZone(ctx, zone)
if err != nil {
    log.Printf("Failed to initialize thermal zone: %v", err)
}
```

### Manual Cooling Control

```go
fan := &thermal.CoolingDevice{
    Name:     "cpu_fan",
    Type:     v1alpha1.CoolingDeviceType_COOLING_DEVICE_TYPE_FAN,
    HwmonPath: "/sys/class/hwmon/hwmon1/pwm1",
    MinPower: 0,
    MaxPower: 255,
}

// Set fan to 75% speed
err := thermal.SetCoolingDevicePower(ctx, fan, 75.0)
if err != nil {
    log.Printf("Failed to set fan speed: %v", err)
}
```

### Temperature Monitoring

```go
temperature, err := thermal.ReadZoneTemperature(ctx, zone)
if err != nil {
    log.Printf("Failed to read temperature: %v", err)
    return
}

output, err := thermal.UpdatePIDControl(ctx, zone, temperature)
if err != nil {
    log.Printf("PID update failed: %v", err)
    return
}

err = thermal.SetCoolingOutput(ctx, zone, output)
if err != nil {
    log.Printf("Failed to set cooling output: %v", err)
}
```

## Emergency Response

### Critical Temperature Handling

The thermal manager implements a multi-stage emergency response:

1. **Warning Threshold**: Increase cooling, log warning
2. **Critical Threshold**: Maximum cooling, alert other services
3. **Emergency Threshold**: Request emergency shutdown via powermgr

### Emergency Actions

```go
// Emergency cooling
err := thermal.SetCoolingOutput(ctx, zone, 100.0)

// Reset PID controller
err := thermal.ResetPIDController(ctx, zone)

// Check for emergency conditions
err := thermal.CheckThermalEmergency(ctx, zone)
if err == thermal.ErrCriticalTemperature {
    // Handle critical condition
}
```

## Hardware Integration

### Hwmon Interface

The thermal manager interfaces with Linux hwmon for hardware control:

- **PWM Outputs**: Fan and pump speed control (0-255 range)
- **Temperature Inputs**: Sensor readings in millidegrees Celsius
- **Device Discovery**: Automatic detection of thermal hardware

### Supported Device Types

- **Fans**: PWM-controlled case and CPU fans
- **Water Pumps**: AIO and custom loop pumps
- **Heat Exchangers**: Active cooling systems
- **Liquid Coolers**: All-in-one cooling solutions

## Monitoring and Observability

### Logging

All thermal operations are logged with structured logging:

```go
slog.InfoContext(ctx, "Thermal control update",
    "zone", zone.Name,
    "temperature", temp,
    "target", zone.TargetTemperature,
    "output", output)
```

### Telemetry

OpenTelemetry integration provides distributed tracing:

- Thermal control loop spans
- PID calculation tracing
- Emergency response tracking
- Service integration spans

### Metrics

Key thermal metrics are exposed:

- Current temperatures by zone
- PID controller outputs
- Cooling device power levels
- Emergency event counts
- Control loop execution times

## Safety Features

### Failsafe Operation

- **Default Cooling**: Safe cooling levels on startup
- **Error Recovery**: Graceful handling of sensor/device failures
- **Emergency Coordination**: Integration with power management
- **Thermal Runaway Protection**: Detection and response

### Reliability

- **Redundant Sensors**: Multiple sensors per zone support
- **Device Fault Tolerance**: Continued operation with failed devices
- **State Persistence**: Optional thermal state persistence
- **Graceful Degradation**: Reduced functionality vs. complete failure

## Development

### Building

```bash
go build ./pkg/thermal
go build ./service/thermalmgr
```

### Testing

```bash
go test ./pkg/thermal
go test ./service/thermalmgr
```

### Dependencies

- `go.einride.tech/pid` - PID controller implementation
- `github.com/u-bmc/u-bmc/pkg/hwmon` - Hardware monitoring interface
- `github.com/nats-io/nats.go` - NATS messaging
- Standard Go libraries

## Troubleshooting

### Common Issues

**High Temperatures:**
- Check PID tuning parameters
- Verify cooling device operation
- Ensure adequate thermal capacity

**Control Oscillation:**
- Reduce proportional gain (Kp)
- Increase sample time
- Check for mechanical issues

**Device Not Found:**
- Verify hwmon path exists
- Check device permissions
- Enable hardware discovery

**Emergency Shutdowns:**
- Review temperature thresholds
- Check thermal zone configuration
- Verify emergency response settings

### Debug Commands

```bash
# List thermal zones
nats request thermalmgr.zones.list '{}'

# Get zone details
nats request thermalmgr.zone.get '{"name":"cpu_zone"}'

# Check control status
nats request thermalmgr.control.status '{}'

# List cooling devices
nats request thermalmgr.devices.list '{}'
```

## License

This software is licensed under the BSD-3-Clause license. See the LICENSE file for details.