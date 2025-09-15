// SPDX-License-Identifier: BSD-3-Clause

package powermgr

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	schemav1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/gpio"
	"github.com/u-bmc/u-bmc/pkg/ipc"
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
	// PowerOn powers on the specified component
	PowerOn(ctx context.Context, componentName string) error
	// PowerOff powers off the specified component
	PowerOff(ctx context.Context, componentName string, force bool) error
	// Reset resets the specified component
	Reset(ctx context.Context, componentName string) error
	// GetPowerStatus returns the current power status of the component
	GetPowerStatus(ctx context.Context, componentName string) (bool, error)
	// Initialize initializes the backend with configuration
	Initialize(ctx context.Context, config *Config) error
	// Close closes the backend and cleans up resources
	Close() error
}

// GPIOBackend implements power control using GPIO lines.
type GPIOBackend struct {
	manager    *gpio.Manager
	config     *Config
	components map[string]ComponentConfig
	mu         sync.RWMutex
}

// NewGPIOBackend creates a new GPIO-based power control backend.
func NewGPIOBackend() *GPIOBackend {
	return &GPIOBackend{
		manager:    gpio.NewManager(),
		components: make(map[string]ComponentConfig),
	}
}

// Initialize initializes the GPIO backend.
func (b *GPIOBackend) Initialize(ctx context.Context, config *Config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config = config
	for name, component := range config.Components {
		b.components[name] = component
	}

	return nil
}

// PowerOn powers on a component using GPIO.
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

	line, err := b.manager.RequestLine(b.config.GPIOChip, component.GPIO.PowerButton.Line,
		gpio.WithDirection(component.GPIO.PowerButton.Direction),
		gpio.WithActiveState(component.GPIO.PowerButton.ActiveState),
		gpio.WithInitialValue(component.GPIO.PowerButton.InitialValue),
		gpio.WithBias(component.GPIO.PowerButton.Bias),
		gpio.WithConsumer("powermgr"),
	)
	if err != nil {
		return fmt.Errorf("%w: failed to request power button GPIO: %w", ErrGPIOOperationFailed, err)
	}
	defer line.Close()

	return line.ToggleCtx(ctx, component.PowerOnDelay)
}

// PowerOff powers off a component using GPIO.
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

	line, err := b.manager.RequestLine(b.config.GPIOChip, component.GPIO.PowerButton.Line,
		gpio.WithDirection(component.GPIO.PowerButton.Direction),
		gpio.WithActiveState(component.GPIO.PowerButton.ActiveState),
		gpio.WithInitialValue(component.GPIO.PowerButton.InitialValue),
		gpio.WithBias(component.GPIO.PowerButton.Bias),
		gpio.WithConsumer("powermgr"),
	)
	if err != nil {
		return fmt.Errorf("%w: failed to request power button GPIO: %w", ErrGPIOOperationFailed, err)
	}
	defer line.Close()

	if force {
		return line.Hold(ctx, component.ForceOffDelay)
	}
	return line.ToggleCtx(ctx, component.PowerOffDelay)
}

// Reset resets a component using GPIO.
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

	line, err := b.manager.RequestLine(b.config.GPIOChip, component.GPIO.ResetButton.Line,
		gpio.WithDirection(component.GPIO.ResetButton.Direction),
		gpio.WithActiveState(component.GPIO.ResetButton.ActiveState),
		gpio.WithInitialValue(component.GPIO.ResetButton.InitialValue),
		gpio.WithBias(component.GPIO.ResetButton.Bias),
		gpio.WithConsumer("powermgr"),
	)
	if err != nil {
		return fmt.Errorf("%w: failed to request reset button GPIO: %w", ErrGPIOOperationFailed, err)
	}
	defer line.Close()

	return line.ToggleCtx(ctx, component.ResetDelay)
}

// GetPowerStatus returns the power status using GPIO.
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

	line, err := b.manager.RequestLine(b.config.GPIOChip, component.GPIO.PowerStatus.Line,
		gpio.WithDirection(component.GPIO.PowerStatus.Direction),
		gpio.WithActiveState(component.GPIO.PowerStatus.ActiveState),
		gpio.WithBias(component.GPIO.PowerStatus.Bias),
		gpio.WithConsumer("powermgr"),
	)
	if err != nil {
		return false, fmt.Errorf("%w: failed to request power status GPIO: %w", ErrGPIOOperationFailed, err)
	}
	defer line.Close()

	value, err := line.GetValue()
	if err != nil {
		return false, fmt.Errorf("%w: failed to read power status: %w", ErrGPIOOperationFailed, err)
	}

	return value == 1, nil
}

// Close closes the GPIO backend.
func (b *GPIOBackend) Close() error {
	return b.manager.Close()
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

// registerEndpoints registers all NATS endpoints for power operations.
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

// createRequestHandler creates a request handler with telemetry support.
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

// handleHostPowerAction handles power action requests for host components.
func (p *PowerMgr) handleHostPowerAction(ctx context.Context, req micro.Request) {
	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleHostPowerAction")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) != 4 || parts[0] != "powermgr" || parts[1] != "host" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	hostID := parts[2]
	componentName := fmt.Sprintf("host.%s", hostID)

	var request schemav1alpha1.ChangeHostStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if request.Action == schemav1alpha1.HostAction_HOST_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidPowerAction, "unspecified action")
		return
	}

	var err error
	switch request.Action {
	case schemav1alpha1.HostAction_HOST_ACTION_ON:
		err = p.backend.PowerOn(ctx, componentName)
	case schemav1alpha1.HostAction_HOST_ACTION_OFF:
		err = p.backend.PowerOff(ctx, componentName, false)
	case schemav1alpha1.HostAction_HOST_ACTION_FORCE_OFF:
		err = p.backend.PowerOff(ctx, componentName, true)
	case schemav1alpha1.HostAction_HOST_ACTION_REBOOT:
		err = p.backend.Reset(ctx, componentName)
	case schemav1alpha1.HostAction_HOST_ACTION_FORCE_RESTART:
		err = p.backend.Reset(ctx, componentName)
	default:
		ipc.RespondWithError(ctx, req, ErrPowerOperationNotSupported, fmt.Sprintf("unsupported action: %v", request.Action))
		return
	}

	if err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	response := &schemav1alpha1.ChangeHostStateResponse{
		CurrentStatus: schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING,
	}

	resp, marshalErr := response.MarshalVT()
	if marshalErr != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, marshalErr.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if p.logger != nil {
		p.logger.InfoContext(ctx, "Host power action completed",
			"component", componentName,
			"action", request.Action.String())
	}
}

// handleChassisPowerAction handles power action requests for chassis components.
func (p *PowerMgr) handleChassisPowerAction(ctx context.Context, req micro.Request) {
	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleChassisPowerAction")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) != 4 || parts[0] != "powermgr" || parts[1] != "chassis" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	chassisID := parts[2]
	componentName := fmt.Sprintf("chassis.%s", chassisID)

	var request schemav1alpha1.ChangeChassisStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if request.Action == schemav1alpha1.ChassisAction_CHASSIS_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidPowerAction, "unspecified action")
		return
	}

	var err error
	switch request.Action {
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_ON:
		err = p.backend.PowerOn(ctx, componentName)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF:
		err = p.backend.PowerOff(ctx, componentName, false)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_EMERGENCY_SHUTDOWN:
		err = p.backend.PowerOff(ctx, componentName, true)
	case schemav1alpha1.ChassisAction_CHASSIS_ACTION_POWER_CYCLE:
		if powerOffErr := p.backend.PowerOff(ctx, componentName, false); powerOffErr != nil {
			err = powerOffErr
		} else {
			time.Sleep(2 * time.Second)
			err = p.backend.PowerOn(ctx, componentName)
		}
	default:
		ipc.RespondWithError(ctx, req, ErrPowerOperationNotSupported, fmt.Sprintf("unsupported action: %v", request.Action))
		return
	}

	if err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	response := &schemav1alpha1.ChangeChassisStateResponse{
		CurrentStatus: schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING,
	}

	resp, marshalErr := response.MarshalVT()
	if marshalErr != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, marshalErr.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if p.logger != nil {
		p.logger.InfoContext(ctx, "Chassis power action completed",
			"component", componentName,
			"action", request.Action.String())
	}
}

// handleBMCPowerAction handles power action requests for BMC components.
func (p *PowerMgr) handleBMCPowerAction(ctx context.Context, req micro.Request) {
	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleBMCPowerAction")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) != 4 || parts[0] != "powermgr" || parts[1] != "bmc" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	bmcID := parts[2]
	componentName := fmt.Sprintf("bmc.%s", bmcID)

	var request schemav1alpha1.ChangeManagementControllerStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if request.Action == schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidPowerAction, "unspecified action")
		return
	}

	var err error
	switch request.Action {
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT:
		err = p.backend.Reset(ctx, componentName)
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_WARM_RESET:
		err = p.backend.Reset(ctx, componentName)
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET:
		err = p.backend.Reset(ctx, componentName)
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_HARD_RESET:
		err = p.backend.Reset(ctx, componentName)
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_FACTORY_RESET:
		err = p.backend.Reset(ctx, componentName)
	default:
		ipc.RespondWithError(ctx, req, ErrPowerOperationNotSupported, fmt.Sprintf("unsupported action: %v", request.Action))
		return
	}

	if err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	response := &schemav1alpha1.ChangeManagementControllerStateResponse{
		CurrentStatus: schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY,
	}

	resp, marshalErr := response.MarshalVT()
	if marshalErr != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, marshalErr.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if p.logger != nil {
		p.logger.InfoContext(ctx, "BMC power action completed",
			"component", componentName,
			"action", request.Action.String())
	}
}

// shutdown gracefully stops the power manager and cleans up resources.
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

// CreateGPIOPowerOnCallback creates a GPIO-based power on callback function.
func CreateGPIOPowerOnCallback(chipPath, lineName string, duration time.Duration, opts ...gpio.Option) func(context.Context) error {
	manager := gpio.NewManager()
	return func(ctx context.Context) error {
		line, err := manager.RequestLine(chipPath, lineName, opts...)
		if err != nil {
			return err
		}
		defer line.Close()
		return line.ToggleCtx(ctx, duration)
	}
}

// CreateGPIOPowerOffCallback creates a GPIO-based power off callback function.
func CreateGPIOPowerOffCallback(chipPath, lineName string, duration time.Duration, opts ...gpio.Option) func(context.Context) error {
	manager := gpio.NewManager()
	return func(ctx context.Context) error {
		line, err := manager.RequestLine(chipPath, lineName, opts...)
		if err != nil {
			return err
		}
		defer line.Close()
		return line.ToggleCtx(ctx, duration)
	}
}

// CreateGPIOResetCallback creates a GPIO-based reset callback function.
func CreateGPIOResetCallback(chipPath, lineName string, duration time.Duration, opts ...gpio.Option) func(context.Context) error {
	manager := gpio.NewManager()
	return func(ctx context.Context) error {
		line, err := manager.RequestLine(chipPath, lineName, opts...)
		if err != nil {
			return err
		}
		defer line.Close()
		return line.ToggleCtx(ctx, duration)
	}
}

// CreateGPIOReset creates a stateless GPIO reset action function for compatibility with state machines.
func CreateGPIOReset(chipPath, lineName string, duration time.Duration, opts ...gpio.Option) func(context.Context, ...any) error {
	return gpio.CreateToggleAction(gpio.NewManager(), chipPath, lineName, duration, opts...)
}
