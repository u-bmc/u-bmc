// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/arunsworld/nursery"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/state"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	operationState      = "state"
	operationControl    = "control"
	operationInfo       = "info"
	operationTransition = "transition"
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
		enableMetrics:             true,
		enableTracing:             true,
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
	if !s.config.enableMetrics {
		return nil
	}

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
				sm, err := s.createHostStateMachine(hostName)
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
				sm, err := s.createChassisStateMachine(chassisName)
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
			sm, err := s.createManagementControllerStateMachine(bmcName)
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
	if s.config.enableHostManagement {
		for i := range s.config.numHosts {
			hostEndpoints := []string{
				fmt.Sprintf("statemgr.host.%d.state", i),
				fmt.Sprintf("statemgr.host.%d.control", i),
				fmt.Sprintf("statemgr.host.%d.info", i),
			}

			for _, endpoint := range hostEndpoints {
				if err := s.microService.AddEndpoint(endpoint,
					micro.HandlerFunc(s.createRequestHandler(s.handleHostStateRequest))); err != nil {
					return fmt.Errorf("failed to register host endpoint %s: %w", endpoint, err)
				}
			}
		}
	}

	if s.config.enableChassisManagement {
		for i := range s.config.numChassis {
			chassisEndpoints := []string{
				fmt.Sprintf("statemgr.chassis.%d.state", i),
				fmt.Sprintf("statemgr.chassis.%d.control", i),
				fmt.Sprintf("statemgr.chassis.%d.info", i),
			}

			for _, endpoint := range chassisEndpoints {
				if err := s.microService.AddEndpoint(endpoint,
					micro.HandlerFunc(s.createRequestHandler(s.handleChassisStateRequest))); err != nil {
					return fmt.Errorf("failed to register chassis endpoint %s: %w", endpoint, err)
				}
			}
		}
	}

	if s.config.enableBMCManagement {
		bmcEndpoints := []string{
			"statemgr.bmc.0.state",
			"statemgr.bmc.0.control",
			"statemgr.bmc.0.info",
		}

		for _, endpoint := range bmcEndpoints {
			if err := s.microService.AddEndpoint(endpoint,
				micro.HandlerFunc(s.createRequestHandler(s.handleManagementControllerStateRequest))); err != nil {
				return fmt.Errorf("failed to register BMC endpoint %s: %w", endpoint, err)
			}
		}
	}

	return nil
}

func (s *StateMgr) createRequestHandler(handler func(context.Context, micro.Request)) micro.HandlerFunc {
	return func(req micro.Request) {
		ctx := telemetry.GetCtxFromReq(req)

		if s.tracer != nil {
			_, span := s.tracer.Start(ctx, "statemgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", s.config.serviceName),
			)
			defer span.End()
		}

		handler(ctx, req)
	}
}

func (s *StateMgr) recordTransition(componentName, fromState, toState, trigger string, duration time.Duration, err error) {
	if !s.config.enableMetrics {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("component", componentName),
		attribute.String("from_state", fromState),
		attribute.String("to_state", toState),
		attribute.String("trigger", trigger),
	}

	if err != nil {
		attrs = append(attrs, attribute.String("status", "error"))
		s.stateTransitionFailures.Add(context.Background(), 1, metric.WithAttributes(attrs...))
	} else {
		attrs = append(attrs, attribute.String("status", "success"))
		if s.stateTransitionDuration != nil {
			s.stateTransitionDuration.Record(context.Background(), duration.Seconds(), metric.WithAttributes(attrs...))
		}
	}

	s.stateTransitionsTotal.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

func (s *StateMgr) updateCurrentState(componentName, stateName string) {
	if !s.config.enableMetrics || s.currentStateGauge == nil {
		return
	}

	// TODO: Implement state to integer encoding for metrics
	s.currentStateGauge.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("component", componentName),
		attribute.String("state", stateName),
	))
}

func (s *StateMgr) requestPowerAction(ctx context.Context, componentName, action string) error {
	if s.config.powerControlSubjectPrefix == "" {
		return nil
	}

	subject := fmt.Sprintf("%s.%s.action", s.config.powerControlSubjectPrefix, componentName)

	// TODO: Create proper protobuf message for power control requests
	// For now, just log the power request
	s.logger.InfoContext(ctx, "Requesting power action",
		"component", componentName,
		"action", action,
		"subject", subject)

	return nil
}

func (s *StateMgr) requestLEDAction(ctx context.Context, componentName, action string) error {
	if s.config.ledControlSubjectPrefix == "" {
		return nil
	}

	subject := fmt.Sprintf("%s.%s.action", s.config.ledControlSubjectPrefix, componentName)

	// TODO: Create proper protobuf message for LED control requests
	// For now, just log the LED request
	s.logger.InfoContext(ctx, "Requesting LED action",
		"component", componentName,
		"action", action,
		"subject", subject)

	return nil
}

func (s *StateMgr) shutdown(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
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
