// SPDX-License-Identifier: BSD-3-Clause

package ledmgr

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/gpio"
	"github.com/u-bmc/u-bmc/pkg/i2c"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var _ service.Service = (*LEDMgr)(nil)

// LEDBackend defines the interface for LED control backends.
type LEDBackend interface {
	SetLEDState(ctx context.Context, componentName string, ledType LEDType, state LEDState) error
	GetLEDState(ctx context.Context, componentName string, ledType LEDType) (LEDState, error)
	Initialize(ctx context.Context, config *config) error
	Close() error
}

// GPIOLEDBackend implements LED control using GPIO lines.
type GPIOLEDBackend struct {
	config     *config
	components map[string]ComponentConfig
	blinkTasks map[string]*blinkTask
	mu         sync.RWMutex
}

type blinkTask struct {
	cancel    context.CancelFunc
	isRunning bool
}

// NewGPIOLEDBackend creates a new GPIO-based LED control backend.
func NewGPIOLEDBackend() *GPIOLEDBackend {
	return &GPIOLEDBackend{
		components: make(map[string]ComponentConfig),
		blinkTasks: make(map[string]*blinkTask),
	}
}

func (b *GPIOLEDBackend) Initialize(ctx context.Context, config *config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config = config
	for name, component := range config.components {
		for _, ledConfig := range component.LEDs {
			if ledConfig.Backend == BackendTypeGPIO {
				b.components[name] = component
				break
			}
		}
	}

	return nil
}

func (b *GPIOLEDBackend) SetLEDState(ctx context.Context, componentName string, ledType LEDType, state LEDState) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	component, exists := b.components[componentName]
	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	ledConfig, exists := component.LEDs[ledType]
	if !exists || !ledConfig.Enabled {
		return fmt.Errorf("%w: LED type '%s' for component '%s'", ErrLEDNotConfigured, ledType, componentName)
	}

	if ledConfig.Backend != BackendTypeGPIO {
		return fmt.Errorf("%w: LED '%s' for component '%s' is not configured for GPIO backend", ErrBackendNotSupported, ledType, componentName)
	}

	blinkKey := fmt.Sprintf("%s:%s", componentName, ledType)

	if task, exists := b.blinkTasks[blinkKey]; exists && task.isRunning {
		task.cancel()
		delete(b.blinkTasks, blinkKey)
	}

	var opts []gpio.Option
	if ledConfig.GPIO.ActiveState == ActiveLow {
		opts = append(opts, gpio.WithActiveLow())
	}

	switch state {
	case LEDStateOff:
		return gpio.SetGPIO(b.config.gpioChip, ledConfig.GPIO.Line, 0, opts...)
	case LEDStateOn:
		return gpio.SetGPIO(b.config.gpioChip, ledConfig.GPIO.Line, 1, opts...)
	case LEDStateBlink:
		return b.startBlinkTask(ctx, blinkKey, ledConfig, component.BlinkInterval, opts...)
	case LEDStateFastBlink:
		return b.startBlinkTask(ctx, blinkKey, ledConfig, component.BlinkInterval/2, opts...)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidLEDState, state)
	}
}

func (b *GPIOLEDBackend) startBlinkTask(ctx context.Context, blinkKey string, ledConfig LEDConfig, interval time.Duration, opts ...gpio.Option) error {
	ctx, cancel := context.WithCancel(ctx)

	task := &blinkTask{
		cancel:    cancel,
		isRunning: true,
	}
	b.blinkTasks[blinkKey] = task

	go func() {
		defer func() {
			b.mu.Lock()
			if t, exists := b.blinkTasks[blinkKey]; exists && t == task {
				t.isRunning = false
			}
			b.mu.Unlock()
		}()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		state := 1
		for {
			select {
			case <-ctx.Done():
				_ = gpio.SetGPIO(b.config.gpioChip, ledConfig.GPIO.Line, 0, opts...)
				return
			case <-ticker.C:
				_ = gpio.SetGPIO(b.config.gpioChip, ledConfig.GPIO.Line, state, opts...)
				state = 1 - state
			}
		}
	}()

	return nil
}

func (b *GPIOLEDBackend) GetLEDState(ctx context.Context, componentName string, ledType LEDType) (LEDState, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	component, exists := b.components[componentName]
	if !exists {
		return LEDStateOff, fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	ledConfig, exists := component.LEDs[ledType]
	if !exists {
		return LEDStateOff, fmt.Errorf("%w: LED type '%s' for component '%s'", ErrLEDNotConfigured, ledType, componentName)
	}

	blinkKey := fmt.Sprintf("%s:%s", componentName, ledType)
	if task, exists := b.blinkTasks[blinkKey]; exists && task.isRunning {
		return LEDStateBlink, nil
	}

	var opts []gpio.Option
	if ledConfig.GPIO.ActiveState == ActiveLow {
		opts = append(opts, gpio.WithActiveLow())
	}

	value, err := gpio.GetGPIO(b.config.gpioChip, ledConfig.GPIO.Line, opts...)
	if err != nil {
		return LEDStateOff, fmt.Errorf("%w: failed to read GPIO: %w", ErrGPIOOperationFailed, err)
	}

	if value == 1 {
		return LEDStateOn, nil
	}
	return LEDStateOff, nil
}

func (b *GPIOLEDBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, task := range b.blinkTasks {
		if task.isRunning {
			task.cancel()
		}
	}
	b.blinkTasks = make(map[string]*blinkTask)

	return nil
}

// I2CLEDBackend implements LED control using I2C communication.
type I2CLEDBackend struct {
	config     *config
	components map[string]ComponentConfig
	blinkTasks map[string]*blinkTask
	mu         sync.RWMutex
}

// NewI2CLEDBackend creates a new I2C-based LED control backend.
func NewI2CLEDBackend() *I2CLEDBackend {
	return &I2CLEDBackend{
		components: make(map[string]ComponentConfig),
		blinkTasks: make(map[string]*blinkTask),
	}
}

func (b *I2CLEDBackend) Initialize(ctx context.Context, config *config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config = config
	for name, component := range config.components {
		for _, ledConfig := range component.LEDs {
			if ledConfig.Backend == BackendTypeI2C {
				b.components[name] = component
				break
			}
		}
	}

	return nil
}

func (b *I2CLEDBackend) SetLEDState(ctx context.Context, componentName string, ledType LEDType, state LEDState) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	component, exists := b.components[componentName]
	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	ledConfig, exists := component.LEDs[ledType]
	if !exists || !ledConfig.Enabled {
		return fmt.Errorf("%w: LED type '%s' for component '%s'", ErrLEDNotConfigured, ledType, componentName)
	}

	if ledConfig.Backend != BackendTypeI2C {
		return fmt.Errorf("%w: LED '%s' for component '%s' is not configured for I2C backend", ErrBackendNotSupported, ledType, componentName)
	}

	blinkKey := fmt.Sprintf("%s:%s", componentName, ledType)

	if task, exists := b.blinkTasks[blinkKey]; exists && task.isRunning {
		task.cancel()
		delete(b.blinkTasks, blinkKey)
	}

	var value uint8
	switch state {
	case LEDStateOff:
		value = ledConfig.I2C.OffValue
	case LEDStateOn:
		value = ledConfig.I2C.OnValue
	case LEDStateBlink, LEDStateFastBlink:
		if ledConfig.I2C.BlinkValue == 0 {
			interval := component.BlinkInterval
			if state == LEDStateFastBlink {
				interval = interval / 2
			}
			return b.startI2CBlinkTask(ctx, blinkKey, ledConfig, interval)
		}
		value = ledConfig.I2C.BlinkValue
	default:
		return fmt.Errorf("%w: %s", ErrInvalidLEDState, state)
	}

	if err := i2c.WriteRegister(ledConfig.I2C.DevicePath, ledConfig.I2C.SlaveAddress, ledConfig.I2C.Register, value); err != nil {
		return fmt.Errorf("%w: failed to write I2C register: %w", ErrI2COperationFailed, err)
	}

	return nil
}

func (b *I2CLEDBackend) startI2CBlinkTask(ctx context.Context, blinkKey string, ledConfig LEDConfig, interval time.Duration) error {
	ctx, cancel := context.WithCancel(ctx)

	task := &blinkTask{
		cancel:    cancel,
		isRunning: true,
	}
	b.blinkTasks[blinkKey] = task

	go func() {
		defer func() {
			b.mu.Lock()
			if t, exists := b.blinkTasks[blinkKey]; exists && t == task {
				t.isRunning = false
			}
			b.mu.Unlock()
		}()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		state := true
		for {
			select {
			case <-ctx.Done():
				_ = i2c.WriteRegister(ledConfig.I2C.DevicePath, ledConfig.I2C.SlaveAddress, ledConfig.I2C.Register, ledConfig.I2C.OffValue)
				return
			case <-ticker.C:
				var value uint8
				if state {
					value = ledConfig.I2C.OnValue
				} else {
					value = ledConfig.I2C.OffValue
				}
				_ = i2c.WriteRegister(ledConfig.I2C.DevicePath, ledConfig.I2C.SlaveAddress, ledConfig.I2C.Register, value)
				state = !state
			}
		}
	}()

	return nil
}

func (b *I2CLEDBackend) GetLEDState(ctx context.Context, componentName string, ledType LEDType) (LEDState, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	component, exists := b.components[componentName]
	if !exists {
		return LEDStateOff, fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	ledConfig, exists := component.LEDs[ledType]
	if !exists {
		return LEDStateOff, fmt.Errorf("%w: LED type '%s' for component '%s'", ErrLEDNotConfigured, ledType, componentName)
	}

	blinkKey := fmt.Sprintf("%s:%s", componentName, ledType)
	if task, exists := b.blinkTasks[blinkKey]; exists && task.isRunning {
		return LEDStateBlink, nil
	}

	value, err := i2c.ReadRegister(ledConfig.I2C.DevicePath, ledConfig.I2C.SlaveAddress, ledConfig.I2C.Register)
	if err != nil {
		return LEDStateOff, fmt.Errorf("%w: failed to read I2C register: %w", ErrI2COperationFailed, err)
	}

	switch value {
	case ledConfig.I2C.OnValue:
		return LEDStateOn, nil
	case ledConfig.I2C.BlinkValue:
		return LEDStateBlink, nil
	default:
		return LEDStateOff, nil
	}
}

func (b *I2CLEDBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, task := range b.blinkTasks {
		if task.isRunning {
			task.cancel()
		}
	}
	b.blinkTasks = make(map[string]*blinkTask)

	return nil
}

// LEDMgr manages LED operations for BMC components.
type LEDMgr struct {
	config       *config
	nc           *nats.Conn
	microService micro.Service
	backends     map[BackendType]LEDBackend
	logger       *slog.Logger
	tracer       trace.Tracer
	meter        metric.Meter
	cancel       context.CancelFunc
	started      bool
	mu           sync.RWMutex

	ledOperationsTotal   metric.Int64Counter
	ledOperationDuration metric.Float64Histogram
	ledFailuresTotal     metric.Int64Counter
	currentLEDStates     metric.Int64UpDownCounter
}

// New creates a new LEDMgr instance with the provided options.
func New(opts ...Option) *LEDMgr {
	cfg := &config{
		serviceName:             DefaultServiceName,
		serviceDescription:      DefaultServiceDescription,
		serviceVersion:          DefaultServiceVersion,
		gpioChip:                DefaultGPIOChip,
		i2cDevice:               DefaultI2CDevice,
		defaultBackend:          BackendTypeGPIO,
		components:              make(map[string]ComponentConfig),
		enableHostManagement:    true,
		enableChassisManagement: true,
		enableBMCManagement:     true,
		numHosts:                1,
		numChassis:              1,
		defaultOperationTimeout: DefaultOperationTimeout,
		defaultBlinkInterval:    DefaultBlinkInterval,
		enableMetrics:           true,
		enableTracing:           true,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &LEDMgr{
		config:   cfg,
		backends: make(map[BackendType]LEDBackend),
	}
}

// Name returns the service name.
func (s *LEDMgr) Name() string {
	return s.config.serviceName
}

// Run starts the LED manager service and registers NATS IPC endpoints.
func (s *LEDMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return ErrServiceAlreadyStarted
	}
	s.started = true
	ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	s.tracer = otel.Tracer(s.config.serviceName)
	s.meter = otel.Meter(s.config.serviceName)

	ctx, span := s.tracer.Start(ctx, "ledmgr.Run")
	defer span.End()

	s.logger = log.GetGlobalLogger().With("service", s.config.serviceName)
	s.logger.InfoContext(ctx, "Starting LED manager service",
		"version", s.config.serviceVersion,
		"hosts", s.config.numHosts,
		"chassis", s.config.numChassis,
		"default_backend", s.config.defaultBackend)

	if err := s.config.Validate(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	if err := s.initializeMetrics(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	s.config.AddDefaultComponents()

	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrNATSConnectionFailed, err)
	}
	s.nc = nc
	defer nc.Drain() //nolint:errcheck

	if err := s.initializeBackends(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrBackendInitializationFailed, err)
	}
	defer s.closeBackends()

	s.microService, err = micro.AddService(nc, micro.Config{
		Name:        s.config.serviceName,
		Description: s.config.serviceDescription,
		Version:     s.config.serviceVersion,
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create micro service: %w", err)
	}

	if err := s.registerEndpoints(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	s.logger.InfoContext(ctx, "LED manager service started successfully",
		"endpoints_registered", true,
		"backends_initialized", len(s.backends))

	span.SetAttributes(
		attribute.String("service.name", s.config.serviceName),
		attribute.String("service.version", s.config.serviceVersion),
		attribute.Int("components.count", len(s.config.components)),
		attribute.String("default.backend", string(s.config.defaultBackend)),
	)

	<-ctx.Done()

	err = ctx.Err()
	ctx = context.WithoutCancel(ctx)
	s.logger.InfoContext(ctx, "Shutting down LED manager service")
	s.shutdown(ctx)

	return err
}

func (s *LEDMgr) initializeMetrics() error {
	if !s.config.enableMetrics {
		return nil
	}

	var err error

	s.ledOperationsTotal, err = s.meter.Int64Counter(
		"ledmgr_operations_total",
		metric.WithDescription("Total number of LED operations"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create LED operations counter: %w", err)
	}

	s.ledOperationDuration, err = s.meter.Float64Histogram(
		"ledmgr_operation_duration_seconds",
		metric.WithDescription("Duration of LED operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create LED operation duration histogram: %w", err)
	}

	s.ledFailuresTotal, err = s.meter.Int64Counter(
		"ledmgr_failures_total",
		metric.WithDescription("Total number of LED operation failures"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create LED failures counter: %w", err)
	}

	s.currentLEDStates, err = s.meter.Int64UpDownCounter(
		"ledmgr_current_states",
		metric.WithDescription("Current LED states (encoded as integer)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create current LED states gauge: %w", err)
	}

	return nil
}

func (s *LEDMgr) initializeBackends(ctx context.Context) error {
	backendTypes := make(map[BackendType]bool)
	for _, component := range s.config.components {
		for _, ledConfig := range component.LEDs {
			backendTypes[ledConfig.Backend] = true
		}
	}

	for backendType := range backendTypes {
		var backend LEDBackend
		switch backendType {
		case BackendTypeGPIO:
			backend = NewGPIOLEDBackend()
		case BackendTypeI2C:
			backend = NewI2CLEDBackend()
		default:
			return fmt.Errorf("%w: unknown backend type '%s'", ErrBackendNotSupported, backendType)
		}

		if err := backend.Initialize(ctx, s.config); err != nil {
			return fmt.Errorf("failed to initialize %s backend: %w", backendType, err)
		}

		s.backends[backendType] = backend
	}

	return nil
}

func (s *LEDMgr) closeBackends() {
	for backendType, backend := range s.backends {
		if err := backend.Close(); err != nil {
			s.logger.Error("Failed to close backend",
				"backend_type", backendType,
				"error", err)
		}
	}
}

func (s *LEDMgr) registerEndpoints(ctx context.Context) error {
	if s.config.enableHostManagement {
		hostGroup := s.microService.AddGroup("host")
		for i := 0; i < s.config.numHosts; i++ {
			controlEndpoint := fmt.Sprintf("%d.control", i)
			statusEndpoint := fmt.Sprintf("%d.status", i)

			if err := hostGroup.AddEndpoint(controlEndpoint,
				micro.HandlerFunc(s.createRequestHandler(ctx, s.handleLEDRequest))); err != nil {
				return fmt.Errorf("failed to register host control endpoint %s: %w", controlEndpoint, err)
			}
			if err := hostGroup.AddEndpoint(statusEndpoint,
				micro.HandlerFunc(s.createRequestHandler(ctx, s.handleLEDRequest))); err != nil {
				return fmt.Errorf("failed to register host status endpoint %s: %w", statusEndpoint, err)
			}
		}
	}

	if s.config.enableChassisManagement {
		chassisGroup := s.microService.AddGroup("chassis")
		for i := 0; i < s.config.numChassis; i++ {
			controlEndpoint := fmt.Sprintf("%d.control", i)
			statusEndpoint := fmt.Sprintf("%d.status", i)

			if err := chassisGroup.AddEndpoint(controlEndpoint,
				micro.HandlerFunc(s.createRequestHandler(ctx, s.handleLEDRequest))); err != nil {
				return fmt.Errorf("failed to register chassis control endpoint %s: %w", controlEndpoint, err)
			}
			if err := chassisGroup.AddEndpoint(statusEndpoint,
				micro.HandlerFunc(s.createRequestHandler(ctx, s.handleLEDRequest))); err != nil {
				return fmt.Errorf("failed to register chassis status endpoint %s: %w", statusEndpoint, err)
			}
		}
	}

	if s.config.enableBMCManagement {
		bmcGroup := s.microService.AddGroup("bmc")

		if err := bmcGroup.AddEndpoint("0.control",
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleLEDRequest))); err != nil {
			return fmt.Errorf("failed to register BMC control endpoint: %w", err)
		}
		if err := bmcGroup.AddEndpoint("0.status",
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleLEDRequest))); err != nil {
			return fmt.Errorf("failed to register BMC status endpoint: %w", err)
		}
	}

	return nil
}

func (s *LEDMgr) createRequestHandler(parentCtx context.Context, handler func(context.Context, micro.Request)) micro.HandlerFunc {
	return func(req micro.Request) {
		ctx := telemetry.GetCtxFromReq(req)

		ctx = context.WithoutCancel(ctx)
		if parentCtx != nil {
			select {
			case <-parentCtx.Done():
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			default:
			}
		}

		if s.tracer != nil {
			_, span := s.tracer.Start(ctx, "ledmgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", s.config.serviceName),
			)
			defer span.End()
		}

		handler(ctx, req) //nolint:contextcheck
	}
}

func (s *LEDMgr) handleLEDRequest(ctx context.Context, req micro.Request) {
	start := time.Now()

	if s.tracer != nil {
		var span trace.Span
		_, span = s.tracer.Start(ctx, "ledmgr.handleLEDRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 4 || parts[0] != "ledmgr" {
		s.respondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	componentType := parts[1]
	componentID := parts[2]
	operation := parts[3]

	componentName := fmt.Sprintf("%s.%s", componentType, componentID)

	switch operation {
	case "control":
		s.handleLEDControl(ctx, req, componentName, start)
	case "status":
		s.handleLEDStatus(ctx, req, componentName, start)
	default:
		s.respondWithError(ctx, req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

func (s *LEDMgr) handleLEDControl(ctx context.Context, req micro.Request, componentName string, start time.Time) {
	var controlReq v1alpha1.LEDControlRequest
	if err := controlReq.UnmarshalVT(req.Data()); err != nil {
		s.respondWithError(ctx, req, ErrLEDOperationFailed, fmt.Sprintf("Failed to unmarshal LED control request: %v", err))
		return
	}

	if controlReq.ComponentName != componentName {
		s.respondWithError(ctx, req, ErrLEDOperationFailed, fmt.Sprintf("Component name mismatch: expected %s, got %s", componentName, controlReq.ComponentName))
		return
	}

	ledType := s.convertProtoLEDType(controlReq.LedType)
	ledState := s.convertProtoLEDState(controlReq.LedState)

	err := s.setLEDState(ctx, componentName, ledType, ledState)

	s.recordOperation(ctx, "control", componentName, string(ledType), string(ledState), err)
	if s.config.enableMetrics && s.ledOperationDuration != nil {
		duration := time.Since(start).Seconds()
		s.ledOperationDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", "control"),
			attribute.String("component", componentName),
			attribute.String("led_type", string(ledType)),
		))
	}

	response := &v1alpha1.LEDControlResponse{
		Success: err == nil,
		Message: "LED control completed successfully",
	}

	if err != nil {
		response.Message = fmt.Sprintf("LED control failed: %v", err)
		response.CurrentState = v1alpha1.LEDState_LED_STATE_UNSPECIFIED
	} else {
		currentState, _ := s.getLEDState(ctx, componentName, ledType)
		response.CurrentState = s.convertToProtoLEDState(currentState)
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		s.respondWithError(ctx, req, ErrLEDOperationFailed, fmt.Sprintf("Failed to marshal response: %v", err))
		return
	}

	if err := req.Respond(responseData); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if s.logger != nil {
		s.logger.InfoContext(ctx, "LED control completed",
			"component", componentName,
			"led_type", ledType,
			"state", ledState,
			"success", response.Success,
			"duration", time.Since(start))
	}
}

func (s *LEDMgr) handleLEDStatus(ctx context.Context, req micro.Request, componentName string, start time.Time) {
	var statusReq v1alpha1.LEDStatusRequest
	if err := statusReq.UnmarshalVT(req.Data()); err != nil {
		s.respondWithError(ctx, req, ErrLEDOperationFailed, fmt.Sprintf("Failed to unmarshal LED status request: %v", err))
		return
	}

	if statusReq.ComponentName != componentName {
		s.respondWithError(ctx, req, ErrLEDOperationFailed, fmt.Sprintf("Component name mismatch: expected %s, got %s", componentName, statusReq.ComponentName))
		return
	}

	ledType := s.convertProtoLEDType(statusReq.LedType)

	state, err := s.getLEDState(ctx, componentName, ledType)

	s.recordOperation(ctx, "status", componentName, string(ledType), string(state), err)
	if s.config.enableMetrics && s.ledOperationDuration != nil {
		duration := time.Since(start).Seconds()
		s.ledOperationDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", "status"),
			attribute.String("component", componentName),
			attribute.String("led_type", string(ledType)),
		))
	}

	if err != nil {
		s.respondWithError(ctx, req, ErrLEDOperationFailed, err.Error())
		return
	}

	hardwareInfo := fmt.Sprintf("LED %s on component %s", ledType, componentName)
	response := &v1alpha1.LEDStatusResponse{
		CurrentState:    s.convertToProtoLEDState(state),
		IsBlinking:      state == LEDStateBlink || state == LEDStateFastBlink,
		BlinkIntervalMs: 1000,
		Controllable:    true,
		HardwareInfo:    &hardwareInfo,
	}

	if state == LEDStateFastBlink {
		response.BlinkIntervalMs = 500
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		s.respondWithError(ctx, req, ErrLEDOperationFailed, fmt.Sprintf("Failed to marshal response: %v", err))
		return
	}

	if err := req.Respond(responseData); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if s.logger != nil {
		s.logger.InfoContext(ctx, "LED status query completed",
			"component", componentName,
			"led_type", ledType,
			"state", state,
			"duration", time.Since(start))
	}
}

func (s *LEDMgr) setLEDState(ctx context.Context, componentName string, ledType LEDType, state LEDState) error {
	component, exists := s.config.GetComponentConfig(componentName)
	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	ledConfig, exists := component.LEDs[ledType]
	if !exists {
		return fmt.Errorf("%w: LED type '%s' for component '%s'", ErrLEDNotConfigured, ledType, componentName)
	}

	backend, exists := s.backends[ledConfig.Backend]
	if !exists {
		return fmt.Errorf("%w: backend '%s' for LED '%s' of component '%s'", ErrBackendNotConfigured, ledConfig.Backend, ledType, componentName)
	}

	return backend.SetLEDState(ctx, componentName, ledType, state)
}

func (s *LEDMgr) getLEDState(ctx context.Context, componentName string, ledType LEDType) (LEDState, error) {
	component, exists := s.config.GetComponentConfig(componentName)
	if !exists {
		return LEDStateOff, fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	ledConfig, exists := component.LEDs[ledType]
	if !exists {
		return LEDStateOff, fmt.Errorf("%w: LED type '%s' for component '%s'", ErrLEDNotConfigured, ledType, componentName)
	}

	backend, exists := s.backends[ledConfig.Backend]
	if !exists {
		return LEDStateOff, fmt.Errorf("%w: backend '%s' for LED '%s' of component '%s'", ErrBackendNotConfigured, ledConfig.Backend, ledType, componentName)
	}

	return backend.GetLEDState(ctx, componentName, ledType)
}

func (s *LEDMgr) recordOperation(ctx context.Context, operation, component, ledType, ledState string, err error) {
	if !s.config.enableMetrics {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("component", component),
		attribute.String("led_type", ledType),
	}

	if ledState != "" {
		attrs = append(attrs, attribute.String("led_state", ledState))
	}

	if err != nil {
		attrs = append(attrs, attribute.String("status", "error"))
		s.ledFailuresTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	} else {
		attrs = append(attrs, attribute.String("status", "success"))
		if s.currentLEDStates != nil {
			s.currentLEDStates.Add(ctx, 1, metric.WithAttributes(attrs...))
		}
	}

	s.ledOperationsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (s *LEDMgr) respondWithError(ctx context.Context, req micro.Request, err error, message string) {
	response := &v1alpha1.LEDControlResponse{
		Success:      false,
		Message:      message,
		CurrentState: v1alpha1.LEDState_LED_STATE_UNSPECIFIED,
	}

	responseData, marshalErr := response.MarshalVT()
	if marshalErr != nil {
		if reqErr := req.Respond([]byte(fmt.Sprintf("ERROR: %s", message))); reqErr != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "Failed to send error response", "error", reqErr)
		}
		return
	}

	if reqErr := req.Respond(responseData); reqErr != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to send error response", "error", reqErr)
	}
}

func (s *LEDMgr) shutdown(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	if ctx.Err() != nil {
		ctx = context.WithoutCancel(ctx)
	}
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.closeBackends()
	s.started = false
}

// WithName is a backward compatibility alias for WithServiceName.
func WithName(name string) Option {
	return WithServiceName(name)
}

// ErrLEDNotConfigured indicates LED is not configured for the component.
var ErrLEDNotConfigured = errors.New("LED not configured for component")

func (s *LEDMgr) convertProtoLEDType(protoType v1alpha1.LEDType) LEDType {
	switch protoType {
	case v1alpha1.LEDType_LED_TYPE_POWER:
		return LEDTypePower
	case v1alpha1.LEDType_LED_TYPE_STATUS:
		return LEDTypeStatus
	case v1alpha1.LEDType_LED_TYPE_ERROR:
		return LEDTypeError
	case v1alpha1.LEDType_LED_TYPE_IDENTIFY:
		return LEDTypeIdentify
	default:
		return LEDTypePower
	}
}

func (s *LEDMgr) convertProtoLEDState(protoState v1alpha1.LEDState) LEDState {
	switch protoState {
	case v1alpha1.LEDState_LED_STATE_OFF:
		return LEDStateOff
	case v1alpha1.LEDState_LED_STATE_ON:
		return LEDStateOn
	case v1alpha1.LEDState_LED_STATE_BLINK:
		return LEDStateBlink
	case v1alpha1.LEDState_LED_STATE_FAST_BLINK:
		return LEDStateFastBlink
	default:
		return LEDStateOff
	}
}

func (s *LEDMgr) convertToProtoLEDState(state LEDState) v1alpha1.LEDState {
	switch state {
	case LEDStateOff:
		return v1alpha1.LEDState_LED_STATE_OFF
	case LEDStateOn:
		return v1alpha1.LEDState_LED_STATE_ON
	case LEDStateBlink:
		return v1alpha1.LEDState_LED_STATE_BLINK
	case LEDStateFastBlink:
		return v1alpha1.LEDState_LED_STATE_FAST_BLINK
	default:
		return v1alpha1.LEDState_LED_STATE_UNSPECIFIED
	}
}
