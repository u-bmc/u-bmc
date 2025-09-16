// SPDX-License-Identifier: BSD-3-Clause

package thermalmgr

import (
	"context"
	"time"

	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/thermal"
	"github.com/u-bmc/u-bmc/service/powermgr"
	"github.com/u-bmc/u-bmc/service/sensormon"
)

// ExampleIntegratedThermalConfig demonstrates how to configure an integrated thermal management system
// with sensormon for temperature monitoring and powermgr for emergency response.
func ExampleIntegratedThermalConfig() (*ThermalMgr, *sensormon.SensorMon, *powermgr.PowerMgr) {
	// Configure thermal manager with PID control and emergency response
	thermalConfig := NewConfig(
		WithServiceName("thermalmgr"),
		WithServiceVersion("1.0.0"),
		WithThermalControlInterval(time.Second),
		WithEmergencyCheckInterval(500*time.Millisecond),
		WithHwmonPath("/sys/class/hwmon"),
		WithDefaultPIDConfig(2.0, 0.5, 0.1),         // Moderate PID gains
		WithTemperatureThresholds(75.0, 85.0, 95.0), // Warning, Critical, Emergency
		WithSensormonEndpoint("sensormon"),
		WithPowermgrEndpoint("powermgr"),
		WithIntegration(true, true),                        // Enable sensor and power integration
		WithEmergencyResponse(true, 2*time.Second, 100.0),  // Emergency cooling after 2s delay
		WithDiscovery(true),                                // Auto-discover cooling devices
		WithPersistence(false, "THERMALMGR", 24*time.Hour), // No persistence in example
	)

	// Configure sensor monitor with thermal integration
	sensorConfig := sensormon.NewConfig(
		sensormon.WithServiceName("sensormon"),
		sensormon.WithServiceVersion("1.0.0"),
		sensormon.WithHwmonPath("/sys/class/hwmon"),
		sensormon.WithEnableHwmonSensors(true),
		sensormon.WithMonitoringInterval(time.Second),
		sensormon.WithThresholdCheckInterval(2*time.Second),
		sensormon.WithEnableThresholdMonitoring(true),
		sensormon.WithBroadcastSensorReadings(true),
		// Thermal integration settings
		sensormon.WithEnableThermalIntegration(true),
		sensormon.WithThermalMgrEndpoint("thermalmgr"),
		sensormon.WithTemperatureUpdateInterval(5*time.Second),
		sensormon.WithEnableThermalAlerts(true),
		sensormon.WithThermalThresholds(75.0, 85.0), // Warning, Critical
		sensormon.WithEmergencyResponseDelay(5*time.Second),
	)

	// Configure power manager with thermal emergency response
	powerConfig := powermgr.NewConfig(
		powermgr.WithServiceName("powermgr"),
		powermgr.WithServiceVersion("1.0.0"),
		powermgr.WithGPIOChip("/dev/gpiochip0"),
		// Thermal emergency response settings
		powermgr.WithEnableThermalResponse(true),
		powermgr.WithEmergencyResponseDelay(5*time.Second),
		powermgr.WithEnableEmergencyShutdown(true),
		powermgr.WithShutdownTemperatureLimit(95.0), // Emergency shutdown at 95°C
		powermgr.WithShutdownComponents([]string{"host.0", "chassis.0"}),
		powermgr.WithMaxEmergencyAttempts(3),
		powermgr.WithEmergencyAttemptInterval(time.Second),
	)

	// Create service instances
	thermalMgr := New(
		WithServiceName(thermalConfig.ServiceName),
		WithServiceVersion(thermalConfig.ServiceVersion),
		WithThermalControlInterval(thermalConfig.ThermalControlInterval),
		WithEmergencyCheckInterval(thermalConfig.EmergencyCheckInterval),
		WithHwmonPath(thermalConfig.HwmonPath),
		WithDefaultPIDConfig(thermalConfig.DefaultPIDKp, thermalConfig.DefaultPIDKi, thermalConfig.DefaultPIDKd),
		WithTemperatureThresholds(thermalConfig.DefaultWarningTemp, thermalConfig.DefaultCriticalTemp, thermalConfig.EmergencyShutdownTemp),
		WithSensormonEndpoint(thermalConfig.SensormonEndpoint),
		WithPowermgrEndpoint(thermalConfig.PowermgrEndpoint),
		WithIntegration(thermalConfig.EnableSensorIntegration, thermalConfig.EnablePowerIntegration),
		WithEmergencyResponse(thermalConfig.EnableEmergencyResponse, thermalConfig.EmergencyResponseDelay, thermalConfig.FailsafeCoolingLevel),
		WithDiscovery(thermalConfig.EnableHwmonDiscovery),
	)

	sensorMon := sensormon.New(
		sensormon.WithServiceName(sensorConfig.ServiceName),
		sensormon.WithServiceVersion(sensorConfig.ServiceVersion),
		sensormon.WithHwmonPath(sensorConfig.HwmonPath),
		sensormon.WithEnableHwmonSensors(sensorConfig.EnableHwmonSensors),
		sensormon.WithMonitoringInterval(sensorConfig.MonitoringInterval),
		sensormon.WithThresholdCheckInterval(sensorConfig.ThresholdCheckInterval),
		sensormon.WithEnableThresholdMonitoring(sensorConfig.EnableThresholdMonitoring),
		sensormon.WithBroadcastSensorReadings(sensorConfig.BroadcastSensorReadings),
		sensormon.WithEnableThermalIntegration(sensorConfig.EnableThermalIntegration),
		sensormon.WithThermalMgrEndpoint(sensorConfig.ThermalMgrEndpoint),
		sensormon.WithTemperatureUpdateInterval(sensorConfig.TemperatureUpdateInterval),
		sensormon.WithEnableThermalAlerts(sensorConfig.EnableThermalAlerts),
		sensormon.WithThermalThresholds(sensorConfig.WarningTempThreshold, sensorConfig.CriticalTempThreshold),
		sensormon.WithEmergencyResponseDelay(sensorConfig.EmergencyResponseDelay),
	)

	powerMgr := powermgr.New(
		powermgr.WithServiceName(powerConfig.ServiceName),
		powermgr.WithServiceVersion(powerConfig.ServiceVersion),
		powermgr.WithGPIOChip(powerConfig.GPIOChip),
		powermgr.WithEnableThermalResponse(powerConfig.EnableThermalResponse),
		powermgr.WithEmergencyResponseDelay(powerConfig.EmergencyResponseDelay),
		powermgr.WithEnableEmergencyShutdown(powerConfig.EnableEmergencyShutdown),
		powermgr.WithShutdownTemperatureLimit(powerConfig.ShutdownTemperatureLimit),
		powermgr.WithShutdownComponents(powerConfig.ShutdownComponents),
		powermgr.WithMaxEmergencyAttempts(powerConfig.MaxEmergencyAttempts),
		powermgr.WithEmergencyAttemptInterval(powerConfig.EmergencyAttemptInterval),
	)

	return thermalMgr, sensorMon, powerMgr
}

// ExampleCustomThermalZone demonstrates creating a custom thermal zone with specific sensors and cooling devices.
func ExampleCustomThermalZone() *thermal.ThermalZone {
	// Create a thermal zone for CPU thermal management
	cpuZone := &thermal.ThermalZone{
		Name: "cpu_thermal_zone",
		SensorPaths: []string{
			"/sys/class/hwmon/hwmon0/temp1_input", // CPU die temperature
			"/sys/class/hwmon/hwmon0/temp2_input", // CPU package temperature
		},
		TargetTemperature:   65.0, // Target 65°C
		WarningTemperature:  75.0, // Warning at 75°C
		CriticalTemperature: 85.0, // Critical at 85°C
		PIDConfig: thermal.PIDConfig{
			Kp:         2.0,         // Proportional gain
			Ki:         0.3,         // Integral gain
			Kd:         0.1,         // Derivative gain
			SampleTime: time.Second, // 1 second sample time
			OutputMin:  10.0,        // Minimum 10% fan speed
			OutputMax:  100.0,       // Maximum 100% fan speed
		},
	}

	// Add cooling devices
	cpuFan := &thermal.CoolingDevice{
		Name:         "cpu_fan",
		Type:         v1alpha1.CoolingDeviceType_COOLING_DEVICE_TYPE_FAN,
		HwmonPath:    "/sys/class/hwmon/hwmon1/pwm1",
		MinPower:     25.0,  // Minimum 25% for bearing health
		MaxPower:     100.0, // Maximum 100%
		CurrentPower: 30.0,  // Start at 30%
	}

	caseFans := &thermal.CoolingDevice{
		Name:         "case_fans",
		Type:         v1alpha1.CoolingDeviceType_COOLING_DEVICE_TYPE_FAN,
		HwmonPath:    "/sys/class/hwmon/hwmon1/pwm2",
		MinPower:     20.0, // Minimum 20%
		MaxPower:     90.0, // Maximum 90% (quieter)
		CurrentPower: 25.0, // Start at 25%
	}

	cpuZone.CoolingDevices = []*thermal.CoolingDevice{cpuFan, caseFans}

	return cpuZone
}

// ExampleLiquidCoolingThermalZone demonstrates a liquid cooling setup with pump and radiator fans.
func ExampleLiquidCoolingThermalZone() *thermal.ThermalZone {
	// Create a thermal zone for liquid cooling system
	liquidZone := &thermal.ThermalZone{
		Name: "liquid_cooling_zone",
		SensorPaths: []string{
			"/sys/class/hwmon/hwmon0/temp1_input", // CPU temperature
			"/sys/class/hwmon/hwmon2/temp1_input", // Coolant temperature
		},
		TargetTemperature:   60.0, // Lower target for liquid cooling
		WarningTemperature:  70.0, // Warning threshold
		CriticalTemperature: 80.0, // Critical threshold
		PIDConfig: thermal.PIDConfig{
			Kp:         1.5,             // Lower proportional gain (liquid has thermal mass)
			Ki:         0.2,             // Lower integral gain
			Kd:         0.05,            // Lower derivative gain
			SampleTime: 2 * time.Second, // Longer sample time for thermal mass
			OutputMin:  15.0,            // Minimum pump/fan speed
			OutputMax:  100.0,           // Maximum pump/fan speed
		},
	}

	// Water pump
	pump := &thermal.CoolingDevice{
		Name:         "aio_pump",
		Type:         v1alpha1.CoolingDeviceType_COOLING_DEVICE_TYPE_WATER_PUMP,
		HwmonPath:    "/sys/class/hwmon/hwmon3/pwm1",
		MinPower:     50.0, // Pumps need minimum speed to avoid cavitation
		MaxPower:     100.0,
		CurrentPower: 60.0,
	}

	// Radiator fans
	radiatorFans := &thermal.CoolingDevice{
		Name:         "radiator_fans",
		Type:         v1alpha1.CoolingDeviceType_COOLING_DEVICE_TYPE_FAN,
		HwmonPath:    "/sys/class/hwmon/hwmon3/pwm2",
		MinPower:     15.0, // Can run very low
		MaxPower:     100.0,
		CurrentPower: 20.0,
	}

	liquidZone.CoolingDevices = []*thermal.CoolingDevice{pump, radiatorFans}

	return liquidZone
}

// ExampleQuietThermalProfile demonstrates a quiet cooling profile for low-noise operation.
func ExampleQuietThermalProfile() thermal.PIDConfig {
	return thermal.PIDConfig{
		Kp:         0.8,             // Lower proportional gain
		Ki:         0.1,             // Lower integral gain
		Kd:         0.02,            // Lower derivative gain
		SampleTime: 3 * time.Second, // Longer sample time (less aggressive)
		OutputMin:  15.0,            // Higher minimum to avoid stopping fans
		OutputMax:  60.0,            // Lower maximum for quiet operation
	}
}

// ExampleAggressiveThermalProfile demonstrates an aggressive cooling profile for maximum performance.
func ExampleAggressiveThermalProfile() thermal.PIDConfig {
	return thermal.PIDConfig{
		Kp:         3.0,                    // High proportional gain
		Ki:         1.0,                    // High integral gain
		Kd:         0.5,                    // High derivative gain
		SampleTime: 500 * time.Millisecond, // Fast response
		OutputMin:  20.0,                   // Higher minimum for better response
		OutputMax:  100.0,                  // Full cooling available
	}
}

// ExampleInitializeIntegratedSystem demonstrates initializing a complete integrated thermal management system.
func ExampleInitializeIntegratedSystem(ctx context.Context) error {
	// Get service instances
	thermalMgr, sensorMon, powerMgr := ExampleIntegratedThermalConfig()

	// Create custom thermal zones
	cpuZone := ExampleCustomThermalZone()
	liquidZone := ExampleLiquidCoolingThermalZone()

	// Initialize thermal zones
	if err := thermal.InitializeThermalZone(ctx, cpuZone); err != nil {
		return err
	}

	if err := thermal.InitializeThermalZone(ctx, liquidZone); err != nil {
		return err
	}

	// Note: In a real implementation, you would start the services using their Run methods
	// with appropriate NATS connections and context handling.

	_ = thermalMgr
	_ = sensorMon
	_ = powerMgr

	return nil
}

// ExampleThermalControlLoop demonstrates a manual thermal control loop implementation.
func ExampleThermalControlLoop(ctx context.Context, zone *thermal.ThermalZone) error {
	ticker := time.NewTicker(zone.PIDConfig.SampleTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Read current temperature
			temperature, err := thermal.ReadZoneTemperature(ctx, zone)
			if err != nil {
				continue // Skip this cycle on error
			}

			// Check for emergency conditions
			if err := thermal.CheckThermalEmergency(ctx, zone); err != nil {
				if err == thermal.ErrCriticalTemperature {
					// Emergency condition - maximum cooling
					thermal.SetCoolingOutput(ctx, zone, 100.0)
					continue
				}
			}

			// Update PID control
			output, err := thermal.UpdatePIDControl(ctx, zone, temperature)
			if err != nil {
				continue // Skip this cycle on error
			}

			// Apply cooling output
			if err := thermal.SetCoolingOutput(ctx, zone, output); err != nil {
				continue // Skip this cycle on error
			}
		}
	}
}
