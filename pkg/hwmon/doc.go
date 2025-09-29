// SPDX-License-Identifier: BSD-3-Clause

// Package hwmon provides simple, stateless functions for accessing Linux hwmon (hardware monitoring)
// subsystem through sysfs. This package enables direct interaction with hwmon devices without
// interpretation or abstraction layers, allowing applications to work directly with raw hwmon endpoints.
//
// # Overview
//
// The Linux hwmon subsystem exposes hardware monitoring devices through the sysfs filesystem under
// /sys/class/hwmon/. Each hwmon device represents a monitoring chip and provides various attributes
// for reading sensor values and device information.
//
// This package provides:
//   - Direct read/write access to hwmon sysfs attributes
//   - Simple device discovery by name
//   - Context-aware operations with timeout support
//   - Stateless, functional API design
//   - Minimal error handling and mapping
//
// # Core Concepts
//
// Hwmon Device: A hardware monitoring device exposed by the kernel at /sys/class/hwmon/hwmonN.
// Each device contains a 'name' file and various numbered sensor attributes.
//
// Attribute: A specific hwmon file such as temp1_input, fan2_label, in3_max, etc.
// Attributes follow the pattern: <type><number>_<property>
//
// # Basic Usage
//
// Reading a temperature sensor:
//
//	value, err := ReadInt("/sys/class/hwmon/hwmon0/temp1_input")
//	if err != nil {
//		log.Printf("Failed to read temperature: %v", err)
//		return
//	}
//	tempC := float64(value) / 1000.0 // Convert millidegrees to Celsius
//
// Writing a fan PWM value:
//
//	err := WriteInt("/sys/class/hwmon/hwmon1/pwm1", 127) // 50% duty cycle
//	if err != nil {
//		log.Printf("Failed to set fan speed: %v", err)
//	}
//
// Reading a sensor label:
//
//	label, err := ReadString("/sys/class/hwmon/hwmon0/temp1_label")
//	if err != nil {
//		log.Printf("Failed to read label: %v", err)
//		return
//	}
//
// Finding a device by name:
//
//	devicePath, err := FindDeviceByName("k10temp")
//	if err != nil {
//		log.Printf("Device not found: %v", err)
//		return
//	}
//	// devicePath will be something like "/sys/class/hwmon/hwmon0"
//
// # Device Discovery
//
// Finding devices and their attributes:
//
//	devices, err := ListDevices()
//	if err != nil {
//		log.Printf("Failed to list devices: %v", err)
//		return
//	}
//
//	for _, device := range devices {
//		name, _ := ReadString(filepath.Join(device, "name"))
//		fmt.Printf("Device: %s at %s\n", name, device)
//
//		attrs, _ := ListAttributes(device, "temp.*_input")
//		for _, attr := range attrs {
//			value, _ := ReadInt(filepath.Join(device, attr))
//			fmt.Printf("  %s: %d\n", attr, value)
//		}
//	}
//
// # Context Support
//
// All operations support context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	value, err := ReadIntCtx(ctx, "/sys/class/hwmon/hwmon0/temp1_input")
//	if err != nil {
//		if errors.Is(err, context.DeadlineExceeded) {
//			log.Println("Read operation timed out")
//		}
//		return
//	}
//
// # Common Attribute Patterns
//
// Temperature sensors (values in millidegrees Celsius):
//   - temp1_input, temp2_input, etc. - current temperature
//   - temp1_max, temp1_crit - maximum and critical thresholds
//   - temp1_label - human-readable name
//
// Voltage sensors (values in millivolts):
//   - in0_input, in1_input, etc. - current voltage
//   - in0_min, in0_max - minimum and maximum thresholds
//   - in0_label - voltage rail name
//
// Fan sensors:
//   - fan1_input - current RPM
//   - pwm1 - PWM duty cycle (0-255)
//   - fan1_label - fan description
//
// Power sensors (values in microwatts):
//   - power1_input - current power consumption
//   - power1_max - maximum power limit
//
// # Error Handling
//
// The package provides specific error types for common conditions:
//
//	value, err := ReadInt(path)
//	if err != nil {
//		switch {
//		case errors.Is(err, ErrFileNotFound):
//			log.Printf("Attribute not available")
//		case errors.Is(err, ErrPermissionDenied):
//			log.Printf("Insufficient permissions")
//		case errors.Is(err, ErrInvalidValue):
//			log.Printf("Value could not be parsed")
//		default:
//			log.Printf("Unexpected error: %v", err)
//		}
//		return
//	}
//
// # Integration Example
//
// Using with sensor monitoring:
//
//	func monitorTemperature(ctx context.Context, deviceName string) error {
//		devicePath, err := FindDeviceByName(deviceName)
//		if err != nil {
//			return err
//		}
//
//		tempPath := filepath.Join(devicePath, "temp1_input")
//		ticker := time.NewTicker(1 * time.Second)
//		defer ticker.Stop()
//
//		for {
//			select {
//			case <-ctx.Done():
//				return ctx.Err()
//			case <-ticker.C:
//				temp, err := ReadIntCtx(ctx, tempPath)
//				if err != nil {
//					log.Printf("Failed to read temperature: %v", err)
//					continue
//				}
//				tempC := float64(temp) / 1000.0
//				if tempC > 80.0 {
//					log.Printf("High temperature warning: %.1f°C", tempC)
//				}
//			}
//		}
//	}
//
// # Performance Considerations
//
// This package provides direct sysfs access without caching. For high-frequency monitoring,
// consider implementing application-level caching or rate limiting to avoid excessive
// filesystem operations.
//
// All functions are stateless and thread-safe, making them suitable for concurrent use
// across multiple goroutines.
package hwmon
