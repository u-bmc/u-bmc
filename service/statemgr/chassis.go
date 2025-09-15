// SPDX-License-Identifier: BSD-3-Clause

package statemgr

import (
	"context"
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
	chassisTriggerTransitionCompleteOn  = "CHASSIS_TRANSITION_COMPLETE_ON"
	chassisTriggerTransitionCompleteOff = "CHASSIS_TRANSITION_COMPLETE_OFF"
	chassisTriggerTransitionWarning     = "CHASSIS_TRANSITION_WARNING"
	chassisTriggerTransitionCritical    = "CHASSIS_TRANSITION_CRITICAL"
	chassisTriggerTransitionFailed      = "CHASSIS_TRANSITION_FAILED"
	chassisTriggerTransitionClear       = "CHASSIS_TRANSITION_CLEAR"
)

func (s *StateMgr) createChassisStateMachine(chassisName string) (*state.Machine, error) {
	sm, err := state.NewStateMachine(
		state.WithName(chassisName),
		state.WithDescription(fmt.Sprintf("Chassis %s state machine", chassisName)),
		state.WithInitialState(schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String()),
		state.WithStates(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
		),
		// API-triggered transitions
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_ON.String(),
			s.createChassisTransitionAction(chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_ON),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF.String(),
			s.createChassisTransitionAction(chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_POWER_CYCLE.String(),
			s.createChassisTransitionAction(chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_POWER_CYCLE),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF.String(),
			s.createChassisTransitionAction(chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF.String(),
			s.createChassisTransitionAction(chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_EMERGENCY_SHUTDOWN.String(),
			s.createChassisTransitionAction(chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_EMERGENCY_SHUTDOWN),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF.String(),
			s.createChassisTransitionAction(chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF),
		),
		// Internal transitions (not exposed via API)
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			chassisTriggerTransitionCompleteOn,
		),
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
			chassisTriggerTransitionCompleteOff,
		),
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
			chassisTriggerTransitionFailed,
		),
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
			chassisTriggerTransitionWarning,
		),
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			chassisTriggerTransitionCritical,
		),
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			chassisTriggerTransitionClear,
		),
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			chassisTriggerTransitionCritical,
		),
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
			chassisTriggerTransitionWarning,
		),
		state.WithTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
			chassisTriggerTransitionFailed,
		),
		state.WithStateTimeout(s.config.StateTimeout),
		state.WithStateEntry(s.createChassisStatusEntryCallback(chassisName)),
		state.WithStateExit(s.createChassisStatusExitCallback(chassisName)),
		state.WithPersistence(s.createChassisPersistenceCallback(chassisName)),
		state.WithBroadcast(s.createChassisBroadcastCallback(chassisName)),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chassis %s state machine: %w", chassisName, err)
	}

	return sm, nil
}

func (s *StateMgr) createChassisStatusEntryCallback(chassisName string) state.StateEntryCallback {
	return func(machineName, stateName string) error {
		if s.logger != nil {
			s.logger.Info("Chassis entering state",
				"chassis_name", chassisName,
				"state", stateName)
		}
		return nil
	}
}

func (s *StateMgr) createChassisStatusExitCallback(chassisName string) state.StateExitCallback {
	return func(machineName, stateName string) error {
		if s.logger != nil {
			s.logger.Info("Chassis exiting state",
				"chassis_name", chassisName,
				"state", stateName)
		}
		return nil
	}
}

func (s *StateMgr) createChassisTransitionAction(chassisName string, action schemav1alpha1.ChassisAction) state.ActionFunc {
	return func(from, to, trigger string) error {
		if s.logger != nil {
			s.logger.Info("Chassis state transition",
				"chassis_name", chassisName,
				"from", from,
				"to", to,
				"action", action.String())
		}
		return nil
	}
}

func (s *StateMgr) createChassisPersistenceCallback(componentName string) state.PersistenceCallback {
	return func(machineName, state string) error {
		if !s.config.PersistStateChanges || s.js == nil {
			return nil
		}

		subject := fmt.Sprintf("statemgr.state.%s", componentName)

		stateEvent := &schemav1alpha1.ChassisStateChange{
			ChassisName:   componentName,
			CurrentStatus: chassisStatusStringToEnum(state),
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

func (s *StateMgr) createChassisBroadcastCallback(componentName string) state.BroadcastCallback {
	return func(machineName, previousState, currentState string, trigger string) error {
		if !s.config.BroadcastStateChanges || s.nc == nil {
			return nil
		}

		subject := fmt.Sprintf("statemgr.event.%s.transition", componentName)

		transitionEvent := &schemav1alpha1.ChassisStateChange{
			ChassisName:    componentName,
			PreviousStatus: chassisStatusStringToEnum(previousState),
			CurrentStatus:  chassisStatusStringToEnum(currentState),
			Cause:          chassisActionStringToEnum(trigger),
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

func (s *StateMgr) handleChassisStateRequest(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		var span trace.Span
		_, span = s.tracer.Start(ctx, "statemgr.handleChassisStateRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 4 {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	chassisID := parts[2]
	chassisName := fmt.Sprintf("chassis.%s", chassisID)
	operation := parts[3]

	switch operation {
	case operationState:
		s.handleGetChassisState(ctx, req, chassisName)
	case operationControl:
		s.handleChassisControl(ctx, req, chassisName)
	case operationInfo:
		s.handleGetChassisInfo(ctx, req, chassisName)
	default:
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

func (s *StateMgr) handleGetChassisState(ctx context.Context, req micro.Request, chassisName string) {
	sm, exists := s.getStateMachine(chassisName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("chassis %s not found", chassisName))
		return
	}

	currentState := sm.State()
	statusEnum := chassisStatusStringToEnum(currentState)

	response := &schemav1alpha1.GetChassisResponse{
		Chassis: []*schemav1alpha1.Chassis{
			{
				Name:   chassisName,
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

func (s *StateMgr) handleChassisControl(ctx context.Context, req micro.Request, chassisName string) {
	var request schemav1alpha1.ChangeChassisStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	sm, exists := s.getStateMachine(chassisName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("chassis %s not found", chassisName))
		return
	}

	trigger := request.Action.String()
	if trigger == "" {
		ipc.RespondWithError(ctx, req, ErrInvalidTrigger, fmt.Sprintf("invalid action: %v", request.Action))
		return
	}

	if err := sm.Fire(ctx, trigger); err != nil {
		ipc.RespondWithError(ctx, req, ErrStateTransitionFailed, err.Error())
		return
	}

	currentState := sm.State()
	statusEnum := chassisStatusStringToEnum(currentState)

	response := &schemav1alpha1.ChangeChassisStateResponse{
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

func (s *StateMgr) handleGetChassisInfo(ctx context.Context, req micro.Request, chassisName string) {
	sm, exists := s.getStateMachine(chassisName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("chassis %s not found", chassisName))
		return
	}

	currentState := sm.State()
	statusEnum := chassisStatusStringToEnum(currentState)

	response := &schemav1alpha1.Chassis{
		Name:   chassisName,
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

func chassisStatusStringToEnum(stateName string) schemav1alpha1.ChassisStatus {
	if stateValue, ok := schemav1alpha1.ChassisStatus_value[stateName]; ok {
		return schemav1alpha1.ChassisStatus(stateValue)
	}
	return schemav1alpha1.ChassisStatus_CHASSIS_STATUS_UNSPECIFIED
}

func chassisActionStringToEnum(actionName string) schemav1alpha1.ChassisAction {
	if actionValue, ok := schemav1alpha1.ChassisAction_value[actionName]; ok {
		return schemav1alpha1.ChassisAction(actionValue)
	}
	return schemav1alpha1.ChassisAction_CHASSIS_ACTION_UNSPECIFIED
}
