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
		enableMetrics:               true,
		enableTracing:               true,
		enableStateReporting:        true,
		stateReportingSubjectPrefix: "statemgr",
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

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
	p.mu.Lock()
	if p.started {
		p.mu.Unlock()
		return ErrServiceAlreadyStarted
	}
	p.started = true
	ctx, p.cancel = context.WithCancel(ctx)
	p.mu.Unlock()

	p.tracer = otel.Tracer(p.config.serviceName)
	p.meter = otel.Meter(p.config.serviceName)

	ctx, span := p.tracer.Start(ctx, "powermgr.Run")
	defer span.End()

	p.logger = log.GetGlobalLogger().With("service", p.config.serviceName)
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

	p.config.AddDefaultComponents()

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

	if !p.config.enableMetrics {
		return nil
	}

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
	if p.config.enableHostManagement {
		hostGroup := p.microService.AddGroup("host")
		for i := 0; i < p.config.numHosts; i++ {
			endpoint := fmt.Sprintf("%d.action", i)
			if err := hostGroup.AddEndpoint(endpoint,
				micro.HandlerFunc(p.createRequestHandler(ctx, p.handleHostPowerAction))); err != nil {
				return fmt.Errorf("failed to register host endpoint %s: %w", endpoint, err)
			}
		}
	}

	if p.config.enableChassisManagement {
		chassisGroup := p.microService.AddGroup("chassis")
		for i := 0; i < p.config.numChassis; i++ {
			endpoint := fmt.Sprintf("%d.action", i)
			if err := chassisGroup.AddEndpoint(endpoint,
				micro.HandlerFunc(p.createRequestHandler(ctx, p.handleChassisPowerAction))); err != nil {
				return fmt.Errorf("failed to register chassis endpoint %s: %w", endpoint, err)
			}
		}
	}

	if p.config.enableBMCManagement {
		bmcGroup := p.microService.AddGroup("bmc")
		if err := bmcGroup.AddEndpoint("0.action",
			micro.HandlerFunc(p.createRequestHandler(ctx, p.handleBMCPowerAction))); err != nil {
			return fmt.Errorf("failed to register BMC endpoint: %w", err)
		}
	}

	return nil
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
	component, exists := p.config.GetComponentConfig(componentName)
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
	if !p.config.enableMetrics {
		return
	}

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

	// Send power operation result to statemgr via NATS
	subject := fmt.Sprintf("%s.%s.power.result", p.config.stateReportingSubjectPrefix, componentName)

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
