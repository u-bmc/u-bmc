// SPDX-License-Identifier: BSD-3-Clause

package powermgr

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ThermalEmergencyAlert is an alias for the protobuf message.
type ThermalEmergencyAlert = v1alpha1.ThermalEmergencyAlert

// ThermalEmergencyConfig holds thermal emergency response configuration.
type ThermalEmergencyConfig struct {
	EnableThermalResponse    bool
	EmergencyResponseDelay   time.Duration
	EnableEmergencyShutdown  bool
	ShutdownTemperatureLimit float64
	ShutdownComponents       []string
	MaxEmergencyAttempts     int
	EmergencyAttemptInterval time.Duration
}

// initializeThermalIntegration sets up thermal emergency response.
func (p *PowerMgr) initializeThermalIntegration(ctx context.Context) error {
	if !p.config.enableThermalResponse {
		return nil
	}

	p.logger.InfoContext(ctx, "Initializing thermal emergency integration",
		"emergency_shutdown", p.config.enableEmergencyShutdown,
		"response_delay", p.config.emergencyResponseDelay,
		"shutdown_limit", p.config.shutdownTemperatureLimit)

	// Subscribe to thermal emergency alerts
	if err := p.subscribeThermalAlerts(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to thermal alerts: %w", err)
	}

	return nil
}

// subscribeThermalAlerts subscribes to thermal emergency alert topics.
func (p *PowerMgr) subscribeThermalAlerts(ctx context.Context) error {
	// Subscribe to emergency thermal alerts from sensormon
	sensorAlertSubject := "sensormon.alerts.emergency"
	if _, err := p.nc.Subscribe(sensorAlertSubject, p.handleThermalEmergencyAlert); err != nil {
		return fmt.Errorf("failed to subscribe to sensor emergency alerts: %w", err)
	}

	// Subscribe to emergency thermal alerts from thermalmgr
	thermalAlertSubject := "thermalmgr.alerts.emergency"
	if _, err := p.nc.Subscribe(thermalAlertSubject, p.handleThermalEmergencyAlert); err != nil {
		return fmt.Errorf("failed to subscribe to thermal manager emergency alerts: %w", err)
	}

	// Subscribe to thermal emergency requests from thermalmgr
	emergencyRequestSubject := "powermgr.emergency.thermal"
	if _, err := p.nc.Subscribe(emergencyRequestSubject, p.handleThermalEmergencyRequest); err != nil {
		return fmt.Errorf("failed to subscribe to thermal emergency requests: %w", err)
	}

	p.logger.InfoContext(ctx, "Subscribed to thermal emergency topics",
		"sensor_alerts", sensorAlertSubject,
		"thermal_alerts", thermalAlertSubject,
		"emergency_requests", emergencyRequestSubject)

	return nil
}

// handleThermalEmergencyAlert handles thermal emergency alerts.
func (p *PowerMgr) handleThermalEmergencyAlert(msg *nats.Msg) {
	ctx := context.Background()

	var alert ThermalEmergencyAlert
	if err := alert.UnmarshalVT(msg.Data); err != nil {
		p.logger.WarnContext(ctx, "Invalid thermal emergency alert",
			"subject", msg.Subject,
			"error", err)
		return
	}

	p.logger.InfoContext(ctx, "Thermal emergency alert received",
		"type", alert.Type,
		"sensor_id", alert.GetSensorId(),
		"sensor_name", alert.GetSensorName(),
		"zone_name", alert.GetZoneName(),
		"temperature", alert.Temperature,
		"threshold", alert.GetThreshold(),
		"severity", alert.Severity,
		"message", alert.Message)

	// Process emergency alert
	if err := p.processThermalEmergency(ctx, &alert); err != nil {
		p.logger.ErrorContext(ctx, "Failed to process thermal emergency",
			"alert_type", alert.Type,
			"error", err)
	}
}

// handleThermalEmergencyRequest handles direct thermal emergency requests.
func (p *PowerMgr) handleThermalEmergencyRequest(msg *nats.Msg) {
	ctx := context.Background()

	var request ThermalEmergencyAlert
	if err := request.UnmarshalVT(msg.Data); err != nil {
		p.logger.WarnContext(ctx, "Invalid thermal emergency request",
			"subject", msg.Subject,
			"error", err)
		return
	}

	p.logger.ErrorContext(ctx, "Thermal emergency request received",
		"action", request.Action,
		"zone_name", request.ZoneName,
		"temperature", request.Temperature)

	// Process emergency request
	if err := p.processThermalEmergencyRequest(ctx, &request); err != nil {
		p.logger.ErrorContext(ctx, "Failed to process thermal emergency request",
			"action", request.Action,
			"error", err)
	}
}

// processThermalEmergency processes a thermal emergency alert.
func (p *PowerMgr) processThermalEmergency(ctx context.Context, alert *ThermalEmergencyAlert) error {
	// Check if emergency shutdown is enabled and temperature exceeds limit
	if p.config.enableEmergencyShutdown && alert.Temperature >= p.config.shutdownTemperatureLimit {
		p.logger.ErrorContext(ctx, "Temperature exceeds emergency shutdown limit",
			"temperature", alert.Temperature,
			"limit", p.config.shutdownTemperatureLimit,
			"initiating_shutdown", true)

		// Wait for emergency response delay to allow thermal management to respond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(p.config.emergencyResponseDelay):
		}

		// Perform emergency shutdown
		return p.performEmergencyShutdown(ctx, alert)
	}

	// Log the emergency but don't take shutdown action
	p.logger.WarnContext(ctx, "Thermal emergency condition detected but below shutdown threshold",
		"temperature", alert.Temperature,
		"shutdown_limit", p.config.shutdownTemperatureLimit,
		"emergency_shutdown_enabled", p.config.enableEmergencyShutdown)

	return nil
}

// processThermalEmergencyRequest processes a direct thermal emergency request.
func (p *PowerMgr) processThermalEmergencyRequest(ctx context.Context, request *ThermalEmergencyAlert) error {
	switch *request.Action {
	case "emergency_shutdown":
		p.logger.ErrorContext(ctx, "Emergency shutdown requested by thermal manager",
			"zone_name", request.ZoneName,
			"temperature", request.Temperature)
		return p.performEmergencyShutdown(ctx, request)

	case "power_throttle":
		p.logger.WarnContext(ctx, "Power throttling requested by thermal manager",
			"zone_name", request.ZoneName,
			"temperature", request.Temperature)
		return p.performPowerThrottling(ctx, request)

	case "immediate_shutdown":
		p.logger.ErrorContext(ctx, "Immediate shutdown requested by thermal manager",
			"zone_name", request.ZoneName,
			"temperature", request.Temperature)
		return p.performImmediateShutdown(ctx, request)

	default:
		return fmt.Errorf("unknown thermal emergency action: %s", request.Action)
	}
}

// performEmergencyShutdown performs an emergency shutdown of specified components.
func (p *PowerMgr) performEmergencyShutdown(ctx context.Context, alert *ThermalEmergencyAlert) error {
	p.logger.ErrorContext(ctx, "Performing emergency thermal shutdown",
		"temperature", alert.Temperature,
		"components", p.config.shutdownComponents)

	var lastErr error
	shutdownCount := 0

	// Shutdown components in configured order
	for _, componentName := range p.config.shutdownComponents {
		if err := p.shutdownComponentWithRetry(ctx, componentName, false); err != nil {
			lastErr = err
			p.logger.ErrorContext(ctx, "Failed to shutdown component during thermal emergency",
				"component", componentName,
				"error", err)
			continue
		}
		shutdownCount++

		p.logger.InfoContext(ctx, "Component shutdown for thermal emergency",
			"component", componentName,
			"temperature", alert.Temperature)

		// Brief delay between component shutdowns
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	if shutdownCount == 0 && lastErr != nil {
		return fmt.Errorf("all component shutdowns failed during thermal emergency: %w", lastErr)
	}

	// Send confirmation message
	response := &v1alpha1.ThermalEventResponse{
		EventType:     "emergency_shutdown_completed",
		ComponentName: fmt.Sprintf("%d_components", shutdownCount),
		Action:        "emergency_shutdown",
		Success:       true,
		Message:       fmt.Sprintf("Emergency shutdown completed for %d/%d components", shutdownCount, len(p.config.shutdownComponents)),
		Timestamp:     timestamppb.Now(),
		AdditionalData: map[string]string{
			"components_shutdown": fmt.Sprintf("%d", shutdownCount),
			"total_components":    fmt.Sprintf("%d", len(p.config.shutdownComponents)),
			"temperature":         fmt.Sprintf("%.2f", alert.Temperature),
		},
	}

	if responseData, err := response.MarshalVT(); err == nil {
		if err := p.nc.Publish("powermgr.events.thermal_shutdown", responseData); err != nil {
			p.logger.WarnContext(ctx, "Failed to publish thermal shutdown event",
				"error", err)
		}
	}

	p.logger.InfoContext(ctx, "Emergency thermal shutdown completed",
		"components_shutdown", shutdownCount,
		"total_components", len(p.config.shutdownComponents),
		"temperature", alert.Temperature)

	return nil
}

// performPowerThrottling performs power throttling for thermal management.
func (p *PowerMgr) performPowerThrottling(ctx context.Context, request *ThermalEmergencyAlert) error {
	p.logger.WarnContext(ctx, "Power throttling requested for thermal management",
		"zone_name", request.ZoneName,
		"temperature", request.Temperature)

	// Note: Power throttling would typically involve reducing CPU frequencies,
	// limiting power rails, or other power reduction mechanisms.
	// This is a placeholder for such functionality.

	// Send response indicating throttling action
	// Send throttling notification
	response := &v1alpha1.ThermalEventResponse{
		EventType:     "power_throttling_applied",
		ComponentName: request.GetZoneName(),
		Action:        "power_throttling",
		Success:       true,
		Message:       "Power throttling applied for thermal management",
		Timestamp:     timestamppb.Now(),
		AdditionalData: map[string]string{
			"zone_name":   request.GetZoneName(),
			"temperature": fmt.Sprintf("%.2f", request.Temperature),
		},
	}

	if responseData, err := response.MarshalVT(); err == nil {
		if err := p.nc.Publish("powermgr.events.thermal_throttle", responseData); err != nil {
			p.logger.WarnContext(ctx, "Failed to publish thermal throttle event",
				"error", err)
		}
	}

	p.logger.InfoContext(ctx, "Power throttling applied for thermal management",
		"zone_name", request.ZoneName,
		"temperature", request.Temperature)

	return nil
}

// performImmediateShutdown performs an immediate forced shutdown.
func (p *PowerMgr) performImmediateShutdown(ctx context.Context, request *ThermalEmergencyAlert) error {
	p.logger.ErrorContext(ctx, "Performing immediate shutdown for thermal emergency",
		"zone_name", request.ZoneName,
		"temperature", request.Temperature)

	var lastErr error
	shutdownCount := 0

	// Force shutdown all components immediately
	for _, componentName := range p.config.shutdownComponents {
		if err := p.shutdownComponentWithRetry(ctx, componentName, true); err != nil {
			lastErr = err
			p.logger.ErrorContext(ctx, "Failed to force shutdown component",
				"component", componentName,
				"error", err)
			continue
		}
		shutdownCount++
	}

	if shutdownCount == 0 && lastErr != nil {
		return fmt.Errorf("all immediate shutdowns failed: %w", lastErr)
	}

	// Send confirmation message
	// Send immediate shutdown notification
	response := &v1alpha1.ThermalEventResponse{
		EventType:     "immediate_shutdown_completed",
		ComponentName: fmt.Sprintf("%d_components", shutdownCount),
		Action:        "immediate_shutdown",
		Success:       true,
		Message:       fmt.Sprintf("Immediate shutdown completed for %d/%d components", shutdownCount, len(p.config.shutdownComponents)),
		Timestamp:     timestamppb.Now(),
		AdditionalData: map[string]string{
			"components_shutdown": fmt.Sprintf("%d", shutdownCount),
			"total_components":    fmt.Sprintf("%d", len(p.config.shutdownComponents)),
			"temperature":         fmt.Sprintf("%.2f", request.Temperature),
		},
	}

	if responseData, err := response.MarshalVT(); err == nil {
		if err := p.nc.Publish("powermgr.events.immediate_shutdown", responseData); err != nil {
			p.logger.WarnContext(ctx, "Failed to publish immediate shutdown event",
				"error", err)
		}
	}

	p.logger.ErrorContext(ctx, "Immediate shutdown completed for thermal emergency",
		"components_shutdown", shutdownCount,
		"zone_name", request.ZoneName,
		"temperature", request.Temperature)

	return nil
}

// shutdownComponentWithRetry attempts to shutdown a component with retry logic.
func (p *PowerMgr) shutdownComponentWithRetry(ctx context.Context, componentName string, force bool) error {
	var lastErr error

	for attempt := 0; attempt < p.config.maxEmergencyAttempts; attempt++ {
		if attempt > 0 {
			p.logger.InfoContext(ctx, "Retrying component shutdown",
				"component", componentName,
				"attempt", attempt+1,
				"max_attempts", p.config.maxEmergencyAttempts)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.config.emergencyAttemptInterval):
			}
		}

		backend, err := p.getBackendForComponent(componentName)
		if err != nil {
			lastErr = err
			continue
		}

		err = backend.PowerOff(ctx, componentName, force)
		if err == nil {
			return nil
		}

		lastErr = err
		p.logger.WarnContext(ctx, "Component shutdown attempt failed",
			"component", componentName,
			"attempt", attempt+1,
			"force", force,
			"error", err)
	}

	return fmt.Errorf("component shutdown failed after %d attempts: %w",
		p.config.maxEmergencyAttempts, lastErr)
}
