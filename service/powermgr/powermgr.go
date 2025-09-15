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
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	operationPowerOn          = "power.on"
	operationPowerOff         = "power.off"
	operationPowerReset       = "power.reset"
	operationPowerStatus      = "power.status"
	operationPowerConsumption = "power.consumption"
	operationPowerCap         = "power.cap"
)

var _ service.Service = (*PowerMgr)(nil)

// Backend defines the interface for power control backends.
type Backend interface {
	// PowerOn powers on the specified component
	PowerOn(ctx context.Context, componentName string) error
	// PowerOff powers off the specified component
	PowerOff(ctx context.Context, componentName string, force bool) error
	// Reset resets the specified component
	Reset(ctx context.Context, componentName string) error
	// GetPowerStatus returns the current power status of the component
	GetPowerStatus(ctx context.Context, componentName string) (bool, error)
	// GetPowerConsumption returns the current power consumption in watts
	GetPowerConsumption(ctx context.Context, componentName string) (float64, error)
	// SetPowerCap sets the power cap for the component
	SetPowerCap(ctx context.Context, componentName string, capWatts float64) error
	// GetPowerCap gets the current power cap for the component
	GetPowerCap(ctx context.Context, componentName string) (float64, error)
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

// GetPowerConsumption returns mock power consumption (GPIO doesn't provide this).
func (b *GPIOBackend) GetPowerConsumption(ctx context.Context, componentName string) (float64, error) {
	b.mu.RLock()
	component, exists := b.components[componentName]
	b.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("%w: component '%s'", ErrComponentNotFound, componentName)
	}

	if !component.EnablePowerMonitoring {
		return 0, fmt.Errorf("%w: power monitoring not enabled for component '%s'", ErrPowerMonitoringDisabled, componentName)
	}

	// GPIO backend doesn't provide actual power consumption
	// This would be implemented by a dedicated power monitoring backend
	return 0, ErrPowerDataUnavailable
}

// SetPowerCap sets power cap (not supported by GPIO backend).
func (b *GPIOBackend) SetPowerCap(ctx context.Context, componentName string, capWatts float64) error {
	return ErrPowerCapNotSupported
}

// GetPowerCap gets power cap (not supported by GPIO backend).
func (b *GPIOBackend) GetPowerCap(ctx context.Context, componentName string) (float64, error) {
	return 0, ErrPowerCapNotSupported
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
	backend      Backend
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

	// Add default components if not configured
	p.config.AddDefaultComponents()

	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrNATSConnectionFailed, err)
	}
	p.nc = nc
	defer nc.Drain()

	// Initialize GPIO backend as default
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
			hostEndpoints := []string{
				fmt.Sprintf("powermgr.host.%d.power.on", i),
				fmt.Sprintf("powermgr.host.%d.power.off", i),
				fmt.Sprintf("powermgr.host.%d.power.reset", i),
				fmt.Sprintf("powermgr.host.%d.power.status", i),
				fmt.Sprintf("powermgr.host.%d.power.consumption", i),
				fmt.Sprintf("powermgr.host.%d.power.cap", i),
			}

			for _, endpoint := range hostEndpoints {
				if err := p.microService.AddEndpoint(endpoint,
					micro.HandlerFunc(p.createRequestHandler(p.handleHostPowerRequest))); err != nil {
					return fmt.Errorf("failed to register host endpoint %s: %w", endpoint, err)
				}
			}
		}
	}

	if p.config.EnableChassisManagement {
		for i := 0; i < p.config.NumChassis; i++ {
			chassisEndpoints := []string{
				fmt.Sprintf("powermgr.chassis.%d.power.on", i),
				fmt.Sprintf("powermgr.chassis.%d.power.off", i),
				fmt.Sprintf("powermgr.chassis.%d.power.status", i),
				fmt.Sprintf("powermgr.chassis.%d.power.consumption", i),
			}

			for _, endpoint := range chassisEndpoints {
				if err := p.microService.AddEndpoint(endpoint,
					micro.HandlerFunc(p.createRequestHandler(p.handleChassisPowerRequest))); err != nil {
					return fmt.Errorf("failed to register chassis endpoint %s: %w", endpoint, err)
				}
			}
		}
	}

	if p.config.EnableBMCManagement {
		bmcEndpoints := []string{
			"powermgr.bmc.0.power.reset",
			"powermgr.bmc.0.power.status",
		}

		for _, endpoint := range bmcEndpoints {
			if err := p.microService.AddEndpoint(endpoint,
				micro.HandlerFunc(p.createRequestHandler(p.handleBMCPowerRequest))); err != nil {
				return fmt.Errorf("failed to register BMC endpoint %s: %w", endpoint, err)
			}
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

// handleHostPowerRequest handles power requests for host components.
func (p *PowerMgr) handleHostPowerRequest(ctx context.Context, req micro.Request) {
	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleHostPowerRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 5 || parts[0] != "powermgr" || parts[1] != "host" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	hostID := parts[2]
	componentName := fmt.Sprintf("host.%s", hostID)
	operation := strings.Join(parts[3:], ".")

	switch operation {
	case operationPowerOn:
		p.handlePowerOn(ctx, req, componentName)
	case operationPowerOff:
		p.handlePowerOff(ctx, req, componentName)
	case operationPowerReset:
		p.handlePowerReset(ctx, req, componentName)
	case operationPowerStatus:
		p.handlePowerStatus(ctx, req, componentName)
	case operationPowerConsumption:
		p.handlePowerConsumption(ctx, req, componentName)
	case operationPowerCap:
		p.handlePowerCap(ctx, req, componentName)
	default:
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

// handleChassisPowerRequest handles power requests for chassis components.
func (p *PowerMgr) handleChassisPowerRequest(ctx context.Context, req micro.Request) {
	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleChassisPowerRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 5 || parts[0] != "powermgr" || parts[1] != "chassis" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	chassisID := parts[2]
	componentName := fmt.Sprintf("chassis.%s", chassisID)
	operation := strings.Join(parts[3:], ".")

	switch operation {
	case operationPowerOn:
		p.handlePowerOn(ctx, req, componentName)
	case operationPowerOff:
		p.handlePowerOff(ctx, req, componentName)
	case operationPowerStatus:
		p.handlePowerStatus(ctx, req, componentName)
	case operationPowerConsumption:
		p.handlePowerConsumption(ctx, req, componentName)
	default:
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

// handleBMCPowerRequest handles power requests for BMC components.
func (p *PowerMgr) handleBMCPowerRequest(ctx context.Context, req micro.Request) {
	if p.tracer != nil {
		var span trace.Span
		_, span = p.tracer.Start(ctx, "powermgr.handleBMCPowerRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 5 || parts[0] != "powermgr" || parts[1] != "bmc" {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	bmcID := parts[2]
	componentName := fmt.Sprintf("bmc.%s", bmcID)
	operation := strings.Join(parts[3:], ".")

	switch operation {
	case operationPowerReset:
		p.handlePowerReset(ctx, req, componentName)
	case operationPowerStatus:
		p.handlePowerStatus(ctx, req, componentName)
	default:
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

// handlePowerOn handles power on requests.
func (p *PowerMgr) handlePowerOn(ctx context.Context, req micro.Request, componentName string) {
	var request schemav1alpha1.PowerOnRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if err := p.backend.PowerOn(ctx, componentName); err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	response := &schemav1alpha1.PowerOnResponse{
		ComponentName: componentName,
		Success:       true,
		Timestamp:     timestamppb.New(time.Now().UTC()),
	}

	resp, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if p.logger != nil {
		p.logger.InfoContext(ctx, "Power on completed", "component", componentName)
	}
}

// handlePowerOff handles power off requests.
func (p *PowerMgr) handlePowerOff(ctx context.Context, req micro.Request, componentName string) {
	var request schemav1alpha1.PowerOffRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	force := request.Force
	if err := p.backend.PowerOff(ctx, componentName, force); err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	response := &schemav1alpha1.PowerOffResponse{
		ComponentName: componentName,
		Success:       true,
		Timestamp:     timestamppb.New(time.Now().UTC()),
	}

	resp, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if p.logger != nil {
		p.logger.InfoContext(ctx, "Power off completed", "component", componentName, "force", force)
	}
}

// handlePowerReset handles reset requests.
func (p *PowerMgr) handlePowerReset(ctx context.Context, req micro.Request, componentName string) {
	var request schemav1alpha1.PowerResetRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if err := p.backend.Reset(ctx, componentName); err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	response := &schemav1alpha1.PowerResetResponse{
		ComponentName: componentName,
		Success:       true,
		Timestamp:     timestamppb.New(time.Now().UTC()),
	}

	resp, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}

	if p.logger != nil {
		p.logger.InfoContext(ctx, "Reset completed", "component", componentName)
	}
}

// handlePowerStatus handles power status requests.
func (p *PowerMgr) handlePowerStatus(ctx context.Context, req micro.Request, componentName string) {
	powered, err := p.backend.GetPowerStatus(ctx, componentName)
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	response := &schemav1alpha1.PowerStatusResponse{
		ComponentName: componentName,
		Powered:       powered,
		Timestamp:     timestamppb.New(time.Now().UTC()),
	}

	resp, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}
}

// handlePowerConsumption handles power consumption requests.
func (p *PowerMgr) handlePowerConsumption(ctx context.Context, req micro.Request, componentName string) {
	consumption, err := p.backend.GetPowerConsumption(ctx, componentName)
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
		return
	}

	response := &schemav1alpha1.PowerConsumptionResponse{
		ComponentName:    componentName,
		ConsumptionWatts: consumption,
		Timestamp:        timestamppb.New(time.Now().UTC()),
	}

	resp, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(resp); err != nil && p.logger != nil {
		p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}
}

// handlePowerCap handles power cap requests.
func (p *PowerMgr) handlePowerCap(ctx context.Context, req micro.Request, componentName string) {
	var request schemav1alpha1.PowerCapRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	if request.GetCap {
		capWatts, err := p.backend.GetPowerCap(ctx, componentName)
		if err != nil {
			ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
			return
		}

		response := &schemav1alpha1.PowerCapResponse{
			ComponentName: componentName,
			CapWatts:      capWatts,
			Timestamp:     timestamppb.New(time.Now().UTC()),
		}

		resp, err := response.MarshalVT()
		if err != nil {
			ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
			return
		}

		if err := req.Respond(resp); err != nil && p.logger != nil {
			p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
		}
	} else {
		if request.CapWatts == nil {
			ipc.RespondWithError(ctx, req, ErrInvalidPowerValue, "cap_watts is required for set operation")
			return
		}

		capWatts := *request.CapWatts
		if err := p.backend.SetPowerCap(ctx, componentName, capWatts); err != nil {
			ipc.RespondWithError(ctx, req, ErrPowerOperationFailed, err.Error())
			return
		}

		response := &schemav1alpha1.PowerCapResponse{
			ComponentName: componentName,
			CapWatts:      capWatts,
			Success:       true,
			Timestamp:     timestamppb.New(time.Now().UTC()),
			Duration:      durationpb.New(time.Hour), // Default cap duration
		}

		resp, err := response.MarshalVT()
		if err != nil {
			ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
			return
		}

		if err := req.Respond(resp); err != nil && p.logger != nil {
			p.logger.ErrorContext(ctx, "Failed to send response", "error", err)
		}

		if p.logger != nil {
			p.logger.InfoContext(ctx, "Power cap set", "component", componentName, "cap_watts", capWatts)
		}
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
