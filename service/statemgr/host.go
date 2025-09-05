// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go/micro"
	schemav1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/state"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

func (s *StateMgr) createHostStateMachine(hostID int) (*state.FSM, error) {
	hostName := fmt.Sprintf("host.%d", hostID)

	config := state.NewConfig(
		state.WithName(hostName),
		state.WithDescription(fmt.Sprintf("Host %d state machine", hostID)),
		state.WithInitialState(schemav1alpha1.HostState_HOST_STATE_OFF.String()),
		state.WithStates(
			state.StateDefinition{
				Name:        schemav1alpha1.HostState_HOST_STATE_OFF.String(),
				Description: "Host is powered off",
				OnEntry:     s.createHostStateEntryAction(hostID, schemav1alpha1.HostState_HOST_STATE_OFF),
				OnExit:      s.createHostStateExitAction(hostID, schemav1alpha1.HostState_HOST_STATE_OFF),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostState_HOST_STATE_ON.String(),
				Description: "Host is powered on and running",
				OnEntry:     s.createHostStateEntryAction(hostID, schemav1alpha1.HostState_HOST_STATE_ON),
				OnExit:      s.createHostStateExitAction(hostID, schemav1alpha1.HostState_HOST_STATE_ON),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostState_HOST_STATE_TRANSITIONING.String(),
				Description: "Host is transitioning between states",
				OnEntry:     s.createHostStateEntryAction(hostID, schemav1alpha1.HostState_HOST_STATE_TRANSITIONING),
				OnExit:      s.createHostStateExitAction(hostID, schemav1alpha1.HostState_HOST_STATE_TRANSITIONING),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostState_HOST_STATE_QUIESCED.String(),
				Description: "Host is in quiesced state",
				OnEntry:     s.createHostStateEntryAction(hostID, schemav1alpha1.HostState_HOST_STATE_QUIESCED),
				OnExit:      s.createHostStateExitAction(hostID, schemav1alpha1.HostState_HOST_STATE_QUIESCED),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostState_HOST_STATE_DIAGNOSTIC_MODE.String(),
				Description: "Host is in diagnostic mode",
				OnEntry:     s.createHostStateEntryAction(hostID, schemav1alpha1.HostState_HOST_STATE_DIAGNOSTIC_MODE),
				OnExit:      s.createHostStateExitAction(hostID, schemav1alpha1.HostState_HOST_STATE_DIAGNOSTIC_MODE),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostState_HOST_STATE_ERROR.String(),
				Description: "Host is in error state",
				OnEntry:     s.createHostStateEntryAction(hostID, schemav1alpha1.HostState_HOST_STATE_ERROR),
				OnExit:      s.createHostStateExitAction(hostID, schemav1alpha1.HostState_HOST_STATE_ERROR),
			},
		),
		state.WithTransitions(
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_OFF.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_TRANSITIONING.String(),
				Trigger: schemav1alpha1.HostTransition_HOST_TRANSITION_ON.String(),
				Action:  s.createHostTransitionAction(hostID, schemav1alpha1.HostTransition_HOST_TRANSITION_ON),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_ON.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_TRANSITIONING.String(),
				Trigger: schemav1alpha1.HostTransition_HOST_TRANSITION_OFF.String(),
				Action:  s.createHostTransitionAction(hostID, schemav1alpha1.HostTransition_HOST_TRANSITION_OFF),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_ON.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_TRANSITIONING.String(),
				Trigger: schemav1alpha1.HostTransition_HOST_TRANSITION_REBOOT.String(),
				Action:  s.createHostTransitionAction(hostID, schemav1alpha1.HostTransition_HOST_TRANSITION_REBOOT),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_ON.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_QUIESCED.String(),
				Trigger: "HOST_TRANSITION_QUIESCE",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_TRANSITIONING.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_ON.String(),
				Trigger: "HOST_TRANSITION_COMPLETE_ON",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_TRANSITIONING.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_OFF.String(),
				Trigger: "HOST_TRANSITION_COMPLETE_OFF",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_TRANSITIONING.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_ERROR.String(),
				Trigger: "HOST_TRANSITION_ERROR",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_ERROR.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_OFF.String(),
				Trigger: schemav1alpha1.HostTransition_HOST_TRANSITION_FORCE_OFF.String(),
				Action:  s.createHostTransitionAction(hostID, schemav1alpha1.HostTransition_HOST_TRANSITION_FORCE_OFF),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_ERROR.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_TRANSITIONING.String(),
				Trigger: schemav1alpha1.HostTransition_HOST_TRANSITION_FORCE_RESTART.String(),
				Action:  s.createHostTransitionAction(hostID, schemav1alpha1.HostTransition_HOST_TRANSITION_FORCE_RESTART),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_QUIESCED.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_ON.String(),
				Trigger: "HOST_TRANSITION_RESUME",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostState_HOST_STATE_QUIESCED.String(),
				To:      schemav1alpha1.HostState_HOST_STATE_OFF.String(),
				Trigger: schemav1alpha1.HostTransition_HOST_TRANSITION_OFF.String(),
				Action:  s.createHostTransitionAction(hostID, schemav1alpha1.HostTransition_HOST_TRANSITION_OFF),
			},
		),
		state.WithPersistState(s.config.PersistStateChanges),
		state.WithStateTimeout(s.config.StateTimeout),
		state.WithMetrics(s.config.EnableMetrics),
		state.WithTracing(s.config.EnableTracing),
	)

	sm, err := state.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create host %d state machine: %w", hostID, err)
	}

	if err := sm.SetPersistenceCallback(s.createHostPersistenceCallback(hostName)); err != nil {
		return nil, fmt.Errorf("failed to set persistence callback: %w", err)
	}
	if err := sm.SetBroadcastCallback(s.createHostBroadcastCallback(hostName)); err != nil {
		return nil, fmt.Errorf("failed to set broadcast callback: %w", err)
	}

	return sm, nil
}

func (s *StateMgr) createHostStateEntryAction(hostID int, hostState schemav1alpha1.HostState) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.Info("Host entering state",
				"host_id", hostID,
				"state", hostState.String())
		}
		return nil
	}
}

func (s *StateMgr) createHostStateExitAction(hostID int, hostState schemav1alpha1.HostState) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.Info("Host exiting state",
				"host_id", hostID,
				"state", hostState.String())
		}
		return nil
	}
}

func (s *StateMgr) createHostTransitionAction(hostID int, transition schemav1alpha1.HostTransition) state.TransitionAction {
	return func(ctx context.Context, from, to string) error {
		if s.logger != nil {
			s.logger.Info("Host state transition",
				"host_id", hostID,
				"from", from,
				"to", to,
				"transition", transition.String())
		}
		return nil
	}
}

func (s *StateMgr) createHostPersistenceCallback(componentName string) state.PersistenceCallback {
	return func(machineName, state string) error {
		if !s.config.PersistStateChanges || s.js == nil {
			return nil
		}

		subject := fmt.Sprintf("statemgr.state.%s", componentName)

		stateEvent := map[string]interface{}{
			"component": componentName,
			"state":     state,
			"timestamp": time.Now().Unix(),
		}

		dataBytes, err := json.Marshal(stateEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal state event: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:contextcheck
		defer cancel()

		_, err = s.js.Publish(ctx, subject, dataBytes)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrStatePersistenceFailed, err)
		}

		return nil
	}
}

func (s *StateMgr) createHostBroadcastCallback(componentName string) state.BroadcastCallback {
	return func(machineName, previousState, currentState string, trigger string) error {
		if !s.config.BroadcastStateChanges || s.nc == nil {
			return nil
		}

		subject := fmt.Sprintf("statemgr.event.%s.transition", componentName)

		transitionEvent := map[string]interface{}{
			"component":      componentName,
			"previous_state": previousState,
			"current_state":  currentState,
			"trigger":        trigger,
			"timestamp":      time.Now().Unix(),
		}

		eventBytes, err := json.Marshal(transitionEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal transition event: %w", err)
		}

		err = s.nc.Publish(subject, eventBytes)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrBroadcastFailed, err)
		}

		return nil
	}
}

func (s *StateMgr) handleHostStateRequest(req micro.Request) {
	ctx := telemetry.GetCtxFromReq(req)
	if s.tracer != nil {
		var span trace.Span
		_, span = s.tracer.Start(ctx, "statemgr.handleHostStateRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 4 {
		s.respondWithError(req, ErrInvalidRequest, "invalid subject format")
		return
	}

	hostIDStr := parts[2]
	operation := parts[3]

	hostID, err := strconv.Atoi(hostIDStr)
	if err != nil {
		s.respondWithError(req, ErrInvalidComponentID, fmt.Sprintf("invalid host ID: %s", hostIDStr))
		return
	}

	hostName := fmt.Sprintf("host.%d", hostID)

	switch operation {
	case operationState:
		s.handleGetHostState(req, hostName)
	case operationTransition:
		s.handleHostTransition(req, hostName)
	case operationInfo:
		s.handleGetHostInfo(req, hostName)
	default:
		s.respondWithError(req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

func (s *StateMgr) handleGetHostState(req micro.Request, hostName string) {
	sm, exists := s.stateMachines[hostName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("host %s not found", hostName))
		return
	}

	currentState := sm.CurrentState()
	stateEnum := s.hostStateStringToEnum(currentState)

	response := &schemav1alpha1.GetHostResponse{
		Hosts: []*schemav1alpha1.Host{
			{
				Name:         hostName,
				CurrentState: &stateEnum,
			},
		},
	}

	data, err := response.MarshalVT()
	if err != nil {
		s.respondWithError(req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to respond to request", "error", err)
		}
	}
}

func (s *StateMgr) handleHostTransition(req micro.Request, hostName string) {
	ctx := telemetry.GetCtxFromReq(req)

	var request schemav1alpha1.HostChangeStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		s.respondWithError(req, ErrUnmarshalingFailed, err.Error())
		return
	}

	sm, exists := s.stateMachines[hostName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("host %s not found", hostName))
		return
	}

	trigger := s.hostTransitionToTrigger(request.Transition)
	if trigger == "" {
		s.respondWithError(req, ErrInvalidTrigger, fmt.Sprintf("invalid transition: %v", request.Transition))
		return
	}

	data := map[string]interface{}{
		"force": request.GetForce(),
	}

	for k, v := range request.Metadata {
		data[k] = v
	}

	err := sm.Fire(ctx, trigger, data)
	if err != nil {
		s.respondWithError(req, ErrStateTransitionFailed, err.Error())
		return
	}

	currentState := sm.CurrentState()
	stateEnum := s.hostStateStringToEnum(currentState)

	response := &schemav1alpha1.HostChangeStateResponse{
		Success:      true,
		CurrentState: &stateEnum,
		TransitionId: proto.String(fmt.Sprintf("%s-%d", hostName, time.Now().UnixNano())),
		Metadata:     request.Metadata,
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		s.respondWithError(req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(responseData); err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to respond to request", "error", err)
		}
	}
}

func (s *StateMgr) handleGetHostInfo(req micro.Request, hostName string) {
	sm, exists := s.stateMachines[hostName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("host %s not found", hostName))
		return
	}

	currentState := sm.CurrentState()
	stateEnum := s.hostStateStringToEnum(currentState)
	triggers := sm.PermittedTriggers()

	response := &schemav1alpha1.Host{
		Name:         hostName,
		CurrentState: &stateEnum,
		Metadata: map[string]string{
			"permitted_triggers": strings.Join(triggers, ","),
		},
	}

	data, err := response.MarshalVT()
	if err != nil {
		s.respondWithError(req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to respond to request", "error", err)
		}
	}
}

func (s *StateMgr) hostStateStringToEnum(stateName string) schemav1alpha1.HostState {
	if stateValue, ok := schemav1alpha1.HostState_value[stateName]; ok {
		return schemav1alpha1.HostState(stateValue)
	}
	return schemav1alpha1.HostState_HOST_STATE_UNSPECIFIED
}

func (s *StateMgr) hostTransitionToTrigger(transition schemav1alpha1.HostTransition) string {
	return transition.String()
}
