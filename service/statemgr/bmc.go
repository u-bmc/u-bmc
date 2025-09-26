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
	managementControllerTriggerTransitionCompleteReady      = "MANAGEMENT_CONTROLLER_TRANSITION_COMPLETE_READY"
	managementControllerTriggerTransitionCompleteNotReady   = "MANAGEMENT_CONTROLLER_TRANSITION_COMPLETE_NOT_READY"
	managementControllerTriggerTransitionCompleteDisabled   = "MANAGEMENT_CONTROLLER_TRANSITION_COMPLETE_DISABLED"
	managementControllerTriggerTransitionCompleteError      = "MANAGEMENT_CONTROLLER_TRANSITION_COMPLETE_ERROR"
	managementControllerTriggerTransitionCompleteDiagnostic = "MANAGEMENT_CONTROLLER_TRANSITION_COMPLETE_DIAGNOSTIC"
	managementControllerTriggerTransitionCompleteQuiesced   = "MANAGEMENT_CONTROLLER_TRANSITION_COMPLETE_QUIESCED"
	managementControllerTriggerTransitionTimeout            = "MANAGEMENT_CONTROLLER_TRANSITION_TIMEOUT"
	managementControllerTriggerTransitionResume             = "MANAGEMENT_CONTROLLER_TRANSITION_RESUME"
)

func (s *StateMgr) createManagementControllerStateMachine(ctx context.Context, controllerName string) (*state.Machine, error) {
	sm, err := state.NewStateMachine(
		state.WithName(controllerName),
		state.WithDescription(fmt.Sprintf("Management Controller %s state machine", controllerName)),
		state.WithInitialState(schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String()),
		state.WithStates(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC.String(),
		),
		// API-triggered transitions
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_WARM_RESET.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_WARM_RESET),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_HARD_RESET.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_HARD_RESET),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_FACTORY_RESET.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_FACTORY_RESET),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_DISABLE.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_DISABLE),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_DISABLE.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_DISABLE),
		),
		state.WithActionTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE.String(),
			s.createManagementControllerTransitionAction(ctx, controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE),
		),
		// Internal transitions (not exposed via API)
		state.WithTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			managementControllerTriggerTransitionCompleteReady,
		),
		state.WithTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
			managementControllerTriggerTransitionCompleteError,
		),
		state.WithTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC.String(),
			managementControllerTriggerTransitionCompleteDiagnostic,
		),
		state.WithTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED.String(),
			managementControllerTriggerTransitionTimeout,
		),
		state.WithTransition(
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED.String(),
			schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
			managementControllerTriggerTransitionResume,
		),
		state.WithStateTimeout(s.config.stateTimeout),
		state.WithStateEntry(s.createManagementControllerStatusEntryCallback(ctx, controllerName)),
		state.WithStateExit(s.createManagementControllerStatusExitCallback(ctx, controllerName)),
		state.WithPersistence(s.createManagementControllerPersistenceCallback(ctx, controllerName)),
		state.WithBroadcast(s.createManagementControllerBroadcastCallback(ctx, controllerName)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create management controller %s state machine: %w", controllerName, err)
	}

	return sm, nil
}

func (s *StateMgr) createManagementControllerStatusEntryCallback(ctx context.Context, controllerName string) state.EntryCallback {
	return func(ctx context.Context, machineName, stateName string) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Management Controller entering status",
				"controller_name", controllerName,
				"status", stateName)
		}
		return nil
	}
}

func (s *StateMgr) createManagementControllerStatusExitCallback(ctx context.Context, controllerName string) state.ExitCallback {
	return func(ctx context.Context, machineName, stateName string) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Management Controller exiting status",
				"controller_name", controllerName,
				"status", stateName)
		}
		return nil
	}
}

func (s *StateMgr) createManagementControllerTransitionAction(ctx context.Context, controllerName string, action schemav1alpha1.ManagementControllerAction) state.ActionFunc {
	return func(from, to, trigger string) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Management Controller state transition",
				"controller_name", controllerName,
				"from", from,
				"to", to,
				"action", action.String())
		}
		return nil
	}
}

func (s *StateMgr) createManagementControllerPersistenceCallback(ctx context.Context, controllerName string) state.PersistenceCallback {
	return func(ctx context.Context, machineName, state string) error {
		if !s.config.persistStateChanges || s.js == nil {
			return nil
		}

		subject := fmt.Sprintf("statemgr.state.%s", controllerName)

		stateEvent := &schemav1alpha1.ManagementControllerStateChange{
			ControllerName: controllerName,
			CurrentStatus:  managementControllerStatusStringToEnum(state),
			ChangedAt:      timestamppb.New(time.Now().UTC()),
		}

		dataBytes, err := stateEvent.MarshalVT()
		if err != nil {
			return fmt.Errorf("failed to marshal state event: %w", err)
		}

		publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if _, err = s.js.Publish(publishCtx, subject, dataBytes); err != nil {
			return fmt.Errorf("%w: %w", ErrStatePersistenceFailed, err)
		}

		return nil
	}
}

func (s *StateMgr) createManagementControllerBroadcastCallback(_ context.Context, componentName string) state.BroadcastCallback {
	return func(ctx context.Context, machineName, previousState, currentState string, trigger string) error {
		if !s.config.broadcastStateChanges || s.nc == nil {
			return nil
		}

		subject := fmt.Sprintf("statemgr.event.%s.transition", componentName)

		transitionEvent := &schemav1alpha1.ManagementControllerStateChange{
			ControllerName: componentName,
			PreviousStatus: managementControllerStatusStringToEnum(previousState),
			CurrentStatus:  managementControllerStatusStringToEnum(currentState),
			Cause:          managementControllerActionStringToEnum(trigger),
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

func (s *StateMgr) handleManagementControllerStateRequest(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		var span trace.Span
		ctx, span = s.tracer.Start(ctx, "statemgr.handleManagementControllerStateRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	// Validate subject format: statemgr.controller.{name}.{operation}
	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 4 || parts[0] != "statemgr" || (parts[1] != "controller" && parts[1] != "bmc") {
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, "invalid subject format")
		return
	}

	controllerID := parts[2]
	controllerName := fmt.Sprintf("bmc.%s", controllerID)
	operation := parts[3]

	switch operation {
	case operationState:
		s.handleGetManagementControllerState(ctx, req, controllerName)
	case operationControl:
		s.handleManagementControllerControl(ctx, req, controllerName)
	case operationInfo:
		s.handleGetManagementControllerInfo(ctx, req, controllerName)
	default:
		ipc.RespondWithError(ctx, req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

func (s *StateMgr) handleGetManagementControllerState(ctx context.Context, req micro.Request, controllerName string) {
	sm, exists := s.getStateMachine(controllerName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("Management Controller %s not found", controllerName))
		return
	}

	currentState := sm.State(ctx)
	statusEnum := managementControllerStatusStringToEnum(currentState)

	response := &schemav1alpha1.GetManagementControllerResponse{
		Controllers: []*schemav1alpha1.ManagementController{
			{
				Name:   controllerName,
				Status: &statusEnum,
			},
		},
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

func (s *StateMgr) handleManagementControllerControl(ctx context.Context, req micro.Request, controllerName string) {
	var request schemav1alpha1.ChangeManagementControllerStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ErrUnmarshalingFailed, err.Error())
		return
	}

	sm, exists := s.getStateMachine(controllerName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("Management Controller %s not found", controllerName))
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

	currentState := sm.State(ctx)
	statusEnum := managementControllerStatusStringToEnum(currentState)

	response := &schemav1alpha1.ChangeManagementControllerStateResponse{
		CurrentStatus: statusEnum,
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

func (s *StateMgr) handleGetManagementControllerInfo(ctx context.Context, req micro.Request, controllerName string) {
	sm, exists := s.getStateMachine(controllerName)
	if !exists {
		ipc.RespondWithError(ctx, req, ErrComponentNotFound, fmt.Sprintf("Management Controller %s not found", controllerName))
		return
	}

	currentState := sm.State(ctx)
	statusEnum := managementControllerStatusStringToEnum(currentState)

	response := &schemav1alpha1.ManagementController{
		Name:   controllerName,
		Status: &statusEnum,
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

func managementControllerStatusStringToEnum(status string) schemav1alpha1.ManagementControllerStatus {
	if stateValue, ok := schemav1alpha1.ManagementControllerStatus_value[status]; ok {
		return schemav1alpha1.ManagementControllerStatus(stateValue)
	}
	return schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNSPECIFIED
}

func managementControllerActionStringToEnum(actionName string) schemav1alpha1.ManagementControllerAction {
	if actionValue, ok := schemav1alpha1.ManagementControllerAction_value[actionName]; ok {
		return schemav1alpha1.ManagementControllerAction(actionValue)
	}
	return schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_UNSPECIFIED
}
