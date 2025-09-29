// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/hwmon"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ service.Service = (*SensorMon)(nil)

// SensorMon manages sensor monitoring operations for BMC systems.
// It provides NATS-based IPC endpoints for sensor management and monitoring.
type SensorMon struct {
	config          *config
	nc              *nats.Conn
	js              jetstream.JetStream
	microService    micro.Service
	sensors         map[string]*sensorInfo
	monitoring      bool
	monitoringStop  chan struct{}
	monitoringStats MonitoringStats
	mu              sync.RWMutex
	logger          *slog.Logger
	tracer          trace.Tracer
	cancel          context.CancelFunc
	started         bool
}

// sensorInfo holds information about a discovered sensor.
type sensorInfo struct {
	Sensor    *v1alpha1.Sensor
	Path      string
	Type      sensorType
	LastRead  time.Time
	LastValue interface{}
}

// sensorType represents the type of sensor backend.
type sensorType int

const (
	sensorTypeHwmon sensorType = iota
	sensorTypeGPIO
)

// New creates a new SensorMon instance with the provided options.
func New(opts ...Option) *SensorMon {
	cfg := &config{
		serviceName:               DefaultServiceName,
		serviceDescription:        DefaultServiceDescription,
		serviceVersion:            DefaultServiceVersion,
		hwmonPath:                 DefaultHwmonPath,
		gpioChipPath:              DefaultGPIOChipPath,
		monitoringInterval:        DefaultMonitoringInterval,
		thresholdCheckInterval:    DefaultThresholdCheckInterval,
		sensorTimeout:             DefaultSensorTimeout,
		enableHwmonSensors:        true,
		enableGPIOSensors:         false,
		enableMetrics:             true,
		enableTracing:             true,
		enableThresholdMonitoring: true,
		broadcastSensorReadings:   false,
		persistSensorData:         false,
		streamName:                "SENSORMON",
		streamSubjects:            []string{"sensormon.data.>", "sensormon.events.>"},
		streamRetention:           24 * time.Hour,
		maxConcurrentReads:        DefaultMaxConcurrentReads,
		sensorDiscoveryTimeout:    DefaultSensorDiscoveryTimeout,
		enableThermalIntegration:  false,
		thermalMgrEndpoint:        "thermalmgr",
		temperatureUpdateInterval: DefaultTemperatureUpdateInterval,
		enableThermalAlerts:       false,
		criticalTempThreshold:     DefaultCriticalTempThreshold,
		warningTempThreshold:      DefaultWarningTempThreshold,
		emergencyResponseDelay:    DefaultEmergencyResponseDelay,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &SensorMon{
		config:         cfg,
		sensors:        make(map[string]*sensorInfo),
		monitoringStop: make(chan struct{}),
	}
}

// Name returns the service name.
func (s *SensorMon) Name() string {
	return s.config.serviceName
}

// Run starts the sensor monitoring service and registers NATS IPC endpoints.
func (s *SensorMon) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	s.tracer = otel.Tracer(s.config.serviceName)

	ctx, span := s.tracer.Start(ctx, "sensormon.Run")
	defer span.End()

	s.logger = log.GetGlobalLogger().With("service", s.config.serviceName)

	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return ErrServiceAlreadyStarted
	}
	s.started = true
	ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	s.logger.InfoContext(ctx, "Starting sensor monitoring service",
		"version", s.config.serviceVersion,
		"hwmon_enabled", s.config.enableHwmonSensors,
		"gpio_enabled", s.config.enableGPIOSensors)

	if err := s.config.Validate(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrNATSConnectionFailed, err)
	}
	s.nc = nc
	defer nc.Drain() //nolint:errcheck

	if s.config.persistSensorData {
		s.js, err = jetstream.New(nc)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("%w: %w", ErrJetStreamInitFailed, err)
		}

		if err := s.setupJetStream(ctx); err != nil {
			span.RecordError(err)
			return err
		}
	}

	if err := s.discoverSensors(ctx); err != nil {
		span.RecordError(err)
		s.logger.WarnContext(ctx, "Sensor discovery failed", "error", err)
	}

	if err := s.initializeThermalIntegration(ctx); err != nil {
		span.RecordError(err)
		s.logger.WarnContext(ctx, "Thermal integration initialization failed", "error", err)
	}

	s.microService, err = micro.AddService(nc, micro.Config{
		Name:        s.config.serviceName,
		Description: s.config.serviceDescription,
		Version:     s.config.serviceVersion,
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrMicroServiceCreationFailed, err)
	}

	if err := s.registerEndpoints(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	s.logger.InfoContext(ctx, "Sensor monitoring service started successfully",
		"endpoints_registered", true,
		"sensors_discovered", len(s.sensors))

	span.SetAttributes(
		attribute.String("service.name", s.config.serviceName),
		attribute.String("service.version", s.config.serviceVersion),
		attribute.Int("sensors.count", len(s.sensors)),
	)

	<-ctx.Done()

	err = ctx.Err()
	ctx = context.WithoutCancel(ctx)
	s.logger.InfoContext(ctx, "Shutting down sensor monitoring service")
	s.shutdown(ctx)

	return err
}

func (s *SensorMon) setupJetStream(ctx context.Context) error {
	streamConfig := jetstream.StreamConfig{
		Name:        s.config.streamName,
		Description: "Sensor monitoring data stream",
		Subjects:    s.config.streamSubjects,
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      s.config.streamRetention,
		Storage:     jetstream.FileStorage,
		Replicas:    1,
		MaxMsgs:     -1,
		MaxBytes:    -1,
	}

	stream, err := s.js.CreateOrUpdateStream(ctx, streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create JetStream stream: %w", err)
	}

	info, err := stream.Info(ctx)
	if err == nil {
		s.logger.InfoContext(ctx, "JetStream stream configured",
			"name", info.Config.Name,
			"subjects", info.Config.Subjects,
			"messages", info.State.Msgs)
	}

	return nil
}

func (s *SensorMon) discoverSensors(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.config.sensorDiscoveryTimeout)
	defer cancel()

	if s.config.enableHwmonSensors {
		if err := s.discoverHwmonSensors(ctx); err != nil {
			s.logger.WarnContext(ctx, "Hwmon sensor discovery failed", "error", err)
		}
	}

	if s.config.enableGPIOSensors {
		if err := s.discoverGPIOSensors(ctx); err != nil {
			s.logger.WarnContext(ctx, "GPIO sensor discovery failed", "error", err)
		}
	}

	return nil
}

func (s *SensorMon) discoverHwmonSensors(ctx context.Context) error {
	devices, err := hwmon.ListDevicesInPathCtx(ctx, s.config.hwmonPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSensorDiscoveryFailed, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, device := range devices {
		deviceName, err := hwmon.ReadStringCtx(ctx, filepath.Join(device, "name"))
		if err != nil {
			continue
		}

		if err := s.discoverHwmonSensorsInDevice(ctx, device, deviceName); err != nil {
			s.logger.WarnContext(ctx, "Failed to discover sensors in device",
				"device", deviceName,
				"path", device,
				"error", err)
		}
	}

	return nil
}

func (s *SensorMon) discoverHwmonSensorsInDevice(ctx context.Context, devicePath, deviceName string) error {
	// Discover temperature sensors
	tempAttrs, err := hwmon.ListAttributesCtx(ctx, devicePath, `temp\d+_input`)
	if err == nil {
		for _, attr := range tempAttrs {
			if err := s.addHwmonSensor(ctx, devicePath, deviceName, attr, v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE); err != nil {
				s.logger.WarnContext(ctx, "Failed to add temperature sensor",
					"device", deviceName,
					"attribute", attr,
					"error", err)
			}
		}
	}

	// Discover voltage sensors
	voltageAttrs, err := hwmon.ListAttributesCtx(ctx, devicePath, `in\d+_input`)
	if err == nil {
		for _, attr := range voltageAttrs {
			if err := s.addHwmonSensor(ctx, devicePath, deviceName, attr, v1alpha1.SensorContext_SENSOR_CONTEXT_VOLTAGE); err != nil {
				s.logger.WarnContext(ctx, "Failed to add voltage sensor",
					"device", deviceName,
					"attribute", attr,
					"error", err)
			}
		}
	}

	// Discover fan sensors
	fanAttrs, err := hwmon.ListAttributesCtx(ctx, devicePath, `fan\d+_input`)
	if err == nil {
		for _, attr := range fanAttrs {
			if err := s.addHwmonSensor(ctx, devicePath, deviceName, attr, v1alpha1.SensorContext_SENSOR_CONTEXT_TACH); err != nil {
				s.logger.WarnContext(ctx, "Failed to add fan sensor",
					"device", deviceName,
					"attribute", attr,
					"error", err)
			}
		}
	}

	// Discover power sensors
	powerAttrs, err := hwmon.ListAttributesCtx(ctx, devicePath, `power\d+_input`)
	if err == nil {
		for _, attr := range powerAttrs {
			if err := s.addHwmonSensor(ctx, devicePath, deviceName, attr, v1alpha1.SensorContext_SENSOR_CONTEXT_POWER); err != nil {
				s.logger.WarnContext(ctx, "Failed to add power sensor",
					"device", deviceName,
					"attribute", attr,
					"error", err)
			}
		}
	}

	// Discover current sensors
	currentAttrs, err := hwmon.ListAttributesCtx(ctx, devicePath, `curr\d+_input`)
	if err == nil {
		for _, attr := range currentAttrs {
			if err := s.addHwmonSensor(ctx, devicePath, deviceName, attr, v1alpha1.SensorContext_SENSOR_CONTEXT_CURRENT); err != nil {
				s.logger.WarnContext(ctx, "Failed to add current sensor",
					"device", deviceName,
					"attribute", attr,
					"error", err)
			}
		}
	}

	return nil
}

func (s *SensorMon) addHwmonSensor(ctx context.Context, devicePath, deviceName, devAttribute string, sensorContext v1alpha1.SensorContext) error {
	sensorPath := filepath.Join(devicePath, devAttribute)
	sensorID := fmt.Sprintf("%s_%s", deviceName, devAttribute)

	// Try to read the label
	labelFile := filepath.Base(devAttribute)
	labelFile = labelFile[:len(labelFile)-len("_input")] + "_label"
	labelPath := filepath.Join(devicePath, labelFile)

	var sensorName string
	if hwmon.FileExistsCtx(ctx, labelPath) {
		if label, err := hwmon.ReadStringCtx(ctx, labelPath); err == nil {
			sensorName = label
		}
	}
	if sensorName == "" {
		sensorName = fmt.Sprintf("%s %s", deviceName, devAttribute)
	}

	sensorStatus := v1alpha1.SensorStatus_SENSOR_STATUS_ENABLED
	sensorUnit := s.getUnitForContext(sensorContext)

	sensor := &v1alpha1.Sensor{
		Id:      sensorID,
		Name:    sensorName,
		Context: &sensorContext,
		Status:  &sensorStatus,
		Unit:    &sensorUnit,
		Reading: &v1alpha1.Sensor_AnalogReading{
			AnalogReading: &v1alpha1.AnalogSensorReading{},
		},
		Location: &v1alpha1.Location{
			ComponentLocation: &v1alpha1.ComponentLocation{
				Name: fmt.Sprintf("Device: %s, Path: %s", deviceName, sensorPath),
			},
		},
		LastReadingTimestamp: timestamppb.Now(),
		CustomAttributes:     make(map[string]string),
	}

	info := &sensorInfo{
		Sensor: sensor,
		Path:   sensorPath,
		Type:   sensorTypeHwmon,
	}

	s.sensors[sensorID] = info

	s.logger.DebugContext(ctx, "Added hwmon sensor",
		"id", sensorID,
		"name", sensorName,
		"path", sensorPath,
		"context", sensorContext.String())

	return nil
}

func (s *SensorMon) discoverGPIOSensors(ctx context.Context) error {
	// GPIO sensor discovery would be implemented based on configuration
	// This is a placeholder for GPIO-based sensors
	s.logger.DebugContext(ctx, "GPIO sensor discovery not yet implemented")
	return nil
}

func (s *SensorMon) getUnitForContext(sensorContext v1alpha1.SensorContext) v1alpha1.SensorUnit {
	switch sensorContext {
	case v1alpha1.SensorContext_SENSOR_CONTEXT_TEMPERATURE:
		return v1alpha1.SensorUnit_SENSOR_UNIT_CELSIUS
	case v1alpha1.SensorContext_SENSOR_CONTEXT_VOLTAGE:
		return v1alpha1.SensorUnit_SENSOR_UNIT_VOLTS
	case v1alpha1.SensorContext_SENSOR_CONTEXT_CURRENT:
		return v1alpha1.SensorUnit_SENSOR_UNIT_AMPS
	case v1alpha1.SensorContext_SENSOR_CONTEXT_TACH:
		return v1alpha1.SensorUnit_SENSOR_UNIT_RPM
	case v1alpha1.SensorContext_SENSOR_CONTEXT_POWER:
		return v1alpha1.SensorUnit_SENSOR_UNIT_WATTS
	case v1alpha1.SensorContext_SENSOR_CONTEXT_ENERGY:
		return v1alpha1.SensorUnit_SENSOR_UNIT_JOULES
	case v1alpha1.SensorContext_SENSOR_CONTEXT_PRESSURE:
		return v1alpha1.SensorUnit_SENSOR_UNIT_PASCALS
	case v1alpha1.SensorContext_SENSOR_CONTEXT_HUMIDITY:
		return v1alpha1.SensorUnit_SENSOR_UNIT_PERCENT
	case v1alpha1.SensorContext_SENSOR_CONTEXT_ALTITUDE:
		return v1alpha1.SensorUnit_SENSOR_UNIT_METERS
	case v1alpha1.SensorContext_SENSOR_CONTEXT_FLOW_RATE:
		return v1alpha1.SensorUnit_SENSOR_UNIT_LITERS_PER_MINUTE
	default:
		return v1alpha1.SensorUnit_SENSOR_UNIT_UNSPECIFIED
	}
}

func (s *SensorMon) registerEndpoints(ctx context.Context) error {
	endpoints := []struct {
		subject string
		handler micro.HandlerFunc
	}{
		{"sensormon.sensors.list", s.createRequestHandler(ctx, s.handleListSensors)},
		{"sensormon.sensor.get", s.createRequestHandler(ctx, s.handleGetSensor)},
		{"sensormon.monitoring.start", s.createRequestHandler(ctx, s.handleStartMonitoring)},
		{"sensormon.monitoring.stop", s.createRequestHandler(ctx, s.handleStopMonitoring)},
		{"sensormon.monitoring.status", s.createRequestHandler(ctx, s.handleMonitoringStatus)},
	}

	for _, ep := range endpoints {
		if err := s.microService.AddEndpoint(ep.subject, ep.handler); err != nil {
			return fmt.Errorf("%w: %s: %w", ErrEndpointRegistrationFailed, ep.subject, err)
		}
	}

	return nil
}

func (s *SensorMon) createRequestHandler(parentCtx context.Context, handler func(context.Context, micro.Request)) micro.HandlerFunc {
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

		if s.tracer != nil {
			_, span := s.tracer.Start(ctx, "sensormon.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", s.config.serviceName),
			)
			defer span.End()
		}

		handler(ctx, req) //nolint:contextcheck
	}
}

func (s *SensorMon) shutdown(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	if s.monitoring {
		close(s.monitoringStop)
		s.monitoring = false
	}

	s.started = false
}

func (s *SensorMon) getSensor(id string) (*sensorInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sensor, exists := s.sensors[id]
	return sensor, exists
}

func (s *SensorMon) getSensorByName(name string) (*sensorInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sensor := range s.sensors {
		if sensor.Sensor.Name == name {
			return sensor, true
		}
	}
	return nil, false
}

// MonitoringStats holds monitoring statistics.
type MonitoringStats struct {
	StartTime    time.Time
	ReadCount    uint64
	ErrorCount   uint64
	LastReadTime time.Time
}
