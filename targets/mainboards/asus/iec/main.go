// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"runtime/debug"
	"time"

	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/cert"
	"github.com/u-bmc/u-bmc/service/operator"
	"github.com/u-bmc/u-bmc/service/powermgr"
	"github.com/u-bmc/u-bmc/service/sensormon"
	"github.com/u-bmc/u-bmc/service/statemgr"
	"github.com/u-bmc/u-bmc/service/thermalmgr"
	"github.com/u-bmc/u-bmc/service/websrv"
)

func main() {
	// The device has only 512MB of RAM; limit memory usage to 256MB
	debug.SetMemoryLimit(256 * 1024 * 1024)

	// Configure enhanced sensor monitoring for ASUS IEC AST2600 IPMI card
	sensorConfig := createASUSIECSensorConfig()

	// Configure enhanced power management for ASUS IEC AST2600 IPMI card
	powerConfig := createASUSIECPowerConfig()

	// Configure enhanced thermal management for ASUS IEC AST2600 IPMI card
	thermalConfig := createASUSIECThermalConfig()

	// Configure state management
	stateConfig := []statemgr.Option{
		statemgr.WithServiceName("statemgr"),
		statemgr.WithServiceDescription("ASUS IEC IPMI Card State Management Service"),
		statemgr.WithStreamName("ASUS_IEC_STATEMGR"),
		statemgr.WithStreamSubjects("asus.iec.statemgr.state.>", "asus.iec.statemgr.event.>"),
		statemgr.WithStreamRetention(0), // Keep forever for audit trail
		statemgr.WithHostManagement(true),
		statemgr.WithChassisManagement(true),
		statemgr.WithBMCManagement(true),
		statemgr.WithNumHosts(1),
		statemgr.WithNumChassis(1),
		statemgr.WithStateTimeout(60 * time.Second),
		statemgr.WithBroadcastStateChanges(true),
		statemgr.WithPersistStateChanges(true),
	}

	// Configure web server with KVM and SOL support for AST2600 IPMI card
	webConfig := []websrv.Option{
		websrv.WithServiceName("websrv"),
		websrv.WithAddr(":8443"),
		websrv.WithWebUI(true),
		websrv.WithWebUIPath("/opt/u-bmc/webui"),
		websrv.WithReadTimeout(30 * time.Second),
		websrv.WithWriteTimeout(30 * time.Second),
		websrv.WithIdleTimeout(120 * time.Second),
		websrv.WithRmemMax("7500000"),
		websrv.WithWmemMax("7500000"),
		websrv.WithCertificateType(cert.CertificateTypeSelfSigned),
		websrv.WithHostname("asus-iec-ipmi-bmc.local"),
		websrv.WithCertPath("/var/cache/cert/asus-iec-cert.pem"),
		websrv.WithKeyPath("/var/cache/cert/asus-iec-key.pem"),
		websrv.WithAlternativeNames("asus-iec-ipmi-bmc", "asus-iec.local"),
	}

	if err := operator.New(
		operator.WithServiceName("asus-iec-ipmi-expansion-card-operator"),
		// Init on this platform handles mounts; keep operator startup resilient.
		operator.WithMountCheck(false),
		// Not implemented or not needed for this hardware
		operator.WithoutConsolesrv(),
		operator.WithoutInventorymgr(),
		operator.WithoutIpmisrv(),
		operator.WithoutTelemetry(),
		operator.WithoutUpdatemgr(),
		operator.WithoutUsermgr(),
		operator.WithoutSecuritymgr(),
		operator.WithoutLedmgr(),
		operator.WithoutKvmsrv(),
		// Implemented services with enhanced configuration
		operator.WithStatemgr(stateConfig...),
		operator.WithWebsrv(webConfig...),
		operator.WithSensormon(sensorConfig...),
		operator.WithPowermgr(powerConfig...),
		operator.WithThermalmgr(thermalConfig...),
	).Run(context.Background(), nil); err != nil {
		panic(err)
	}
}

// createASUSIECSensorConfig creates sensor configuration for ASUS IEC AST2600 IPMI card.
// This card has access to motherboard sensors through hwmon and can monitor system temperatures,
// voltages, fan speeds, and power consumption.
func createASUSIECSensorConfig() []sensormon.Option {
	// Define comprehensive sensor definitions for ASUS IEC card
	sensors := []sensormon.SensorDefinition{
		// CPU temperature sensors
		{
			ID:          "cpu_temp",
			Name:        "CPU Temperature",
			Description: "Main CPU die temperature from motherboard",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:      "cpu",
				Position:  "die",
				Component: "CPU_MAIN",
				Coordinates: map[string]string{
					"socket": "main",
				},
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(75.0),
				Critical: ptrFloat64(85.0),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"coretemp",
				"temp1_input",
				"temp1_input", "temp1_label",
			),
			CustomAttributes: map[string]string{
				"thermal_zone": "cpu_zone",
				"sensor_type":  "cpu_die",
			},
		},
		// System temperature sensors
		{
			ID:          "sys_temp_1",
			Name:        "System Temperature 1",
			Description: "System ambient temperature sensor 1",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:      "system",
				Position:  "inlet",
				Component: "SYS_TEMP_1",
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(45.0),
				Critical: ptrFloat64(55.0),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"temp1_input",
				"temp1_input", "temp1_label",
			),
			CustomAttributes: map[string]string{
				"thermal_zone": "system_zone",
				"sensor_type":  "ambient",
			},
		},
		{
			ID:          "sys_temp_2",
			Name:        "System Temperature 2",
			Description: "System ambient temperature sensor 2",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:      "system",
				Position:  "outlet",
				Component: "SYS_TEMP_2",
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(50.0),
				Critical: ptrFloat64(60.0),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"temp2_input",
				"temp2_input", "temp2_label",
			),
			CustomAttributes: map[string]string{
				"thermal_zone": "system_zone",
				"sensor_type":  "ambient",
			},
		},
		// Memory temperature sensor
		{
			ID:          "dimm_temp",
			Name:        "DIMM Temperature",
			Description: "Memory module temperature",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:      "memory",
				Position:  "center",
				Component: "DIMM_MAIN",
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(70.0),
				Critical: ptrFloat64(80.0),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"i2c-.*-1a",
				"temp1_input",
				"temp1_input",
			),
			CustomAttributes: map[string]string{
				"thermal_zone": "memory_zone",
				"sensor_type":  "memory",
			},
		},
		// Fan speed sensors (8 fan controls on AST2600)
		{
			ID:          "fan1",
			Name:        "System Fan 1",
			Description: "CPU cooling fan 1",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TACH,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_RPM,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:     "cooling",
				Position: "cpu_front",
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(800.0),
				Critical: ptrFloat64(500.0),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"fan1_input",
				"fan1_input", "fan1_label",
			),
			CustomAttributes: map[string]string{
				"cooling_device": "fan1_control",
				"fan_type":       "cpu_cooler",
			},
		},
		{
			ID:          "fan2",
			Name:        "System Fan 2",
			Description: "CPU cooling fan 2",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TACH,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_RPM,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:     "cooling",
				Position: "cpu_rear",
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(800.0),
				Critical: ptrFloat64(500.0),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"fan2_input",
				"fan2_input", "fan2_label",
			),
			CustomAttributes: map[string]string{
				"cooling_device": "fan2_control",
				"fan_type":       "cpu_cooler",
			},
		},
		{
			ID:          "fan3",
			Name:        "System Fan 3",
			Description: "Case cooling fan 1",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TACH,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_RPM,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:     "cooling",
				Position: "case_front",
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(600.0),
				Critical: ptrFloat64(300.0),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"fan3_input",
				"fan3_input", "fan3_label",
			),
			CustomAttributes: map[string]string{
				"cooling_device": "fan3_control",
				"fan_type":       "case_fan",
			},
		},
		{
			ID:          "fan4",
			Name:        "System Fan 4",
			Description: "Case cooling fan 2",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TACH,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_RPM,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:     "cooling",
				Position: "case_rear",
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(600.0),
				Critical: ptrFloat64(300.0),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"fan4_input",
				"fan4_input", "fan4_label",
			),
			CustomAttributes: map[string]string{
				"cooling_device": "fan4_control",
				"fan_type":       "case_fan",
			},
		},
		// Voltage sensors
		{
			ID:          "12v_rail",
			Name:        "12V Power Rail",
			Description: "Main 12V power supply rail",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_VOLTAGE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_VOLTS,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:      "psu",
				Component: "PSU_MAIN",
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(12.6),
				Critical: ptrFloat64(13.2),
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(11.4),
				Critical: ptrFloat64(10.8),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"in0_input",
				"in0_input", "in0_label",
			),
		},
		{
			ID:          "5v_rail",
			Name:        "5V Power Rail",
			Description: "5V power supply rail",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_VOLTAGE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_VOLTS,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:      "psu",
				Component: "PSU_MAIN",
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(5.25),
				Critical: ptrFloat64(5.5),
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(4.75),
				Critical: ptrFloat64(4.5),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"in1_input",
				"in1_input", "in1_label",
			),
		},
		{
			ID:          "3v3_rail",
			Name:        "3.3V Power Rail",
			Description: "3.3V power supply rail",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_VOLTAGE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_VOLTS,
			Backend:     sensormon.BackendTypeHwmon,
			Location: sensormon.Location{
				Zone:      "psu",
				Component: "PSU_MAIN",
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(3.465),
				Critical: ptrFloat64(3.63),
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(3.135),
				Critical: ptrFloat64(2.97),
			},
			Enabled: true,
			HwmonConfig: sensormon.NewHwmonSensorConfigWithPattern(
				"nct6775",
				"in2_input",
				"in2_input", "in2_label",
			),
		},
	}

	// Configure sensor event callbacks
	callbacks := sensormon.SensorCallbacks{
		OnSensorRead: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
			// Log sensor reads for debugging
			return nil
		},
		OnThresholdWarning: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
			// Handle warning thresholds - could adjust fan speeds
			return nil
		},
		OnThresholdCritical: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
			// Handle critical thresholds - could trigger emergency cooling
			return nil
		},
		OnSensorError: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
			// Handle sensor read errors
			return nil
		},
	}

	return []sensormon.Option{
		sensormon.WithServiceName("asus-iec-sensormon"),
		sensormon.WithServiceDescription("ASUS IEC IPMI Card Sensor Monitoring Service"),
		sensormon.WithHwmonPath("/sys/class/hwmon"),
		sensormon.WithGPIOChipPath("/dev/gpiochip0"),
		sensormon.WithMonitoringInterval(2 * time.Second),
		sensormon.WithThresholdCheckInterval(5 * time.Second),
		sensormon.WithSensorTimeout(3 * time.Second),
		// Enable hwmon sensors with auto-discovery
		sensormon.WithHwmonSensors(true),
		sensormon.WithoutGPIOSensors(),
		sensormon.WithoutMockSensors(),
		sensormon.WithThresholdMonitoring(true),
		sensormon.WithAutoDiscovery(true), // Auto-discover additional sensors
		sensormon.WithBroadcastSensorReadings(true),
		sensormon.WithPersistSensorData(true),
		sensormon.WithStreamName("ASUS_IEC_SENSORMON"),
		sensormon.WithStreamSubjects("asus.iec.sensormon.data.>", "asus.iec.sensormon.events.>"),
		sensormon.WithStreamRetention(168 * time.Hour), // 7 days
		sensormon.WithMaxConcurrentReads(8),
		// Enhanced configuration
		sensormon.WithSensorDefinitions(sensors...),
		sensormon.WithCallbacks(callbacks),
		// Thermal integration
		sensormon.WithThermalIntegration(true),
		sensormon.WithThermalMgrEndpoint("asus-iec-thermalmgr"),
		sensormon.WithTemperatureUpdateInterval(3 * time.Second),
		sensormon.WithThermalAlerts(true),
		sensormon.WithWarningTempThreshold(75.0),
		sensormon.WithCriticalTempThreshold(85.0),
		sensormon.WithEmergencyResponseDelay(3 * time.Second),
	}
}

// createASUSIECPowerConfig creates power management configuration for ASUS IEC AST2600 IPMI card.
// This card has GPIO-based power control for the host system and chassis.
func createASUSIECPowerConfig() []powermgr.Option {
	// Define power management components with GPIO control
	hostComponents := map[string]powermgr.ComponentConfig{
		"main-host": {
			Name:    "Main Host System",
			Type:    "host",
			Enabled: true,
			Backend: powermgr.BackendTypeGPIO,
			GPIO: powermgr.NewGPIOConfig(
				18, // Power button GPIO (AST2600 GPIO pin)
				19, // Reset button GPIO
				20, // Power status GPIO
			),
			OperationTimeout: 30 * time.Second,
			PowerOnDelay:     3 * time.Second,
			PowerOffDelay:    10 * time.Second,
			ResetDelay:       1 * time.Second,
			ForceOffDelay:    15 * time.Second,
		},
	}

	chassisComponents := map[string]powermgr.ComponentConfig{
		"main-chassis": {
			Name:    "Main Chassis",
			Type:    "chassis",
			Enabled: true,
			Backend: powermgr.BackendTypeGPIO,
			GPIO: powermgr.NewGPIOConfig(
				21, // Chassis power GPIO
				22, // Chassis reset GPIO
				23, // Chassis status GPIO
			),
			OperationTimeout: 15 * time.Second,
			PowerOnDelay:     2 * time.Second,
			PowerOffDelay:    5 * time.Second,
			ResetDelay:       500 * time.Millisecond,
			ForceOffDelay:    2 * time.Second,
		},
	}

	// Configure power event callbacks
	callbacks := powermgr.PowerCallbacks{
		OnPowerOn: func(componentName string, event powermgr.PowerEvent, data interface{}) error {
			// Handle power on events - could update LED status
			return nil
		},
		OnPowerOff: func(componentName string, event powermgr.PowerEvent, data interface{}) error {
			// Handle power off events
			return nil
		},
		OnReset: func(componentName string, event powermgr.PowerEvent, data interface{}) error {
			// Handle reset events
			return nil
		},
		OnEmergencyShutdown: func(componentName string, event powermgr.PowerEvent, data interface{}) error {
			// Handle emergency thermal shutdown
			return nil
		},
		OnThermalThrottling: func(componentName string, event powermgr.PowerEvent, data interface{}) error {
			// Handle thermal throttling events
			return nil
		},
	}

	return []powermgr.Option{
		powermgr.WithServiceName("asus-iec-powermgr"),
		powermgr.WithServiceDescription("ASUS IEC IPMI Card Power Management Service"),
		powermgr.WithGPIOChip("/dev/gpiochip0"), // AST2600 GPIO chip
		powermgr.WithDefaultBackend(powermgr.BackendTypeGPIO),
		// Component configuration
		powermgr.WithComponents(hostComponents),
		powermgr.WithComponents(chassisComponents),
		powermgr.WithHostManagement(true),
		powermgr.WithChassisManagement(true),
		powermgr.WithoutBMCManagement(), // Don't control BMC power
		powermgr.WithNumHosts(1),
		powermgr.WithNumChassis(1),
		powermgr.WithDefaultOperationTimeout(30 * time.Second),
		// State reporting
		powermgr.WithStateReporting(true),
		powermgr.WithStateReportingSubjectPrefix("asus.iec.powermgr.state"),
		// Thermal response
		powermgr.WithThermalResponse(true),
		powermgr.WithEmergencyResponseDelay(5 * time.Second),
		powermgr.WithEmergencyShutdown(true),
		powermgr.WithShutdownTemperatureLimit(90.0),
		powermgr.WithShutdownComponents([]string{"main-host"}),
		powermgr.WithMaxEmergencyAttempts(3),
		powermgr.WithEmergencyAttemptInterval(30 * time.Second),
		// Callbacks
		powermgr.WithCallbacks(callbacks),
	}
}

// createASUSIECThermalConfig creates thermal management configuration for ASUS IEC AST2600 IPMI card.
// This card supports 8 fan controls for comprehensive thermal management.
func createASUSIECThermalConfig() []thermalmgr.Option {
	// Define cooling devices (8 fan controls on AST2600)
	coolingDevices := []thermalmgr.CoolingDeviceConfig{
		thermalmgr.NewFanDevice("fan1_control", "CPU Fan 1 Control", "/sys/class/hwmon/hwmon1", 1, 20.0, 100.0, 30.0),
		thermalmgr.NewFanDevice("fan2_control", "CPU Fan 2 Control", "/sys/class/hwmon/hwmon1", 2, 20.0, 100.0, 30.0),
		thermalmgr.NewFanDevice("fan3_control", "Case Fan 1 Control", "/sys/class/hwmon/hwmon1", 3, 10.0, 100.0, 25.0),
		thermalmgr.NewFanDevice("fan4_control", "Case Fan 2 Control", "/sys/class/hwmon/hwmon1", 4, 10.0, 100.0, 25.0),
		thermalmgr.NewFanDevice("fan5_control", "Auxiliary Fan 1 Control", "/sys/class/hwmon/hwmon1", 5, 0.0, 100.0, 20.0),
		thermalmgr.NewFanDevice("fan6_control", "Auxiliary Fan 2 Control", "/sys/class/hwmon/hwmon1", 6, 0.0, 100.0, 20.0),
		thermalmgr.NewFanDevice("fan7_control", "PSU Fan Control", "/sys/class/hwmon/hwmon1", 7, 15.0, 100.0, 35.0),
		thermalmgr.NewFanDevice("fan8_control", "Extra Fan Control", "/sys/class/hwmon/hwmon1", 8, 0.0, 100.0, 15.0),
	}

	// Define thermal zones with PID control
	thermalZones := []thermalmgr.ThermalZoneConfig{
		// CPU thermal zone
		{
			ID:               "cpu_zone",
			Name:             "CPU Thermal Zone",
			Description:      "Primary CPU thermal management zone",
			Enabled:          true,
			SensorIDs:        []string{"cpu_temp"},
			CoolingDeviceIDs: []string{"fan1_control", "fan2_control"},
			TargetTemp:       65.0,
			WarningTemp:      75.0,
			CriticalTemp:     85.0,
			EmergencyTemp:    95.0,
			PIDConfig: thermalmgr.NewPIDConfig(
				1.5, 0.15, 0.1, // Kp, Ki, Kd - tuned for CPU cooling
				2*time.Second, // Sample time
				20.0, 100.0,   // Output range (20-100% fan speed)
			),
			CustomAttributes: map[string]string{
				"priority":  "high",
				"zone_type": "cpu",
			},
		},
		// System thermal zone
		{
			ID:               "system_zone",
			Name:             "System Thermal Zone",
			Description:      "General system thermal management zone",
			Enabled:          true,
			SensorIDs:        []string{"sys_temp_1", "sys_temp_2"},
			CoolingDeviceIDs: []string{"fan3_control", "fan4_control", "fan5_control", "fan6_control"},
			TargetTemp:       40.0,
			WarningTemp:      45.0,
			CriticalTemp:     55.0,
			EmergencyTemp:    65.0,
			PIDConfig: thermalmgr.NewPIDConfig(
				1.2, 0.1, 0.05, // Kp, Ki, Kd - tuned for case cooling
				3*time.Second, // Sample time
				10.0, 80.0,    // Output range (10-80% fan speed)
			),
			CustomAttributes: map[string]string{
				"priority":  "medium",
				"zone_type": "system",
			},
		},
		// Memory thermal zone
		{
			ID:               "memory_zone",
			Name:             "Memory Thermal Zone",
			Description:      "Memory thermal management zone",
			Enabled:          true,
			SensorIDs:        []string{"dimm_temp"},
			CoolingDeviceIDs: []string{"fan3_control", "fan5_control"},
			TargetTemp:       50.0,
			WarningTemp:      70.0,
			CriticalTemp:     80.0,
			EmergencyTemp:    90.0,
			PIDConfig: thermalmgr.NewPIDConfig(
				1.0, 0.08, 0.02, // Kp, Ki, Kd - gentle for memory cooling
				4*time.Second, // Sample time
				15.0, 60.0,    // Output range (15-60% fan speed)
			),
			CustomAttributes: map[string]string{
				"priority":  "low",
				"zone_type": "memory",
			},
		},
	}

	// Configure thermal event callbacks
	callbacks := thermalmgr.ThermalCallbacks{
		OnTemperatureWarning: func(zoneName string, event thermalmgr.ThermalEvent, data interface{}) error {
			// Handle temperature warnings - could increase fan speeds
			return nil
		},
		OnTemperatureCritical: func(zoneName string, event thermalmgr.ThermalEvent, data interface{}) error {
			// Handle critical temperatures - could trigger emergency cooling
			return nil
		},
		OnEmergencyShutdown: func(zoneName string, event thermalmgr.ThermalEvent, data interface{}) error {
			// Handle emergency shutdown - power off system
			return nil
		},
		OnCoolingEngaged: func(zoneName string, event thermalmgr.ThermalEvent, data interface{}) error {
			// Handle cooling engagement events
			return nil
		},
		OnPIDControllerUpdated: func(zoneName string, event thermalmgr.ThermalEvent, data interface{}) error {
			// Handle PID controller updates
			return nil
		},
	}

	return []thermalmgr.Option{
		thermalmgr.WithServiceName("asus-iec-thermalmgr"),
		thermalmgr.WithServiceDescription("ASUS IEC IPMI Card Thermal Management Service"),
		thermalmgr.WithThermalControl(true),
		thermalmgr.WithThermalControlInterval(2 * time.Second),
		thermalmgr.WithEmergencyCheckInterval(500 * time.Millisecond),
		thermalmgr.WithDefaultPIDSampleTime(2 * time.Second),
		thermalmgr.WithMaxThermalZones(16),
		thermalmgr.WithMaxCoolingDevices(8),
		thermalmgr.WithHwmonPath("/sys/class/hwmon"),
		thermalmgr.WithDiscovery(true), // Auto-discover additional thermal devices
		// Global temperature thresholds
		thermalmgr.WithTemperatureThresholds(75.0, 85.0, 95.0), // Warning, Critical, Emergency
		// Default PID configuration
		thermalmgr.WithDefaultPIDConfig(1.2, 0.1, 0.05), // Conservative defaults
		thermalmgr.WithOutputRange(0.0, 100.0),
		// Integration with other services
		thermalmgr.WithSensorIntegration(true),
		thermalmgr.WithSensormonEndpoint("asus-iec-sensormon"),
		thermalmgr.WithPowerIntegration(true),
		thermalmgr.WithPowermgrEndpoint("asus-iec-powermgr"),
		// Data persistence
		thermalmgr.WithPersistence(true, "ASUS_IEC_THERMALMGR", 72*time.Hour),
		// Emergency response
		thermalmgr.WithEmergencyResponse(true),
		thermalmgr.WithEmergencyResponseDelay(2 * time.Second),
		thermalmgr.WithFailsafeCoolingLevel(100.0), // Full cooling in emergency
		// Enhanced configuration
		thermalmgr.WithThermalZones(thermalZones...),
		thermalmgr.WithCoolingDevices(coolingDevices...),
		thermalmgr.WithCallbacks(callbacks),
		thermalmgr.WithoutMockMode(), // Use real hardware
	}
}

// Helper function to create float64 pointers
func ptrFloat64(f float64) *float64 {
	return &f
}
