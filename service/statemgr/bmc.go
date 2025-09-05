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

func (s *StateMgr) createManagementControllerStateMachine(ctx context.Context, controllerName string) (*state.FSM, error) {
	config := state.NewConfig(
		state.WithName(controllerName),
		state.WithDescription(fmt.Sprintf("Management Controller %s state machine", controllerName)),
		state.WithInitialState(schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String()),
		state.WithStates(
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				Description: "Management Controller is not ready",
				OnEntry:     s.createManagementControllerStatusEntryAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY),
				OnExit:      s.createManagementControllerStatusExitAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				Description: "Management Controller is ready and operational",
				OnEntry:     s.createManagementControllerStatusEntryAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY),
				OnExit:      s.createManagementControllerStatusExitAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
				Description: "Management Controller is disabled",
				OnEntry:     s.createManagementControllerStatusEntryAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED),
				OnExit:      s.createManagementControllerStatusExitAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				Description: "Management Controller is in error state",
				OnEntry:     s.createManagementControllerStatusEntryAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR),
				OnExit:      s.createManagementControllerStatusExitAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED.String(),
				Description: "Management Controller is quiesced",
				OnEntry:     s.createManagementControllerStatusEntryAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED),
				OnExit:      s.createManagementControllerStatusExitAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC.String(),
				Description: "Management Controller is in diagnostic mode",
				OnEntry:     s.createManagementControllerStatusEntryAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC),
				OnExit:      s.createManagementControllerStatusExitAction(controllerName, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC),
			},
		),
		state.WithTransitions(
			// API-triggered transitions
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_WARM_RESET.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_WARM_RESET),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_HARD_RESET.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_HARD_RESET),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_FACTORY_RESET.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_FACTORY_RESET),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_DISABLE.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_DISABLE),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_COLD_RESET),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_DISABLE.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_DISABLE),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				Trigger: schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE.String(),
				Action:  s.createManagementControllerTransitionAction(controllerName, schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_ENABLE),
			},
			// Internal transitions (not exposed via API)
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				Trigger: managementControllerTriggerTransitionCompleteReady,
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				Trigger: managementControllerTriggerTransitionCompleteError,
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DIAGNOSTIC.String(),
				Trigger: managementControllerTriggerTransitionCompleteDiagnostic,
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_NOT_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED.String(),
				Trigger: managementControllerTriggerTransitionTimeout,
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_QUIESCED.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_READY.String(),
				Trigger: managementControllerTriggerTransitionResume,
			},
		),
		state.WithPersistState(s.config.PersistStateChanges),
		state.WithStateTimeout(s.config.StateTimeout),
		state.WithMetrics(s.config.EnableMetrics),
		state.WithTracing(s.config.EnableTracing),
	)

	sm, err := state.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create management controller %s state machine: %w", controllerName, err)
	}

	if err := sm.SetPersistenceCallback(s.createManagementControllerPersistenceCallback(ctx, controllerName)); err != nil {
		return nil, fmt.Errorf("failed to set persistence callback for management controller %s: %w", controllerName, err)
	}
	if err := sm.SetBroadcastCallback(s.createManagementControllerBroadcastCallback(ctx, controllerName)); err != nil {
		return nil, fmt.Errorf("failed to set broadcast callback for management controller %s: %w", controllerName, err)
	}

	return sm, nil
}

func (s *StateMgr) createManagementControllerStatusEntryAction(controllerName string, status schemav1alpha1.ManagementControllerStatus) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Management Controller entering status",
				"controller_name", controllerName,
				"status", status.String())
		}
		return nil
	}
}

func (s *StateMgr) createManagementControllerStatusExitAction(controllerName string, status schemav1alpha1.ManagementControllerStatus) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "Management Controller exiting status",
				"controller_name", controllerName,
				"status", status.String())
		}
		return nil
	}
}

func (s *StateMgr) createManagementControllerTransitionAction(controllerName string, action schemav1alpha1.ManagementControllerAction) state.TransitionAction {
	return func(ctx context.Context, from, to string) error {
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
	return func(machineName, state string) error {
		if !s.config.PersistStateChanges || s.js == nil {
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

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if _, err = s.js.Publish(ctx, subject, dataBytes); err != nil {
			return fmt.Errorf("%w: %w", ErrStatePersistenceFailed, err)
		}

		return nil
	}
}

func (s *StateMgr) createManagementControllerBroadcastCallback(ctx context.Context, componentName string) state.BroadcastCallback {
	return func(machineName, previousState, currentState string, trigger string) error {
		if !s.config.BroadcastStateChanges || s.nc == nil {
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
		_, span = s.tracer.Start(ctx, "statemgr.handleManagementControllerStateRequest")
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

	currentState := sm.CurrentState()
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

	currentState := sm.CurrentState()
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

	currentState := sm.CurrentState()
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
