// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go/micro"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// MonitoringStatus represents the current monitoring status.
type MonitoringStatus struct {
	Active       bool
	StartTime    time.Time
	SensorCount  int
	ReadCount    uint64
	ErrorCount   uint64
	LastReadTime time.Time
	ReadInterval time.Duration
}

// handleStartMonitoring handles requests to start continuous sensor monitoring.
func (s *SensorMon) handleStartMonitoring(ctx context.Context, req micro.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.monitoring {
		s.logger.WarnContext(ctx, "Monitoring already started")
		_ = req.Error("409", "Monitoring already active", nil)
		return
	}

	s.monitoring = true
	s.monitoringStop = make(chan struct{})
	s.monitoringStats.StartTime = time.Now()

	// Start monitoring goroutine
	go s.runMonitoring(ctx)

	// Send a simple success response
	if err := req.Respond([]byte(`{"status":"started","active":true,"message":"Sensor monitoring started successfully"}`)); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send start monitoring response", "error", err)
	}

	s.logger.InfoContext(ctx, "Sensor monitoring started")
}

// handleStopMonitoring handles requests to stop continuous sensor monitoring.
func (s *SensorMon) handleStopMonitoring(ctx context.Context, req micro.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.monitoring {
		s.logger.WarnContext(ctx, "Monitoring not active")
		_ = req.Error("409", "Monitoring not active", nil)
		return
	}

	s.monitoring = false
	close(s.monitoringStop)

	// Send a simple success response
	if err := req.Respond([]byte(`{"status":"stopped","active":false,"message":"Sensor monitoring stopped successfully"}`)); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send stop monitoring response", "error", err)
	}

	s.logger.InfoContext(ctx, "Sensor monitoring stopped")
}

// handleMonitoringStatus handles requests to get the current monitoring status.
func (s *SensorMon) handleMonitoringStatus(ctx context.Context, req micro.Request) {
	s.mu.RLock()
	status := s.getMonitoringStatus()
	s.mu.RUnlock()

	response := fmt.Sprintf(`{
		"status":"%s",
		"active":%t,
		"sensor_count":%d,
		"read_count":%d,
		"error_count":%d,
		"read_interval":"%s",
		"message":"%s"
	}`,
		s.getStatusString(status.Active),
		status.Active,
		status.SensorCount,
		status.ReadCount,
		status.ErrorCount,
		status.ReadInterval.String(),
		s.getStatusMessage(status),
	)

	if err := req.Respond([]byte(response)); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send monitoring status response", "error", err)
	}

	s.logger.DebugContext(ctx, "Retrieved monitoring status",
		"active", status.Active,
		"sensor_count", status.SensorCount,
		"read_count", status.ReadCount,
		"error_count", status.ErrorCount)
}

// runMonitoring runs the continuous sensor monitoring loop.
func (s *SensorMon) runMonitoring(ctx context.Context) {
	s.logger.InfoContext(ctx, "Starting monitoring loop",
		"interval", s.config.monitoringInterval,
		"threshold_check_interval", s.config.thresholdCheckInterval)

	ticker := time.NewTicker(s.config.monitoringInterval)
	defer ticker.Stop()

	var thresholdTicker *time.Ticker
	if s.config.enableThresholdMonitoring {
		thresholdTicker = time.NewTicker(s.config.thresholdCheckInterval)
		defer thresholdTicker.Stop()
	}

	for {
		select {
		case <-s.monitoringStop:
			s.logger.InfoContext(ctx, "Monitoring loop stopped")
			return
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "Monitoring loop canceled", "reason", ctx.Err())
			return
		case <-ticker.C:
			s.performSensorReads(ctx)
		case <-func() <-chan time.Time {
			if thresholdTicker != nil {
				return thresholdTicker.C
			}
			return make(chan time.Time) // Never fires if threshold monitoring disabled
		}():
			s.performThresholdChecks(ctx)
		}
	}
}

// performSensorReads reads all sensors concurrently.
func (s *SensorMon) performSensorReads(ctx context.Context) {
	s.mu.RLock()
	sensors := make([]*sensorInfo, 0, len(s.sensors))
	for _, sensor := range s.sensors {
		sensors = append(sensors, sensor)
	}
	s.mu.RUnlock()

	if len(sensors) == 0 {
		return
	}

	// Use a semaphore to limit concurrent reads
	semaphore := make(chan struct{}, s.config.maxConcurrentReads)
	var wg sync.WaitGroup
	var readCount, errorCount uint64

	for _, sensor := range sensors {
		wg.Add(1)
		go func(sensorInfo *sensorInfo) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			if err := s.readSensorValue(ctx, sensorInfo); err != nil {
				s.logger.WarnContext(ctx, "Failed to read sensor during monitoring",
					"sensor_id", sensorInfo.Sensor.Id,
					"error", err)
				errorCount++
			} else {
				readCount++
			}
		}(sensor)
	}

	wg.Wait()

	// Update monitoring statistics
	s.mu.Lock()
	s.monitoringStats.ReadCount += readCount
	s.monitoringStats.ErrorCount += errorCount
	s.monitoringStats.LastReadTime = time.Now()
	s.mu.Unlock()

	// Broadcast sensor readings if enabled
	if s.config.broadcastSensorReadings {
		s.broadcastSensorReadings(ctx, sensors)
	}

	// Persist sensor data if enabled
	if s.config.persistSensorData {
		s.persistSensorData(ctx, sensors)
	}

	s.logger.DebugContext(ctx, "Completed sensor read cycle",
		"sensors_read", readCount,
		"errors", errorCount,
		"total_sensors", len(sensors))
}

// performThresholdChecks checks all sensors for threshold violations.
func (s *SensorMon) performThresholdChecks(ctx context.Context) {
	s.mu.RLock()
	sensors := make([]*sensorInfo, 0, len(s.sensors))
	for _, sensor := range s.sensors {
		if sensor.Sensor.Reading != nil {
			sensors = append(sensors, sensor)
		}
	}
	s.mu.RUnlock()

	violationCount := 0

	for _, sensorInfo := range sensors {
		if violations := s.checkSensorThresholds(sensorInfo); len(violations) > 0 {
			violationCount++
			s.handleThresholdViolations(ctx, sensorInfo, violations)
		}
	}

	if violationCount > 0 {
		s.logger.WarnContext(ctx, "Threshold violations detected",
			"violation_count", violationCount,
			"total_sensors", len(sensors))
	}
}

// checkSensorThresholds checks if a sensor has violated any thresholds.
func (s *SensorMon) checkSensorThresholds(sensorInfo *sensorInfo) []ThresholdViolation {
	var violations []ThresholdViolation

	analogReading, ok := sensorInfo.Sensor.Reading.(*v1alpha1.Sensor_AnalogReading)
	if !ok || analogReading.AnalogReading == nil {
		return violations
	}

	currentValue := analogReading.AnalogReading.Value

	// Check upper thresholds
	if thresholds := analogReading.AnalogReading.UpperThresholds; thresholds != nil {
		if thresholds.Critical != nil && currentValue > *thresholds.Critical {
			violations = append(violations, ThresholdViolation{
				Type:      "upper_critical",
				Current:   currentValue,
				Threshold: *thresholds.Critical,
				Severity:  "critical",
			})
		} else if thresholds.Warning != nil && currentValue > *thresholds.Warning {
			violations = append(violations, ThresholdViolation{
				Type:      "upper_warning",
				Current:   currentValue,
				Threshold: *thresholds.Warning,
				Severity:  "warning",
			})
		}
	}

	// Check lower thresholds
	if thresholds := analogReading.AnalogReading.LowerThresholds; thresholds != nil {
		if thresholds.Critical != nil && currentValue < *thresholds.Critical {
			violations = append(violations, ThresholdViolation{
				Type:      "lower_critical",
				Current:   currentValue,
				Threshold: *thresholds.Critical,
				Severity:  "critical",
			})
		} else if thresholds.Warning != nil && currentValue < *thresholds.Warning {
			violations = append(violations, ThresholdViolation{
				Type:      "lower_warning",
				Current:   currentValue,
				Threshold: *thresholds.Warning,
				Severity:  "warning",
			})
		}
	}

	return violations
}

// handleThresholdViolations handles detected threshold violations.
func (s *SensorMon) handleThresholdViolations(ctx context.Context, sensorInfo *sensorInfo, violations []ThresholdViolation) {
	for _, violation := range violations {
		s.logger.WarnContext(ctx, "Sensor threshold violation detected",
			"sensor_id", sensorInfo.Sensor.Id,
			"sensor_name", sensorInfo.Sensor.Name,
			"violation_type", violation.Type,
			"current_value", violation.Current,
			"threshold_value", violation.Threshold,
			"severity", violation.Severity)

		// Update sensor status based on violation severity
		if violation.Severity == "critical" {
			status := v1alpha1.SensorStatus_SENSOR_STATUS_CRITICAL
			sensorInfo.Sensor.Status = &status
		} else {
			status := v1alpha1.SensorStatus_SENSOR_STATUS_WARNING
			sensorInfo.Sensor.Status = &status
		}

		// Broadcast threshold violation event if enabled
		if s.config.broadcastSensorReadings {
			s.broadcastThresholdViolation(ctx, sensorInfo, violation)
		}
	}
}

// broadcastSensorReadings broadcasts sensor readings via NATS.
func (s *SensorMon) broadcastSensorReadings(ctx context.Context, sensors []*sensorInfo) {
	for _, sensorInfo := range sensors {
		subject := fmt.Sprintf("sensormon.data.%s", sensorInfo.Sensor.Id)

		data, err := proto.Marshal(sensorInfo.Sensor)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to marshal sensor for broadcast",
				"sensor_id", sensorInfo.Sensor.Id,
				"error", err)
			continue
		}

		if err := s.nc.Publish(subject, data); err != nil {
			s.logger.WarnContext(ctx, "Failed to broadcast sensor reading",
				"sensor_id", sensorInfo.Sensor.Id,
				"subject", subject,
				"error", err)
		}
	}
}

// broadcastThresholdViolation broadcasts a threshold violation event.
func (s *SensorMon) broadcastThresholdViolation(ctx context.Context, sensorInfo *sensorInfo, violation ThresholdViolation) {
	subject := fmt.Sprintf("sensormon.events.threshold.%s", violation.Severity)

	// Create a simple JSON message for threshold violations
	message := fmt.Sprintf(`{
		"sensor_id": "%s",
		"sensor_name": "%s",
		"violation_type": "%s",
		"current_value": %f,
		"threshold_value": %f,
		"severity": "%s",
		"timestamp": "%s"
	}`,
		sensorInfo.Sensor.Id,
		sensorInfo.Sensor.Name,
		violation.Type,
		violation.Current,
		violation.Threshold,
		violation.Severity,
		time.Now().Format(time.RFC3339),
	)

	if err := s.nc.Publish(subject, []byte(message)); err != nil {
		s.logger.WarnContext(ctx, "Failed to broadcast threshold violation",
			"sensor_id", sensorInfo.Sensor.Id,
			"subject", subject,
			"error", err)
	}
}

// persistSensorData persists sensor data to JetStream.
func (s *SensorMon) persistSensorData(ctx context.Context, sensors []*sensorInfo) {
	if s.js == nil {
		return
	}

	for _, sensorInfo := range sensors {
		subject := fmt.Sprintf("sensormon.data.%s", sensorInfo.Sensor.Id)

		data, err := proto.Marshal(sensorInfo.Sensor)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to marshal sensor for persistence",
				"sensor_id", sensorInfo.Sensor.Id,
				"error", err)
			continue
		}

		if _, err := s.js.Publish(ctx, subject, data); err != nil {
			s.logger.WarnContext(ctx, "Failed to persist sensor data",
				"sensor_id", sensorInfo.Sensor.Id,
				"subject", subject,
				"error", err)
		}
	}
}

// getMonitoringStatus returns the current monitoring status.
func (s *SensorMon) getMonitoringStatus() MonitoringStatus {
	return MonitoringStatus{
		Active:       s.monitoring,
		StartTime:    s.monitoringStats.StartTime,
		SensorCount:  len(s.sensors),
		ReadCount:    s.monitoringStats.ReadCount,
		ErrorCount:   s.monitoringStats.ErrorCount,
		LastReadTime: s.monitoringStats.LastReadTime,
		ReadInterval: s.config.monitoringInterval,
	}
}

// getStatusString returns a string representation of the monitoring status.
func (s *SensorMon) getStatusString(active bool) string { //nolint:revive
	if active {
		return "active"
	}
	return "inactive"
}

// getStatusMessage returns a descriptive message about the monitoring status.
func (s *SensorMon) getStatusMessage(status MonitoringStatus) string {
	if status.Active {
		return fmt.Sprintf("Monitoring %d sensors with %d reads and %d errors",
			status.SensorCount, status.ReadCount, status.ErrorCount)
	}
	return "Sensor monitoring is not active"
}

// ThresholdViolation represents a sensor threshold violation.
type ThresholdViolation struct {
	Type      string
	Current   float64
	Threshold float64
	Severity  string
}
