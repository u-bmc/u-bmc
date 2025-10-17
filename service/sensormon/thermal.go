// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ThermalIntegrationConfig holds thermal integration settings.
type ThermalIntegrationConfig struct {
	EnableThermalIntegration  bool
	ThermalMgrEndpoint        string
	TemperatureUpdateInterval time.Duration
	ThermalAlertThresholds    map[string]float64
	EnableThermalAlerts       bool
	CriticalTempThreshold     float64
	WarningTempThreshold      float64
	EmergencyResponseDelay    time.Duration
}

// SensorReading is an alias for the protobuf message.
type SensorReading = v1alpha1.SensorReading

// SensorAlert is an alias for the protobuf message.
type SensorAlert = v1alpha1.SensorAlert

// initializeThermalIntegration sets up thermal management integration.
func (s *SensorMon) initializeThermalIntegration(ctx context.Context) error {
	if !s.config.enableThermalIntegration {
		return nil
	}

	s.logger.InfoContext(ctx, "Initializing thermal integration",
		"endpoint", s.config.thermalMgrEndpoint,
		"update_interval", s.config.temperatureUpdateInterval)

	// Subscribe to thermal management requests
	if err := s.subscribeThermalRequests(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to thermal requests: %w", err)
	}

	// Start thermal monitoring if enabled
	if s.config.enableThermalAlerts {
		go s.runThermalMonitoring(ctx)
	}

	return nil
}

// subscribeThermalRequests subscribes to thermal management request topics.
func (s *SensorMon) subscribeThermalRequests(ctx context.Context) error {
	// Subscribe to temperature data requests from thermalmgr
	tempDataSubject := fmt.Sprintf("%s.sensors.temperature", s.config.thermalMgrEndpoint)
	if _, err := s.nc.Subscribe(tempDataSubject, s.handleTemperatureDataRequest); err != nil {
		return fmt.Errorf("failed to subscribe to temperature data requests: %w", err)
	}

	// Subscribe to sensor configuration requests from thermalmgr
	configSubject := fmt.Sprintf("%s.sensors.config", s.config.thermalMgrEndpoint)
	if _, err := s.nc.Subscribe(configSubject, s.handleSensorConfigRequest); err != nil {
		return fmt.Errorf("failed to subscribe to sensor config requests: %w", err)
	}

	s.logger.InfoContext(ctx, "Subscribed to thermal management topics",
		"temp_data_subject", tempDataSubject,
		"config_subject", configSubject)

	return nil
}

// runThermalMonitoring runs continuous thermal monitoring and alerting.
func (s *SensorMon) runThermalMonitoring(ctx context.Context) {
	s.logger.InfoContext(ctx, "Starting thermal monitoring",
		"critical_threshold", s.config.criticalTempThreshold,
		"warning_threshold", s.config.warningTempThreshold)

	ticker := time.NewTicker(s.config.temperatureUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "Thermal monitoring stopped", "reason", ctx.Err())
			return
		case <-ticker.C:
			s.performThermalChecks(ctx)
		}
	}
}

// performThermalChecks performs thermal threshold checks and sends alerts.
func (s *SensorMon) performThermalChecks(ctx context.Context) {
	s.mu.RLock()
	tempSensors := s.getTemperatureSensors()
	s.mu.RUnlock()

	if len(tempSensors) == 0 {
		return
	}

	criticalCount := 0
	warningCount := 0

	for _, sensorInfo := range tempSensors {
		temp, err := s.getTemperatureValue(sensorInfo)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to get temperature value for thermal check",
				"sensor_id", sensorInfo.Sensor.Id,
				"error", err)
			continue
		}

		// Check critical threshold
		if temp >= s.config.criticalTempThreshold {
			criticalCount++
			s.sendThermalAlert(ctx, sensorInfo, temp, s.config.criticalTempThreshold, "critical")
		} else if temp >= s.config.warningTempThreshold {
			warningCount++
			s.sendThermalAlert(ctx, sensorInfo, temp, s.config.warningTempThreshold, "warning")
		}

		// Send temperature update to thermal manager
		s.sendTemperatureUpdate(ctx, sensorInfo, temp)
	}

	if criticalCount > 0 || warningCount > 0 {
		s.logger.WarnContext(ctx, "Thermal threshold violations detected",
			"critical_count", criticalCount,
			"warning_count", warningCount,
			"total_temp_sensors", len(tempSensors))
	}
}

// getTemperatureSensors returns all temperature sensors.
func (s *SensorMon) getTemperatureSensors() []*sensorInfo {
	var tempSensors []*sensorInfo

	for _, sensor := range s.sensors {
		if s.isTemperatureSensor(sensor) {
			tempSensors = append(tempSensors, sensor)
		}
	}

	return tempSensors
}

// isTemperatureSensor checks if a sensor is a temperature sensor.
func (s *SensorMon) isTemperatureSensor(sensorInfo *sensorInfo) bool {
	if sensorInfo.Sensor.Context == nil {
		return false
	}
	return *sensorInfo.Sensor.Context == v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE
}

// getTemperatureValue extracts temperature value from sensor reading.
func (s *SensorMon) getTemperatureValue(sensorInfo *sensorInfo) (float64, error) {
	analogReading, ok := sensorInfo.Sensor.Reading.(*v1alpha1.Sensor_AnalogReading)
	if !ok || analogReading.AnalogReading == nil {
		return 0, fmt.Errorf("sensor does not have analog reading")
	}

	return analogReading.AnalogReading.Value, nil
}

// sendThermalAlert sends a thermal alert to the thermal manager.
func (s *SensorMon) sendThermalAlert(ctx context.Context, sensorInfo *sensorInfo, temperature, threshold float64, severity string) {
	alert := &v1alpha1.SensorAlert{
		Type:       "temperature_threshold",
		SensorId:   sensorInfo.Sensor.Id,
		SensorName: sensorInfo.Sensor.Name,
		Value:      temperature,
		Threshold:  &threshold,
		Severity:   severity,
		Timestamp:  timestamppb.Now(),
		Message:    fmt.Sprintf("Temperature %.1f°C exceeds %s threshold %.1f°C", temperature, severity, threshold),
	}

	// Add zone information if available
	if sensorInfo.Sensor.CustomAttributes != nil {
		// Add thermal zone if configured
		if s.config.enableThermalIntegration {
			if zoneName, exists := sensorInfo.Sensor.CustomAttributes["thermal_zone"]; exists {
				alert.ZoneName = &zoneName
			}
		}
	}

	alertData, err := alert.MarshalVT()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal thermal alert",
			"error", err)
		return
	}

	// Send alert to thermal manager
	subject := fmt.Sprintf("thermalmgr.alerts.%s", severity)
	if err := s.nc.Publish(subject, alertData); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send thermal alert",
			"sensor_id", sensorInfo.Sensor.Id,
			"subject", subject,
			"error", err)
		return
	}

	s.logger.WarnContext(ctx, "Thermal alert sent",
		"sensor_id", sensorInfo.Sensor.Id,
		"sensor_name", sensorInfo.Sensor.Name,
		"temperature", temperature,
		"threshold", threshold,
		"severity", severity)

	// For critical alerts, also send emergency notification with delay
	if severity == "critical" {
		go s.sendDelayedEmergencyNotification(ctx, sensorInfo, temperature)
	}
}

// sendTemperatureUpdate sends a temperature update to the thermal manager.
func (s *SensorMon) sendTemperatureUpdate(ctx context.Context, sensorInfo *sensorInfo, temperature float64) {
	reading := &v1alpha1.SensorReading{
		SensorId:   sensorInfo.Sensor.Id,
		SensorName: sensorInfo.Sensor.Name,
		Value:      temperature,
		Unit:       "celsius",
		Timestamp:  timestamppb.Now(),
	}

	// Add location if available
	if sensorInfo.Sensor.Location != nil {
		if sensorInfo.Sensor.Location.ComponentLocation.Position != nil {
			reading.Location = sensorInfo.Sensor.Location.ComponentLocation.Position
		}
	}

	// Add zone information if available
	if sensorInfo.Sensor.CustomAttributes != nil {
		// Add location and zone info if available
		if s.config.enableThermalIntegration {
			if zoneName, exists := sensorInfo.Sensor.CustomAttributes["thermal_zone"]; exists {
				reading.ZoneName = &zoneName
			}
		}
	}

	readingData, err := reading.MarshalVT()
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to marshal temperature reading",
			"error", err)
		return
	}

	// Send reading to thermal manager
	subject := fmt.Sprintf("thermalmgr.data.temperature.%s", sensorInfo.Sensor.Id)
	if err := s.nc.Publish(subject, readingData); err != nil {
		s.logger.WarnContext(ctx, "Failed to send temperature update",
			"sensor_id", sensorInfo.Sensor.Id,
			"subject", subject,
			"error", err)
	}
}

// sendDelayedEmergencyNotification sends an emergency notification after a delay if condition persists.
func (s *SensorMon) sendDelayedEmergencyNotification(ctx context.Context, sensorInfo *sensorInfo, initialTemp float64) {
	// Wait for emergency response delay
	select {
	case <-ctx.Done():
		return
	case <-time.After(s.config.emergencyResponseDelay):
	}

	// Re-check temperature to see if condition persists
	currentTemp, err := s.getTemperatureValue(sensorInfo)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to re-check temperature for emergency notification",
			"sensor_id", sensorInfo.Sensor.Id,
			"error", err)
		return
	}

	// If temperature is still critical, send emergency notification
	if currentTemp >= s.config.criticalTempThreshold {
		emergencyAlert := &v1alpha1.SensorAlert{
			Type:       "emergency_thermal",
			SensorId:   sensorInfo.Sensor.Id,
			SensorName: sensorInfo.Sensor.Name,
			Value:      currentTemp,
			Threshold:  &s.config.criticalTempThreshold,
			Severity:   "emergency",
			Timestamp:  timestamppb.Now(),
			Message:    fmt.Sprintf("Emergency: Critical temperature %.1f°C persists after %v delay", currentTemp, s.config.emergencyResponseDelay),
		}

		alertData, err := emergencyAlert.MarshalVT()
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to marshal emergency thermal alert",
				"sensor_id", sensorInfo.Sensor.Id,
				"error", err)
			return
		}

		// Send to both thermal manager and power manager
		subjects := []string{
			"thermalmgr.alerts.emergency",
			"powermgr.alerts.thermal_emergency",
		}

		for _, subject := range subjects {
			if err := s.nc.Publish(subject, alertData); err != nil {
				s.logger.ErrorContext(ctx, "Failed to send emergency thermal alert",
					"sensor_id", sensorInfo.Sensor.Id,
					"subject", subject,
					"error", err)
			}
		}

		s.logger.ErrorContext(ctx, "Emergency thermal alert sent",
			"sensor_id", sensorInfo.Sensor.Id,
			"sensor_name", sensorInfo.Sensor.Name,
			"temperature", currentTemp,
			"initial_temp", initialTemp,
			"threshold", s.config.criticalTempThreshold)
	}
}

// handleTemperatureDataRequest handles requests for temperature data from thermal manager.
func (s *SensorMon) handleTemperatureDataRequest(msg *nats.Msg) {
	ctx := context.Background()

	var request v1alpha1.SensorDataRequest
	if err := request.UnmarshalVT(msg.Data); err != nil {
		s.logger.WarnContext(ctx, "Invalid temperature data request", "error", err)
		return
	}

	s.mu.RLock()
	var readings []*v1alpha1.SensorReading

	if len(request.SensorIds) > 0 {
		// Return specific sensors
		for _, sensorID := range request.SensorIds {
			if sensorInfo, exists := s.sensors[sensorID]; exists && s.isTemperatureSensor(sensorInfo) {
				if temp, err := s.getTemperatureValue(sensorInfo); err == nil {
					readings = append(readings, &v1alpha1.SensorReading{
						SensorId:   sensorInfo.Sensor.Id,
						SensorName: sensorInfo.Sensor.Name,
						Value:      temp,
						Unit:       "celsius",
						Timestamp:  timestamppb.Now(),
					})
				}
			}
		}
	} else {
		// Return all temperature sensors
		for _, sensorInfo := range s.sensors {
			if s.isTemperatureSensor(sensorInfo) {
				if temp, err := s.getTemperatureValue(sensorInfo); err == nil {
					readings = append(readings, &v1alpha1.SensorReading{
						SensorId:   sensorInfo.Sensor.Id,
						SensorName: sensorInfo.Sensor.Name,
						Value:      temp,
						Unit:       "celsius",
						Timestamp:  timestamppb.Now(),
					})
				}
			}
		}
	}
	s.mu.RUnlock()

	response := &v1alpha1.SensorDataResponse{
		Readings: readings,
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal temperature data response", "error", err)
		return
	}

	// Send response
	subject := "sensormon.response.temperature_data"
	if err := s.nc.Publish(subject, responseData); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send temperature data response",
			"subject", subject,
			"error", err)
	}

	s.logger.DebugContext(ctx, "Sent temperature data response",
		"readings_count", len(readings),
		"requested_sensors", len(request.SensorIds))
}

// handleSensorConfigRequest handles sensor configuration requests from thermal manager.
func (s *SensorMon) handleSensorConfigRequest(msg *nats.Msg) {
	ctx := context.Background()

	var request v1alpha1.SensorConfigRequest
	if err := request.UnmarshalVT(msg.Data); err != nil {
		s.logger.WarnContext(ctx, "Invalid sensor config request", "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sensorInfo, exists := s.sensors[request.SensorId]
	if !exists {
		s.logger.WarnContext(ctx, "Sensor not found for config request",
			"sensor_id", request.SensorId)
		return
	}

	switch request.Action {
	case "assign_zone":
		// Assign sensor to thermal zone
		if sensorInfo.Sensor.CustomAttributes == nil {
			sensorInfo.Sensor.CustomAttributes = make(map[string]string)
		}
		sensorInfo.Sensor.CustomAttributes["thermal_zone"] = request.GetZoneName()

		s.logger.InfoContext(ctx, "Assigned sensor to thermal zone",
			"sensor_id", request.SensorId,
			"zone_name", request.GetZoneName())

	case "update_attributes":
		if sensorInfo.Sensor.CustomAttributes == nil {
			sensorInfo.Sensor.CustomAttributes = make(map[string]string)
		}
		for key, value := range request.Attributes {
			sensorInfo.Sensor.CustomAttributes[key] = value
		}

		s.logger.InfoContext(ctx, "Updated sensor attributes",
			"sensor_id", request.SensorId,
			"attributes", request.Attributes)

	default:
		s.logger.WarnContext(ctx, "Unknown sensor configuration action",
			"action", request.Action,
			"sensor_id", request.SensorId)
	}
}
