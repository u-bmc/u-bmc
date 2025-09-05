// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go/micro"
	schemav1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/ipc"
	"github.com/u-bmc/u-bmc/pkg/state"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Internal triggers for state transitions not exposed via API.
const (
	hostTriggerTransitionCompleteOn  = "HOST_TRANSITION_COMPLETE_ON"
	hostTriggerTransitionCompleteOff = "HOST_TRANSITION_COMPLETE_OFF"
	hostTriggerTransitionError       = "HOST_TRANSITION_ERROR"
	hostTriggerTransitionResume      = "HOST_TRANSITION_RESUME"
	hostTriggerTransitionTimeout     = "HOST_TRANSITION_TIMEOUT"
)

func (s *StateMgr) createHostStateMachine(ctx context.Context, hostName string) (*state.FSM, error) {
	config := state.NewConfig(
		state.WithName(hostName),
		state.WithDescription(fmt.Sprintf("Host %s state machine", hostName)),
		state.WithInitialState(schemav1alpha1.HostStatus_HOST_STATUS_OFF.String()),
		state.WithStates(
			state.StateDefinition{
				Name:        schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
				Description: "Host is powered off",
				OnEntry:     s.createHostStatusEntryAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_OFF),
				OnExit:      s.createHostStatusExitAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_OFF),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
				Description: "Host is powered on and running",
				OnEntry:     s.createHostStatusEntryAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_ON),
				OnExit:      s.createHostStatusExitAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_ON),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				Description: "Host is transitioning between states",
				OnEntry:     s.createHostStatusEntryAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING),
				OnExit:      s.createHostStatusExitAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED.String(),
				Description: "Host is in quiesced state (entered automatically on timeout)",
				OnEntry:     s.createHostStatusEntryAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED),
				OnExit:      s.createHostStatusExitAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostStatus_HOST_STATUS_DIAGNOSTIC.String(),
				Description: "Host is in diagnostic mode",
				OnEntry:     s.createHostStatusEntryAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_DIAGNOSTIC),
				OnExit:      s.createHostStatusExitAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_DIAGNOSTIC),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.HostStatus_HOST_STATUS_ERROR.String(),
				Description: "Host is in error state (entered automatically on error)",
				OnEntry:     s.createHostStatusEntryAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_ERROR),
				OnExit:      s.createHostStatusExitAction(hostName, schemav1alpha1.HostStatus_HOST_STATUS_ERROR),
			},
		),
		state.WithTransitions(
			// API-triggered transitions
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				Trigger: schemav1alpha1.HostAction_HOST_ACTION_ON.String(),
				Action:  s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_ON),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				Trigger: schemav1alpha1.HostAction_HOST_ACTION_OFF.String(),
				Action:  s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_OFF),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				Trigger: schemav1alpha1.HostAction_HOST_ACTION_REBOOT.String(),
				Action:  s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_REBOOT),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_ERROR.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
				Trigger: schemav1alpha1.HostAction_HOST_ACTION_FORCE_OFF.String(),
				Action:  s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_FORCE_OFF),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_ERROR.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				Trigger: schemav1alpha1.HostAction_HOST_ACTION_FORCE_RESTART.String(),
				Action:  s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_FORCE_RESTART),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
				Trigger: schemav1alpha1.HostAction_HOST_ACTION_OFF.String(),
				Action:  s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_OFF),
			},
			// Internal transitions (not exposed via API)
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
				Trigger: hostTriggerTransitionCompleteOn,
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
				Trigger: hostTriggerTransitionCompleteOff,
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_ERROR.String(),
				Trigger: hostTriggerTransitionError,
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED.String(),
				Trigger: hostTriggerTransitionTimeout,
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED.String(),
				To:      schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
				Trigger: hostTriggerTransitionResume,
			},
		),
		state.WithPersistState(s.config.PersistStateChanges),
		state.WithStateTimeout(s.config.StateTimeout),
		state.WithMetrics(s.config.EnableMetrics),
		state.WithTracing(s.config.EnableTracing),
	)

	sm, err := state.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create host %s state machine: %w", hostName, err)
	}

	if err := sm.SetPersistenceCallback(s.createHostPersistenceCallback(ctx, hostName)); err != nil {
		return nil, fmt.Errorf("failed to set persistence callback: %w", err)
	}
	if err := sm.SetBroadcastCallback(s.createHostBroadcastCallback(ctx, hostName)); err != nil {
		return nil, fmt.Errorf("failed to set broadcast callback: %w", err)
	}

	return sm, nil
}

func (s *StateMgr) createHostStatusEntryAction(hostName string, hostStatus schemav1alpha1.HostStatus) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Host entering state",
				"host_name", hostName,
				"state", hostStatus.String())
		}
		return nil
	}
}

func (s *StateMgr) createHostStatusExitAction(hostName string, hostStatus schemav1alpha1.HostStatus) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Host exiting state",
				"host_name", hostName,
				"state", hostStatus.String())
		}
		return nil
	}
}

func (s *StateMgr) createHostTransitionAction(hostName string, action schemav1alpha1.HostAction) state.TransitionAction {
	return func(ctx context.Context, from, to string) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Host state transition",
				"host_name", hostName,
				"from", from,
				"to", to,
				"action", action.String())
		}
		return nil
	}
}

func (s *StateMgr) createHostPersistenceCallback(ctx context.Context, componentName string) state.PersistenceCallback {
	return func(machineName, state string) error {
		if !s.config.PersistStateChanges || s.js == nil {
			return nil
		}

		subject := fmt.Sprintf("statemgr.state.%s", componentName)

		stateEvent := &schemav1alpha1.HostStateChange{
			HostName:      componentName,
			CurrentStatus: hostStatusStringToEnum(state),
			ChangedAt:     timestamppb.New(time.Now().UTC()),
		}

		dataBytes, err := stateEvent.MarshalVT()
		if err != nil {
			return fmt.Errorf("failed to marshal state event: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if _, err = s.js.Publish(ctx, subject, dataBytes); err != nil {
			return fmt.Errorf("%w: %w", ErrStatePersistenceFailed, err)
		}

		return nil
	}
}

func (s *StateMgr) createHostBroadcastCallback(ctx context.Context, componentName string) state.BroadcastCallback {
	return func(machineName, previousState, currentState string, trigger string) error {
		if !s.config.BroadcastStateChanges || s.nc == nil {
			return nil
		}

		subject := fmt.Sprintf("statemgr.event.%s.transition", componentName)

		transitionEvent := &schemav1alpha1.HostStateChange{
			HostName:       componentName,
			PreviousStatus: hostStatusStringToEnum(previousState),
			CurrentStatus:  hostStatusStringToEnum(currentState),
			Cause:          hostActionStringToEnum(trigger),
			ChangedAt:      timestamppb.New(time.Now().UTC()),
		}

		eventBytes, err := transitionEvent.MarshalVT()
		if err != nil {
			return fmt.Errorf("failed to marshal transition event: %w", err)
		}

		if err = s.nc.Publish(subject, eventBytes); err != nil {
			return fmt.Errorf("%w: %w", ErrBroadcastFailed, err)
		}

		return nil
	}
}

func (s *StateMgr) handleHostStateRequest(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		var span trace.Span
		_, span = s.tracer.Start(ctx, "statemgr.handleHostStateRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 4 {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	hostID := parts[2]
	hostName := fmt.Sprintf("host.%s", hostID)
	operation := parts[3]

	switch operation {
	case operationState:
		s.handleGetHostState(ctx, req, hostName)
	case operationControl:
		s.handleHostControl(ctx, req, hostName)
	case operationInfo:
		s.handleGetHostInfo(ctx, req, hostName)
	default:
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

func (s *StateMgr) handleGetHostState(ctx context.Context, req micro.Request, hostName string) {
	sm, exists := s.getStateMachine(hostName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("host %s not found", hostName))
		return
	}

	currentState := sm.CurrentState()
	statusEnum := hostStatusStringToEnum(currentState)

	response := &schemav1alpha1.GetHostResponse{
		Hosts: []*schemav1alpha1.Host{
			{
				Name:   hostName,
				Status: &statusEnum,
			},
		},
	}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *StateMgr) handleHostControl(ctx context.Context, req micro.Request, hostName string) {
	var request schemav1alpha1.ChangeHostStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	sm, exists := s.getStateMachine(hostName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("host %s not found", hostName))
		return
	}

	// Explicitly reject the UNSPECIFIED enum value
	if request.Action == schemav1alpha1.HostAction_HOST_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidTrigger, fmt.Sprintf("invalid action: %v", request.Action))
		return
	}

	trigger := request.Action.String()
	if trigger == "" {
		ipc.RespondWithError(ctx, req, ErrInvalidTrigger, fmt.Sprintf("invalid action: %v", request.Action))
		return
	}

	if err := sm.Fire(ctx, trigger); err != nil {
		if !errors.Is(err, state.ErrTransitionTimeout) {
			ipc.RespondWithError(ctx, req, ErrStateTransitionFailed, err.Error())
			return
		}

		if s.logger != nil {
			s.logger.WarnContext(ctx, "Host transition timed out, triggering timeout transition",
				"host_name", hostName,
				"trigger", trigger)
		}

		// Trigger the internal timeout transition to move to QUIESCED state
		if timeoutErr := sm.Fire(ctx, hostTriggerTransitionTimeout); timeoutErr != nil {
			if s.logger != nil {
				s.logger.ErrorContext(ctx, "Failed to trigger timeout transition",
					"host_name", hostName,
					"error", timeoutErr)
			}
			// Return the original timeout error since the timeout transition also failed
			ipc.RespondWithError(ctx, req, ErrStateTransitionFailed, err.Error())
			return
		}
	}

	currentState := sm.CurrentState()
	statusEnum := hostStatusStringToEnum(currentState)

	response := &schemav1alpha1.ChangeHostStateResponse{
		CurrentStatus: statusEnum,
	}

	responseData, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(responseData); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *StateMgr) handleGetHostInfo(ctx context.Context, req micro.Request, hostName string) {
	sm, exists := s.getStateMachine(hostName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("host %s not found", hostName))
		return
	}

	currentState := sm.CurrentState()
	statusEnum := hostStatusStringToEnum(currentState)

	response := &schemav1alpha1.Host{
		Name:   hostName,
		Status: &statusEnum,
	}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func hostStatusStringToEnum(stateName string) schemav1alpha1.HostStatus {
	if stateValue, ok := schemav1alpha1.HostStatus_value[stateName]; ok {
		return schemav1alpha1.HostStatus(stateValue)
	}
	return schemav1alpha1.HostStatus_HOST_STATUS_UNSPECIFIED
}

func hostActionStringToEnum(actionName string) schemav1alpha1.HostAction {
	if actionValue, ok := schemav1alpha1.HostAction_value[actionName]; ok {
		return schemav1alpha1.HostAction(actionValue)
	}
	return schemav1alpha1.HostAction_HOST_ACTION_UNSPECIFIED
}
