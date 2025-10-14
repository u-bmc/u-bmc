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
	// Most BMCs have only 512MB of RAM; limit memory usage to 256MB
	debug.SetMemoryLimit(256 * 1024 * 1024)

	// Configure mock sensors for testing
	sensorConfig := createMockSensorConfig()

	// Configure mock power management for testing
	powerConfig := createMockPowerConfig()

	// Configure mock thermal management for testing
	thermalConfig := createMockThermalConfig()

	// Configure state management
	stateConfig := []statemgr.Option{
		statemgr.WithStreamRetention(0), // Keep forever for audit trail
		statemgr.WithHostManagement(true),
		statemgr.WithChassisManagement(true),
		statemgr.WithBMCManagement(true),
		statemgr.WithNumHosts(1),
		statemgr.WithNumChassis(1),
		statemgr.WithStateTimeout(20 * time.Second),
		statemgr.WithBroadcastStateChanges(true),
		statemgr.WithPersistStateChanges(false),
	}

	webConfig := []websrv.Option{
		websrv.WithAddr(":443"),
		websrv.WithWebUI(false),
		websrv.WithWebUIPath("/opt/u-bmc/webui"),
		websrv.WithReadTimeout(30 * time.Second),
		websrv.WithWriteTimeout(30 * time.Second),
		websrv.WithIdleTimeout(120 * time.Second),
		websrv.WithRmemMax("7500000"),
		websrv.WithWmemMax("7500000"),
		websrv.WithCertificateType(cert.CertificateTypeSelfSigned),
		websrv.WithHostname("u-bmc-local.test"),
		websrv.WithCertPath("/var/cache/cert/ubmc-cert.pem"),
		websrv.WithKeyPath("/var/cache/cert/ubmc-key.pem"),
		websrv.WithAlternativeNames("u-bmc-local"),
	}

	if err := operator.New(
		// Init on this platform handles mounts; keep operator startup resilient.
		operator.WithMountCheck(false),
		// Not implemented or not needed for local testing
		operator.WithoutConsolesrv(),
		operator.WithoutInventorymgr(),
		operator.WithoutIpmisrv(),
		operator.WithoutTelemetry(),
		operator.WithoutUpdatemgr(),
		operator.WithoutUsermgr(),
		operator.WithoutSecuritymgr(),
		operator.WithoutLedmgr(),
		operator.WithoutKvmsrv(),
		// Implemented with mock backends for testing
		operator.WithStatemgr(stateConfig...),
		operator.WithWebsrv(webConfig...),
		operator.WithSensormon(sensorConfig...),
		operator.WithPowermgr(powerConfig...),
		operator.WithThermalmgr(thermalConfig...),
	).Run(context.Background(), nil); err != nil {
		panic(err)
	}
}

// createMockSensorConfig creates a comprehensive mock sensor configuration for testing.
func createMockSensorConfig() []sensormon.Option {
	// Define mock sensor definitions
	sensors := []sensormon.SensorDefinition{
		// CPU temperature sensors
		{
			ID:          "cpu0_temp",
			Name:        "CPU 0 Temperature",
			Description: "Main CPU die temperature",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS,
			Backend:     sensormon.BackendTypeMock,
			Location: sensormon.Location{
				Zone:      "cpu",
				Position:  "die",
				Component: "CPU0",
				Coordinates: map[string]string{
					"socket": "0",
					"core":   "all",
				},
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(75.0),
				Critical: ptrFloat64(85.0),
			},
			Enabled:    true,
			MockConfig: sensormon.NewMockTemperatureSensor(45.0), // Base temperature 45°C
		},
		{
			ID:          "cpu1_temp",
			Name:        "CPU 1 Temperature",
			Description: "Secondary CPU die temperature",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS,
			Backend:     sensormon.BackendTypeMock,
			Location: sensormon.Location{
				Zone:      "cpu",
				Position:  "die",
				Component: "CPU1",
				Coordinates: map[string]string{
					"socket": "1",
					"core":   "all",
				},
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(75.0),
				Critical: ptrFloat64(85.0),
			},
			Enabled:    true,
			MockConfig: sensormon.NewMockTemperatureSensor(42.0), // Base temperature 42°C
		},
		// Memory temperature sensor
		{
			ID:          "dimm_temp",
			Name:        "DIMM Temperature",
			Description: "Memory module temperature",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS,
			Backend:     sensormon.BackendTypeMock,
			Location: sensormon.Location{
				Zone:      "memory",
				Position:  "center",
				Component: "DIMM_A1",
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(65.0),
				Critical: ptrFloat64(75.0),
			},
			Enabled:    true,
			MockConfig: sensormon.NewMockTemperatureSensor(35.0), // Base temperature 35°C
		},
		// Fan sensors
		{
			ID:          "fan1",
			Name:        "System Fan 1",
			Description: "Main system cooling fan",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TACH,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_RPM,
			Backend:     sensormon.BackendTypeMock,
			Location: sensormon.Location{
				Zone:     "cooling",
				Position: "front",
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(800.0),
				Critical: ptrFloat64(500.0),
			},
			Enabled:    true,
			MockConfig: sensormon.NewMockFanSensor(1200.0), // Base speed 1200 RPM
		},
		{
			ID:          "fan2",
			Name:        "System Fan 2",
			Description: "Secondary system cooling fan",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_TACH,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_RPM,
			Backend:     sensormon.BackendTypeMock,
			Location: sensormon.Location{
				Zone:     "cooling",
				Position: "rear",
			},
			LowerThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(800.0),
				Critical: ptrFloat64(500.0),
			},
			Enabled:    true,
			MockConfig: sensormon.NewMockFanSensor(1100.0), // Base speed 1100 RPM
		},
		// Voltage sensors
		{
			ID:          "12v_rail",
			Name:        "12V Power Rail",
			Description: "Main 12V power supply rail",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_VOLTAGE,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_VOLTS,
			Backend:     sensormon.BackendTypeMock,
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
			Enabled:    true,
			MockConfig: sensormon.NewMockVoltageSensor(12.0), // Base voltage 12V
		},
		// Power sensors
		{
			ID:          "system_power",
			Name:        "System Power Consumption",
			Description: "Total system power draw",
			Context:     v1alpha1.SensorContext_SENSOR_CONTEXT_POWER,
			Unit:        v1alpha1.SensorUnit_SENSOR_UNIT_WATTS,
			Backend:     sensormon.BackendTypeMock,
			Location: sensormon.Location{
				Zone: "system",
			},
			UpperThresholds: &sensormon.Threshold{
				Warning:  ptrFloat64(400.0),
				Critical: ptrFloat64(450.0),
			},
			Enabled:    true,
			MockConfig: sensormon.NewMockPowerSensor(250.0), // Base power 250W
		},
	}

	// Configure sensor callbacks
	callbacks := sensormon.SensorCallbacks{
		OnSensorRead: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
			// Log sensor reads in test mode
			return nil
		},
		OnThresholdWarning: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
			// Handle warning thresholds
			return nil
		},
		OnThresholdCritical: func(sensorID string, event sensormon.SensorEvent, data interface{}) error {
			// Handle critical thresholds
			return nil
		},
	}

	return []sensormon.Option{
		sensormon.WithServiceName("sensormon"),
		sensormon.WithServiceDescription("Mock sensor monitoring service for testing"),
		sensormon.WithMonitoringInterval(2 * time.Second),
		sensormon.WithThresholdCheckInterval(5 * time.Second),
		sensormon.WithSensorTimeout(1 * time.Second),
		// Enable only mock sensors for testing
		sensormon.WithoutHwmonSensors(),
		sensormon.WithoutGPIOSensors(),
		sensormon.WithMockSensors(true),
		sensormon.WithThresholdMonitoring(true),
		sensormon.WithAutoDiscovery(false), // Use only configured sensors
		sensormon.WithBroadcastSensorReadings(true),
		sensormon.WithoutPersistSensorData(),
		// Enhanced configuration
		sensormon.WithSensorDefinitions(sensors...),
		sensormon.WithCallbacks(callbacks),
		sensormon.WithMockFailureSimulation(false, 0.0), // No failures for testing
		// Thermal integration
		sensormon.WithThermalIntegration(true),
		sensormon.WithThermalMgrEndpoint("thermalmgr"),
		sensormon.WithTemperatureUpdateInterval(3 * time.Second),
		sensormon.WithThermalAlerts(true),
		sensormon.WithWarningTempThreshold(70.0),
		sensormon.WithCriticalTempThreshold(80.0),
		sensormon.WithEmergencyResponseDelay(2 * time.Second),
	}
}

// createMockPowerConfig creates a mock power management configuration for testing.
func createMockPowerConfig() []powermgr.Option {
	// Define mock power components
	hostComponents := map[string]powermgr.ComponentConfig{
		"host.0": {
			Name:             "host.0",
			Type:             "host",
			Enabled:          true,
			Backend:          powermgr.BackendTypeMock,
			Mock:             powermgr.NewReliableMockConfig(false), // Start powered off
			OperationTimeout: 10 * time.Second,
			PowerOnDelay:     2 * time.Second,
			PowerOffDelay:    3 * time.Second,
			ResetDelay:       1 * time.Second,
			ForceOffDelay:    500 * time.Millisecond,
		},
	}

	chassisComponents := map[string]powermgr.ComponentConfig{
		"chassis.0": {
			Name:             "chassis.0",
			Type:             "chassis",
			Enabled:          true,
			Backend:          powermgr.BackendTypeMock,
			Mock:             powermgr.NewReliableMockConfig(true), // Start powered on
			OperationTimeout: 5 * time.Second,
			PowerOnDelay:     1 * time.Second,
			PowerOffDelay:    2 * time.Second,
			ResetDelay:       500 * time.Millisecond,
			ForceOffDelay:    200 * time.Millisecond,
		},
	}

	// Configure power callbacks
	callbacks := powermgr.PowerCallbacks{
		OnPowerOn: func(componentName string, event powermgr.PowerEvent, data interface{}) error {
			// Handle power on events
			return nil
		},
		OnPowerOff: func(componentName string, event powermgr.PowerEvent, data interface{}) error {
			// Handle power off events
			return nil
		},
		OnEmergencyShutdown: func(componentName string, event powermgr.PowerEvent, data interface{}) error {
			// Handle emergency shutdown
			return nil
		},
	}

	return []powermgr.Option{
		powermgr.WithServiceName("powermgr"),
		powermgr.WithServiceDescription("Mock power management service for testing"),
		powermgr.WithDefaultBackend(powermgr.BackendTypeMock),
		powermgr.WithMockBackends(true),
		// Component configuration
		powermgr.WithHostManagement(true),
		powermgr.WithChassisManagement(true),
		powermgr.WithBMCManagement(false), // Don't manage BMC power in test
		powermgr.WithNumHosts(1),
		powermgr.WithNumChassis(1),
		powermgr.WithDefaultOperationTimeout(10 * time.Second),
		// Add components
		powermgr.WithComponents(hostComponents),
		powermgr.WithComponents(chassisComponents),
		// State reporting
		powermgr.WithStateReporting(true),
		powermgr.WithStateReportingSubjectPrefix("powermgr.state"),
		// Thermal response
		powermgr.WithThermalResponse(true),
		powermgr.WithEmergencyResponseDelay(3 * time.Second),
		powermgr.WithEmergencyShutdown(true),
		powermgr.WithShutdownTemperatureLimit(90.0),
		powermgr.WithShutdownComponents([]string{"host.0"}),
		powermgr.WithMaxEmergencyAttempts(3),
		powermgr.WithEmergencyAttemptInterval(1 * time.Second),
		// Callbacks
		powermgr.WithCallbacks(callbacks),
	}
}

// createMockThermalConfig creates a mock thermal management configuration for testing.
func createMockThermalConfig() []thermalmgr.Option {
	// Define mock cooling devices
	coolingDevices := []thermalmgr.CoolingDeviceConfig{
		thermalmgr.NewMockCoolingDevice("fan1_control", "System Fan 1 Control", 20.0, 100.0, 40.0),
		thermalmgr.NewMockCoolingDevice("fan2_control", "System Fan 2 Control", 20.0, 100.0, 40.0),
	}

	// Define thermal zones
	thermalZones := []thermalmgr.ThermalZoneConfig{
		thermalmgr.NewThermalZone(
			"cpu_zone",
			"CPU Thermal Zone",
			[]string{"cpu0_temp", "cpu1_temp"},       // Monitor CPU temperatures
			[]string{"fan1_control", "fan2_control"}, // Control both fans
			60.0,                                     // Target temperature
			70.0,                                     // Warning temperature
			80.0,                                     // Critical temperature
		),
		thermalmgr.NewThermalZone(
			"system_zone",
			"System Thermal Zone",
			[]string{"dimm_temp", "system_power"}, // Monitor memory and power
			[]string{"fan2_control"},              // Control rear fan
			50.0,                                  // Target temperature
			60.0,                                  // Warning temperature
			70.0,                                  // Critical temperature
		),
	}

	// Configure thermal callbacks
	callbacks := thermalmgr.ThermalCallbacks{
		OnTemperatureWarning: func(zoneName string, event thermalmgr.ThermalEvent, data interface{}) error {
			// Handle temperature warnings
			return nil
		},
		OnTemperatureCritical: func(zoneName string, event thermalmgr.ThermalEvent, data interface{}) error {
			// Handle critical temperatures
			return nil
		},
		OnEmergencyShutdown: func(zoneName string, event thermalmgr.ThermalEvent, data interface{}) error {
			// Handle emergency shutdown
			return nil
		},
	}

	return []thermalmgr.Option{
		thermalmgr.WithServiceName("thermalmgr"),
		thermalmgr.WithServiceDescription("Mock thermal management service for testing"),
		thermalmgr.WithThermalControl(true),
		thermalmgr.WithThermalControlInterval(3 * time.Second),
		thermalmgr.WithEmergencyCheckInterval(1 * time.Second),
		thermalmgr.WithDefaultPIDSampleTime(2 * time.Second),
		thermalmgr.WithMaxThermalZones(10),
		thermalmgr.WithMaxCoolingDevices(20),
		thermalmgr.WithoutDiscovery(), // Use only configured zones and devices
		// Temperature thresholds for global monitoring
		thermalmgr.WithTemperatureThresholds(70.0, 80.0, 90.0), // Warning, Critical, Emergency
		// PID controller defaults
		thermalmgr.WithDefaultPIDConfig(1.0, 0.1, 0.05), // Kp, Ki, Kd
		thermalmgr.WithOutputRange(0.0, 100.0),
		// Integration with other services
		thermalmgr.WithSensorIntegration(true),
		thermalmgr.WithSensormonEndpoint("sensormon"),
		thermalmgr.WithPowerIntegration(true),
		thermalmgr.WithPowermgrEndpoint("powermgr"),
		// Data persistence (disabled for testing)
		thermalmgr.WithoutPersistThermalData(),
		// Emergency response
		thermalmgr.WithEmergencyResponse(true),
		thermalmgr.WithEmergencyResponseDelay(2 * time.Second),
		thermalmgr.WithFailsafeCoolingLevel(100.0), // Max cooling in emergency
		// Enhanced configuration
		thermalmgr.WithThermalZones(thermalZones...),
		thermalmgr.WithCoolingDevices(coolingDevices...),
		thermalmgr.WithCallbacks(callbacks),
		thermalmgr.WithMockMode(true), // Enable mock mode for testing
	}
}

// Helper function to create float64 pointers
func ptrFloat64(f float64) *float64 {
	return &f
}
