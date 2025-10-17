// SPDX-License-Identifier: BSD-3-Clause

package thermalmgr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	"github.com/u-bmc/u-bmc/pkg/ipc"
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
	config         *config
	nc             *nats.Conn
	js             jetstream.JetStream
	microService   micro.Service
	thermalZones   map[string]*thermal.Zone
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
	cfg := &config{
		serviceName:             DefaultServiceName,
		serviceDescription:      DefaultServiceDescription,
		serviceVersion:          DefaultServiceVersion,
		enableThermalControl:    true,
		thermalControlInterval:  DefaultThermalControlInterval,
		emergencyCheckInterval:  DefaultEmergencyCheckInterval,
		defaultPIDSampleTime:    DefaultDefaultPIDSampleTime,
		maxThermalZones:         DefaultMaxThermalZones,
		maxCoolingDevices:       DefaultMaxCoolingDevices,
		hwmonPath:               DefaultHwmonPath,
		enableHwmonDiscovery:    true,
		defaultWarningTemp:      DefaultDefaultWarningTemp,
		defaultCriticalTemp:     DefaultDefaultCriticalTemp,
		emergencyShutdownTemp:   DefaultEmergencyShutdownTemp,
		defaultPIDKp:            DefaultDefaultPIDKp,
		defaultPIDKi:            DefaultDefaultPIDKi,
		defaultPIDKd:            DefaultDefaultPIDKd,
		defaultOutputMin:        DefaultDefaultOutputMin,
		defaultOutputMax:        DefaultDefaultOutputMax,
		sensormonEndpoint:       "sensormon",
		powermgrEndpoint:        "powermgr",
		enableSensorIntegration: true,
		enablePowerIntegration:  true,
		persistThermalData:      false,
		streamName:              "THERMALMGR",
		streamSubjects:          []string{"thermalmgr.>"},
		streamRetention:         24 * time.Hour,
		enableEmergencyResponse: true,
		emergencyResponseDelay:  DefaultEmergencyResponseDelay,
		failsafeCoolingLevel:    DefaultFailsafeCoolingLevel,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &ThermalMgr{
		config:         cfg,
		thermalZones:   make(map[string]*thermal.Zone),
		coolingDevices: make(map[string]*thermal.CoolingDevice),
		controlStop:    make(chan struct{}),
		emergencyStop:  make(chan struct{}),
	}
}

// Name returns the service name.
func (t *ThermalMgr) Name() string {
	return t.config.serviceName
}

// Run starts the thermal manager service and registers NATS IPC endpoints.
func (t *ThermalMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	t.tracer = otel.Tracer(t.config.serviceName)

	ctx, span := t.tracer.Start(ctx, "thermalmgr.Run")
	defer span.End()

	t.logger = log.GetGlobalLogger().With("service", t.config.serviceName)

	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return ErrServiceAlreadyStarted
	}
	t.started = true
	ctx, t.cancel = context.WithCancel(ctx)
	t.mu.Unlock()

	t.logger.InfoContext(ctx, "Starting thermal manager service",
		"version", t.config.serviceVersion,
		"thermal_control", t.config.enableThermalControl,
		"hwmon_discovery", t.config.enableHwmonDiscovery)

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

	if t.config.persistThermalData {
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
		Name:        t.config.serviceName,
		Description: t.config.serviceDescription,
		Version:     t.config.serviceVersion,
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrMicroServiceCreationFailed, err)
	}

	if err := t.registerEndpoints(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	if t.config.enableThermalControl {
		go t.runThermalControl(ctx)
	}

	if t.config.enableEmergencyResponse {
		go t.runEmergencyMonitoring(ctx)
	}

	t.logger.InfoContext(ctx, "Thermal manager service started successfully",
		"endpoints_registered", true,
		"thermal_zones", len(t.thermalZones),
		"cooling_devices", len(t.coolingDevices))

	span.SetAttributes(
		attribute.String("service.name", t.config.serviceName),
		attribute.String("service.version", t.config.serviceVersion),
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
		Name:        t.config.streamName,
		Description: "Thermal manager data stream",
		Subjects:    t.config.streamSubjects,
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      t.config.streamRetention,
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
	if t.config.enableHwmonDiscovery {
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
	devices, err := thermal.DiscoverCoolingDevices(ctx, t.config.hwmonPath)
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

	defaultZone := &thermal.Zone{
		Name:                "default_zone",
		SensorPaths:         []string{}, // Will be populated by sensor integration
		TargetTemperature:   65.0,
		WarningTemperature:  t.config.defaultWarningTemp,
		CriticalTemperature: t.config.defaultCriticalTemp,
		PIDConfig: thermal.PIDConfig{
			Kp:         t.config.defaultPIDKp,
			Ki:         t.config.defaultPIDKi,
			Kd:         t.config.defaultPIDKd,
			SampleTime: t.config.defaultPIDSampleTime,
			OutputMin:  t.config.defaultOutputMin,
			OutputMax:  t.config.defaultOutputMax,
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
	groups := make(map[string]micro.Group)

	if err := ipc.RegisterEndpointWithGroupCache(t.microService, ipc.SubjectThermalZoneList,
		micro.HandlerFunc(t.createRequestHandler(ctx, t.handleListThermalZones)), groups); err != nil {
		return fmt.Errorf("failed to register thermal zone list endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(t.microService, ipc.SubjectThermalZoneInfo,
		micro.HandlerFunc(t.createRequestHandler(ctx, t.handleGetThermalZone)), groups); err != nil {
		return fmt.Errorf("failed to register thermal zone info endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(t.microService, ipc.SubjectThermalZoneSet,
		micro.HandlerFunc(t.createRequestHandler(ctx, t.handleSetThermalZone)), groups); err != nil {
		return fmt.Errorf("failed to register thermal zone set endpoint: %w", err)
	}

	return nil
}

func (t *ThermalMgr) createRequestHandler(parentCtx context.Context, handler func(context.Context, micro.Request)) micro.HandlerFunc {
	return func(req micro.Request) {
		ctx := telemetry.GetCtxFromReq(req)

		// Merge parent context values/cancellation with telemetry context
		ctx = context.WithoutCancel(ctx)
		if parentCtx != nil {
			// Create a new context that inherits from parent but uses telemetry values
			select {
			case <-parentCtx.Done():
				// If parent is already canceled, use a canceled context
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			default:
				// Parent is still active
			}
		}

		if t.tracer != nil {
			_, span := t.tracer.Start(ctx, "thermalmgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", t.config.serviceName),
			)
			defer span.End()
		}

		handler(ctx, req) //nolint:contextcheck
	}
}

func (t *ThermalMgr) runThermalControl(ctx context.Context) {
	t.logger.InfoContext(ctx, "Starting thermal control loop")

	ticker := time.NewTicker(t.config.thermalControlInterval)
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
	zones := make([]*thermal.Zone, 0, len(t.thermalZones))
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

func (t *ThermalMgr) updateThermalZone(ctx context.Context, zone *thermal.Zone) error {
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

	ticker := time.NewTicker(t.config.emergencyCheckInterval)
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
	zones := make([]*thermal.Zone, 0, len(t.thermalZones))
	for _, zone := range t.thermalZones {
		zones = append(zones, zone)
	}
	t.mu.RUnlock()

	for _, zone := range zones {
		if err := thermal.CheckThermalEmergency(ctx, zone); err != nil {
			if errors.Is(err, thermal.ErrCriticalTemperature) {
				t.logger.ErrorContext(ctx, "Critical temperature detected",
					"zone", zone.Name)
				t.handleCriticalThermalCondition(ctx, zone)
			}
		}
	}
}

func (t *ThermalMgr) handleCriticalThermalCondition(ctx context.Context, zone *thermal.Zone) {
	// Apply maximum cooling immediately
	if err := thermal.SetCoolingOutput(ctx, zone, t.config.failsafeCoolingLevel); err != nil {
		t.logger.ErrorContext(ctx, "Failed to apply emergency cooling",
			"zone", zone.Name,
			"error", err)
	}

	// If power integration is enabled, request emergency action
	if t.config.enablePowerIntegration {
		go t.requestEmergencyPowerAction(ctx, zone)
	}
}

func (t *ThermalMgr) requestEmergencyPowerAction(ctx context.Context, zone *thermal.Zone) {
	// Wait for emergency response delay to allow cooling to take effect
	select {
	case <-ctx.Done():
		return
	case <-time.After(t.config.emergencyResponseDelay):
	}

	// Check if thermal condition persists
	if err := thermal.CheckThermalEmergency(ctx, zone); errors.Is(err, thermal.ErrCriticalTemperature) {
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
			subject := fmt.Sprintf("%s.emergency.thermal", t.config.powermgrEndpoint)
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
func (t *ThermalMgr) getThermalZone(name string) (*thermal.Zone, bool) {
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
