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
	"go.opentelemetry.io/otel/metric"
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

func (s *StateMgr) createChassisStateMachine(ctx context.Context, chassisName string) (*state.Machine, error) {
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
			s.createChassisTransitionAction(ctx, chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_ON),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF.String(),
			s.createChassisTransitionAction(ctx, chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_POWER_CYCLE.String(),
			s.createChassisTransitionAction(ctx, chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_POWER_CYCLE),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF.String(),
			s.createChassisTransitionAction(ctx, chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF.String(),
			s.createChassisTransitionAction(ctx, chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_EMERGENCY_SHUTDOWN.String(),
			s.createChassisTransitionAction(ctx, chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_EMERGENCY_SHUTDOWN),
		),
		state.WithActionTransition(
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
			schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
			schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF.String(),
			s.createChassisTransitionAction(ctx, chassisName, schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF),
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
		state.WithStateTimeout(s.config.stateTimeout),
		state.WithStateEntry(s.createChassisStatusEntryCallback(ctx, chassisName)),
		state.WithStateExit(s.createChassisStatusExitCallback(ctx, chassisName)),
		state.WithPersistence(s.createChassisPersistenceCallback(ctx, chassisName)),
		state.WithBroadcast(s.createChassisBroadcastCallback(ctx, chassisName)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chassis %s state machine: %w", chassisName, err)
	}

	return sm, nil
}

func (s *StateMgr) createChassisStatusEntryCallback(ctx context.Context, chassisName string) state.EntryCallback {
	return func(ctx context.Context, machineName, stateName string) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Chassis entering state",
				"chassis_name", chassisName,
				"state", stateName)
		}

		s.updateCurrentState(ctx, chassisName, stateName)

		switch stateName {
		case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String():
			if err := s.requestLEDAction(ctx, chassisName, "power_on"); err != nil && s.logger != nil {
				s.logger.ErrorContext(ctx, "Failed to request LED action",
					"chassis_name", chassisName,
					"action", "power_on",
					"error", err)
			}
		case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String():
			if err := s.requestLEDAction(ctx, chassisName, "power_off"); err != nil && s.logger != nil {
				s.logger.ErrorContext(ctx, "Failed to request LED action",
					"chassis_name", chassisName,
					"action", "power_off",
					"error", err)
			}
		case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String():
			if err := s.requestLEDAction(ctx, chassisName, "critical_error"); err != nil && s.logger != nil {
				s.logger.ErrorContext(ctx, "Failed to request LED action",
					"chassis_name", chassisName,
					"action", "critical_error",
					"error", err)
			}
		case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String():
			if err := s.requestLEDAction(ctx, chassisName, "warning"); err != nil && s.logger != nil {
				s.logger.ErrorContext(ctx, "Failed to request LED action",
					"chassis_name", chassisName,
					"action", "warning",
					"error", err)
			}
		case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String():
			if err := s.requestLEDAction(ctx, chassisName, "failed"); err != nil && s.logger != nil {
				s.logger.ErrorContext(ctx, "Failed to request LED action",
					"chassis_name", chassisName,
					"action", "failed",
					"error", err)
			}
		}

		return nil
	}
}

func (s *StateMgr) createChassisStatusExitCallback(ctx context.Context, chassisName string) state.ExitCallback {
	return func(ctx context.Context, machineName, stateName string) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Chassis exiting state",
				"chassis_name", chassisName,
				"state", stateName)
		}
		return nil
	}
}

func (s *StateMgr) createChassisTransitionAction(ctx context.Context, chassisName string, action schemav1alpha1.ChassisAction) state.ActionFunc {
	return func(from, to, trigger string) error {
		start := time.Now()

		if s.logger != nil {
			s.logger.InfoContext(ctx, "Chassis state transition",
				"chassis_name", chassisName,
				"from", from,
				"to", to,
				"action", action.String())
		}

		var powerAction string
		switch action {
		case schemav1alpha1.ChassisAction_CHASSIS_ACTION_ON:
			powerAction = "power_on"
		case schemav1alpha1.ChassisAction_CHASSIS_ACTION_OFF:
			powerAction = "power_off"
		case schemav1alpha1.ChassisAction_CHASSIS_ACTION_POWER_CYCLE:
			powerAction = "power_cycle"
		case schemav1alpha1.ChassisAction_CHASSIS_ACTION_EMERGENCY_SHUTDOWN:
			powerAction = "emergency_shutdown"
		default:
			return fmt.Errorf("unsupported chassis action: %v", action)
		}

		actionCtx, cancel := context.WithTimeout(ctx, s.config.stateTimeout)
		defer cancel()

		err := s.requestPowerAction(actionCtx, chassisName, powerAction)
		duration := time.Since(start)

		s.recordTransition(ctx, chassisName, from, to, trigger, duration, err)

		return err
	}
}

func (s *StateMgr) createChassisPersistenceCallback(ctx context.Context, componentName string) state.PersistenceCallback {
	return func(ctx context.Context, machineName, state string) error {
		if !s.config.persistStateChanges || s.js == nil {
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

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if _, err = s.js.Publish(ctx, subject, dataBytes); err != nil {
			return fmt.Errorf("%w: %w", ErrStatePersistenceFailed, err)
		}

		return nil
	}
}

func (s *StateMgr) createChassisBroadcastCallback(_ context.Context, componentName string) state.BroadcastCallback {
	return func(ctx context.Context, machineName, previousState, currentState string, trigger string) error {
		if !s.config.broadcastStateChanges || s.nc == nil {
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
	start := time.Now()

	if s.tracer != nil {
		var span trace.Span
		ctx, span = s.tracer.Start(ctx, "statemgr.handleChassisStateRequest")
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
		s.handleChassisControl(ctx, req, chassisName, start)
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

	currentState := sm.State(ctx)
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

func (s *StateMgr) handleChassisControl(ctx context.Context, req micro.Request, chassisName string, start time.Time) {
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

	if request.Action == schemav1alpha1.ChassisAction_CHASSIS_ACTION_UNSPECIFIED {
		ipc.RespondWithError(ctx, req, ErrInvalidTrigger, fmt.Sprintf("invalid action: %v", request.Action))
		return
	}

	previousState := sm.State(ctx)
	trigger := request.Action.String()
	if trigger == "" {
		ipc.RespondWithError(ctx, req, ErrInvalidTrigger, fmt.Sprintf("invalid action: %v", request.Action))
		return
	}

	if err := sm.Fire(ctx, trigger); err != nil {
		duration := time.Since(start)
		s.recordTransition(ctx, chassisName, previousState, sm.State(ctx), trigger, duration, err)

		if s.config.enableMetrics && s.stateTransitionDuration != nil {
			s.stateTransitionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
				attribute.String("component", chassisName),
				attribute.String("operation", "control"),
				attribute.String("status", "error"),
			))
		}

		ipc.RespondWithError(ctx, req, ErrStateTransitionFailed, err.Error())
		return
	}

	duration := time.Since(start)
	currentState := sm.State(ctx)
	s.recordTransition(ctx, chassisName, previousState, currentState, trigger, duration, nil)

	if s.config.enableMetrics && s.stateTransitionDuration != nil {
		s.stateTransitionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
			attribute.String("component", chassisName),
			attribute.String("operation", "control"),
			attribute.String("status", "success"),
		))
	}

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

	currentState := sm.State(ctx)
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
