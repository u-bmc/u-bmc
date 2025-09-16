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
	"github.com/u-bmc/u-bmc/pkg/gpio"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ service.Service = (*PowerMgr)(nil)

// PowerBackend defines the interface for power control backends.
type PowerBackend interface {
	PowerOn(ctx context.Context, componentName string) error
	PowerOff(ctx context.Context, componentName string, force bool) error
	Reset(ctx context.Context, componentName string) error
	GetPowerStatus(ctx context.Context, componentName string) (bool, error)
	Initialize(ctx context.Context, config *Config) error
	Close() error
}

// GPIOBackend implements power control using GPIO lines.
type GPIOBackend struct {
	config     *Config
	components map[string]ComponentConfig
	mu         sync.RWMutex
}

// NewGPIOBackend creates a new GPIO-based power control backend.
func NewGPIOBackend() *GPIOBackend {
	return &GPIOBackend{
		components: make(map[string]ComponentConfig),
	}
}

func (b *GPIOBackend) Initialize(ctx context.Context, config *Config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config = config
	for name, component := range config.Components {
		b.components[name] = component
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

	return gpio.ToggleGPIOCtx(ctx, b.config.GPIOChip, component.GPIO.PowerButton.Line, component.PowerOnDelay, opts...)
}

func (b *GPIOBackend) PowerOff(ctx context.Context, componentName string, force bool) error {
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
		line, err := gpio.RequestLine(b.config.GPIOChip, component.GPIO.PowerButton.Line, append(opts, gpio.AsOutput())...)
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
			line.SetValue(0)
			return ctx.Err()
		}

		if err := line.SetValue(0); err != nil {
			return fmt.Errorf("%w: failed to set GPIO low: %w", ErrGPIOOperationFailed, err)
		}
		return nil
	}

	return gpio.ToggleGPIOCtx(ctx, b.config.GPIOChip, component.GPIO.PowerButton.Line, component.PowerOffDelay, opts...)
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

	return gpio.ToggleGPIOCtx(ctx, b.config.GPIOChip, component.GPIO.ResetButton.Line, component.ResetDelay, opts...)
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

	value, err := gpio.GetGPIO(b.config.GPIOChip, component.GPIO.PowerStatus.Line, opts...)
	if err != nil {
		return false, fmt.Errorf("%w: failed to read power status: %w", ErrGPIOOperationFailed, err)
	}

	return value == 1, nil
}

func (b *GPIOBackend) Close() error {
	return nil
}

// PowerMgr manages power operations for BMC components.
type PowerMgr struct {
	config       *Config
	nc           *nats.Conn
	microService micro.Service
	backend      PowerBackend
	logger       *slog.Logger
	tracer       trace.Tracer
	cancel       context.CancelFunc
	started      bool
	mu           sync.RWMutex
}

// New creates a new PowerMgr instance with the provided options.
func New(opts ...Option) *PowerMgr {
	config := NewConfig(opts...)

	return &PowerMgr{
		config: config,
		tracer: otel.Tracer("powermgr"),
	}
}

// Name returns the service name.
func (p *PowerMgr) Name() string {
	return p.config.ServiceName
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

	ctx, span := p.tracer.Start(ctx, "powermgr.Run")
	defer span.End()

	p.logger = log.GetGlobalLogger().With("service", p.config.ServiceName)
	p.logger.InfoContext(ctx, "Starting power manager service",
		"version", p.config.ServiceVersion,
		"hosts", p.config.NumHosts,
		"chassis", p.config.NumChassis)

	if err := p.config.Validate(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	p.config.AddDefaultComponents()

	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrNATSConnectionFailed, err)
	}
	p.nc = nc
	defer nc.Drain()

	p.backend = NewGPIOBackend()
	if err := p.backend.Initialize(ctx, p.config); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrBackendInitializationFailed, err)
	}
	defer p.backend.Close()

	if err := p.initializeThermalIntegration(ctx); err != nil {
		span.RecordError(err)
		p.logger.WarnContext(ctx, "Thermal integration initialization failed", "error", err)
	}

	p.microService, err = micro.AddService(nc, micro.Config{
		Name:        p.config.ServiceName,
		Description: p.config.ServiceDescription,
		Version:     p.config.ServiceVersion,
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
		"backend_initialized", true)

	span.SetAttributes(
		attribute.String("service.name", p.config.ServiceName),
		attribute.String("service.version", p.config.ServiceVersion),
		attribute.Int("components.count", len(p.config.Components)),
	)

	<-ctx.Done()

	err = ctx.Err()
	ctx = context.WithoutCancel(ctx)
	p.logger.InfoContext(ctx, "Shutting down power manager service")
	p.shutdown(ctx)

	return err
}

func (p *PowerMgr) registerEndpoints(ctx context.Context) error {
	if p.config.EnableHostManagement {
		for i := 0; i < p.config.NumHosts; i++ {
			endpoint := fmt.Sprintf("powermgr.host.%d.action", i)
			if err := p.microService.AddEndpoint(endpoint,
				micro.HandlerFunc(p.createRequestHandler(p.handleHostPowerAction))); err != nil {
				return fmt.Errorf("failed to register host endpoint %s: %w", endpoint, err)
			}
		}
	}

	if p.config.EnableChassisManagement {
		for i := 0; i < p.config.NumChassis; i++ {
			endpoint := fmt.Sprintf("powermgr.chassis.%d.action", i)
			if err := p.microService.AddEndpoint(endpoint,
				micro.HandlerFunc(p.createRequestHandler(p.handleChassisPowerAction))); err != nil {
				return fmt.Errorf("failed to register chassis endpoint %s: %w", endpoint, err)
			}
		}
	}

	if p.config.EnableBMCManagement {
		endpoint := "powermgr.bmc.0.action"
		if err := p.microService.AddEndpoint(endpoint,
			micro.HandlerFunc(p.createRequestHandler(p.handleBMCPowerAction))); err != nil {
			return fmt.Errorf("failed to register BMC endpoint %s: %w", endpoint, err)
		}
	}

	return nil
}

func (p *PowerMgr) createRequestHandler(handler func(context.Context, micro.Request)) micro.HandlerFunc {
	return func(req micro.Request) {
		ctx := telemetry.GetCtxFromReq(req)

		if p.tracer != nil {
			_, span := p.tracer.Start(ctx, "powermgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", p.config.ServiceName),
			)
			defer span.End()
		}

		handler(ctx, req)
	}
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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if p.backend != nil {
		if err := p.backend.Close(); err != nil {
			p.logger.Error("Failed to close power backend", "error", err)
		}
	}

	p.started = false
}
