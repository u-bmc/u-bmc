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

func (s *StateMgr) createHostStateMachine(hostName string) (*state.Machine, error) {
	sm, err := state.NewStateMachine(
		state.WithName(hostName),
		state.WithDescription(fmt.Sprintf("Host %s state machine", hostName)),
		state.WithInitialState(schemav1alpha1.HostStatus_HOST_STATUS_OFF.String()),
		state.WithStates(
			schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_DIAGNOSTIC.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_ERROR.String(),
		),
		// API-triggered transitions
		state.WithActionTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostAction_HOST_ACTION_ON.String(),
			s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_ON),
		),
		state.WithActionTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostAction_HOST_ACTION_OFF.String(),
			s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_OFF),
		),
		state.WithActionTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostAction_HOST_ACTION_REBOOT.String(),
			s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_REBOOT),
		),
		state.WithActionTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_ERROR.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
			schemav1alpha1.HostAction_HOST_ACTION_FORCE_OFF.String(),
			s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_FORCE_OFF),
		),
		state.WithActionTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_ERROR.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostAction_HOST_ACTION_FORCE_RESTART.String(),
			s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_FORCE_RESTART),
		),
		state.WithActionTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
			schemav1alpha1.HostAction_HOST_ACTION_OFF.String(),
			s.createHostTransitionAction(hostName, schemav1alpha1.HostAction_HOST_ACTION_OFF),
		),
		// Internal transitions (not exposed via API)
		state.WithTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
			hostTriggerTransitionCompleteOn,
		),
		state.WithTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_OFF.String(),
			hostTriggerTransitionCompleteOff,
		),
		state.WithTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_ERROR.String(),
			hostTriggerTransitionError,
		),
		state.WithTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_TRANSITIONING.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED.String(),
			hostTriggerTransitionTimeout,
		),
		state.WithTransition(
			schemav1alpha1.HostStatus_HOST_STATUS_QUIESCED.String(),
			schemav1alpha1.HostStatus_HOST_STATUS_ON.String(),
			hostTriggerTransitionResume,
		),
		state.WithStateTimeout(s.config.StateTimeout),
		state.WithStateEntry(s.createHostStatusEntryCallback(hostName)),
		state.WithStateExit(s.createHostStatusExitCallback(hostName)),
		state.WithPersistence(s.createHostPersistenceCallback(hostName)),
		state.WithBroadcast(s.createHostBroadcastCallback(hostName)),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create host %s state machine: %w", hostName, err)
	}

	return sm, nil
}

func (s *StateMgr) createHostStatusEntryCallback(hostName string) state.StateEntryCallback {
	return func(machineName, stateName string) error {
		if s.logger != nil {
			s.logger.Info("Host entering state",
				"host_name", hostName,
				"state", stateName)
		}
		return nil
	}
}

func (s *StateMgr) createHostStatusExitCallback(hostName string) state.StateExitCallback {
	return func(machineName, stateName string) error {
		if s.logger != nil {
			s.logger.Info("Host exiting state",
				"host_name", hostName,
				"state", stateName)
		}
		return nil
	}
}

func (s *StateMgr) createHostTransitionAction(hostName string, action schemav1alpha1.HostAction) state.ActionFunc {
	return func(from, to, trigger string) error {
		if s.logger != nil {
			s.logger.Info("Host state transition",
				"host_name", hostName,
				"from", from,
				"to", to,
				"action", action.String())
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

		stateEvent := &schemav1alpha1.HostStateChange{
			HostName:      componentName,
			CurrentStatus: hostStatusStringToEnum(state),
			ChangedAt:     timestamppb.New(time.Now().UTC()),
		}

		dataBytes, err := stateEvent.MarshalVT()
		if err != nil {
			return fmt.Errorf("failed to marshal state event: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err = s.js.Publish(ctx, subject, dataBytes); err != nil {
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

	currentState := sm.State()
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

	currentState := sm.State()
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

	currentState := sm.State()
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
