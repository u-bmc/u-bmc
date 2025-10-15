// SPDX-License-Identifier: BSD-3-Clause

package thermalmgr

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/micro"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/thermal"
)

// ThermalControlStatusResponse represents the thermal control status.
type ThermalControlStatusResponse struct {
	Running            bool     `json:"running"`
	ActiveZones        []string `json:"active_zones"`
	EmergencyCondition bool     `json:"emergency_condition"`
	LastUpdate         string   `json:"last_update"`
}

// handleListThermalZones handles requests to list all thermal zones.
func (t *ThermalMgr) handleListThermalZones(ctx context.Context, req micro.Request) {
	t.mu.RLock()
	zones := make([]*v1alpha1.ThermalZone, 0, len(t.thermalZones))
	for _, zone := range t.thermalZones {
		protoZone := t.convertThermalZoneToProto(zone)
		zones = append(zones, protoZone)
	}
	t.mu.RUnlock()

	response := &v1alpha1.ListThermalZonesResponse{
		ThermalZones: zones,
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal thermal zones list response",
			"error", err)
		_ = req.Error("500", "failed to marshal response", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send thermal zones list response",
			"error", err)
	}

	t.logger.DebugContext(ctx, "Listed thermal zones",
		"count", len(zones))
}

// handleGetThermalZone handles requests to get a specific thermal zone.
func (t *ThermalMgr) handleGetThermalZone(ctx context.Context, req micro.Request) {
	var request v1alpha1.GetThermalZoneRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		t.logger.WarnContext(ctx, "Invalid get thermal zone request",
			"error", err)
		_ = req.Error("400", "invalid request format", nil)
		return
	}

	var zoneName string
	switch id := request.Identifier.(type) {
	case *v1alpha1.GetThermalZoneRequest_Name:
		zoneName = id.Name
	default:
		_ = req.Error("400", "unsupported identifier type", nil)
		return
	}

	zone, exists := t.getThermalZone(zoneName)
	if !exists {
		_ = req.Error("404", fmt.Sprintf("thermal zone not found: %s", zoneName), nil)
		return
	}

	protoZone := t.convertThermalZoneToProto(zone)
	response := &v1alpha1.GetThermalZoneResponse{
		ThermalZones: []*v1alpha1.ThermalZone{protoZone},
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal thermal zone response",
			"zone", zoneName,
			"error", err)
		_ = req.Error("500", "failed to marshal response", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send thermal zone response",
			"zone", zoneName,
			"error", err)
	}

	t.logger.DebugContext(ctx, "Retrieved thermal zone",
		"zone", zoneName)
}

// handleSetThermalZone handles requests to update a thermal zone.
func (t *ThermalMgr) handleSetThermalZone(ctx context.Context, req micro.Request) {
	var request v1alpha1.SetThermalZoneRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		t.logger.WarnContext(ctx, "Invalid set thermal zone request",
			"error", err)
		_ = req.Error("400", "invalid request format", nil)
		return
	}

	zoneName := request.Name
	if zoneName == "" {
		_ = req.Error("400", "zone name required", nil)
		return
	}

	zone, exists := t.getThermalZone(zoneName)
	if !exists {
		_ = req.Error("404", fmt.Sprintf("thermal zone not found: %s", zoneName), nil)
		return
	}

	// Update zone target temperature if provided
	if request.TargetTemperature != nil {
		zone.TargetTemperature = *request.TargetTemperature
		t.logger.InfoContext(ctx, "Updated thermal zone target temperature",
			"zone", zoneName,
			"target_temp", *request.TargetTemperature)
	}

	// Update PID settings if provided
	if request.PidSettings != nil {
		zone.PIDConfig = thermal.PIDConfig{
			Kp:         request.PidSettings.Kp,
			Ki:         request.PidSettings.Ki,
			Kd:         request.PidSettings.Kd,
			SampleTime: t.config.defaultPIDSampleTime,
			OutputMin:  request.PidSettings.GetOutputMin(),
			OutputMax:  request.PidSettings.GetOutputMax(),
		}

		// Reinitialize PID controller with new settings
		if err := thermal.InitializeThermalZone(ctx, zone); err != nil {
			t.logger.ErrorContext(ctx, "Failed to reinitialize thermal zone",
				"zone", zoneName,
				"error", err)
			_ = req.Error("500", "failed to reinitialize thermal zone", nil)
			return
		}

		t.logger.InfoContext(ctx, "Updated thermal zone PID settings",
			"zone", zoneName,
			"kp", request.PidSettings.Kp,
			"ki", request.PidSettings.Ki,
			"kd", request.PidSettings.Kd)
	}

	protoZone := t.convertThermalZoneToProto(zone)
	response := &v1alpha1.SetThermalZoneResponse{
		ThermalZone: protoZone,
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal set thermal zone response",
			"zone", zoneName,
			"error", err)
		_ = req.Error("500", "failed to marshal response", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send set thermal zone response",
			"zone", zoneName,
			"error", err)
	}

	t.logger.DebugContext(ctx, "Updated thermal zone",
		"zone", zoneName)
}

// handleListCoolingDevices handles requests to list all cooling devices.
func (t *ThermalMgr) handleListCoolingDevices(ctx context.Context, req micro.Request) {
	t.mu.RLock()
	devices := make([]*v1alpha1.CoolingDevice, 0, len(t.coolingDevices))
	for _, device := range t.coolingDevices {
		protoDevice := t.convertCoolingDeviceToProto(device)
		devices = append(devices, protoDevice)
	}
	t.mu.RUnlock()

	response := &v1alpha1.ListCoolingDevicesResponse{
		CoolingDevices: devices,
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal cooling devices list response",
			"error", err)
		_ = req.Error("500", "failed to marshal response", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send cooling devices list response",
			"error", err)
	}

	t.logger.DebugContext(ctx, "Listed cooling devices",
		"count", len(devices))
}

// handleGetCoolingDevice handles requests to get a specific cooling device.
func (t *ThermalMgr) handleGetCoolingDevice(ctx context.Context, req micro.Request) {
	var request v1alpha1.GetCoolingDeviceRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		t.logger.WarnContext(ctx, "Invalid get cooling device request",
			"error", err)
		_ = req.Error("400", "invalid request format", nil)
		return
	}

	if request.Name == "" {
		_ = req.Error("400", "device name required", nil)
		return
	}

	device, exists := t.getCoolingDevice(request.Name)
	if !exists {
		_ = req.Error("404", fmt.Sprintf("cooling device not found: %s", request.Name), nil)
		return
	}

	response := &v1alpha1.GetCoolingDeviceResponse{
		CoolingDevice: t.convertCoolingDeviceToProto(device),
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal cooling device response",
			"device", request.Name,
			"error", err)
		_ = req.Error("500", "internal server error", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send cooling device response",
			"device", request.Name,
			"error", err)
	}

	t.logger.DebugContext(ctx, "Retrieved cooling device",
		"device", request.Name)
}

// handleSetCoolingDevice handles requests to update a cooling device.
func (t *ThermalMgr) handleSetCoolingDevice(ctx context.Context, req micro.Request) {
	var request v1alpha1.SetCoolingDeviceRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		t.logger.WarnContext(ctx, "Invalid set cooling device request",
			"error", err)
		_ = req.Error("400", "invalid request format", nil)
		return
	}

	if request.Name == "" {
		_ = req.Error("400", "device name required", nil)
		return
	}

	device, exists := t.getCoolingDevice(request.Name)
	if !exists {
		_ = req.Error("404", fmt.Sprintf("cooling device not found: %s", request.Name), nil)
		return
	}

	// Update device power if provided
	// Update power percentage if provided
	if request.PowerPercent != nil {
		if err := thermal.SetCoolingDevicePower(ctx, device, *request.PowerPercent); err != nil {
			t.logger.ErrorContext(ctx, "Failed to set cooling device power",
				"device", request.Name,
				"power", *request.PowerPercent,
				"error", err)
			_ = req.Error("500", "failed to set cooling device power", nil)
			return
		}

		t.logger.InfoContext(ctx, "Updated cooling device power",
			"device", request.Name,
			"power_percent", *request.PowerPercent)
	}

	response := &v1alpha1.SetCoolingDeviceResponse{
		CoolingDevice: t.convertCoolingDeviceToProto(device),
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal set cooling device response",
			"device", request.Name,
			"error", err)
		_ = req.Error("500", "internal server error", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send set cooling device response",
			"error", err)
	}

	t.logger.DebugContext(ctx, "Updated cooling device",
		"device", request.Name,
		"power_percent", request.PowerPercent)
}

// handleSensorDataForThermalZone demonstrates proper cross-schema usage.
// This handler shows how the thermal service can consume sensor data (sensor.proto)
// to make thermal management decisions (thermal.proto).
func (t *ThermalMgr) handleSensorDataForThermalZone(ctx context.Context, req micro.Request) {
	var request v1alpha1.SensorDataRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		t.logger.WarnContext(ctx, "Invalid sensor data request",
			"error", err)
		_ = req.Error("400", "invalid request format", nil)
		return
	}

	// Example: Thermal service using sensor schema messages to collect temperature data
	// for thermal zone management decisions
	t.mu.RLock()
	zone, exists := t.thermalZones[request.GetZoneName()]
	t.mu.RUnlock()

	if !exists {
		_ = req.Error("404", fmt.Sprintf("thermal zone not found: %s", request.GetZoneName()), nil)
		return
	}

	// Create thermal zone temperature response using thermal schema
	// Note: Using available fields from the zone object and providing defaults for missing data
	currentTemp := 0.0 // Would get actual current temperature from sensors
	targetTemp := zone.TargetTemperature

	zoneTemp := &v1alpha1.ThermalZoneTemperature{
		ZoneName:              zone.Name,
		CurrentTemperature:    currentTemp,
		TargetTemperature:     &targetTemp,
		ContributingSensorIds: zone.SensorPaths, // Use SensorPaths instead of SensorNames
		// LastUpdated would be set from actual timestamp data
	}

	response := &v1alpha1.ThermalZoneTemperatureResponse{
		ZoneTemperatures: []*v1alpha1.ThermalZoneTemperature{zoneTemp},
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal thermal zone temperature response",
			"zone", request.GetZoneName(),
			"error", err)
		_ = req.Error("500", "internal server error", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send thermal zone temperature response",
			"error", err)
	}

	t.logger.DebugContext(ctx, "Provided thermal zone temperature data",
		"zone", request.GetZoneName(),
		"target_temp", zone.TargetTemperature)
}

// handleStartThermalControl handles requests to start thermal control.
func (t *ThermalMgr) handleStartThermalControl(ctx context.Context, req micro.Request) {
	t.mu.Lock()
	if t.controlRunning {
		t.mu.Unlock()
		_ = req.Error("409", "thermal control already running", nil)
		return
	}

	// Start thermal control loop
	go t.runThermalControl(ctx)
	t.mu.Unlock()

	response := map[string]interface{}{
		"status":  "started",
		"message": "thermal control started successfully",
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal start thermal control response",
			"error", err)
		_ = req.Error("500", "failed to marshal response", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send start thermal control response",
			"error", err)
	}

	t.logger.InfoContext(ctx, "Thermal control started")
}

// handleStopThermalControl handles requests to stop thermal control.
func (t *ThermalMgr) handleStopThermalControl(ctx context.Context, req micro.Request) {
	t.mu.Lock()
	if !t.controlRunning {
		t.mu.Unlock()
		_ = req.Error("409", "thermal control not running", nil)
		return
	}

	// Stop thermal control loop
	close(t.controlStop)
	t.controlRunning = false
	t.controlStop = make(chan struct{})
	t.mu.Unlock()

	response := map[string]interface{}{
		"status":  "stopped",
		"message": "thermal control stopped successfully",
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal stop thermal control response",
			"error", err)
		_ = req.Error("500", "failed to marshal response", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send stop thermal control response",
			"error", err)
	}

	t.logger.InfoContext(ctx, "Thermal control stopped")
}

// handleThermalControlStatus handles requests for thermal control status.
func (t *ThermalMgr) handleThermalControlStatus(ctx context.Context, req micro.Request) {
	t.mu.RLock()
	running := t.controlRunning
	zoneNames := make([]string, 0, len(t.thermalZones))
	for name := range t.thermalZones {
		zoneNames = append(zoneNames, name)
	}
	t.mu.RUnlock()

	response := ThermalControlStatusResponse{
		Running:            running,
		ActiveZones:        zoneNames,
		EmergencyCondition: false, // Would check actual emergency status
		LastUpdate:         "now", // Would use actual last update time
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal thermal control status response",
			"error", err)
		_ = req.Error("500", "failed to marshal response", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send thermal control status response",
			"error", err)
	}

	t.logger.DebugContext(ctx, "Retrieved thermal control status",
		"running", running,
		"zones", len(zoneNames))
}

// handleEmergencyThermal handles emergency thermal condition requests.
func (t *ThermalMgr) handleEmergencyThermal(ctx context.Context, req micro.Request) {
	type EmergencyThermalRequest struct {
		ZoneName string `json:"zone_name"`
		Action   string `json:"action"`
		Force    bool   `json:"force,omitempty"`
	}

	var request EmergencyThermalRequest
	if err := json.Unmarshal(req.Data(), &request); err != nil {
		t.logger.WarnContext(ctx, "Invalid emergency thermal request",
			"error", err)
		_ = req.Error("400", "invalid request format", nil)
		return
	}

	zone, exists := t.getThermalZone(request.ZoneName)
	if !exists {
		_ = req.Error("404", fmt.Sprintf("thermal zone not found: %s", request.ZoneName), nil)
		return
	}

	switch request.Action {
	case "emergency_cooling":
		if err := thermal.SetCoolingOutput(ctx, zone, t.config.failsafeCoolingLevel); err != nil {
			t.logger.ErrorContext(ctx, "Failed to apply emergency cooling",
				"zone", request.ZoneName,
				"error", err)
			_ = req.Error("500", "failed to apply emergency cooling", nil)
			return
		}
		t.logger.WarnContext(ctx, "Emergency cooling applied",
			"zone", request.ZoneName,
			"cooling_level", t.config.failsafeCoolingLevel)

	case "reset_pid":
		if err := thermal.ResetPIDController(ctx, zone); err != nil {
			t.logger.ErrorContext(ctx, "Failed to reset PID controller",
				"zone", request.ZoneName,
				"error", err)
			_ = req.Error("500", "failed to reset PID controller", nil)
			return
		}
		t.logger.InfoContext(ctx, "PID controller reset",
			"zone", request.ZoneName)

	default:
		_ = req.Error("400", fmt.Sprintf("unsupported action: %s", request.Action), nil)
		return
	}

	response := map[string]interface{}{
		"status":    "completed",
		"zone_name": request.ZoneName,
		"action":    request.Action,
		"message":   fmt.Sprintf("emergency action %s completed for zone %s", request.Action, request.ZoneName),
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal emergency thermal response",
			"zone", request.ZoneName,
			"action", request.Action,
			"error", err)
		_ = req.Error("500", "failed to marshal response", nil)
		return
	}

	if err := req.Respond(responseData); err != nil {
		t.logger.ErrorContext(ctx, "Failed to send emergency thermal response",
			"zone", request.ZoneName,
			"action", request.Action,
			"error", err)
	}

	t.logger.InfoContext(ctx, "Emergency thermal action completed",
		"zone", request.ZoneName,
		"action", request.Action)
}

// convertThermalZoneToProto converts a thermal zone to protobuf format.
func (t *ThermalMgr) convertThermalZoneToProto(zone *thermal.Zone) *v1alpha1.ThermalZone {
	protoZone := &v1alpha1.ThermalZone{
		Name:               zone.Name,
		SensorNames:        zone.SensorPaths,
		CurrentTemperature: zone.TargetTemperature, // Would use actual current temp
		TargetTemperature:  &zone.TargetTemperature,
		Status:             v1alpha1.ThermalZoneStatus_THERMAL_ZONE_STATUS_NORMAL,
		CustomAttributes:   make(map[string]string),
	}

	// Add cooling device names
	for _, device := range zone.CoolingDevices {
		protoZone.CoolingDeviceNames = append(protoZone.CoolingDeviceNames, device.Name)
	}

	// Convert PID settings
	protoZone.PidSettings = &v1alpha1.PIDSettings{
		Kp:         zone.PIDConfig.Kp,
		Ki:         zone.PIDConfig.Ki,
		Kd:         zone.PIDConfig.Kd,
		SampleTime: zone.PIDConfig.SampleTime.Seconds(),
		OutputMin:  &zone.PIDConfig.OutputMin,
		OutputMax:  &zone.PIDConfig.OutputMax,
	}

	return protoZone
}

// convertCoolingDeviceToProto converts a cooling device to protobuf format.
func (t *ThermalMgr) convertCoolingDeviceToProto(device *thermal.CoolingDevice) *v1alpha1.CoolingDevice {
	deviceType := device.Type
	currentPower := device.CurrentPower
	minPower := device.MinPower
	maxPower := device.MaxPower
	status := device.Status
	controlMode := v1alpha1.CoolingDeviceControlMode_COOLING_DEVICE_CONTROL_MODE_AUTOMATIC

	protoDevice := &v1alpha1.CoolingDevice{
		Name:                   device.Name,
		Type:                   &deviceType,
		CoolingPowerPercent:    &currentPower,
		MinCoolingPowerPercent: &minPower,
		MaxCoolingPowerPercent: &maxPower,
		Status:                 &status,
		ControlMode:            &controlMode,
		CustomAttributes:       make(map[string]string),
	}

	// Add hwmon path to custom attributes
	if device.HwmonPath != "" {
		protoDevice.CustomAttributes["hwmon_path"] = device.HwmonPath
	}

	return protoDevice
}
