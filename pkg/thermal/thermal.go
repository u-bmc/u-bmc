// SPDX-License-Identifier: BSD-3-Clause

package thermal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"strings"
	"time"

	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/hwmon"
	"go.einride.tech/pid"
)

// PIDConfig holds configuration parameters for a PID controller.
type PIDConfig struct {
	Kp         float64       // Proportional gain
	Ki         float64       // Integral gain
	Kd         float64       // Derivative gain
	SampleTime time.Duration // Control loop sample time
	OutputMin  float64       // Minimum output value
	OutputMax  float64       // Maximum output value
}

// Zone represents a thermal management zone with sensors and cooling devices.
type Zone struct {
	Name                string
	SensorPaths         []string
	CoolingDevices      []*CoolingDevice
	TargetTemperature   float64
	WarningTemperature  float64
	CriticalTemperature float64
	PIDConfig           PIDConfig
	pidController       *pid.Controller
	lastTemperature     float64
	lastUpdate          time.Time
	currentOutput       float64
}

// CoolingDevice represents a controllable cooling device.
type CoolingDevice struct {
	Name         string
	Type         v1alpha1.CoolingDeviceType
	HwmonPath    string
	MinPower     float64
	MaxPower     float64
	CurrentPower float64
	Status       v1alpha1.CoolingDeviceStatus
}

// TemperatureReading holds a temperature measurement with metadata.
type TemperatureReading struct {
	Value     float64
	Timestamp time.Time
	SensorID  string
}

// ValidatePIDConfig validates PID configuration parameters.
func ValidatePIDConfig(config PIDConfig) error {
	if config.SampleTime <= 0 {
		return ErrInvalidSampleTime
	}
	if config.OutputMin >= config.OutputMax {
		return ErrOutputLimitsInvalid
	}
	if math.IsNaN(config.Kp) || math.IsNaN(config.Ki) || math.IsNaN(config.Kd) {
		return ErrInvalidPIDConfig
	}
	if math.IsInf(config.Kp, 0) || math.IsInf(config.Ki, 0) || math.IsInf(config.Kd, 0) {
		return ErrInvalidPIDConfig
	}
	return nil
}

// InitializeThermalZone initializes a thermal zone with its PID controller.
func InitializeThermalZone(ctx context.Context, zone *Zone) error {
	if zone == nil {
		return ErrInvalidZoneConfiguration
	}

	if err := ValidatePIDConfig(zone.PIDConfig); err != nil {
		slog.ErrorContext(ctx, "Invalid PID configuration",
			"zone", zone.Name,
			"error", err)
		return fmt.Errorf("%w: %w", ErrInvalidZoneConfiguration, err)
	}

	if len(zone.SensorPaths) == 0 {
		return fmt.Errorf("%w: no sensors configured", ErrInvalidZoneConfiguration)
	}

	if len(zone.CoolingDevices) == 0 {
		return fmt.Errorf("%w: no cooling devices configured", ErrInvalidZoneConfiguration)
	}

	zone.pidController = &pid.Controller{
		Config: pid.ControllerConfig{
			ProportionalGain: zone.PIDConfig.Kp,
			IntegralGain:     zone.PIDConfig.Ki,
			DerivativeGain:   zone.PIDConfig.Kd,
		},
	}

	zone.lastUpdate = time.Now()
	zone.currentOutput = 0.0

	slog.InfoContext(ctx, "Thermal zone initialized",
		"zone", zone.Name,
		"sensors", len(zone.SensorPaths),
		"cooling_devices", len(zone.CoolingDevices),
		"target_temp", zone.TargetTemperature)

	return nil
}

// ReadZoneTemperature reads and aggregates temperature from all sensors in a zone.
func ReadZoneTemperature(ctx context.Context, zone *Zone) (float64, error) {
	if zone == nil {
		return 0, ErrThermalZoneNotFound
	}

	if len(zone.SensorPaths) == 0 {
		return 0, ErrSensorReadFailure
	}

	var temperatures []float64 //nolint:prealloc
	var errs []error

	for _, sensorPath := range zone.SensorPaths {
		temp, err := ReadTemperatureFromPath(ctx, sensorPath)
		if err != nil {
			errs = append(errs, err)
			slog.WarnContext(ctx, "Failed to read sensor",
				"zone", zone.Name,
				"sensor", sensorPath,
				"error", err)
			continue
		}
		temperatures = append(temperatures, temp)
	}

	if len(temperatures) == 0 {
		return 0, fmt.Errorf("%w: all sensors failed: %w", ErrSensorReadFailure, errors.Join(errs...))
	}

	// Use maximum temperature for thermal management
	maxTemp := FindMaximumTemperature(temperatures)
	zone.lastTemperature = maxTemp

	return maxTemp, nil
}

// ReadTemperatureFromPath reads temperature from a hwmon sensor path.
func ReadTemperatureFromPath(ctx context.Context, sensorPath string) (float64, error) {
	if !hwmon.FileExistsCtx(ctx, sensorPath) {
		return 0, fmt.Errorf("%w: %s", ErrSensorReadFailure, sensorPath)
	}

	value, err := hwmon.ReadIntCtx(ctx, sensorPath)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrSensorReadFailure, err)
	}

	// Convert millidegrees to degrees Celsius
	tempC := float64(value) / 1000.0

	if tempC < -273.15 || tempC > 200.0 {
		return 0, fmt.Errorf("%w: %f°C", ErrInvalidTemperature, tempC)
	}

	return tempC, nil
}

// ReadMultipleTemperatures reads temperature from multiple sensor paths concurrently.
func ReadMultipleTemperatures(ctx context.Context, sensorPaths []string) ([]TemperatureReading, error) {
	if len(sensorPaths) == 0 {
		return nil, ErrSensorReadFailure
	}

	readings := make([]TemperatureReading, 0, len(sensorPaths))
	timestamp := time.Now()

	for _, path := range sensorPaths {
		temp, err := ReadTemperatureFromPath(ctx, path)
		if err != nil {
			slog.WarnContext(ctx, "Failed to read temperature sensor",
				"path", path,
				"error", err)
			continue
		}

		readings = append(readings, TemperatureReading{
			Value:     temp,
			Timestamp: timestamp,
			SensorID:  filepath.Base(path),
		})
	}

	if len(readings) == 0 {
		return nil, ErrSensorReadFailure
	}

	return readings, nil
}

// UpdatePIDControl updates the PID controller and returns the new output value.
func UpdatePIDControl(ctx context.Context, zone *Zone, currentTemperature float64) (float64, error) {
	if zone == nil {
		return 0, ErrThermalZoneNotFound
	}

	if zone.pidController == nil {
		return 0, ErrPIDNotInitialized
	}

	if math.IsNaN(currentTemperature) || math.IsInf(currentTemperature, 0) {
		return 0, ErrInvalidTemperature
	}

	now := time.Now()
	sampleInterval := now.Sub(zone.lastUpdate)

	// Ensure minimum sample time
	if sampleInterval < zone.PIDConfig.SampleTime/2 {
		return zone.currentOutput, nil
	}

	zone.pidController.Update(pid.ControllerInput{
		ReferenceSignal:  zone.TargetTemperature,
		ActualSignal:     currentTemperature,
		SamplingInterval: sampleInterval,
	})

	// Get the control signal and apply output limits
	output := zone.pidController.State.ControlSignal
	output = math.Max(zone.PIDConfig.OutputMin, math.Min(zone.PIDConfig.OutputMax, output))

	zone.currentOutput = output
	zone.lastUpdate = now

	slog.DebugContext(ctx, "PID control update",
		"zone", zone.Name,
		"current_temp", currentTemperature,
		"target_temp", zone.TargetTemperature,
		"error", zone.pidController.State.ControlError,
		"output", output,
		"sample_interval", sampleInterval)

	return output, nil
}

// SetCoolingOutput applies the specified cooling output to all devices in a zone.
func SetCoolingOutput(ctx context.Context, zone *Zone, outputPercent float64) error {
	if zone == nil {
		return ErrThermalZoneNotFound
	}

	if outputPercent < 0 || outputPercent > 100 {
		return ErrInvalidCoolingPower
	}

	var lastErr error
	successCount := 0

	for _, device := range zone.CoolingDevices {
		err := SetCoolingDevicePower(ctx, device, outputPercent)
		if err != nil {
			lastErr = err
			slog.WarnContext(ctx, "Failed to set cooling device power",
				"zone", zone.Name,
				"device", device.Name,
				"power", outputPercent,
				"error", err)
			continue
		}
		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("%w: %w", ErrCoolingControlFailure, lastErr)
	}

	slog.DebugContext(ctx, "Cooling output applied",
		"zone", zone.Name,
		"output_percent", outputPercent,
		"devices_updated", successCount,
		"total_devices", len(zone.CoolingDevices))

	return nil
}

// SetCoolingDevicePower sets the power level of a cooling device.
func SetCoolingDevicePower(ctx context.Context, device *CoolingDevice, powerPercent float64) error {
	if device == nil {
		return ErrCoolingDeviceUnavailable
	}

	if powerPercent < 0 || powerPercent > 100 {
		return ErrInvalidCoolingPower
	}

	if device.HwmonPath == "" {
		return ErrHwmonPathInvalid
	}

	// Calculate actual power value based on device range
	powerRange := device.MaxPower - device.MinPower
	actualPower := device.MinPower + (powerRange * powerPercent / 100.0)

	// Round to nearest integer for hwmon
	powerValue := int(actualPower + 0.5)

	err := hwmon.WriteIntCtx(ctx, device.HwmonPath, powerValue)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCoolingControlFailure, err)
	}

	device.CurrentPower = powerPercent

	slog.DebugContext(ctx, "Cooling device power set",
		"device", device.Name,
		"power_percent", powerPercent,
		"actual_value", powerValue,
		"path", device.HwmonPath)

	return nil
}

// GetCoolingDeviceStatus retrieves the current status of a cooling device.
func GetCoolingDeviceStatus(ctx context.Context, device *CoolingDevice) (v1alpha1.CoolingDeviceStatus, error) {
	if device == nil {
		return v1alpha1.CoolingDeviceStatus_COOLING_DEVICE_STATUS_ERROR, ErrCoolingDeviceUnavailable
	}

	if device.HwmonPath == "" {
		return v1alpha1.CoolingDeviceStatus_COOLING_DEVICE_STATUS_ERROR, ErrHwmonPathInvalid
	}

	// Check if the device path exists
	if !hwmon.FileExistsCtx(ctx, device.HwmonPath) {
		device.Status = v1alpha1.CoolingDeviceStatus_COOLING_DEVICE_STATUS_NOT_PRESENT
		return device.Status, nil
	}

	// Try to read current value to verify device is accessible
	_, err := hwmon.ReadIntCtx(ctx, device.HwmonPath)
	if err != nil {
		device.Status = v1alpha1.CoolingDeviceStatus_COOLING_DEVICE_STATUS_ERROR
		return device.Status, nil
	}

	device.Status = v1alpha1.CoolingDeviceStatus_COOLING_DEVICE_STATUS_ENABLED
	return device.Status, nil
}

// CalculateAverageTemperature calculates the average temperature from a slice of readings.
func CalculateAverageTemperature(temperatures []float64) float64 {
	if len(temperatures) == 0 {
		return 0
	}

	sum := 0.0
	for _, temp := range temperatures {
		sum += temp
	}

	return sum / float64(len(temperatures))
}

// FindMaximumTemperature finds the maximum temperature from a slice of readings.
func FindMaximumTemperature(temperatures []float64) float64 {
	if len(temperatures) == 0 {
		return 0
	}

	maxTemp := temperatures[0]
	for _, temp := range temperatures[1:] {
		if temp > maxTemp {
			maxTemp = temp
		}
	}

	return maxTemp
}

// CheckThermalEmergency checks if a thermal zone is in an emergency state.
func CheckThermalEmergency(ctx context.Context, zone *Zone) error {
	if zone == nil {
		return ErrThermalZoneNotFound
	}

	temp, err := ReadZoneTemperature(ctx, zone)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSensorReadFailure, err)
	}

	if temp >= zone.CriticalTemperature {
		slog.ErrorContext(ctx, "Critical temperature exceeded",
			"zone", zone.Name,
			"temperature", temp,
			"critical_threshold", zone.CriticalTemperature)

		// Set maximum cooling immediately
		if err := SetCoolingOutput(ctx, zone, 100.0); err != nil {
			slog.ErrorContext(ctx, "Failed to apply emergency cooling",
				"zone", zone.Name,
				"error", err)
		}

		return ErrCriticalTemperature
	}

	if temp >= zone.WarningTemperature {
		slog.WarnContext(ctx, "Warning temperature exceeded",
			"zone", zone.Name,
			"temperature", temp,
			"warning_threshold", zone.WarningTemperature)
	}

	return nil
}

// ResetPIDController resets the PID controller state for a thermal zone.
func ResetPIDController(ctx context.Context, zone *Zone) error {
	if zone == nil {
		return ErrThermalZoneNotFound
	}

	if zone.pidController == nil {
		return ErrPIDNotInitialized
	}

	zone.pidController.Reset()
	zone.currentOutput = 0.0
	zone.lastUpdate = time.Now()

	slog.InfoContext(ctx, "PID controller reset",
		"zone", zone.Name)

	return nil
}

// CreateCoolingDeviceFromProto creates a CoolingDevice from protobuf definition.
func CreateCoolingDeviceFromProto(proto *v1alpha1.CoolingDevice) *CoolingDevice {
	if proto == nil {
		return nil
	}

	// Extract hwmon path from location or custom attributes
	hwmonPath := ""
	if proto.CustomAttributes != nil {
		hwmonPath = proto.CustomAttributes["hwmon_path"]
	}

	deviceType := v1alpha1.CoolingDeviceType_COOLING_DEVICE_TYPE_UNSPECIFIED
	if proto.Type != nil {
		deviceType = *proto.Type
	}

	minPower := 0.0
	if proto.MinCoolingPowerPercent != nil {
		minPower = *proto.MinCoolingPowerPercent
	}

	maxPower := 100.0
	if proto.MaxCoolingPowerPercent != nil {
		maxPower = *proto.MaxCoolingPowerPercent
	}

	currentPower := 0.0
	if proto.CoolingPowerPercent != nil {
		currentPower = *proto.CoolingPowerPercent
	}

	status := v1alpha1.CoolingDeviceStatus_COOLING_DEVICE_STATUS_UNSPECIFIED
	if proto.Status != nil {
		status = *proto.Status
	}

	return &CoolingDevice{
		Name:         proto.Name,
		Type:         deviceType,
		HwmonPath:    hwmonPath,
		MinPower:     minPower,
		MaxPower:     maxPower,
		CurrentPower: currentPower,
		Status:       status,
	}
}

// CreateThermalZoneFromProto creates a ThermalZone from protobuf definition.
func CreateThermalZoneFromProto(proto *v1alpha1.ThermalZone) *Zone {
	if proto == nil {
		return nil
	}

	zone := &Zone{
		Name:                proto.Name,
		SensorPaths:         make([]string, len(proto.SensorNames)),
		TargetTemperature:   proto.GetTargetTemperature(),
		WarningTemperature:  proto.GetTargetTemperature() + 10.0, // Default warning threshold
		CriticalTemperature: proto.GetTargetTemperature() + 20.0, // Default critical threshold
	}

	// Convert sensor names to paths (this would need to be mapped from sensor registry)
	copy(zone.SensorPaths, proto.SensorNames)

	// Convert PID settings if present
	if pidSettings := proto.GetPidSettings(); pidSettings != nil {
		zone.PIDConfig = PIDConfig{
			Kp:         pidSettings.Kp,
			Ki:         pidSettings.Ki,
			Kd:         pidSettings.Kd,
			SampleTime: time.Duration(pidSettings.SampleTime * float64(time.Second)),
			OutputMin:  pidSettings.GetOutputMin(),
			OutputMax:  pidSettings.GetOutputMax(),
		}
	} else {
		// Default PID configuration
		zone.PIDConfig = PIDConfig{
			Kp:         1.0,
			Ki:         0.1,
			Kd:         0.05,
			SampleTime: time.Second,
			OutputMin:  0.0,
			OutputMax:  100.0,
		}
	}

	return zone
}

// DiscoverCoolingDevices discovers available cooling devices from hwmon.
func DiscoverCoolingDevices(ctx context.Context, hwmonPath string) ([]*CoolingDevice, error) {
	devices, err := hwmon.ListDevicesInPathCtx(ctx, hwmonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list hwmon devices: %w", err)
	}

	var coolingDevices []*CoolingDevice

	for _, devicePath := range devices {
		// Look for PWM outputs (fans, pumps)
		pwmAttrs, err := hwmon.ListAttributesCtx(ctx, devicePath, `pwm\d+`)
		if err != nil {
			continue
		}

		deviceName, _ := hwmon.ReadStringCtx(ctx, filepath.Join(devicePath, "name"))
		if deviceName == "" {
			deviceName = filepath.Base(devicePath)
		}

		for _, pwmAttr := range pwmAttrs {
			pwmPath := filepath.Join(devicePath, pwmAttr)

			// Extract PWM number from attribute name
			pwmNum := strings.TrimPrefix(pwmAttr, "pwm")

			device := &CoolingDevice{
				Name:      fmt.Sprintf("%s_pwm%s", deviceName, pwmNum),
				Type:      v1alpha1.CoolingDeviceType_COOLING_DEVICE_TYPE_FAN,
				HwmonPath: pwmPath,
				MinPower:  0,
				MaxPower:  255, // Standard PWM range
				Status:    v1alpha1.CoolingDeviceStatus_COOLING_DEVICE_STATUS_ENABLED,
			}

			coolingDevices = append(coolingDevices, device)
		}
	}

	slog.InfoContext(ctx, "Discovered cooling devices",
		"count", len(coolingDevices),
		"hwmon_path", hwmonPath)

	return coolingDevices, nil
}
