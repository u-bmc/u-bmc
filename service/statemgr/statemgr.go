// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	"github.com/u-bmc/u-bmc/pkg/state"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

const (
	operationState      = "state"
	operationControl    = "control"
	operationInfo       = "info"
	operationTransition = "transition"
)

var _ service.Service = (*StateMgr)(nil)

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

func New(opts ...Option) *StateMgr {
	config := NewConfig(opts...)

	return &StateMgr{
		config:        config,
		stateMachines: make(map[string]*state.FSM),
		tracer:        otel.Tracer("statemgr"),
	}
}

func (s *StateMgr) Name() string {
	return s.config.ServiceName
}

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

	s.logger = slog.Default().With("service", s.config.ServiceName)
	s.logger.Info("Starting state manager service",
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

	if err := s.registerEndpoints(); err != nil {
		span.RecordError(err)
		return err
	}

	s.logger.Info("State manager service started successfully",
		"endpoints_registered", true,
		"state_machines_initialized", len(s.stateMachines))

	span.SetAttributes(
		attribute.String("service.name", s.config.ServiceName),
		attribute.String("service.version", s.config.ServiceVersion),
		attribute.Int("state_machines.count", len(s.stateMachines)),
	)

	<-ctx.Done()

	s.logger.Info("Shutting down state manager service")
	s.shutdown(ctx)

	return ctx.Err()
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
		s.logger.Info("JetStream stream configured",
			"name", info.Config.Name,
			"subjects", info.Config.Subjects,
			"messages", info.State.Msgs)
	}

	return nil
}

func (s *StateMgr) initializeStateMachines(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, s.config.NumHosts+s.config.NumChassis+1)

	if s.config.EnableHostManagement {
		for i := 0; i < s.config.NumHosts; i++ {
			wg.Add(1)
			go func(hostID int) {
				defer wg.Done()
				sm, err := s.createHostStateMachine(hostID) //nolint:contextcheck
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
			}(i)
		}
	}

	if s.config.EnableChassisManagement {
		for i := 0; i < s.config.NumChassis; i++ {
			wg.Add(1)
			go func(chassisID int) {
				defer wg.Done()
				sm, err := s.createChassisStateMachine(chassisID) //nolint:contextcheck
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
			}(i)
		}
	}

	if s.config.EnableBMCManagement {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm, err := s.createBMCStateMachine(0) //nolint:contextcheck
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
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StateMgr) registerEndpoints() error {
	if s.config.EnableHostManagement {
		for i := 0; i < s.config.NumHosts; i++ {
			if err := s.microService.AddEndpoint(fmt.Sprintf("statemgr.host.%d.state", i),
				micro.HandlerFunc(s.handleHostStateRequest)); err != nil {
				return fmt.Errorf("failed to register host %d state endpoint: %w", i, err)
			}
			if err := s.microService.AddEndpoint(fmt.Sprintf("statemgr.host.%d.transition", i),
				micro.HandlerFunc(s.handleHostStateRequest)); err != nil {
				return fmt.Errorf("failed to register host %d transition endpoint: %w", i, err)
			}
			if err := s.microService.AddEndpoint(fmt.Sprintf("statemgr.host.%d.info", i),
				micro.HandlerFunc(s.handleHostStateRequest)); err != nil {
				return fmt.Errorf("failed to register host %d info endpoint: %w", i, err)
			}
		}
	}

	if s.config.EnableChassisManagement {
		for i := 0; i < s.config.NumChassis; i++ {
			if err := s.microService.AddEndpoint(fmt.Sprintf("statemgr.chassis.%d.state", i),
				micro.HandlerFunc(s.handleChassisStateRequest)); err != nil {
				return fmt.Errorf("failed to register chassis %d state endpoint: %w", i, err)
			}
			if err := s.microService.AddEndpoint(fmt.Sprintf("statemgr.chassis.%d.transition", i),
				micro.HandlerFunc(s.handleChassisStateRequest)); err != nil {
				return fmt.Errorf("failed to register chassis %d transition endpoint: %w", i, err)
			}
			if err := s.microService.AddEndpoint(fmt.Sprintf("statemgr.chassis.%d.control", i),
				micro.HandlerFunc(s.handleChassisStateRequest)); err != nil {
				return fmt.Errorf("failed to register chassis %d control endpoint: %w", i, err)
			}
			if err := s.microService.AddEndpoint(fmt.Sprintf("statemgr.chassis.%d.info", i),
				micro.HandlerFunc(s.handleChassisStateRequest)); err != nil {
				return fmt.Errorf("failed to register chassis %d info endpoint: %w", i, err)
			}
		}
	}

	if s.config.EnableBMCManagement {
		if err := s.microService.AddEndpoint("statemgr.bmc.0.state",
			micro.HandlerFunc(s.handleBMCStateRequest)); err != nil {
			return fmt.Errorf("failed to register BMC state endpoint: %w", err)
		}
		if err := s.microService.AddEndpoint("statemgr.bmc.0.control",
			micro.HandlerFunc(s.handleBMCStateRequest)); err != nil {
			return fmt.Errorf("failed to register BMC control endpoint: %w", err)
		}
		if err := s.microService.AddEndpoint("statemgr.bmc.0.info",
			micro.HandlerFunc(s.handleBMCStateRequest)); err != nil {
			return fmt.Errorf("failed to register BMC info endpoint: %w", err)
		}
	}

	if err := s.microService.AddEndpoint("statemgr.list",
		micro.HandlerFunc(s.handleListRequest)); err != nil {
		return fmt.Errorf("failed to register list endpoint: %w", err)
	}

	return nil
}

func (s *StateMgr) handleListRequest(req micro.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type componentInfo struct {
		Name     string            `json:"name"`
		State    string            `json:"state"`
		Metadata map[string]string `json:"metadata"`
	}

	type listResponse struct {
		Service    string          `json:"service"`
		Version    string          `json:"version"`
		Components []componentInfo `json:"components"`
		Timestamp  int64           `json:"timestamp"`
	}

	components := make([]componentInfo, 0, len(s.stateMachines))
	for name, sm := range s.stateMachines {
		triggers := sm.PermittedTriggers()
		metadata := make(map[string]string)
		metadata["can_fire"] = fmt.Sprintf("%t", len(triggers) > 0)

		for i, trigger := range triggers {
			metadata[fmt.Sprintf("trigger_%d", i)] = trigger
		}

		components = append(components, componentInfo{
			Name:     name,
			State:    sm.CurrentState(),
			Metadata: metadata,
		})
	}

	response := listResponse{
		Service:    s.config.ServiceName,
		Version:    s.config.ServiceVersion,
		Components: components,
		Timestamp:  time.Now().Unix(),
	}

	data, err := json.Marshal(response)
	if err != nil {
		s.respondWithError(req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil {
		s.logger.Error("Failed to respond to list request",
			"error", err,
			"subject", req.Subject())
	}
}

func (s *StateMgr) respondWithProtobuf(req micro.Request, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		s.respondWithError(req, ErrMarshalingFailed, err.Error())
		return
	}
	if err := req.Respond(data); err != nil {
		s.logger.Error("Failed to send protobuf response",
			"subject", req.Subject(),
			"error", err)
	}
}

func (s *StateMgr) respondWithError(req micro.Request, err error, details string) {
	s.logger.Error("Request failed",
		"subject", req.Subject(),
		"error", err,
		"details", details)

	if err := req.Error("500", fmt.Sprintf("%v: %s", err, details), nil); err != nil {
		s.logger.Error("Failed to send error response",
			"subject", req.Subject(),
			"error", err)
	}
}

func (s *StateMgr) shutdown(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	for name, sm := range s.stateMachines {
		if err := sm.Stop(ctx); err != nil {
			s.logger.Error("Failed to stop state machine",
				"name", name,
				"error", err)
		}
	}

	s.started = false
}
