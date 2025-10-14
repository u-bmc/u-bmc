// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/arunsworld/nursery"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/ipc"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/state"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ service.Service = (*StateMgr)(nil)

// StateMgr manages state machines for BMC, chassis, and host components.
// It provides NATS-based IPC endpoints for state management operations.
type StateMgr struct {
	config        *config
	nc            *nats.Conn
	js            jetstream.JetStream
	microService  micro.Service
	stateMachines map[string]*state.Machine
	mu            sync.RWMutex
	logger        *slog.Logger
	tracer        trace.Tracer
	meter         metric.Meter
	cancel        context.CancelFunc
	started       bool

	// Subscriptions
	powerResultSub *nats.Subscription

	// Metrics
	stateTransitionsTotal   metric.Int64Counter
	stateTransitionDuration metric.Float64Histogram
	stateTransitionFailures metric.Int64Counter
	currentStateGauge       metric.Int64UpDownCounter
}

// New creates a new StateMgr instance with the provided options.
func New(opts ...Option) *StateMgr {
	cfg := &config{
		serviceName:               DefaultServiceName,
		serviceDescription:        DefaultServiceDescription,
		serviceVersion:            DefaultServiceVersion,
		streamName:                DefaultStreamName,
		streamSubjects:            []string{"statemgr.state.>", "statemgr.event.>"},
		streamRetention:           0,
		enableHostManagement:      true,
		enableChassisManagement:   true,
		enableBMCManagement:       true,
		numHosts:                  1,
		numChassis:                1,
		stateTimeout:              DefaultStateTimeout,
		broadcastStateChanges:     true,
		persistStateChanges:       true,
		powerControlSubjectPrefix: "powermgr",
		ledControlSubjectPrefix:   "ledmgr",
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &StateMgr{
		config:        cfg,
		stateMachines: make(map[string]*state.Machine),
	}
}

// Name returns the service name.
func (s *StateMgr) Name() string {
	return s.config.serviceName
}

// Run starts the state manager service and registers NATS IPC endpoints.
// It initializes state machines for enabled components and handles graceful shutdown.
func (s *StateMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	// s.mu.Lock()
	// if s.started {
	// 	s.mu.Unlock()
	// 	return ErrServiceAlreadyStarted
	// }
	// s.started = true
	// ctx, s.cancel = context.WithCancel(ctx)
	// s.mu.Unlock()

	s.tracer = otel.Tracer(s.config.serviceName)
	s.meter = otel.Meter(s.config.serviceName)

	ctx, span := s.tracer.Start(ctx, "statemgr.Run")
	defer span.End()

	s.logger = log.GetGlobalLogger().With("service", s.config.serviceName)
	s.logger.InfoContext(ctx, "Starting state manager service",
		"version", s.config.serviceVersion,
		"hosts", s.config.numHosts,
		"chassis", s.config.numChassis,
		"persistence_enabled", s.config.persistStateChanges)

	if err := s.config.Validate(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	if err := s.initializeMetrics(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	nc, err := nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrNATSConnectionFailed, err)
	}
	s.nc = nc
	defer nc.Drain() //nolint:errcheck

	s.js, err = jetstream.New(nc)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrJetStreamInitFailed, err)
	}

	if s.config.persistStateChanges {
		if err := s.setupJetStream(ctx); err != nil {
			span.RecordError(err)
			return err
		}
	}

	if err := s.initializeStateMachines(ctx); err != nil {
		span.RecordError(err)
		return err
	}

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

	if err := s.setupSubscriptions(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	s.logger.InfoContext(ctx, "State manager service started successfully",
		"endpoints_registered", true,
		"state_machines_initialized", len(s.stateMachines))

	span.SetAttributes(
		attribute.String("service.name", s.config.serviceName),
		attribute.String("service.version", s.config.serviceVersion),
		attribute.Int("state_machines.count", len(s.stateMachines)),
		attribute.Bool("persistence.enabled", s.config.persistStateChanges),
	)

	<-ctx.Done()

	err = ctx.Err()
	ctx = context.WithoutCancel(ctx)
	s.logger.InfoContext(ctx, "Shutting down state manager service")
	s.shutdown(ctx)

	return err
}

func (s *StateMgr) initializeMetrics() error {
	var err error

	s.stateTransitionsTotal, err = s.meter.Int64Counter(
		"statemgr_transitions_total",
		metric.WithDescription("Total number of state transitions"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create state transitions counter: %w", err)
	}

	s.stateTransitionDuration, err = s.meter.Float64Histogram(
		"statemgr_transition_duration_seconds",
		metric.WithDescription("Duration of state transitions"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create state transition duration histogram: %w", err)
	}

	s.stateTransitionFailures, err = s.meter.Int64Counter(
		"statemgr_transition_failures_total",
		metric.WithDescription("Total number of failed state transitions"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create state transition failures counter: %w", err)
	}

	s.currentStateGauge, err = s.meter.Int64UpDownCounter(
		"statemgr_current_state",
		metric.WithDescription("Current state of components (encoded as integer)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create current state gauge: %w", err)
	}

	return nil
}

func (s *StateMgr) setupJetStream(ctx context.Context) error {
	streamConfig := jetstream.StreamConfig{
		Name:        s.config.streamName,
		Description: "State manager persistence stream",
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
		return fmt.Errorf("%w: %w", ErrStreamCreationFailed, err)
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

func (s *StateMgr) initializeStateMachines(ctx context.Context) error {
	var tasks []nursery.ConcurrentJob

	if s.config.enableHostManagement {
		for i := range s.config.numHosts {
			hostID := i
			tasks = append(tasks, func(ctx context.Context, errChan chan error) {
				hostName := fmt.Sprintf("host.%d", hostID)
				sm, err := s.createHostStateMachine(ctx, hostName)
				if err != nil {
					errChan <- err
					return
				}
				if err := sm.Start(ctx); err != nil {
					errChan <- err
					return
				}
				s.mu.Lock()
				s.stateMachines[sm.Name()] = sm
				s.mu.Unlock()
			})
		}
	}

	if s.config.enableChassisManagement {
		for i := range s.config.numChassis {
			chassisID := i
			tasks = append(tasks, func(ctx context.Context, errChan chan error) {
				chassisName := fmt.Sprintf("chassis.%d", chassisID)
				sm, err := s.createChassisStateMachine(ctx, chassisName)
				if err != nil {
					errChan <- err
					return
				}
				if err := sm.Start(ctx); err != nil {
					errChan <- err
					return
				}
				s.mu.Lock()
				s.stateMachines[sm.Name()] = sm
				s.mu.Unlock()
			})
		}
	}

	if s.config.enableBMCManagement {
		tasks = append(tasks, func(ctx context.Context, errChan chan error) {
			bmcName := "bmc.0"
			sm, err := s.createManagementControllerStateMachine(ctx, bmcName)
			if err != nil {
				errChan <- err
				return
			}
			if err := sm.Start(ctx); err != nil {
				errChan <- err
				return
			}
			s.mu.Lock()
			s.stateMachines[sm.Name()] = sm
			s.mu.Unlock()
		})
	}

	return nursery.RunConcurrentlyWithContext(ctx, tasks...)
}

func (s *StateMgr) registerEndpoints(ctx context.Context) error {
	groups := make(map[string]micro.Group)

	if s.config.enableHostManagement {
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectHostList,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleListHosts)), groups); err != nil {
			return fmt.Errorf("failed to register host list endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectHostState,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleHostState)), groups); err != nil {
			return fmt.Errorf("failed to register host state endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectHostControl,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleHostControl)), groups); err != nil {
			return fmt.Errorf("failed to register host control endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectHostInfo,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleHostInfo)), groups); err != nil {
			return fmt.Errorf("failed to register host info endpoint: %w", err)
		}
	}

	if s.config.enableChassisManagement {
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectChassisList,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleListChassis)), groups); err != nil {
			return fmt.Errorf("failed to register chassis list endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectChassisState,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleChassisState)), groups); err != nil {
			return fmt.Errorf("failed to register chassis state endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectChassisControl,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleChassisControl)), groups); err != nil {
			return fmt.Errorf("failed to register chassis control endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectChassisInfo,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleChassisInfo)), groups); err != nil {
			return fmt.Errorf("failed to register chassis info endpoint: %w", err)
		}
	}

	if s.config.enableBMCManagement {
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectBMCList,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleListManagementControllers)), groups); err != nil {
			return fmt.Errorf("failed to register BMC list endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectBMCState,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleManagementControllerState)), groups); err != nil {
			return fmt.Errorf("failed to register BMC state endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectBMCControl,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleManagementControllerControl)), groups); err != nil {
			return fmt.Errorf("failed to register BMC control endpoint: %w", err)
		}
		if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectBMCInfo,
			micro.HandlerFunc(s.createRequestHandler(ctx, s.handleManagementControllerInfo)), groups); err != nil {
			return fmt.Errorf("failed to register BMC info endpoint: %w", err)
		}
	}

	return nil
}

func (s *StateMgr) createRequestHandler(parentCtx context.Context, handler func(context.Context, micro.Request)) micro.HandlerFunc {
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
			_, span := s.tracer.Start(ctx, "statemgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", s.config.serviceName),
			)
			defer span.End()
		}

		handler(ctx, req) //nolint:contextcheck
	}
}

func (s *StateMgr) recordTransition(ctx context.Context, componentName, fromState, toState, trigger string, duration time.Duration, err error) {
	attrs := []attribute.KeyValue{
		attribute.String("component", componentName),
		attribute.String("from_state", fromState),
		attribute.String("to_state", toState),
		attribute.String("trigger", trigger),
	}

	if err != nil {
		attrs = append(attrs, attribute.String("status", "error"))
		s.stateTransitionFailures.Add(ctx, 1, metric.WithAttributes(attrs...))
	} else {
		attrs = append(attrs, attribute.String("status", "success"))
		if s.stateTransitionDuration != nil {
			s.stateTransitionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
		}
	}

	s.stateTransitionsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (s *StateMgr) updateCurrentState(ctx context.Context, componentName, stateName string) {
	if s.currentStateGauge == nil {
		return
	}

	// TODO: Implement state to integer encoding for metrics
	s.currentStateGauge.Add(ctx, 1, metric.WithAttributes(
		attribute.String("component", componentName),
		attribute.String("state", stateName),
	))
}

func (s *StateMgr) requestPowerAction(ctx context.Context, componentName, action string) error {
	if s.config.powerControlSubjectPrefix == "" {
		return nil
	}

	subject := ipc.InternalPowerAction

	timeoutMs := uint32(ipc.DefaultCommandTimeout)
	powerReq := &v1alpha1.PowerControlRequest{
		ComponentName: componentName,
		Action:        action,
		TimeoutMs:     &timeoutMs,
	}

	data, err := proto.Marshal(powerReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal power control request",
			"component", componentName,
			"action", action,
			"error", err)
		return fmt.Errorf("failed to marshal power control request: %w", err)
	}

	if err := s.nc.Publish(subject, data); err != nil {
		s.logger.ErrorContext(ctx, "Failed to publish power control request",
			"component", componentName,
			"action", action,
			"subject", subject,
			"error", err)
		return fmt.Errorf("failed to publish power control request: %w", err)
	}

	s.logger.InfoContext(ctx, "Power action request sent",
		"component", componentName,
		"action", action,
		"subject", subject)

	return nil
}

func (s *StateMgr) requestLEDAction(ctx context.Context, componentName, action string) error {
	if s.config.ledControlSubjectPrefix == "" {
		return nil
	}

	// Parse action to determine LED type and state
	ledType, ledState := s.parseLEDAction(action)
	subject := ipc.InternalLEDControl

	ledReq := &v1alpha1.LEDControlRequest{
		ComponentName: componentName,
		LedType:       ledType,
		LedState:      ledState,
	}

	data, err := proto.Marshal(ledReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal LED control request",
			"component", componentName,
			"action", action,
			"error", err)
		return fmt.Errorf("failed to marshal LED control request: %w", err)
	}

	if err := s.nc.Publish(subject, data); err != nil {
		s.logger.ErrorContext(ctx, "Failed to publish LED control request",
			"component", componentName,
			"action", action,
			"subject", subject,
			"error", err)
		return fmt.Errorf("failed to publish LED control request: %w", err)
	}

	s.logger.InfoContext(ctx, "LED action request sent",
		"component", componentName,
		"action", action,
		"led_type", ledType,
		"led_state", ledState,
		"subject", subject)

	return nil
}

func (s *StateMgr) shutdown(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	// Clean up subscription
	if s.powerResultSub != nil {
		_ = s.powerResultSub.Unsubscribe()
	}

	if ctx.Err() != nil {
		ctx = context.WithoutCancel(ctx)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for name, sm := range s.stateMachines {
		if err := sm.Stop(ctx); err != nil {
			s.logger.Error("Failed to stop state machine",
				"name", name,
				"error", err)
		}
	}

	s.started = false
}

func (s *StateMgr) getStateMachine(name string) (*state.Machine, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sm, exists := s.stateMachines[name]
	return sm, exists
}

func (s *StateMgr) parseLEDAction(action string) (v1alpha1.LEDType, v1alpha1.LEDState) {
	switch action {
	case "power_on": //nolint:goconst
		return v1alpha1.LEDType_LED_TYPE_POWER, v1alpha1.LEDState_LED_STATE_ON
	case "power_off": //nolint:goconst
		return v1alpha1.LEDType_LED_TYPE_POWER, v1alpha1.LEDState_LED_STATE_OFF
	case "power_blink":
		return v1alpha1.LEDType_LED_TYPE_POWER, v1alpha1.LEDState_LED_STATE_BLINK
	case "status_on":
		return v1alpha1.LEDType_LED_TYPE_STATUS, v1alpha1.LEDState_LED_STATE_ON
	case "status_off":
		return v1alpha1.LEDType_LED_TYPE_STATUS, v1alpha1.LEDState_LED_STATE_OFF
	case "status_blink":
		return v1alpha1.LEDType_LED_TYPE_STATUS, v1alpha1.LEDState_LED_STATE_BLINK
	case "error", "error_on":
		return v1alpha1.LEDType_LED_TYPE_ERROR, v1alpha1.LEDState_LED_STATE_ON
	case "error_off":
		return v1alpha1.LEDType_LED_TYPE_ERROR, v1alpha1.LEDState_LED_STATE_OFF
	case "error_blink":
		return v1alpha1.LEDType_LED_TYPE_ERROR, v1alpha1.LEDState_LED_STATE_BLINK
	case "critical_error":
		return v1alpha1.LEDType_LED_TYPE_ERROR, v1alpha1.LEDState_LED_STATE_FAST_BLINK
	case "warning":
		return v1alpha1.LEDType_LED_TYPE_STATUS, v1alpha1.LEDState_LED_STATE_BLINK
	case "failed":
		return v1alpha1.LEDType_LED_TYPE_ERROR, v1alpha1.LEDState_LED_STATE_FAST_BLINK
	case "identify_on":
		return v1alpha1.LEDType_LED_TYPE_IDENTIFY, v1alpha1.LEDState_LED_STATE_ON
	case "identify_off":
		return v1alpha1.LEDType_LED_TYPE_IDENTIFY, v1alpha1.LEDState_LED_STATE_OFF
	case "identify_blink":
		return v1alpha1.LEDType_LED_TYPE_IDENTIFY, v1alpha1.LEDState_LED_STATE_BLINK
	default:
		// Default to power LED off
		return v1alpha1.LEDType_LED_TYPE_POWER, v1alpha1.LEDState_LED_STATE_OFF
	}
}

func (s *StateMgr) ledTypeToString(ledType v1alpha1.LEDType) string {
	switch ledType {
	case v1alpha1.LEDType_LED_TYPE_POWER:
		return "power"
	case v1alpha1.LEDType_LED_TYPE_STATUS:
		return "status"
	case v1alpha1.LEDType_LED_TYPE_ERROR:
		return "error"
	case v1alpha1.LEDType_LED_TYPE_IDENTIFY:
		return "identify"
	default:
		return "power"
	}
}

func (s *StateMgr) setupSubscriptions(ctx context.Context) error {
	if s.config.powerControlSubjectPrefix == "" {
		s.logger.InfoContext(ctx, "State reporting disabled, skipping subscriptions")
		return nil
	}

	subject := ipc.InternalPowerResult

	sub, err := s.nc.Subscribe(subject, func(msg *nats.Msg) {
		s.handlePowerOperationResult(ctx, msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to power operation results: %w", err)
	}

	s.logger.InfoContext(ctx, "Subscribed to power operation results",
		"subject", subject)

	// Store subscription for cleanup
	s.powerResultSub = sub

	return nil
}

func (s *StateMgr) handlePowerOperationResult(ctx context.Context, msg *nats.Msg) {
	var result v1alpha1.PowerOperationResult
	if err := proto.Unmarshal(msg.Data, &result); err != nil {
		s.logger.ErrorContext(ctx, "Failed to unmarshal power operation result",
			"subject", msg.Subject,
			"error", err)
		return
	}

	s.logger.InfoContext(ctx, "Received power operation result",
		"component", result.ComponentName,
		"operation", result.Operation,
		"success", result.Success,
		"duration_ms", result.DurationMs)

	sm, exists := s.getStateMachine(result.ComponentName)
	if !exists {
		s.logger.WarnContext(ctx, "No state machine found for component",
			"component", result.ComponentName)
		return
	}

	var trigger string
	if result.Success {
		trigger = fmt.Sprintf("%s_completed", result.Operation)
	} else {
		trigger = fmt.Sprintf("%s_failed", result.Operation)
	}

	if err := sm.Fire(ctx, trigger); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send trigger to state machine",
			"component", result.ComponentName,
			"trigger", trigger,
			"error", err)
		return
	}

	s.sendStateTransitionNotification(ctx, result.ComponentName, result.Operation, result.Success)
}

func (s *StateMgr) sendStateTransitionNotification(ctx context.Context, componentName, operation string, success bool) {
	sm, exists := s.getStateMachine(componentName)
	if !exists {
		return
	}

	currentState := sm.State(ctx)

	notification := &v1alpha1.StateTransitionNotification{
		ComponentName:        componentName,
		ComponentType:        s.getComponentType(componentName),
		CurrentState:         currentState,
		Trigger:              fmt.Sprintf("power_%s", operation),
		Success:              success,
		ChangedAt:            timestamppb.Now(),
		TransitionDurationMs: 0, // Will be filled by caller if available
	}

	data, err := proto.Marshal(notification)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal state transition notification",
			"component", componentName,
			"error", err)
		return
	}

	subject := ipc.SubjectStateEvent
	if err := s.nc.Publish(subject, data); err != nil {
		s.logger.ErrorContext(ctx, "Failed to publish state transition notification",
			"component", componentName,
			"subject", subject,
			"error", err)
		return
	}

	s.logger.InfoContext(ctx, "State transition notification sent",
		"component", componentName,
		"state", currentState,
		"trigger", notification.Trigger,
		"success", success)
}

func (s *StateMgr) getComponentType(componentName string) string {
	parts := strings.Split(componentName, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func (s *StateMgr) handleListHosts(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		var span trace.Span
		ctx, span = s.tracer.Start(ctx, "statemgr.handleListHosts")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	var request v1alpha1.ListHostsRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	hosts := make([]*v1alpha1.Host, 0, s.config.numHosts)

	for i := 0; i < s.config.numHosts; i++ {
		hostName := fmt.Sprintf("host.%d", i)
		sm, exists := s.getStateMachine(hostName)
		if !exists {
			continue
		}

		currentState := sm.State(ctx)
		statusEnum := hostStatusStringToEnum(currentState)

		host := &v1alpha1.Host{
			Name:   hostName,
			Status: &statusEnum,
		}
		hosts = append(hosts, host)
	}

	response := &v1alpha1.ListHostsResponse{
		Hosts: hosts,
	}

	resp, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(resp); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}
}

func (s *StateMgr) handleListChassis(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		var span trace.Span
		ctx, span = s.tracer.Start(ctx, "statemgr.handleListChassis")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	var request v1alpha1.ListChassisRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	chassis := make([]*v1alpha1.Chassis, 0, s.config.numChassis)

	for i := 0; i < s.config.numChassis; i++ {
		chassisName := fmt.Sprintf("chassis.%d", i)
		sm, exists := s.getStateMachine(chassisName)
		if !exists {
			continue
		}

		currentState := sm.State(ctx)
		statusEnum := chassisStatusStringToEnum(currentState)

		chassisItem := &v1alpha1.Chassis{
			Name:   chassisName,
			Status: &statusEnum,
		}
		chassis = append(chassis, chassisItem)
	}

	response := &v1alpha1.ListChassisResponse{
		Chassis: chassis,
	}

	resp, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(resp); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}
}

func (s *StateMgr) handleListManagementControllers(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		var span trace.Span
		ctx, span = s.tracer.Start(ctx, "statemgr.handleListManagementControllers")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	var request v1alpha1.ListManagementControllersRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	controllers := make([]*v1alpha1.ManagementController, 0, 1)

	bmcName := "bmc.0"
	sm, exists := s.getStateMachine(bmcName)
	if exists {
		currentState := sm.State(ctx)
		statusEnum := managementControllerStatusStringToEnum(currentState)

		controller := &v1alpha1.ManagementController{
			Name:   bmcName,
			Status: &statusEnum,
		}
		controllers = append(controllers, controller)
	}

	response := &v1alpha1.ListManagementControllersResponse{
		Controllers: controllers,
	}

	resp, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(resp); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to send response", "error", err)
	}
}
