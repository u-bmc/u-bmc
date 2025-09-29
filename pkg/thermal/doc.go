// SPDX-License-Identifier: BSD-3-Clause

// Package thermal provides stateless functions for thermal management and PID-based cooling control.
// This package enables direct control of cooling devices through Linux hwmon interfaces and implements
// software-based PID control loops for thermal regulation.
//
// # Overview
//
// The thermal package provides abstractions for thermal zone management, cooling device control,
// and PID-based temperature regulation. It integrates with Linux hwmon subsystem for reading
// temperature sensors and controlling cooling devices such as fans, pumps, and other thermal
// management hardware.
//
// This package provides:
//   - Thermal zone management with PID control loops
//   - Cooling device abstraction and control
//   - Temperature monitoring and threshold management
//   - Stateless, functional API design
//   - Context-aware operations with timeout support
//   - Integration with hwmon and other thermal interfaces
//
// # Core Concepts
//
// Thermal Zone: A logical grouping of temperature sensors and cooling devices that work together
// to maintain temperature within specified limits. Each zone has its own PID controller and
// target temperature.
//
// Cooling Device: A hardware component capable of reducing temperature, such as fans, water pumps,
// heat exchangers, or liquid coolers. Each device has controllable power levels and operational
// status.
//
// PID Controller: A software control loop that continuously adjusts cooling device output based
// on the difference between target and actual temperature, using proportional, integral, and
// derivative terms.
//
// # Basic Usage
//
// Creating and managing a thermal zone:
//
//	pidConfig := PIDConfig{
//		Kp:         2.0,
//		Ki:         0.5,
//		Kd:         0.1,
//		SampleTime: time.Second,
//		OutputMin:  0.0,
//		OutputMax:  100.0,
//	}
//
//	zone := &ThermalZone{
//		Name:              "cpu_zone",
//		TargetTemperature: 65.0,
//		PIDConfig:         pidConfig,
//	}
//
//	err := InitializeThermalZone(ctx, zone)
//	if err != nil {
//		log.Printf("Failed to initialize thermal zone: %v", err)
//	}
//
// Reading temperature and updating PID control:
//
//	temperature, err := ReadZoneTemperature(ctx, zone)
//	if err != nil {
//		log.Printf("Failed to read temperature: %v", err)
//		return
//	}
//
//	output, err := UpdatePIDControl(ctx, zone, temperature)
//	if err != nil {
//		log.Printf("PID update failed: %v", err)
//		return
//	}
//
//	err = SetCoolingOutput(ctx, zone, output)
//	if err != nil {
//		log.Printf("Failed to set cooling output: %v", err)
//	}
//
// # Cooling Device Control
//
// Managing individual cooling devices:
//
//	fan := &CoolingDevice{
//		Name:     "cpu_fan",
//		Type:     CoolingDeviceTypeFan,
//		HwmonPath: "/sys/class/hwmon/hwmon1/pwm1",
//		MinPower: 0,
//		MaxPower: 255,
//	}
//
//	err := SetCoolingDevicePower(ctx, fan, 50.0) // 50% power
//	if err != nil {
//		log.Printf("Failed to set fan speed: %v", err)
//	}
//
//	status, err := GetCoolingDeviceStatus(ctx, fan)
//	if err != nil {
//		log.Printf("Failed to get device status: %v", err)
//	}
//
// # Temperature Monitoring
//
// Reading from multiple temperature sensors:
//
//	sensors := []string{
//		"/sys/class/hwmon/hwmon0/temp1_input",
//		"/sys/class/hwmon/hwmon0/temp2_input",
//	}
//
//	temperatures, err := ReadMultipleTemperatures(ctx, sensors)
//	if err != nil {
//		log.Printf("Failed to read temperatures: %v", err)
//		return
//	}
//
//	avgTemp := CalculateAverageTemperature(temperatures)
//	maxTemp := FindMaximumTemperature(temperatures)
//
// # PID Control Configuration
//
// Configuring PID parameters for different thermal profiles:
//
//	// Aggressive cooling profile
//	aggressiveConfig := PIDConfig{
//		Kp:         3.0,  // High proportional gain
//		Ki:         1.0,  // Moderate integral gain
//		Kd:         0.5,  // Higher derivative gain
//		SampleTime: 500 * time.Millisecond,
//		OutputMin:  0.0,
//		OutputMax:  100.0,
//	}
//
//	// Quiet cooling profile
//	quietConfig := PIDConfig{
//		Kp:         1.0,  // Lower proportional gain
//		Ki:         0.2,  // Lower integral gain
//		Kd:         0.05, // Lower derivative gain
//		SampleTime: 2 * time.Second,
//		OutputMin:  0.0,
//		OutputMax:  60.0, // Limit maximum cooling
//	}
//
// # Thermal Zone Control Loop
//
// Implementing a complete thermal management loop:
//
//	func RunThermalControl(ctx context.Context, zone *ThermalZone) error {
//		ticker := time.NewTicker(zone.PIDConfig.SampleTime)
//		defer ticker.Stop()
//
//		for {
//			select {
//			case <-ctx.Done():
//				return ctx.Err()
//			case <-ticker.C:
//				temp, err := ReadZoneTemperature(ctx, zone)
//				if err != nil {
//					slog.WarnContext(ctx, "Failed to read temperature", "error", err)
//					continue
//				}
//
//				output, err := UpdatePIDControl(ctx, zone, temp)
//				if err != nil {
//					slog.ErrorContext(ctx, "PID control update failed", "error", err)
//					continue
//				}
//
//				if err := SetCoolingOutput(ctx, zone, output); err != nil {
//					slog.ErrorContext(ctx, "Failed to set cooling output", "error", err)
//					continue
//				}
//
//				slog.DebugContext(ctx, "Thermal control update",
//					"zone", zone.Name,
//					"temperature", temp,
//					"target", zone.TargetTemperature,
//					"output", output)
//			}
//		}
//	}
//
// # Emergency Thermal Management
//
// Handling critical temperature conditions:
//
//	func CheckThermalEmergency(ctx context.Context, zone *ThermalZone) error {
//		temp, err := ReadZoneTemperature(ctx, zone)
//		if err != nil {
//			return err
//		}
//
//		if temp > zone.CriticalTemperature {
//			slog.ErrorContext(ctx, "Critical temperature exceeded",
//				"zone", zone.Name,
//				"temperature", temp,
//				"critical", zone.CriticalTemperature)
//
//			// Set maximum cooling immediately
//			return SetCoolingOutput(ctx, zone, 100.0)
//		}
//
//		if temp > zone.WarningTemperature {
//			slog.WarnContext(ctx, "Warning temperature exceeded",
//				"zone", zone.Name,
//				"temperature", temp,
//				"warning", zone.WarningTemperature)
//		}
//
//		return nil
//	}
//
// # Context Support
//
// All operations support context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	err := SetCoolingDevicePower(ctx, device, 75.0)
//	if err != nil {
//		if errors.Is(err, context.DeadlineExceeded) {
//			log.Println("Cooling device operation timed out")
//		}
//		return
//	}
//
// # Error Handling
//
// The package provides specific error types for thermal management conditions:
//
//	err := UpdatePIDControl(ctx, zone, temperature)
//	if err != nil {
//		switch {
//		case errors.Is(err, ErrPIDNotInitialized):
//			log.Printf("PID controller not initialized")
//		case errors.Is(err, ErrInvalidTemperature):
//			log.Printf("Invalid temperature reading")
//		case errors.Is(err, ErrCoolingDeviceUnavailable):
//			log.Printf("Cooling device not available")
//		default:
//			log.Printf("Unexpected thermal error: %v", err)
//		}
//		return
//	}
//
// # Performance Considerations
//
// This package provides stateless functions suitable for high-frequency thermal control loops.
// PID controllers maintain internal state but can be safely used across multiple goroutines
// with proper synchronization at the application level.
//
// For optimal thermal control, sample times should be chosen based on thermal time constants
// of the system being controlled. Typical values range from 100ms to 5 seconds depending on
// the thermal mass and response characteristics.
//
// All functions are designed to be thread-safe when used with separate device instances,
// making them suitable for concurrent thermal zone management.
package thermal
