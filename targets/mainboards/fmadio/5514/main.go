// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/u-bmc/u-bmc/pkg/cert"
	"github.com/u-bmc/u-bmc/service/kvmsrv"
	"github.com/u-bmc/u-bmc/service/ledmgr"
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

	// Configure state management
	stateConfig := []statemgr.Option{
		statemgr.WithServiceDescription("FMADIO 5514 State Management Service"),
		statemgr.WithStreamName("STATEMGR"),
		statemgr.WithStreamSubjects("fmadio.statemgr.state.>", "fmadio.statemgr.event.>"),
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
		websrv.WithHostname("fmadio-5514-bmc.local"),
		websrv.WithCertPath("/var/cache/cert/fmadio-cert.pem"),
		websrv.WithKeyPath("/var/cache/cert/fmadio-key.pem"),
		websrv.WithAlternativeNames("fmadio-5514-bmc"),
	}

	powerComponents := map[string]powermgr.ComponentConfig{
		"host": {
			Name:    "host",
			Type:    "host",
			Enabled: true,
			Backend: powermgr.BackendTypeGPIO,
			GPIO: powermgr.GPIOConfig{
				PowerButton: powermgr.GPIOLineConfig{
					Line:         "PWR_BTN_OUT",
					Direction:    powermgr.DirectionOutput,
					ActiveState:  powermgr.ActiveLow,
					InitialValue: 1,
					Bias:         powermgr.BiasDisabled,
				},
				ResetButton: powermgr.GPIOLineConfig{
					Line:         "RST_BTN_OUT",
					Direction:    powermgr.DirectionOutput,
					ActiveState:  powermgr.ActiveLow,
					InitialValue: 1,
					Bias:         powermgr.BiasDisabled,
				},
				PowerStatus: powermgr.GPIOLineConfig{
					Line:        "PWROK_PS_BMC",
					Direction:   powermgr.DirectionInput,
					ActiveState: powermgr.ActiveHigh,
					Bias:        powermgr.BiasPullDown,
				},
			},
			OperationTimeout: 30 * time.Second,
			PowerOnDelay:     2 * time.Second,
			PowerOffDelay:    10 * time.Second,
			ResetDelay:       1 * time.Second,
			ForceOffDelay:    15 * time.Second,
		},
	}

	powerConfig := []powermgr.Option{
		powermgr.WithServiceDescription("FMADIO 5514 Power Management Service"),
		powermgr.WithGPIOChip("/dev/gpiochip0"),
		powermgr.WithComponents(powerComponents),
		powermgr.WithHostManagement(true),
		powermgr.WithChassisManagement(true),
		powermgr.WithBMCManagement(true),
		powermgr.WithNumHosts(1),
		powermgr.WithNumChassis(1),
		powermgr.WithDefaultOperationTimeout(30 * time.Second),
		powermgr.WithMetrics(true),
		powermgr.WithTracing(true),
		powermgr.WithThermalResponse(true),
		powermgr.WithEmergencyResponseDelay(5 * time.Second),
		powermgr.WithEmergencyShutdown(true),
		powermgr.WithShutdownTemperatureLimit(100.0),
		powermgr.WithShutdownComponents([]string{"host"}),
		powermgr.WithMaxEmergencyAttempts(3),
		powermgr.WithEmergencyAttemptInterval(5 * time.Second),
	}

	ledComponents := map[string]ledmgr.ComponentConfig{
		"host.0": {
			Name:    "host.0",
			Type:    "host",
			Enabled: true,
			LEDs: map[ledmgr.LEDType]ledmgr.LEDConfig{
				ledmgr.LEDTypePower: {
					Type:    ledmgr.LEDTypePower,
					Enabled: true,
					Backend: ledmgr.BackendTypeGPIO,
					GPIO: ledmgr.LEDGPIOConfig{
						Line:        "POWER_LED",
						ActiveState: ledmgr.ActiveHigh,
					},
				},
				ledmgr.LEDTypeStatus: {
					Type:    ledmgr.LEDTypeStatus,
					Enabled: true,
					Backend: ledmgr.BackendTypeGPIO,
					GPIO: ledmgr.LEDGPIOConfig{
						Line:        "STATUS_LED",
						ActiveState: ledmgr.ActiveHigh,
					},
				},
				ledmgr.LEDTypeError: {
					Type:    ledmgr.LEDTypeError,
					Enabled: true,
					Backend: ledmgr.BackendTypeGPIO,
					GPIO: ledmgr.LEDGPIOConfig{
						Line:        "ERROR_LED",
						ActiveState: ledmgr.ActiveHigh,
					},
				},
				ledmgr.LEDTypeIdentify: {
					Type:    ledmgr.LEDTypeIdentify,
					Enabled: true,
					Backend: ledmgr.BackendTypeGPIO,
					GPIO: ledmgr.LEDGPIOConfig{
						Line:        "IDENTIFY_LED",
						ActiveState: ledmgr.ActiveHigh,
					},
				},
			},
			OperationTimeout: ledmgr.DefaultOperationTimeout,
			BlinkInterval:    ledmgr.DefaultBlinkInterval,
		},
	}

	ledConfig := []ledmgr.Option{
		ledmgr.WithServiceDescription("FMADIO 5514 LED Management Service"),
		ledmgr.WithGPIOChip("/dev/gpiochip0"),
		ledmgr.WithDefaultBackend(ledmgr.BackendTypeGPIO),
		ledmgr.WithComponents(ledComponents),
		ledmgr.WithHostManagement(true),
		ledmgr.WithChassisManagement(false),
		ledmgr.WithBMCManagement(false),
		ledmgr.WithNumHosts(1),
		ledmgr.WithDefaultOperationTimeout(5 * time.Second),
		ledmgr.WithDefaultBlinkInterval(500 * time.Millisecond),
		ledmgr.WithMetrics(true),
		ledmgr.WithTracing(true),
	}

	kvmConfig := []kvmsrv.Option{
		kvmsrv.WithServiceName("fmadio-kvm"),
		kvmsrv.WithVideoDevice("/dev/video0"),
		kvmsrv.WithVideoResolution(1920, 1080),
		kvmsrv.WithVideoFPS(30),
		kvmsrv.WithVNCPort(5900),
		kvmsrv.WithHTTPPort(8080),
		kvmsrv.WithVNC(true),
		kvmsrv.WithHTTP(true),
		kvmsrv.WithUSB(false),
		kvmsrv.WithVNCMaxClients(2),
		kvmsrv.WithHTTPMaxClients(5),
		kvmsrv.WithJPEGQuality(80),
		kvmsrv.WithUSBGadgetName("fmadio-kvm"),
		kvmsrv.WithClientTimeout(10 * time.Minute),
		kvmsrv.WithFrameTimeout(5 * time.Second),
		kvmsrv.WithBufferCount(4),
	}

	sensorConfig := []sensormon.Option{
		sensormon.WithServiceDescription("FMADIO 5514 Sensor Monitoring Service"),
		sensormon.WithHwmonPath("/sys/class/hwmon"),
		sensormon.WithGPIOChipPath("/dev/gpiochip0"),
		sensormon.WithMonitoringInterval(2 * time.Second),
		sensormon.WithThresholdCheckInterval(10 * time.Second),
		sensormon.WithHwmonSensors(true),
		sensormon.WithGPIOSensors(true),
		sensormon.WithThresholdMonitoring(true),
		sensormon.WithBroadcastSensorReadings(true),
		sensormon.WithPersistSensorData(false),
		sensormon.WithStreamName("FMADIO_SENSORMON"),
		sensormon.WithStreamSubjects("fmadio.sensormon.data.>", "fmadio.sensormon.events.>"),
		sensormon.WithStreamRetention(168 * time.Hour), // 7 days
		sensormon.WithMaxConcurrentReads(16),
		sensormon.WithThermalIntegration(true),
		sensormon.WithThermalMgrEndpoint("thermalmgr"),
		sensormon.WithTemperatureUpdateInterval(5 * time.Second),
		sensormon.WithThermalAlerts(true),
		sensormon.WithThermalThresholds(75.0, 85.0),
		sensormon.WithEmergencyResponseDelay(3 * time.Second),
	}

	thermalConfig := []thermalmgr.Option{
		thermalmgr.WithServiceDescription("FMADIO 5514 Thermal Management Service"),
		thermalmgr.WithThermalControlInterval(2 * time.Second),
		thermalmgr.WithEmergencyCheckInterval(500 * time.Millisecond),
		thermalmgr.WithHwmonPath("/sys/class/hwmon"),
		thermalmgr.WithDiscovery(true),
		thermalmgr.WithDefaultPIDConfig(1.2, 0.1, 0.05),
		thermalmgr.WithTemperatureThresholds(75.0, 85.0, 95.0),
		thermalmgr.WithSensormonEndpoint("sensormon"),
		thermalmgr.WithPowermgrEndpoint("powermgr"),
		thermalmgr.WithIntegration(true, true),
		thermalmgr.WithPersistence(true, "FMADIO_THERMALMGR", 72*time.Hour),
		thermalmgr.WithEmergencyResponseConfig(true, 2*time.Second, 100.0),
	}

	if err := operator.New(
		operator.WithName("fmadio-5514-operator"),
		// Init on this platform handles mounts; keep operator startup resilient.
		operator.WithMountCheck(false),
		// Not implemented yet
		operator.WithoutConsolesrv(),
		operator.WithoutInventorymgr(),
		operator.WithoutIpmisrv(),
		operator.WithoutTelemetry(),
		operator.WithoutUpdatemgr(),
		operator.WithoutUsermgr(),
		operator.WithoutSecuritymgr(),
		// Implemented
		operator.WithStatemgr(stateConfig...),
		operator.WithWebsrv(webConfig...),
		operator.WithPowermgr(powerConfig...),
		operator.WithLedmgr(ledConfig...),
		operator.WithKvmsrv(kvmConfig...),
		operator.WithSensormon(sensorConfig...),
		operator.WithThermalmgr(thermalConfig...),
	).Run(context.Background(), nil); err != nil {
		panic(err)
	}
}
