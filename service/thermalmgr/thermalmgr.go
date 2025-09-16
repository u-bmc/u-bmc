// SPDX-License-Identifier: BSD-3-Clause

package thermalmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/pkg/thermal"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ service.Service = (*ThermalMgr)(nil)

// ThermalMgr manages thermal zones and cooling devices for BMC systems.
// It provides NATS-based IPC endpoints for thermal management operations.
type ThermalMgr struct {
	config         *Config
	nc             *nats.Conn
	js             jetstream.JetStream
	microService   micro.Service
	thermalZones   map[string]*thermal.ThermalZone
	coolingDevices map[string]*thermal.CoolingDevice
	controlRunning bool
	controlStop    chan struct{}
	emergencyStop  chan struct{}
	mu             sync.RWMutex
	logger         *slog.Logger
	tracer         trace.Tracer
	cancel         context.CancelFunc
	started        bool
}

// New creates a new ThermalMgr instance with the provided options.
func New(opts ...Option) *ThermalMgr {
	config := NewConfig(opts...)

	return &ThermalMgr{
		config:         config,
		thermalZones:   make(map[string]*thermal.ThermalZone),
		coolingDevices: make(map[string]*thermal.CoolingDevice),
		controlStop:    make(chan struct{}),
		emergencyStop:  make(chan struct{}),
		tracer:         otel.Tracer("thermalmgr"),
	}
}

// Name returns the service name.
func (t *ThermalMgr) Name() string {
	return t.config.ServiceName
}

// Run starts the thermal manager service and registers NATS IPC endpoints.
func (t *ThermalMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return ErrServiceAlreadyStarted
	}
	t.started = true
	ctx, t.cancel = context.WithCancel(ctx)
	t.mu.Unlock()

	ctx, span := t.tracer.Start(ctx, "thermalmgr.Run")
	defer span.End()

	t.logger = log.GetGlobalLogger().With("service", t.config.ServiceName)
	t.logger.InfoContext(ctx, "Starting thermal manager service",
		"version", t.config.ServiceVersion,
		"thermal_control", t.config.EnableThermalControl,
		"hwmon_discovery", t.config.EnableHwmonDiscovery)

	if err := t.config.Validate(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrNATSConnectionFailed, err)
	}
	t.nc = nc
	defer nc.Drain() //nolint:errcheck

	if t.config.PersistThermalData {
		t.js, err = jetstream.New(nc)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("%w: %w", ErrJetStreamInitFailed, err)
		}

		if err := t.setupJetStream(ctx); err != nil {
			span.RecordError(err)
			return err
		}
	}

	if err := t.initializeThermalSystem(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	t.microService, err = micro.AddService(nc, micro.Config{
		Name:        t.config.ServiceName,
		Description: t.config.ServiceDescription,
		Version:     t.config.ServiceVersion,
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrMicroServiceCreationFailed, err)
	}

	if err := t.registerEndpoints(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	if t.config.EnableThermalControl {
		go t.runThermalControl(ctx)
	}

	if t.config.EnableEmergencyResponse {
		go t.runEmergencyMonitoring(ctx)
	}

	t.logger.InfoContext(ctx, "Thermal manager service started successfully",
		"endpoints_registered", true,
		"thermal_zones", len(t.thermalZones),
		"cooling_devices", len(t.coolingDevices))

	span.SetAttributes(
		attribute.String("service.name", t.config.ServiceName),
		attribute.String("service.version", t.config.ServiceVersion),
		attribute.Int("thermal_zones.count", len(t.thermalZones)),
		attribute.Int("cooling_devices.count", len(t.coolingDevices)),
	)

	<-ctx.Done()

	err = ctx.Err()
	ctx = context.WithoutCancel(ctx)
	t.logger.InfoContext(ctx, "Shutting down thermal manager service")
	t.shutdown(ctx)

	return err
}

func (t *ThermalMgr) setupJetStream(ctx context.Context) error {
	streamConfig := jetstream.StreamConfig{
		Name:        t.config.StreamName,
		Description: "Thermal manager data stream",
		Subjects:    t.config.StreamSubjects,
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      t.config.StreamRetention,
		Storage:     jetstream.FileStorage,
		Replicas:    1,
		MaxMsgs:     -1,
		MaxBytes:    -1,
	}

	stream, err := t.js.CreateOrUpdateStream(ctx, streamConfig)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStreamCreationFailed, err)
	}

	info, err := stream.Info(ctx)
	if err == nil {
		t.logger.InfoContext(ctx, "JetStream stream configured",
			"name", info.Config.Name,
			"subjects", info.Config.Subjects,
			"messages", info.State.Msgs)
	}

	return nil
}

func (t *ThermalMgr) initializeThermalSystem(ctx context.Context) error {
	if t.config.EnableHwmonDiscovery {
		if err := t.discoverCoolingDevices(ctx); err != nil {
			t.logger.WarnContext(ctx, "Cooling device discovery failed", "error", err)
		}
	}

	// Initialize default thermal zones if none are configured
	if len(t.thermalZones) == 0 {
		if err := t.createDefaultThermalZones(ctx); err != nil {
			t.logger.WarnContext(ctx, "Failed to create default thermal zones", "error", err)
		}
	}

	return nil
}

func (t *ThermalMgr) discoverCoolingDevices(ctx context.Context) error {
	devices, err := thermal.DiscoverCoolingDevices(ctx, t.config.HwmonPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrThermalDiscoveryFailed, err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, device := range devices {
		t.coolingDevices[device.Name] = device
		t.logger.DebugContext(ctx, "Discovered cooling device",
			"name", device.Name,
			"type", device.Type.String(),
			"path", device.HwmonPath)
	}

	t.logger.InfoContext(ctx, "Cooling device discovery completed",
		"devices_found", len(devices))

	return nil
}

func (t *ThermalMgr) createDefaultThermalZones(ctx context.Context) error {
	// Create a default thermal zone if we have cooling devices
	t.mu.RLock()
	deviceCount := len(t.coolingDevices)
	t.mu.RUnlock()

	if deviceCount == 0 {
		t.logger.InfoContext(ctx, "No cooling devices found, skipping default thermal zone creation")
		return nil
	}

	defaultZone := &thermal.ThermalZone{
		Name:                "default_zone",
		SensorPaths:         []string{}, // Will be populated by sensor integration
		TargetTemperature:   65.0,
		WarningTemperature:  t.config.DefaultWarningTemp,
		CriticalTemperature: t.config.DefaultCriticalTemp,
		PIDConfig: thermal.PIDConfig{
			Kp:         t.config.DefaultPIDKp,
			Ki:         t.config.DefaultPIDKi,
			Kd:         t.config.DefaultPIDKd,
			SampleTime: t.config.DefaultPIDSampleTime,
			OutputMin:  t.config.DefaultOutputMin,
			OutputMax:  t.config.DefaultOutputMax,
		},
	}

	// Add all discovered cooling devices to the default zone
	t.mu.Lock()
	for _, device := range t.coolingDevices {
		defaultZone.CoolingDevices = append(defaultZone.CoolingDevices, device)
	}
	t.mu.Unlock()

	if err := thermal.InitializeThermalZone(ctx, defaultZone); err != nil {
		return fmt.Errorf("%w: %w", ErrThermalZoneInitFailed, err)
	}

	t.mu.Lock()
	t.thermalZones[defaultZone.Name] = defaultZone
	t.mu.Unlock()

	t.logger.InfoContext(ctx, "Created default thermal zone",
		"name", defaultZone.Name,
		"target_temp", defaultZone.TargetTemperature,
		"cooling_devices", len(defaultZone.CoolingDevices))

	return nil
}

func (t *ThermalMgr) registerEndpoints(ctx context.Context) error {
	endpoints := []struct {
		subject string
		handler micro.HandlerFunc
	}{
		// Thermal zone management
		{"thermalmgr.zones.list", t.createRequestHandler(t.handleListThermalZones)},
		{"thermalmgr.zone.get", t.createRequestHandler(t.handleGetThermalZone)},
		{"thermalmgr.zone.set", t.createRequestHandler(t.handleSetThermalZone)},

		// Cooling device management
		{"thermalmgr.devices.list", t.createRequestHandler(t.handleListCoolingDevices)},
		{"thermalmgr.device.get", t.createRequestHandler(t.handleGetCoolingDevice)},
		{"thermalmgr.device.set", t.createRequestHandler(t.handleSetCoolingDevice)},

		// Thermal control
		{"thermalmgr.control.start", t.createRequestHandler(t.handleStartThermalControl)},
		{"thermalmgr.control.stop", t.createRequestHandler(t.handleStopThermalControl)},
		{"thermalmgr.control.status", t.createRequestHandler(t.handleThermalControlStatus)},
		{"thermalmgr.control.emergency", t.createRequestHandler(t.handleEmergencyThermal)},
	}

	for _, ep := range endpoints {
		if err := t.microService.AddEndpoint(ep.subject, ep.handler); err != nil {
			return fmt.Errorf("%w: %s: %w", ErrEndpointRegistrationFailed, ep.subject, err)
		}
	}

	return nil
}

func (t *ThermalMgr) createRequestHandler(handler func(context.Context, micro.Request)) micro.HandlerFunc {
	return func(req micro.Request) {
		ctx := telemetry.GetCtxFromReq(req)

		if t.tracer != nil {
			_, span := t.tracer.Start(ctx, "thermalmgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", t.config.ServiceName),
			)
			defer span.End()
		}

		handler(ctx, req)
	}
}

func (t *ThermalMgr) runThermalControl(ctx context.Context) {
	t.logger.InfoContext(ctx, "Starting thermal control loop")

	ticker := time.NewTicker(t.config.ThermalControlInterval)
	defer ticker.Stop()

	t.mu.Lock()
	t.controlRunning = true
	t.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			t.logger.InfoContext(ctx, "Thermal control loop stopped by context")
			return
		case <-t.controlStop:
			t.logger.InfoContext(ctx, "Thermal control loop stopped by signal")
			return
		case <-ticker.C:
			t.performThermalControl(ctx)
		}
	}
}

func (t *ThermalMgr) performThermalControl(ctx context.Context) {
	t.mu.RLock()
	zones := make([]*thermal.ThermalZone, 0, len(t.thermalZones))
	for _, zone := range t.thermalZones {
		zones = append(zones, zone)
	}
	t.mu.RUnlock()

	for _, zone := range zones {
		if err := t.updateThermalZone(ctx, zone); err != nil {
			t.logger.WarnContext(ctx, "Thermal zone update failed",
				"zone", zone.Name,
				"error", err)
		}
	}
}

func (t *ThermalMgr) updateThermalZone(ctx context.Context, zone *thermal.ThermalZone) error {
	// Read temperature from zone sensors
	temperature, err := thermal.ReadZoneTemperature(ctx, zone)
	if err != nil {
		return fmt.Errorf("failed to read zone temperature: %w", err)
	}

	// Update PID control
	output, err := thermal.UpdatePIDControl(ctx, zone, temperature)
	if err != nil {
		return fmt.Errorf("PID control update failed: %w", err)
	}

	// Apply cooling output
	if err := thermal.SetCoolingOutput(ctx, zone, output); err != nil {
		return fmt.Errorf("failed to set cooling output: %w", err)
	}

	t.logger.DebugContext(ctx, "Thermal zone updated",
		"zone", zone.Name,
		"temperature", temperature,
		"target", zone.TargetTemperature,
		"output", output)

	return nil
}

func (t *ThermalMgr) runEmergencyMonitoring(ctx context.Context) {
	t.logger.InfoContext(ctx, "Starting emergency thermal monitoring")

	ticker := time.NewTicker(t.config.EmergencyCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.emergencyStop:
			return
		case <-ticker.C:
			t.checkEmergencyConditions(ctx)
		}
	}
}

func (t *ThermalMgr) checkEmergencyConditions(ctx context.Context) {
	t.mu.RLock()
	zones := make([]*thermal.ThermalZone, 0, len(t.thermalZones))
	for _, zone := range t.thermalZones {
		zones = append(zones, zone)
	}
	t.mu.RUnlock()

	for _, zone := range zones {
		if err := thermal.CheckThermalEmergency(ctx, zone); err != nil {
			if err == thermal.ErrCriticalTemperature {
				t.logger.ErrorContext(ctx, "Critical temperature detected",
					"zone", zone.Name)
				t.handleCriticalThermalCondition(ctx, zone)
			}
		}
	}
}

func (t *ThermalMgr) handleCriticalThermalCondition(ctx context.Context, zone *thermal.ThermalZone) {
	// Apply maximum cooling immediately
	if err := thermal.SetCoolingOutput(ctx, zone, t.config.FailsafeCoolingLevel); err != nil {
		t.logger.ErrorContext(ctx, "Failed to apply emergency cooling",
			"zone", zone.Name,
			"error", err)
	}

	// If power integration is enabled, request emergency action
	if t.config.EnablePowerIntegration {
		go t.requestEmergencyPowerAction(ctx, zone)
	}
}

func (t *ThermalMgr) requestEmergencyPowerAction(ctx context.Context, zone *thermal.ThermalZone) {
	// Wait for emergency response delay to allow cooling to take effect
	select {
	case <-ctx.Done():
		return
	case <-time.After(t.config.EmergencyResponseDelay):
	}

	// Check if thermal condition persists
	if err := thermal.CheckThermalEmergency(ctx, zone); err == thermal.ErrCriticalTemperature {
		t.logger.ErrorContext(ctx, "Thermal emergency persists, requesting power intervention",
			"zone", zone.Name)

		// Send emergency thermal message to powermgr
		emergencyMsg := map[string]interface{}{
			"type":        "thermal_emergency",
			"zone":        zone.Name,
			"temperature": zone.TargetTemperature, // Would be actual temp in real implementation
			"action":      "emergency_shutdown",
		}

		if data, err := json.Marshal(emergencyMsg); err == nil {
			subject := fmt.Sprintf("%s.emergency.thermal", t.config.PowermgrEndpoint)
			if err := t.nc.Publish(subject, data); err != nil {
				t.logger.ErrorContext(ctx, "Failed to send emergency message to powermgr",
					"error", err)
			}
		}
	}
}

func (t *ThermalMgr) shutdown(ctx context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}

	if t.controlRunning {
		close(t.controlStop)
		t.controlRunning = false
	}

	close(t.emergencyStop)
	t.started = false
}

// getThermalZone safely retrieves a thermal zone by name.
func (t *ThermalMgr) getThermalZone(name string) (*thermal.ThermalZone, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	zone, exists := t.thermalZones[name]
	return zone, exists
}

// getCoolingDevice safely retrieves a cooling device by name.
func (t *ThermalMgr) getCoolingDevice(name string) (*thermal.CoolingDevice, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	device, exists := t.coolingDevices[name]
	return device, exists
}
