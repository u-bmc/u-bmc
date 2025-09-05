// SPDX-License-Identifier: BSD-3-Clause

// Package hwmon provides a comprehensive interface for interacting with the Linux hwmon (hardware monitoring)
// subsystem through sysfs. This package enables BMC (Baseboard Management Controller) systems and other
// applications to read sensor values, control fans, and manage hardware monitoring devices in a safe and
// efficient manner.
//
// # Overview
//
// The Linux hwmon subsystem exposes hardware monitoring devices through the sysfs filesystem under
// /sys/class/hwmon/. Each hwmon device typically represents a monitoring chip (temperature sensors,
// voltage sensors, fan controllers, etc.) and provides various attributes for reading sensor values
// and controlling device behavior.
//
// This package provides the following key features:
//   - Automatic discovery of hwmon devices by name and label
//   - Type-safe reading and writing of sensor values
//   - Context-aware operations with timeout support
//   - Configurable retry mechanisms and error handling
//   - Support for various sensor types (temperature, voltage, fan, power, etc.)
//   - Thread-safe operations for concurrent access
//   - Structured error handling with specific error types
//   - Comprehensive validation of sensor paths and values
//
// # Core Concepts
//
// Hwmon Device: A hardware monitoring device exposed by the kernel through sysfs. Each device
// has a unique hwmon number (e.g., hwmon0, hwmon1) and typically contains multiple sensors.
//
// Sensor: An individual monitoring point within a hwmon device. Sensors have types (temp, fan,
// voltage, etc.) and provide various attributes like input values, labels, limits, and alarms.
//
// Attribute: A specific property of a sensor exposed as a file in sysfs. Common attributes include:
//   - _input: Current sensor reading
//   - _label: Human-readable sensor name
//   - _min/_max: Minimum/maximum thresholds
//   - _crit: Critical threshold
//   - _alarm: Alarm status
//   - _enable: Enable/disable sensor
//
// Sensor Path: The complete sysfs path to a sensor attribute, constructed from the hwmon device
// path and the specific attribute file.
//
// # Basic Usage
//
// Reading a temperature sensor:
//
//	config := NewConfig(
//		WithDevice("k10temp"),
//		WithSensorLabel("Tctl"),
//		WithSensorType(SensorTypeTemperature),
//		WithTimeout(5*time.Second),
//	)
//
//	sensor, err := NewSensor(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	value, err := sensor.ReadValue(ctx)
//	if err != nil {
//		log.Printf("Failed to read sensor: %v", err)
//		return
//	}
//
//	temp := value.AsTemperature()
//	fmt.Printf("CPU temperature: %.1f°C\n", temp.Celsius())
//
// Discovering all temperature sensors:
//
//	discoverer := NewDiscoverer()
//	sensors, err := discoverer.DiscoverSensors(ctx, SensorTypeTemperature)
//	if err != nil {
//		log.Printf("Discovery failed: %v", err)
//		return
//	}
//
//	for _, sensor := range sensors {
//		value, err := sensor.ReadValue(ctx)
//		if err != nil {
//			continue
//		}
//		fmt.Printf("%s: %.1f°C\n", sensor.Label(), value.AsTemperature().Celsius())
//	}
//
// Controlling a fan:
//
//	config := NewConfig(
//		WithDevice("nct6775"),
//		WithSensorLabel("fan1"),
//		WithSensorType(SensorTypeFan),
//		WithWritable(true),
//	)
//
//	fan, err := NewSensor(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Set fan to 50% speed (assuming PWM control)
//	if err := fan.WriteValue(ctx, NewFanValue(127)); err != nil {
//		log.Printf("Failed to set fan speed: %v", err)
//	}
//
// # Device Discovery
//
// The package provides automatic discovery of hwmon devices and sensors:
//
//	discoverer := NewDiscoverer(
//		WithDiscoveryPath("/sys/class/hwmon"),
//		WithTimeout(10*time.Second),
//	)
//
//	// Find device by name
//	device, err := discoverer.FindDevice(ctx, "k10temp")
//	if err != nil {
//		log.Printf("Device not found: %v", err)
//		return
//	}
//
//	// Get all sensors for the device
//	sensors, err := device.GetSensors(ctx)
//	if err != nil {
//		log.Printf("Failed to get sensors: %v", err)
//		return
//	}
//
// # Sensor Types and Values
//
// The package supports various sensor types with type-safe value handling:
//
//	// Temperature sensors (millidegree Celsius)
//	tempValue := value.AsTemperature()
//	fmt.Printf("Temperature: %.1f°C\n", tempValue.Celsius())
//
//	// Voltage sensors (millivolts)
//	voltageValue := value.AsVoltage()
//	fmt.Printf("Voltage: %.3fV\n", voltageValue.Volts())
//
//	// Fan sensors (RPM)
//	fanValue := value.AsFan()
//	fmt.Printf("Fan speed: %d RPM\n", fanValue.RPM())
//
//	// Power sensors (microwatts)
//	powerValue := value.AsPower()
//	fmt.Printf("Power: %.2fW\n", powerValue.Watts())
//
// # Configuration and Options
//
// Comprehensive configuration support for various use cases:
//
//	config := NewConfig(
//		WithDevice("acpi-0"),
//		WithSensorIndex(1),  // Use temp1_input instead of label
//		WithSensorType(SensorTypeTemperature),
//		WithTimeout(5*time.Second),
//		WithRetryCount(3),
//		WithRetryDelay(100*time.Millisecond),
//		WithValidationEnabled(true),
//		WithCaching(true, 1*time.Second),
//	)
//
// # Error Handling
//
// The package provides structured error handling with specific error types:
//
//	sensor, err := NewSensor(config)
//	if err != nil {
//		switch {
//		case errors.Is(err, ErrDeviceNotFound):
//			log.Printf("Hwmon device not found")
//		case errors.Is(err, ErrSensorNotFound):
//			log.Printf("Sensor not found on device")
//		case errors.Is(err, ErrPermissionDenied):
//			log.Printf("Permission denied accessing sensor")
//		default:
//			log.Printf("Unexpected error: %v", err)
//		}
//		return
//	}
//
//	value, err := sensor.ReadValue(ctx)
//	if err != nil {
//		switch {
//		case errors.Is(err, ErrReadTimeout):
//			log.Printf("Read operation timed out")
//		case errors.Is(err, ErrInvalidValue):
//			log.Printf("Invalid sensor value read")
//		case errors.Is(err, ErrDeviceUnavailable):
//			log.Printf("Device became unavailable")
//		default:
//			log.Printf("Read error: %v", err)
//		}
//		return
//	}
//
// # Thread Safety
//
// All operations in this package are thread-safe. Multiple goroutines can safely:
//   - Read from the same or different sensors simultaneously
//   - Perform device discovery operations concurrently
//   - Access sensor metadata and configuration
//   - Write to different sensors (writes to the same sensor should be coordinated by the caller)
//
// The implementation uses appropriate synchronization mechanisms to ensure data consistency
// while allowing concurrent access where safe.
//
// # BMC Integration
//
// This package is designed specifically for BMC systems where hardware monitoring is critical:
//
// Temperature Monitoring: Monitor CPU, GPU, motherboard, and ambient temperatures for thermal
// management and protection.
//
//	tempSensors, _ := discoverer.DiscoverSensors(ctx, SensorTypeTemperature)
//	for _, sensor := range tempSensors {
//		value, _ := sensor.ReadValue(ctx)
//		temp := value.AsTemperature()
//		if temp.Celsius() > 85.0 {
//			// Trigger thermal protection
//		}
//	}
//
// Fan Control: Monitor and control system fans for cooling optimization.
//
//	fanSensors, _ := discoverer.DiscoverSensors(ctx, SensorTypeFan)
//	for _, sensor := range fanSensors {
//		if sensor.IsWritable() {
//			// Implement fan curve based on temperature
//			fanSpeed := calculateFanSpeed(currentTemp)
//			sensor.WriteValue(ctx, NewFanValue(fanSpeed))
//		}
//	}
//
// Power Monitoring: Track power consumption for efficiency and protection.
//
//	powerSensors, _ := discoverer.DiscoverSensors(ctx, SensorTypePower)
//	totalPower := 0.0
//	for _, sensor := range powerSensors {
//		value, _ := sensor.ReadValue(ctx)
//		totalPower += value.AsPower().Watts()
//	}
//
// Voltage Monitoring: Monitor power rail voltages for stability.
//
//	voltageSensors, _ := discoverer.DiscoverSensors(ctx, SensorTypeVoltage)
//	for _, sensor := range voltageSensors {
//		value, _ := sensor.ReadValue(ctx)
//		voltage := value.AsVoltage()
//		if voltage.Volts() < 11.4 || voltage.Volts() > 12.6 {
//			// Voltage out of acceptable range
//		}
//	}
//
// # Performance Considerations
//
// The package includes several performance optimizations:
//   - Optional caching of sensor values to reduce sysfs access
//   - Efficient device discovery with early termination
//   - Minimal memory allocations during normal operations
//   - Context-aware operations that can be canceled
//   - Configurable retry mechanisms to handle transient errors
//
// For high-frequency monitoring, consider enabling caching:
//
//	config := NewConfig(
//		WithCaching(true, 100*time.Millisecond),
//		// other options...
//	)
//
// # Integration with Other Packages
//
// This package is designed to integrate seamlessly with other u-bmc packages:
//
//	// Integration with state machine for thermal management
//	tempSensor, _ := hwmon.NewSensor(tempConfig)
//	thermalSM, _ := state.New(thermalConfig)
//
//	go func() {
//		for {
//			temp, _ := tempSensor.ReadValue(ctx)
//			celsius := temp.AsTemperature().Celsius()
//
//			switch {
//			case celsius > 90:
//				thermalSM.Fire(ctx, "critical", nil)
//			case celsius > 80:
//				thermalSM.Fire(ctx, "warning", nil)
//			default:
//				thermalSM.Fire(ctx, "normal", nil)
//			}
//
//			time.Sleep(1 * time.Second)
//		}
//	}()
package hwmon
