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
	config        *Config
	nc            *nats.Conn
	js            jetstream.JetStream
	microService  micro.Service
	stateMachines map[string]*state.FSM
	mu            sync.RWMutex
	logger        *slog.Logger
	tracer        trace.Tracer
	cancel        context.CancelFunc
	started       bool
}

// New creates a new StateMgr instance with the provided options.
func New(opts ...Option) *StateMgr {
	config := NewConfig(opts...)

	return &StateMgr{
		config:        config,
		stateMachines: make(map[string]*state.FSM),
		tracer:        otel.Tracer("statemgr"),
	}
}

// Name returns the service name.
func (s *StateMgr) Name() string {
	return s.config.ServiceName
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

	ctx, span := s.tracer.Start(ctx, "statemgr.Run")
	defer span.End()

	s.logger = log.GetGlobalLogger().With("service", s.config.ServiceName)
	s.logger.InfoContext(ctx, "Starting state manager service",
		"version", s.config.ServiceVersion,
		"hosts", s.config.NumHosts,
		"chassis", s.config.NumChassis)

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

	s.js, err = jetstream.New(nc)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: %w", ErrJetStreamInitFailed, err)
	}

	if s.config.PersistStateChanges {
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
		Name:        s.config.ServiceName,
		Description: s.config.ServiceDescription,
		Version:     s.config.ServiceVersion,
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
		attribute.String("service.name", s.config.ServiceName),
		attribute.String("service.version", s.config.ServiceVersion),
		attribute.Int("state_machines.count", len(s.stateMachines)),
	)

	<-ctx.Done()

	// Capture the context error before shutdown
	// to return it after shutdown completes
	err = ctx.Err()

	// We need to create a new context here because the passed-in ctx is already canceled
	ctx = context.WithoutCancel(ctx)
	s.logger.InfoContext(ctx, "Shutting down state manager service")
	s.shutdown(ctx)

	return err
}

func (s *StateMgr) setupJetStream(ctx context.Context) error {
	streamConfig := jetstream.StreamConfig{
		Name:        s.config.StreamName,
		Description: "State manager persistence stream",
		Subjects:    s.config.StreamSubjects,
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      s.config.StreamRetention,
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

	if s.config.EnableHostManagement {
		for i := range s.config.NumHosts {
			hostID := i // capture loop variable
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

	if s.config.EnableChassisManagement {
		for i := range s.config.NumChassis {
			chassisID := i // capture loop variable
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

	if s.config.EnableBMCManagement {
		tasks = append(tasks, func(ctx context.Context, errChan chan error) {
			bmcName := "bmc.0" // Assuming a single BMC for simplicity, can be extended if needed
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
	if s.config.EnableHostManagement {
		for i := range s.config.NumHosts {
			hostEndpoints := []string{
				fmt.Sprintf("statemgr.host.%d.state", i),
				fmt.Sprintf("statemgr.host.%d.control", i),
				fmt.Sprintf("statemgr.host.%d.info", i),
			}

			for _, endpoint := range hostEndpoints {
				if err := s.microService.AddEndpoint(endpoint,
					micro.HandlerFunc(s.createRequestHandler(s.handleHostStateRequest))); err != nil { //nolint:contextcheck
					return fmt.Errorf("failed to register host endpoint %s: %w", endpoint, err)
				}
			}
		}
	}

	if s.config.EnableChassisManagement {
		for i := range s.config.NumChassis {
			chassisEndpoints := []string{
				fmt.Sprintf("statemgr.chassis.%d.state", i),
				fmt.Sprintf("statemgr.chassis.%d.control", i),
				fmt.Sprintf("statemgr.chassis.%d.info", i),
			}

			for _, endpoint := range chassisEndpoints {
				if err := s.microService.AddEndpoint(endpoint,
					micro.HandlerFunc(s.createRequestHandler(s.handleChassisStateRequest))); err != nil { //nolint:contextcheck
					return fmt.Errorf("failed to register chassis endpoint %s: %w", endpoint, err)
				}
			}
		}
	}

	if s.config.EnableBMCManagement {
		bmcEndpoints := []string{
			"statemgr.bmc.0.state",
			"statemgr.bmc.0.control",
			"statemgr.bmc.0.info",
		}

		for _, endpoint := range bmcEndpoints {
			if err := s.microService.AddEndpoint(endpoint,
				micro.HandlerFunc(s.createRequestHandler(s.handleManagementControllerStateRequest))); err != nil { //nolint:contextcheck
				return fmt.Errorf("failed to register BMC endpoint %s: %w", endpoint, err)
			}
		}
	}

	return nil
}

func (s *StateMgr) createRequestHandler(handler func(context.Context, micro.Request)) micro.HandlerFunc {
	return func(req micro.Request) {
		// Extract telemetry context from the NATS request
		ctx := telemetry.GetCtxFromReq(req)

		// Add the context to the request for downstream handlers
		if s.tracer != nil {
			_, span := s.tracer.Start(ctx, "statemgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", s.config.ServiceName),
			)
			defer span.End()
		}

		handler(ctx, req)
	}
}

// shutdown gracefully stops all state machines and cleans up resources.
func (s *StateMgr) shutdown(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	// In case the context is already canceled, create a new one for shutdown operations
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

// getStateMachine safely retrieves a state machine by name.
func (s *StateMgr) getStateMachine(name string) (*state.FSM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sm, exists := s.stateMachines[name]
	return sm, exists
}
