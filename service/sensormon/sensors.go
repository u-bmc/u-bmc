// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/nats-io/nats.go/micro"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/hwmon"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleListSensors handles requests to list all available sensors.
func (s *SensorMon) handleListSensors(ctx context.Context, req micro.Request) {
	var request v1alpha1.ListSensorsRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		s.logger.WarnContext(ctx, "Failed to unmarshal list sensors request", "error", err)
		_ = req.Error("400", "Invalid request format", nil)
		return
	}

	s.mu.RLock()
	sensors := make([]*v1alpha1.Sensor, 0, len(s.sensors))
	for _, sensorInfo := range s.sensors {
		sensor := proto.Clone(sensorInfo.Sensor).(*v1alpha1.Sensor)

		// Apply field mask if provided
		if request.FieldMask != nil {
			sensor = s.applySensorFieldMask(sensor, request.FieldMask)
		}

		sensors = append(sensors, sensor)
	}
	s.mu.RUnlock()

	response := &v1alpha1.ListSensorsResponse{
		Sensor: sensors,
	}

	data, err := response.MarshalVT()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal list sensors response", "error", err)
		_ = req.Error("500", "Internal server error", nil)
		return
	}

	if err := req.Respond(data); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send list sensors response", "error", err)
	}

	s.logger.DebugContext(ctx, "Listed sensors", "count", len(sensors))
}

// handleGetSensor handles requests to get specific sensor information.
func (s *SensorMon) handleGetSensor(ctx context.Context, req micro.Request) {
	var request v1alpha1.GetSensorRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		s.logger.WarnContext(ctx, "Failed to unmarshal get sensor request", "error", err)
		_ = req.Error("400", "Invalid request format", nil)
		return
	}

	// Find sensor by identifier
	var sensorInfo *sensorInfo
	var found bool

	switch identifier := request.Identifier.(type) {
	case *v1alpha1.GetSensorRequest_Id:
		sensorInfo, found = s.getSensor(identifier.Id)
	case *v1alpha1.GetSensorRequest_Name:
		sensorInfo, found = s.getSensorByName(identifier.Name)
	case *v1alpha1.GetSensorRequest_Context:
		sensorInfo, found = s.getSensorByContext(identifier.Context)
	case *v1alpha1.GetSensorRequest_Status:
		sensorInfo, found = s.getSensorByStatus(identifier.Status)
	case *v1alpha1.GetSensorRequest_Location:
		sensorInfo, found = s.getSensorByLocation(identifier.Location)
	default:
		s.logger.WarnContext(ctx, "Invalid sensor identifier in get request")
		_ = req.Error("400", "Invalid sensor identifier", nil)
		return
	}

	if !found {
		s.logger.WarnContext(ctx, "Sensor not found", "request", request.String())
		_ = req.Error("404", "Sensor not found", nil)
		return
	}

	// Read current sensor value
	if err := s.readSensorValue(ctx, sensorInfo); err != nil {
		s.logger.WarnContext(ctx, "Failed to read sensor value",
			"sensor_id", sensorInfo.Sensor.Id,
			"error", err)
		// Continue with cached value
	}

	sensor := proto.Clone(sensorInfo.Sensor).(*v1alpha1.Sensor)

	// Apply field mask if provided
	if request.FieldMask != nil {
		sensor = s.applySensorFieldMask(sensor, request.FieldMask)
	}

	response := &v1alpha1.GetSensorResponse{
		Sensors: []*v1alpha1.Sensor{sensor},
	}

	data, err := response.MarshalVT()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal get sensor response", "error", err)
		_ = req.Error("500", "Internal server error", nil)
		return
	}

	if err := req.Respond(data); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send get sensor response", "error", err)
	}

	s.logger.DebugContext(ctx, "Retrieved sensor",
		"sensor_id", sensorInfo.Sensor.Id,
		"sensor_name", sensorInfo.Sensor.Name)
}

// readSensorValue reads the current value from a sensor and updates the sensor information.
func (s *SensorMon) readSensorValue(ctx context.Context, sensorInfo *sensorInfo) error {
	ctx, cancel := context.WithTimeout(ctx, s.config.sensorTimeout)
	defer cancel()

	switch sensorInfo.Type {
	case sensorTypeHwmon:
		return s.readHwmonSensorValue(ctx, sensorInfo)
	case sensorTypeGPIO:
		return s.readGPIOSensorValue(ctx, sensorInfo)
	case sensorTypeMock:
		return s.readMockSensorValue(ctx, sensorInfo)
	default:
		return fmt.Errorf("%w: unknown sensor type", ErrSensorTypeUnsupported)
	}
}

// readHwmonSensorValue reads a value from an hwmon sensor.
func (s *SensorMon) readHwmonSensorValue(ctx context.Context, sensorInfo *sensorInfo) error {
	rawValue, err := hwmon.ReadIntCtx(ctx, sensorInfo.Path)
	if err != nil {
		status := v1alpha1.SensorStatus_SENSOR_STATUS_ERROR
		sensorInfo.Sensor.Status = &status
		return fmt.Errorf("%w: %w", ErrSensorReadFailed, err)
	}

	// Convert raw value based on sensor sensorContext
	var sensorContext v1alpha1.SensorContext
	if sensorInfo.Sensor.Context != nil {
		sensorContext = *sensorInfo.Sensor.Context
	}
	value := s.convertHwmonValue(rawValue, sensorContext)

	// Update analog reading
	analogReading := &v1alpha1.AnalogSensorReading{
		Value: value,
	}

	// Read thresholds if available
	s.readHwmonThresholds(ctx, sensorInfo, analogReading)

	sensorInfo.Sensor.Reading = &v1alpha1.Sensor_AnalogReading{
		AnalogReading: analogReading,
	}
	sensorInfo.Sensor.LastReadingTimestamp = timestamppb.Now()
	status := v1alpha1.SensorStatus_SENSOR_STATUS_ENABLED
	sensorInfo.Sensor.Status = &status
	sensorInfo.LastRead = time.Now()
	sensorInfo.LastValue = value

	return nil
}

// readGPIOSensorValue reads a value from a GPIO sensor.
func (s *SensorMon) readGPIOSensorValue(ctx context.Context, sensorInfo *sensorInfo) error {
	// GPIO sensor reading would be implemented based on GPIO configuration
	// This is a placeholder for GPIO-based sensors

	stateDesc := "GPIO sensor reading not implemented"
	discreteReading := &v1alpha1.DiscreteSensorReading{
		State:            "enabled",
		StateDescription: &stateDesc,
	}

	sensorInfo.Sensor.Reading = &v1alpha1.Sensor_DiscreteReading{
		DiscreteReading: discreteReading,
	}
	sensorInfo.Sensor.LastReadingTimestamp = timestamppb.Now()
	sensorInfo.LastRead = time.Now()

	return nil
}

// convertHwmonValue converts raw hwmon values to standard units.
func (s *SensorMon) convertHwmonValue(rawValue int, sensorContext v1alpha1.SensorContext) float64 {
	switch sensorContext {
	case v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE:
		// Temperature values are in millidegrees Celsius
		return float64(rawValue) / 1000.0
	case v1alpha1.SensorContext_SENSOR_CONTEXT_VOLTAGE:
		// Voltage values are in millivolts
		return float64(rawValue) / 1000.0
	case v1alpha1.SensorContext_SENSOR_CONTEXT_CURRENT:
		// Current values are in milliamps
		return float64(rawValue) / 1000.0
	case v1alpha1.SensorContext_SENSOR_CONTEXT_POWER:
		// Power values are in microwatts
		return float64(rawValue) / 1000000.0
	case v1alpha1.SensorContext_SENSOR_CONTEXT_TACH:
		// Fan RPM values are direct
		return float64(rawValue)
	default:
		// Return raw value for unknown contexts
		return float64(rawValue)
	}
}

// readHwmonThresholds reads threshold values for an hwmon sensor.
func (s *SensorMon) readHwmonThresholds(ctx context.Context, sensorInfo *sensorInfo, analogReading *v1alpha1.AnalogSensorReading) {
	basePath := filepath.Dir(sensorInfo.Path)
	baseName := filepath.Base(sensorInfo.Path)

	// Remove _input suffix to get base sensor name
	baseName = strings.TrimSuffix(baseName, "_input")

	var sensorContext v1alpha1.SensorContext
	if sensorInfo.Sensor.Context != nil {
		sensorContext = *sensorInfo.Sensor.Context
	}

	// Read upper thresholds
	upperThresholds := &v1alpha1.Threshold{}
	hasUpperThresholds := false

	// Try to read max threshold
	maxPath := filepath.Join(basePath, baseName+"_max")
	if hwmon.FileExistsCtx(ctx, maxPath) {
		if rawValue, err := hwmon.ReadIntCtx(ctx, maxPath); err == nil {
			value := s.convertHwmonValue(rawValue, sensorContext)
			upperThresholds.Warning = &value
			hasUpperThresholds = true
		}
	}

	// Try to read critical threshold
	critPath := filepath.Join(basePath, baseName+"_crit")
	if hwmon.FileExistsCtx(ctx, critPath) {
		if rawValue, err := hwmon.ReadIntCtx(ctx, critPath); err == nil {
			value := s.convertHwmonValue(rawValue, sensorContext)
			upperThresholds.Critical = &value
			hasUpperThresholds = true
		}
	}

	if hasUpperThresholds {
		analogReading.UpperThresholds = upperThresholds
	}

	// Read lower thresholds
	lowerThresholds := &v1alpha1.Threshold{}
	hasLowerThresholds := false

	// Try to read min threshold
	minPath := filepath.Join(basePath, baseName+"_min")
	if hwmon.FileExistsCtx(ctx, minPath) {
		if rawValue, err := hwmon.ReadIntCtx(ctx, minPath); err == nil {
			value := s.convertHwmonValue(rawValue, sensorContext)
			lowerThresholds.Warning = &value
			hasLowerThresholds = true
		}
	}

	if hasLowerThresholds {
		analogReading.LowerThresholds = lowerThresholds
	}
}

// getSensorByContext finds a sensor by its context type.
func (s *SensorMon) getSensorByContext(sensorContext v1alpha1.SensorContext) (*sensorInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sensor := range s.sensors {
		if sensor.Sensor.Context != nil && *sensor.Sensor.Context == sensorContext {
			return sensor, true
		}
	}
	return nil, false
}

// getSensorByStatus finds a sensor by its status.
func (s *SensorMon) getSensorByStatus(status v1alpha1.SensorStatus) (*sensorInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sensor := range s.sensors {
		if sensor.Sensor.Status != nil && *sensor.Sensor.Status == status {
			return sensor, true
		}
	}
	return nil, false
}

// getSensorByLocation finds a sensor by its location.
func (s *SensorMon) getSensorByLocation(location *v1alpha1.Location) (*sensorInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if location == nil {
		return nil, false
	}

	// For now, just match on component location name
	var searchString string
	if location.ComponentLocation != nil {
		searchString = location.ComponentLocation.Name
	}

	if searchString == "" {
		return nil, false
	}

	for _, sensor := range s.sensors {
		if sensor.Sensor.Location != nil &&
			sensor.Sensor.Location.ComponentLocation != nil &&
			strings.Contains(sensor.Sensor.Location.ComponentLocation.Name, searchString) {
			return sensor, true
		}
	}
	return nil, false
}

// readMockSensorValue generates mock sensor values using the mock backend.
func (s *SensorMon) readMockSensorValue(ctx context.Context, sensorInfo *sensorInfo) error {
	// Find the sensor definition to get mock configuration
	var mockConfig *MockSensorConfig
	for _, definition := range s.config.sensorDefinitions {
		if definition.ID == sensorInfo.Sensor.Id && definition.MockConfig != nil {
			mockConfig = definition.MockConfig
			break
		}
	}

	if mockConfig == nil {
		return fmt.Errorf("no mock configuration found for sensor %s", sensorInfo.Sensor.Id)
	}

	// Create mock backend and read value
	backend, err := NewMockBackend(mockConfig)
	if err != nil {
		return fmt.Errorf("failed to create mock backend for sensor %s: %w", sensorInfo.Sensor.Id, err)
	}

	value, err := backend.ReadValue()
	if err != nil {
		status := v1alpha1.SensorStatus_SENSOR_STATUS_ERROR
		sensorInfo.Sensor.Status = &status
		return fmt.Errorf("%w: %w", ErrSensorReadFailed, err)
	}

	// Convert value to float64
	var floatValue float64
	switch v := value.(type) {
	case float64:
		floatValue = v
	case int:
		floatValue = float64(v)
	case int64:
		floatValue = float64(v)
	default:
		return fmt.Errorf("unsupported mock value type: %T", value)
	}

	// Update analog reading
	analogReading := &v1alpha1.AnalogSensorReading{
		Value: floatValue,
	}

	// Copy existing thresholds if they exist
	if sensorInfo.Sensor.Reading != nil {
		if existing, ok := sensorInfo.Sensor.Reading.(*v1alpha1.Sensor_AnalogReading); ok && existing.AnalogReading != nil {
			analogReading.UpperThresholds = existing.AnalogReading.UpperThresholds
			analogReading.LowerThresholds = existing.AnalogReading.LowerThresholds
		}
	}

	sensorInfo.Sensor.Reading = &v1alpha1.Sensor_AnalogReading{
		AnalogReading: analogReading,
	}
	sensorInfo.Sensor.LastReadingTimestamp = timestamppb.Now()
	status := v1alpha1.SensorStatus_SENSOR_STATUS_ENABLED
	sensorInfo.Sensor.Status = &status
	sensorInfo.LastRead = time.Now()
	sensorInfo.LastValue = floatValue

	return nil
}

// applySensorFieldMask applies a field mask to a sensor to return only requested fields.
func (s *SensorMon) applySensorFieldMask(sensor *v1alpha1.Sensor, fieldMask *fieldmaskpb.FieldMask) *v1alpha1.Sensor {
	// Field mask implementation would filter the sensor fields based on the mask
	// For now, return the complete sensor
	return sensor
}
