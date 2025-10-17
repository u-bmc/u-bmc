// SPDX-License-Identifier: BSD-3-Clause

package powermgr

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/gpio"
	"github.com/u-bmc/u-bmc/pkg/i2c"
	"github.com/u-bmc/u-bmc/pkg/ipc"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ service.Service = (*PowerMgr)(nil)

// PowerBackend defines the interface for power control backends.
type PowerBackend interface {
	PowerOn(ctx context.Context, componentName string) error
	PowerOff(ctx context.Context, componentName string, force bool) error
	Reset(ctx context.Context, componentName string) error
	GetPowerStatus(ctx context.Context, componentName string) (bool, error)
	Initialize(ctx context.Context, config *config) error
	Close() error
}

// GPIOBackend implements power control using GPIO lines.
type GPIOBackend struct {
	config     *config
	components map[string]ComponentConfig
	mu         sync.RWMutex
}

// NewGPIOBackend creates a new GPIO-based power control backend.
func NewGPIOBackend() *GPIOBackend {
	return &GPIOBackend{
		components: make(map[string]ComponentConfig),
	}
}

func (b *GPIOBackend) Initialize(ctx context.Context, config *config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config = config
	for name, component := range config.components {
		if component.Backend == BackendTypeGPIO {
			b.components[name] = component
		}
	}

	return nil
}

func (b *GPIOBackend) PowerOn(ctx context.Context, componentName string) error {
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	if !component.Enabled {
		return fmt.Errorf("%w: component '%s'", ErrComponentDisabled, componentName)
	}

	if component.GPIO.PowerButton.Line == "" {
		return fmt.Errorf("%w: power button not configured for component '%s'", ErrGPIONotConfigured, componentName)
	}

	var opts []gpio.Option
	if component.GPIO.PowerButton.ActiveState == ActiveLow {
		opts = append(opts, gpio.WithActiveLow())
	}

	return gpio.ToggleGPIOCtx(ctx, b.config.gpioChip, component.GPIO.PowerButton.Line, component.PowerOnDelay, opts...)
}

func (b *GPIOBackend) PowerOff(ctx context.Context, componentName string, force bool) error { //nolint:revive
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	if !component.Enabled {
		return fmt.Errorf("%w: component '%s'", ErrComponentDisabled, componentName)
	}

	if component.GPIO.PowerButton.Line == "" {
		return fmt.Errorf("%w: power button not configured for component '%s'", ErrGPIONotConfigured, componentName)
	}

	var opts []gpio.Option
	if component.GPIO.PowerButton.ActiveState == ActiveLow {
		opts = append(opts, gpio.WithActiveLow())
	}

	if force {
		line, err := gpio.RequestLine(b.config.gpioChip, component.GPIO.PowerButton.Line, append(opts, gpio.AsOutput())...)
		if err != nil {
			return fmt.Errorf("%w: failed to request power button GPIO: %w", ErrGPIOOperationFailed, err)
		}
		defer line.Close()

		if err := line.SetValue(1); err != nil {
			return fmt.Errorf("%w: failed to set GPIO high: %w", ErrGPIOOperationFailed, err)
		}

		select {
		case <-time.After(component.ForceOffDelay):
		case <-ctx.Done():
			_ = line.SetValue(0)
			return ctx.Err()
		}

		if err := line.SetValue(0); err != nil {
			return fmt.Errorf("%w: failed to set GPIO low: %w", ErrGPIOOperationFailed, err)
		}
		return nil
	}

	return gpio.ToggleGPIOCtx(ctx, b.config.gpioChip, component.GPIO.PowerButton.Line, component.PowerOffDelay, opts...)
}

func (b *GPIOBackend) Reset(ctx context.Context, componentName string) error {
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	if !component.Enabled {
		return fmt.Errorf("%w: component '%s'", ErrComponentDisabled, componentName)
	}

	if component.GPIO.ResetButton.Line == "" {
		return fmt.Errorf("%w: reset button not configured for component '%s'", ErrGPIONotConfigured, componentName)
	}

	var opts []gpio.Option
	if component.GPIO.ResetButton.ActiveState == ActiveLow {
		opts = append(opts, gpio.WithActiveLow())
	}

	return gpio.ToggleGPIOCtx(ctx, b.config.gpioChip, component.GPIO.ResetButton.Line, component.ResetDelay, opts...)
}

func (b *GPIOBackend) GetPowerStatus(ctx context.Context, componentName string) (bool, error) {
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	if component.GPIO.PowerStatus.Line == "" {
		return false, fmt.Errorf("%w: power status not configured for component '%s'", ErrGPIONotConfigured, componentName)
	}

	var opts []gpio.Option
	if component.GPIO.PowerStatus.ActiveState == ActiveLow {
		opts = append(opts, gpio.WithActiveLow())
	}

	value, err := gpio.GetGPIO(b.config.gpioChip, component.GPIO.PowerStatus.Line, opts...)
	if err != nil {
		return false, fmt.Errorf("%w: failed to read power status: %w", ErrGPIOOperationFailed, err)
	}

	return value == 1, nil
}

func (b *GPIOBackend) Close() error {
	return nil
}

// I2CBackend implements power control using I2C communication.
type I2CBackend struct {
	config     *config
	components map[string]ComponentConfig
	mu         sync.RWMutex
}

// NewI2CBackend creates a new I2C-based power control backend.
func NewI2CBackend() *I2CBackend {
	return &I2CBackend{
		components: make(map[string]ComponentConfig),
	}
}

func (b *I2CBackend) Initialize(ctx context.Context, config *config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config = config
	for name, component := range config.components {
		if component.Backend == BackendTypeI2C {
			b.components[name] = component
		}
	}

	return nil
}

func (b *I2CBackend) PowerOn(ctx context.Context, componentName string) error {
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	if !component.Enabled {
		return fmt.Errorf("%w: component '%s'", ErrComponentDisabled, componentName)
	}

	if err := i2c.WriteRegister(component.I2C.DevicePath, component.I2C.SlaveAddress, component.I2C.PowerOnReg, component.I2C.PowerOnValue); err != nil {
		return fmt.Errorf("%w: failed to write power on register: %w", ErrI2COperationFailed, err)
	}

	return nil
}

func (b *I2CBackend) PowerOff(ctx context.Context, componentName string, force bool) error {
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	if !component.Enabled {
		return fmt.Errorf("%w: component '%s'", ErrComponentDisabled, componentName)
	}

	if err := i2c.WriteRegister(component.I2C.DevicePath, component.I2C.SlaveAddress, component.I2C.PowerOffReg, component.I2C.PowerOffValue); err != nil {
		return fmt.Errorf("%w: failed to write power off register: %w", ErrI2COperationFailed, err)
	}

	return nil
}

func (b *I2CBackend) Reset(ctx context.Context, componentName string) error {
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	if !component.Enabled {
		return fmt.Errorf("%w: component '%s'", ErrComponentDisabled, componentName)
	}

	if err := i2c.WriteRegister(component.I2C.DevicePath, component.I2C.SlaveAddress, component.I2C.ResetReg, component.I2C.ResetValue); err != nil {
		return fmt.Errorf("%w: failed to write reset register: %w", ErrI2COperationFailed, err)
	}

	return nil
}

func (b *I2CBackend) GetPowerStatus(ctx context.Context, componentName string) (bool, error) {
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	value, err := i2c.ReadRegister(component.I2C.DevicePath, component.I2C.SlaveAddress, component.I2C.StatusReg)
	if err != nil {
		return false, fmt.Errorf("%w: failed to read status register: %w", ErrI2COperationFailed, err)
	}

	return value != 0, nil
}

func (b *I2CBackend) Close() error {
	return nil
}

// MockBackend implements a mock power management backend for testing.
type MockBackend struct {
	config      *config
	components  map[string]ComponentConfig
	powerStates map[string]bool
	mu          sync.RWMutex
	callbacks   PowerCallbacks
}

// NewMockBackend creates a new mock power control backend.
func NewMockBackend() *MockBackend {
	return &MockBackend{
		components:  make(map[string]ComponentConfig),
		powerStates: make(map[string]bool),
	}
}

// Initialize configures the mock backend.
func (b *MockBackend) Initialize(ctx context.Context, cfg *config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config = cfg
	b.callbacks = cfg.callbacks

	// Initialize power states for all components
	for name, component := range cfg.components {
		b.components[name] = component
		if component.Mock != nil {
			b.powerStates[name] = component.Mock.InitialPowerState
		} else {
			b.powerStates[name] = false
		}
	}

	return nil
}

// PowerOn simulates powering on a component.
func (b *MockBackend) PowerOn(ctx context.Context, componentName string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	component, exists := b.components[componentName]
	if !exists {
		return fmt.Errorf("component %s not found", componentName)
	}

	if component.Mock == nil {
		return fmt.Errorf("component %s has no mock configuration", componentName)
	}

	// Simulate operation delay
	if component.Mock.OperationDelay > 0 {
		time.Sleep(component.Mock.OperationDelay)
	}

	// Simulate failure if configured
	if !component.Mock.AlwaysSucceed && component.Mock.FailureRate > 0 {
		if shouldSimulateFailure(component.Mock.FailureRate) {
			err := fmt.Errorf("simulated power on failure for %s", componentName)
			if b.callbacks.OnOperationFailed != nil {
				b.callbacks.OnOperationFailed(componentName, EventPowerOn, err)
			}
			return err
		}
	}

	// Simulate power state change delay
	if component.Mock.PowerStateDelay > 0 {
		go func() {
			time.Sleep(component.Mock.PowerStateDelay)
			b.mu.Lock()
			b.powerStates[componentName] = true
			b.mu.Unlock()
			if b.callbacks.OnPowerStateChanged != nil {
				b.callbacks.OnPowerStateChanged(componentName, EventPowerStateChanged, true)
			}
		}()
	} else {
		b.powerStates[componentName] = true
		if b.callbacks.OnPowerStateChanged != nil {
			b.callbacks.OnPowerStateChanged(componentName, EventPowerStateChanged, true)
		}
	}

	if b.callbacks.OnPowerOn != nil {
		b.callbacks.OnPowerOn(componentName, EventPowerOn, nil)
	}

	return nil
}

// PowerOff simulates powering off a component.
func (b *MockBackend) PowerOff(ctx context.Context, componentName string, force bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	component, exists := b.components[componentName]
	if !exists {
		return fmt.Errorf("component %s not found", componentName)
	}

	if component.Mock == nil {
		return fmt.Errorf("component %s has no mock configuration", componentName)
	}

	operationDelay := component.Mock.OperationDelay
	failureRate := component.Mock.FailureRate

	// Force operations are faster and more reliable
	if force {
		operationDelay = operationDelay / 2
		failureRate = failureRate * 0.1
	}

	// Simulate operation delay
	if operationDelay > 0 {
		time.Sleep(operationDelay)
	}

	// Simulate failure if configured
	if !component.Mock.AlwaysSucceed && failureRate > 0 {
		if shouldSimulateFailure(failureRate) {
			err := fmt.Errorf("simulated power off failure for %s", componentName)
			if b.callbacks.OnOperationFailed != nil {
				b.callbacks.OnOperationFailed(componentName, EventPowerOff, err)
			}
			return err
		}
	}

	// Force operations have immediate effect
	if force {
		b.powerStates[componentName] = false
		if b.callbacks.OnPowerStateChanged != nil {
			b.callbacks.OnPowerStateChanged(componentName, EventPowerStateChanged, false)
		}
		if b.callbacks.OnForceOff != nil {
			b.callbacks.OnForceOff(componentName, EventForceOff, nil)
		}
	} else {
		// Simulate power state change delay
		if component.Mock.PowerStateDelay > 0 {
			go func() {
				time.Sleep(component.Mock.PowerStateDelay)
				b.mu.Lock()
				b.powerStates[componentName] = false
				b.mu.Unlock()
				if b.callbacks.OnPowerStateChanged != nil {
					b.callbacks.OnPowerStateChanged(componentName, EventPowerStateChanged, false)
				}
			}()
		} else {
			b.powerStates[componentName] = false
			if b.callbacks.OnPowerStateChanged != nil {
				b.callbacks.OnPowerStateChanged(componentName, EventPowerStateChanged, false)
			}
		}
		if b.callbacks.OnPowerOff != nil {
			b.callbacks.OnPowerOff(componentName, EventPowerOff, nil)
		}
	}

	return nil
}

// Reset simulates resetting a component.
func (b *MockBackend) Reset(ctx context.Context, componentName string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	component, exists := b.components[componentName]
	if !exists {
		return fmt.Errorf("component %s not found", componentName)
	}

	if component.Mock == nil {
		return fmt.Errorf("component %s has no mock configuration", componentName)
	}

	// Simulate operation delay
	if component.Mock.OperationDelay > 0 {
		time.Sleep(component.Mock.OperationDelay)
	}

	// Simulate failure if configured
	if !component.Mock.AlwaysSucceed && component.Mock.FailureRate > 0 {
		if shouldSimulateFailure(component.Mock.FailureRate) {
			err := fmt.Errorf("simulated reset failure for %s", componentName)
			if b.callbacks.OnOperationFailed != nil {
				b.callbacks.OnOperationFailed(componentName, EventReset, err)
			}
			return err
		}
	}

	// For reset, simulate a brief power cycle
	originalState := b.powerStates[componentName]
	if component.Mock.PowerStateDelay > 0 {
		go func() {
			// Brief power down
			time.Sleep(component.Mock.PowerStateDelay / 2)
			b.mu.Lock()
			b.powerStates[componentName] = false
			b.mu.Unlock()

			// Power back up to original state
			time.Sleep(component.Mock.PowerStateDelay / 2)
			b.mu.Lock()
			b.powerStates[componentName] = originalState
			b.mu.Unlock()
			if b.callbacks.OnPowerStateChanged != nil {
				b.callbacks.OnPowerStateChanged(componentName, EventPowerStateChanged, originalState)
			}
		}()
	}

	if b.callbacks.OnReset != nil {
		b.callbacks.OnReset(componentName, EventReset, nil)
	}

	return nil
}

// GetPowerStatus returns the simulated power status of a component.
func (b *MockBackend) GetPowerStatus(ctx context.Context, componentName string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	component, exists := b.components[componentName]
	if !exists {
		return false, fmt.Errorf("component %s not found", componentName)
	}

	if component.Mock == nil {
		return false, fmt.Errorf("component %s has no mock configuration", componentName)
	}

	// Simulate small read failure rate
	if !component.Mock.AlwaysSucceed && component.Mock.FailureRate > 0 {
		readFailureRate := component.Mock.FailureRate * 0.05 // Very low failure rate for reads
		if shouldSimulateFailure(readFailureRate) {
			return false, fmt.Errorf("simulated power status read failure for %s", componentName)
		}
	}

	state, exists := b.powerStates[componentName]
	if !exists {
		return false, nil
	}

	return state, nil
}

// Close cleans up the mock backend.
func (b *MockBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.components = make(map[string]ComponentConfig)
	b.powerStates = make(map[string]bool)
	return nil
}

// shouldSimulateFailure determines if a failure should be simulated based on failure rate.
func shouldSimulateFailure(failureRate float64) bool {
	// Simple pseudo-random based on current time
	return float64(time.Now().UnixNano()%1000)/1000.0 < failureRate
}

// PowerMgr manages power operations for BMC components.
type PowerMgr struct {
	config       *config
	nc           *nats.Conn
	microService micro.Service
	backends     map[BackendType]PowerBackend
	logger       *slog.Logger
	tracer       trace.Tracer
	meter        metric.Meter
	cancel       context.CancelFunc
	started      bool
	mu           sync.RWMutex

	// Metrics
	powerOperationsTotal   metric.Int64Counter
	powerOperationDuration metric.Float64Histogram
	powerCyclesTotal       metric.Int64Counter
	powerFailuresTotal     metric.Int64Counter
}

// New creates a new PowerMgr instance with the provided options.
func New(opts ...Option) *PowerMgr {
	cfg := &config{
		serviceName:                 DefaultServiceName,
		serviceDescription:          DefaultServiceDescription,
		serviceVersion:              DefaultServiceVersion,
		gpioChip:                    DefaultGPIOChip,
		i2cDevice:                   DefaultI2CDevice,
		defaultBackend:              BackendTypeGPIO,
		components:                  make(map[string]ComponentConfig),
		enableHostManagement:        true,
		enableChassisManagement:     true,
		enableBMCManagement:         true,
		numHosts:                    1,
		numChassis:                  1,
		defaultOperationTimeout:     DefaultOperationTimeout,
		enableStateReporting:        true,
		stateReportingSubjectPrefix: "statemgr",
		enableThermalResponse:       false,
		emergencyResponseDelay:      5 * time.Second,
		enableEmergencyShutdown:     false,
		shutdownTemperatureLimit:    85.0,
		shutdownComponents:          []string{},
		maxEmergencyAttempts:        3,
		emergencyAttemptInterval:    30 * time.Second,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	addDefaultComponents(cfg)

	return &PowerMgr{
		config:   cfg,
		backends: make(map[BackendType]PowerBackend),
	}
}

// Name returns the service name.
func (p *PowerMgr) Name() string {
	return p.config.serviceName
}

// Run starts the power manager service and registers NATS IPC endpoints.
func (p *PowerMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	p.tracer = otel.Tracer(p.config.serviceName)
	p.meter = otel.Meter(p.config.serviceName)

	ctx, span := p.tracer.Start(ctx, "powermgr.Run")
	defer span.End()

	p.logger = log.GetGlobalLogger().With("service", p.config.serviceName)

	// p.mu.Lock()
	// if p.started {
	// 	p.mu.Unlock()
	// 	return ErrServiceAlreadyStarted
	// }
	// p.started = true
	// ctx, p.cancel = context.WithCancel(ctx)
	// p.mu.Unlock()
	p.logger.InfoContext(ctx, "Starting power manager service",
		"version", p.config.serviceVersion,
		"hosts", p.config.numHosts,
		"chassis", p.config.numChassis,
		"default_backend", p.config.defaultBackend)

	if err := p.config.Validate(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	if err := p.initializeMetrics(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	addDefaultComponents(p.config)

	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrNATSConnectionFailed, err)
	}
	p.nc = nc
	defer nc.Drain() //nolint:errcheck

	if err := p.initializeBackends(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrBackendInitializationFailed, err)
	}
	defer p.closeBackends()

	if err := p.initializeThermalIntegration(ctx); err != nil {
		span.RecordError(err)
		p.logger.WarnContext(ctx, "Thermal integration initialization failed", "error", err)
	}

	p.microService, err = micro.AddService(nc, micro.Config{
		Name:        p.config.serviceName,
		Description: p.config.serviceDescription,
		Version:     p.config.serviceVersion,
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create micro service: %w", err)
	}

	if err := p.registerEndpoints(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	if err := p.subscribeInternalSubjects(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	p.logger.InfoContext(ctx, "Power manager service started successfully",
		"endpoints_registered", true,
		"backends_initialized", len(p.backends))

	span.SetAttributes(
		attribute.String("service.name", p.config.serviceName),
		attribute.String("service.version", p.config.serviceVersion),
		attribute.Int("components.count", len(p.config.components)),
		attribute.String("default.backend", string(p.config.defaultBackend)),
	)

	<-ctx.Done()

	err = ctx.Err()
	ctx = context.WithoutCancel(ctx)
	p.logger.InfoContext(ctx, "Shutting down power manager service")
	p.shutdown(ctx)

	return err
}

func (p *PowerMgr) initializeMetrics() error {
	var err error

	p.powerOperationsTotal, err = p.meter.Int64Counter(
		"powermgr_operations_total",
		metric.WithDescription("Total number of power operations"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create power operations counter: %w", err)
	}

	p.powerOperationDuration, err = p.meter.Float64Histogram(
		"powermgr_operation_duration_seconds",
		metric.WithDescription("Duration of power operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create power operation duration histogram: %w", err)
	}

	p.powerCyclesTotal, err = p.meter.Int64Counter(
		"powermgr_power_cycles_total",
		metric.WithDescription("Total number of power cycles"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create power cycles counter: %w", err)
	}

	p.powerFailuresTotal, err = p.meter.Int64Counter(
		"powermgr_failures_total",
		metric.WithDescription("Total number of power operation failures"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create power failures counter: %w", err)
	}

	return nil
}

func (p *PowerMgr) initializeBackends(ctx context.Context) error {
	backendTypes := make(map[BackendType]bool)
	for _, component := range p.config.components {
		backendTypes[component.Backend] = true
	}

	for backendType := range backendTypes {
		var backend PowerBackend
		switch backendType {
		case BackendTypeGPIO:
			backend = NewGPIOBackend()
		case BackendTypeI2C:
			backend = NewI2CBackend()
		case BackendTypeMock:
			backend = NewMockBackend()
		default:
			return fmt.Errorf("%w: unknown backend type '%s'", ErrBackendNotSupported, backendType)
		}

		if err := backend.Initialize(ctx, p.config); err != nil {
			return fmt.Errorf("failed to initialize %s backend: %w", backendType, err)
		}

		p.backends[backendType] = backend
	}

	return nil
}

func (p *PowerMgr) closeBackends() {
	for backendType, backend := range p.backends {
		if err := backend.Close(); err != nil {
			p.logger.Error("Failed to close backend",
				"backend_type", backendType,
				"error", err)
		}
	}
}

func (p *PowerMgr) registerEndpoints(ctx context.Context) error {
	groups := make(map[string]micro.Group)

	// Register general power endpoints that handle all instances
	if err := ipc.RegisterEndpointWithGroupCache(p.microService, ipc.SubjectPowerAction,
		micro.HandlerFunc(p.createRequestHandler(ctx, p.handleGeneralPowerAction)), groups); err != nil {
		return fmt.Errorf("failed to register power action endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(p.microService, ipc.SubjectPowerResult,
		micro.HandlerFunc(p.createRequestHandler(ctx, p.handleGeneralPowerResult)), groups); err != nil {
		return fmt.Errorf("failed to register power result endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(p.microService, ipc.SubjectPowerStatus,
		micro.HandlerFunc(p.createRequestHandler(ctx, p.handleGeneralPowerStatus)), groups); err != nil {
		return fmt.Errorf("failed to register power status endpoint: %w", err)
	}

	return nil
}

// subscribeInternalSubjects subscribes to internal coordination subjects.
func (p *PowerMgr) subscribeInternalSubjects(ctx context.Context) error {
	// Subscribe to internal power action requests from statemgr
	if _, err := p.nc.Subscribe(ipc.InternalPowerAction, p.handleInternalPowerAction); err != nil {
		return fmt.Errorf("failed to subscribe to internal power action: %w", err)
	}

	p.logger.InfoContext(ctx, "Subscribed to internal coordination subjects",
		"power_actions", ipc.InternalPowerAction)

	return nil
}

// handleInternalPowerAction handles internal power action requests from statemgr.
func (p *PowerMgr) handleInternalPowerAction(msg *nats.Msg) {
	ctx := context.Background()

	var request v1alpha1.PowerControlRequest
	if err := proto.Unmarshal(msg.Data, &request); err != nil {
		p.logger.WarnContext(ctx, "Invalid internal power control request", "error", err)
		return
	}

	componentName := request.ComponentName
	action := request.Action

	p.logger.InfoContext(ctx, "Received internal power action request",
		"component", componentName,
		"action", action)

	var err error
	switch action {
	case "power_on":
		err = p.performPowerAction(ctx, componentName, "on", false)
	case "power_off":
		err = p.performPowerAction(ctx, componentName, "off", false)
	case "force_off":
		err = p.performPowerAction(ctx, componentName, "off", true)
	case "reset":
		err = p.performPowerAction(ctx, componentName, "reset", false)
	default:
		p.logger.WarnContext(ctx, "Unknown power action",
			"component", componentName,
			"action", action)
		return
	}

	// Report the result back to statemgr
	p.reportStateChange(ctx, componentName, action, err == nil)

	if err != nil {
		p.logger.ErrorContext(ctx, "Internal power action failed",
			"component", componentName,
			"action", action,
			"error", err)
	} else {
		p.logger.InfoContext(ctx, "Internal power action completed",
			"component", componentName,
			"action", action)
	}
}

// performPowerAction performs a power action on a component using the appropriate backend.
func (p *PowerMgr) performPowerAction(ctx context.Context, componentName, action string, force bool) error {
	backend, err := p.getBackendForComponent(componentName)
	if err != nil {
		return fmt.Errorf("backend not configured for component %s: %w", componentName, err)
	}

	switch action {
	case "on", "power_on":
		return backend.PowerOn(ctx, componentName)
	case "off", "power_off":
		return backend.PowerOff(ctx, componentName, force)
	case "reset", "reboot":
		return backend.Reset(ctx, componentName)
	default:
		return fmt.Errorf("unsupported power action: %s", action)
	}
}

// handleGeneralPowerAction is a general handler that dispatches to specific component handlers
// based on the message content instead of the subject
func (p *PowerMgr) handleGeneralPowerAction(ctx context.Context, req micro.Request) {
	// Try to parse as different message types and dispatch accordingly

	// Try host power action first
	var hostRequest v1alpha1.ChangeHostStateRequest
	if err := hostRequest.UnmarshalVT(req.Data()); err == nil && hostRequest.HostName != "" {
		p.handleHostPowerActionFromGeneral(ctx, req, &hostRequest)
		return
	}

	// Try chassis power action
	var chassisRequest v1alpha1.ChangeChassisStateRequest
	if err := chassisRequest.UnmarshalVT(req.Data()); err == nil && chassisRequest.ChassisName != "" {
		p.handleChassisPowerActionFromGeneral(ctx, req, &chassisRequest)
		return
	}

	// Try BMC power action - BMC requests might have a different structure
	// For now, assume any other valid power request is for BMC
	var bmcRequest v1alpha1.ChangeHostStateRequest // BMC might use similar structure
	if err := bmcRequest.UnmarshalVT(req.Data()); err == nil {
		p.handleBMCPowerActionFromGeneral(ctx, req, &bmcRequest)
		return
	}

	ipc.RespondWithError(ctx, req, ipc.ErrInvalidRequest, "unable to parse power action request")
}

// handleGeneralPowerResult handles power operation results
func (p *PowerMgr) handleGeneralPowerResult(ctx context.Context, req micro.Request) {
	// Implementation for handling power operation results
	// This would typically receive results from internal power operations
	ipc.RespondWithError(ctx, req, ipc.ErrInternalError, "power result handling not implemented")
}

// handleGeneralPowerStatus handles power status requests
func (p *PowerMgr) handleGeneralPowerStatus(ctx context.Context, req micro.Request) {
	// Implementation for handling power status requests
	// This would return current power status for components
	ipc.RespondWithError(ctx, req, ipc.ErrInternalError, "power status handling not implemented")
}

// Helper methods that adapt the existing component-specific handlers

func (p *PowerMgr) handleHostPowerActionFromGeneral(ctx context.Context, req micro.Request, request *v1alpha1.ChangeHostStateRequest) {
	// Extract host ID from the host name if needed, or use name directly
	// For now, we'll simulate the old behavior by creating a fake subject
	hostID := request.HostName
	componentName := fmt.Sprintf("host.%s", hostID)

	// Call the existing handler logic but adapted for general use
	p.processHostPowerAction(ctx, req, request, componentName)
}

func (p *PowerMgr) handleChassisPowerActionFromGeneral(ctx context.Context, req micro.Request, request *v1alpha1.ChangeChassisStateRequest) {
	chassisID := request.ChassisName
	componentName := fmt.Sprintf("chassis.%s", chassisID)

	// Call the existing handler logic but adapted for general use
	p.processChassisPowerAction(ctx, req, request, componentName)
}

func (p *PowerMgr) handleBMCPowerActionFromGeneral(ctx context.Context, req micro.Request, request *v1alpha1.ChangeHostStateRequest) {
	componentName := "bmc.0" // Assuming single BMC

	// Call the existing handler logic but adapted for general use
	p.processBMCPowerAction(ctx, req, request, componentName)
}

// These methods would contain the core logic from the existing handlers
// but without the subject parsing since we now get the info from message content

func (p *PowerMgr) processHostPowerAction(ctx context.Context, req micro.Request, request *v1alpha1.ChangeHostStateRequest, componentName string) {
	// This would contain the core logic from handleHostPowerAction
	// but without the subject parsing part
	ipc.RespondWithError(ctx, req, ipc.ErrInternalError, "host power action processing not fully implemented")
}

func (p *PowerMgr) processChassisPowerAction(ctx context.Context, req micro.Request, request *v1alpha1.ChangeChassisStateRequest, componentName string) {
	// This would contain the core logic from handleChassisPowerAction
	// but without the subject parsing part
	ipc.RespondWithError(ctx, req, ipc.ErrInternalError, "chassis power action processing not fully implemented")
}

func (p *PowerMgr) processBMCPowerAction(ctx context.Context, req micro.Request, request *v1alpha1.ChangeHostStateRequest, componentName string) {
	// This would contain the core logic from handleBMCPowerAction
	// but without the subject parsing part
	ipc.RespondWithError(ctx, req, ipc.ErrInternalError, "BMC power action processing not fully implemented")
}

func (p *PowerMgr) createRequestHandler(parentCtx context.Context, handler func(context.Context, micro.Request)) micro.HandlerFunc {
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

		if p.tracer != nil {
			_, span := p.tracer.Start(ctx, "powermgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", p.config.serviceName),
			)
			defer span.End()
		}

		handler(ctx, req) //nolint:contextcheck
	}
}

func (p *PowerMgr) getBackendForComponent(componentName string) (PowerBackend, error) {
	component, exists := p.config.components[componentName]
	if !exists {
		return nil, fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	backend, exists := p.backends[component.Backend]
	if !exists {
		return nil, fmt.Errorf("%w: backend '%s' for component '%s'", ErrBackendNotConfigured, component.Backend, componentName)
	}

	return backend, nil
}

func (p *PowerMgr) recordOperation(ctx context.Context, operation, component string, err error) {
	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("component", component),
	}

	if err != nil {
		attrs = append(attrs, attribute.String("status", "error"))
		p.powerFailuresTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	} else {
		attrs = append(attrs, attribute.String("status", "success"))
	}

	p.powerOperationsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

	if operation == "power_cycle" && err == nil {
		p.powerCyclesTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("component", component),
		))
	}
}

func (p *PowerMgr) reportStateChange(ctx context.Context, componentName, operation string, success bool) { //nolint:revive
	if !p.config.enableStateReporting {
		return
	}

	// Send power operation result to statemgr via NATS using internal coordination subject
	subject := ipc.InternalPowerResult

	// Create protobuf message for power operation results
	result := &v1alpha1.PowerOperationResult{
		ComponentName: componentName,
		Operation:     operation,
		Success:       success,
		CompletedAt:   timestamppb.Now(),
		DurationMs:    uint32(time.Since(time.Now()).Milliseconds()), //nolint:gosec // This will be updated by caller
	}

	if !success {
		result.ErrorMessage = fmt.Sprintf("Power operation %s failed for component %s", operation, componentName)
	}

	// Marshal and send the protobuf message
	data, err := proto.Marshal(result)
	if err != nil {
		p.logger.ErrorContext(ctx, "Failed to marshal power operation result",
			"component", componentName,
			"operation", operation,
			"error", err)
		return
	}

	if err := p.nc.Publish(subject, data); err != nil {
		p.logger.ErrorContext(ctx, "Failed to publish power operation result",
			"component", componentName,
			"operation", operation,
			"subject", subject,
			"error", err)
		return
	}

	p.logger.InfoContext(ctx, "Power operation result published",
		"component", componentName,
		"operation", operation,
		"success", success,
		"subject", subject)
}

func (p *PowerMgr) shutdown(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}

	if ctx.Err() != nil {
		ctx = context.WithoutCancel(ctx)
	}
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	p.closeBackends()
	p.started = false
}
