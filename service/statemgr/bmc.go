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
)

func (s *StateMgr) createBMCStateMachine(bmcID int) (*state.FSM, error) {
	bmcName := fmt.Sprintf("bmc.%d", bmcID)

	config := state.NewConfig(
		state.WithName(bmcName),
		state.WithDescription(fmt.Sprintf("BMC %d state machine", bmcID)),
		state.WithInitialState(schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY.String()),
		state.WithStates(
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY.String(),
				Description: "BMC is not ready",
				OnEntry:     s.createBMCStateEntryAction(bmcID, schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY),
				OnExit:      s.createBMCStateExitAction(bmcID, schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY.String(),
				Description: "BMC is ready and operational",
				OnEntry:     s.createBMCStateEntryAction(bmcID, schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY),
				OnExit:      s.createBMCStateExitAction(bmcID, schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String(),
				Description: "BMC is enabled",
				OnEntry:     s.createBMCStatusEntryAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED),
				OnExit:      s.createBMCStatusExitAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
				Description: "BMC is disabled",
				OnEntry:     s.createBMCStatusEntryAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED),
				OnExit:      s.createBMCStatusExitAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				Description: "BMC is in error state",
				OnEntry:     s.createBMCStatusEntryAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR),
				OnExit:      s.createBMCStatusExitAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_IN_TEST.String(),
				Description: "BMC is in test mode",
				OnEntry:     s.createBMCStatusEntryAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_IN_TEST),
				OnExit:      s.createBMCStatusExitAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_IN_TEST),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				Description: "BMC is unavailable or offline",
				OnEntry:     s.createBMCStatusEntryAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE),
				OnExit:      s.createBMCStatusExitAction(bmcID, schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE),
			},
		),
		state.WithTransitions(
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String(),
				Trigger: "BMC_TRANSITION_ENABLE",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_IN_TEST.String(),
				Trigger: "BMC_TRANSITION_TEST",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				Trigger: schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_REBOOT.String(),
				Action:  s.createBMCTransitionAction(bmcID, schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_REBOOT),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				Trigger: schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_WARM_RESET.String(),
				Action:  s.createBMCTransitionAction(bmcID, schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_WARM_RESET),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				Trigger: schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_COLD_RESET.String(),
				Action:  s.createBMCTransitionAction(bmcID, schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_COLD_RESET),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
				Trigger: "BMC_TRANSITION_DISABLE",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String(),
				To:      schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY.String(),
				Trigger: "BMC_TRANSITION_READY",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
				Trigger: "BMC_TRANSITION_DISABLE",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				Trigger: schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_REBOOT.String(),
				Action:  s.createBMCTransitionAction(bmcID, schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_REBOOT),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String(),
				Trigger: "BMC_TRANSITION_ENABLE",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				Trigger: schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_REBOOT.String(),
				Action:  s.createBMCTransitionAction(bmcID, schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_REBOOT),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				Trigger: schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_COLD_RESET.String(),
				Action:  s.createBMCTransitionAction(bmcID, schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_COLD_RESET),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_IN_TEST.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String(),
				Trigger: "BMC_TRANSITION_ENABLE",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_IN_TEST.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String(),
				Trigger: "BMC_TRANSITION_ERROR",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				To:      schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY.String(),
				Trigger: "BMC_TRANSITION_COMPLETE_NOT_READY",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String(),
				To:      schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String(),
				Trigger: "BMC_TRANSITION_COMPLETE_ENABLED",
			},
		),
		state.WithPersistState(s.config.PersistStateChanges),
		state.WithStateTimeout(s.config.StateTimeout),
		state.WithMetrics(s.config.EnableMetrics),
		state.WithTracing(s.config.EnableTracing),
	)

	sm, err := state.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create BMC %d state machine: %w", bmcID, err)
	}

	if err := sm.SetPersistenceCallback(s.createBMCPersistenceCallback(bmcName)); err != nil {
		return nil, fmt.Errorf("failed to set persistence callback for BMC %d: %w", bmcID, err)
	}
	if err := sm.SetBroadcastCallback(s.createBMCBroadcastCallback(bmcName)); err != nil {
		return nil, fmt.Errorf("failed to set broadcast callback for BMC %d: %w", bmcID, err)
	}

	return sm, nil
}

func (s *StateMgr) createBMCStateEntryAction(bmcID int, stateValue schemav1alpha1.ManagementControllerState) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.Info("BMC entering state",
				"bmc_id", bmcID,
				"state", stateValue.String(),
				"timestamp", time.Now())
		}
		return nil
	}
}

func (s *StateMgr) createBMCStateExitAction(bmcID int, stateValue schemav1alpha1.ManagementControllerState) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.Info("BMC exiting state",
				"bmc_id", bmcID,
				"state", stateValue.String(),
				"timestamp", time.Now())
		}
		return nil
	}
}

func (s *StateMgr) createBMCStatusEntryAction(bmcID int, status schemav1alpha1.ManagementControllerStatus) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.Info("BMC entering status",
				"bmc_id", bmcID,
				"status", status.String(),
				"timestamp", time.Now())
		}
		return nil
	}
}

func (s *StateMgr) createBMCStatusExitAction(bmcID int, status schemav1alpha1.ManagementControllerStatus) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.Info("BMC exiting status",
				"bmc_id", bmcID,
				"status", status.String(),
				"timestamp", time.Now())
		}
		return nil
	}
}

func (s *StateMgr) createBMCTransitionAction(bmcID int, transition schemav1alpha1.ManagementControllerTransition) state.TransitionAction {
	return func(ctx context.Context, from, to string) error {
		if s.logger != nil {
			s.logger.Info("BMC state transition",
				"bmc_id", bmcID,
				"from", from,
				"to", to,
				"transition", transition.String(),
				"timestamp", time.Now())
		}
		return nil
	}
}

func (s *StateMgr) createBMCPersistenceCallback(componentName string) state.PersistenceCallback {
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

func (s *StateMgr) createBMCBroadcastCallback(componentName string) state.BroadcastCallback {
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

func (s *StateMgr) handleBMCStateRequest(req micro.Request) {
	ctx := telemetry.GetCtxFromReq(req)

	if s.tracer != nil {
		var span trace.Span
		_, span = s.tracer.Start(ctx, "statemgr.handleBMCStateRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 4 {
		s.respondWithError(req, ErrInvalidRequest, "invalid subject format")
		return
	}

	bmcIDStr := parts[2]
	operation := parts[3]

	bmcID, err := strconv.Atoi(bmcIDStr)
	if err != nil {
		s.respondWithError(req, ErrInvalidComponentID, fmt.Sprintf("invalid BMC ID: %s", bmcIDStr))
		return
	}

	bmcName := fmt.Sprintf("bmc.%d", bmcID)

	switch operation {
	case operationState:
		s.handleGetBMCState(req, bmcName)
	case operationControl:
		s.handleBMCControl(req, bmcName)
	case operationInfo:
		s.handleGetBMCInfo(req, bmcName)
	default:
		s.respondWithError(req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

func (s *StateMgr) handleGetBMCState(req micro.Request, bmcName string) {
	sm, exists := s.stateMachines[bmcName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("BMC %s not found", bmcName))
		return
	}

	currentState := sm.CurrentState()
	stateEnum := s.bmcStateStringToEnum(currentState)
	statusEnum := s.bmcStatusStringToEnum(currentState)

	response := &schemav1alpha1.GetManagementControllerResponse{
		Controllers: []*schemav1alpha1.ManagementController{
			{
				Name:         bmcName,
				CurrentState: &stateEnum,
				Status:       &statusEnum,
			},
		},
	}

	s.respondWithProtobuf(req, response)
}

func (s *StateMgr) handleBMCControl(req micro.Request, bmcName string) {
	var request schemav1alpha1.ManagementControllerControlRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		s.respondWithError(req, ErrUnmarshalingFailed, err.Error())
		return
	}

	sm, exists := s.stateMachines[bmcName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("BMC %s not found", bmcName))
		return
	}

	trigger := s.bmcActionToTrigger(request.Action)
	if trigger == "" {
		s.respondWithError(req, ErrInvalidRequest, fmt.Sprintf("unsupported action: %v", request.Action))
		return
	}

	data := map[string]interface{}{
		"force":      request.GetForce(),
		"parameters": request.Parameters,
	}

	err := sm.Fire(context.Background(), trigger, data)
	if err != nil {
		s.respondWithError(req, ErrBMCResetFailed, err.Error())
		return
	}

	currentState := sm.CurrentState()
	stateEnum := s.bmcStateStringToEnum(currentState)

	response := &schemav1alpha1.ManagementControllerControlResponse{
		Success:      true,
		CurrentState: &stateEnum,
	}

	s.respondWithProtobuf(req, response)
}

func (s *StateMgr) handleGetBMCInfo(req micro.Request, bmcName string) {
	sm, exists := s.stateMachines[bmcName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("BMC %s not found", bmcName))
		return
	}

	currentState := sm.CurrentState()
	stateEnum := s.bmcStateStringToEnum(currentState)
	statusEnum := s.bmcStatusStringToEnum(currentState)
	triggers := sm.PermittedTriggers()

	response := &schemav1alpha1.ManagementController{
		Name:         bmcName,
		CurrentState: &stateEnum,
		Status:       &statusEnum,
		Metadata: map[string]string{
			"permitted_triggers": strings.Join(triggers, ","),
			"state_machine":      "active",
		},
	}

	s.respondWithProtobuf(req, response)
}

func (s *StateMgr) bmcStateStringToEnum(stateStr string) schemav1alpha1.ManagementControllerState {
	switch stateStr {
	case schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY.String():
		return schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_READY
	case schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY.String():
		return schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_NOT_READY
	default:
		return schemav1alpha1.ManagementControllerState_MANAGEMENT_CONTROLLER_STATE_UNSPECIFIED
	}
}

func (s *StateMgr) bmcStatusStringToEnum(status string) schemav1alpha1.ManagementControllerStatus {
	switch status {
	case schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED.String():
		return schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ENABLED
	case schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED.String():
		return schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_DISABLED
	case schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR.String():
		return schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_ERROR
	case schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_IN_TEST.String():
		return schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_IN_TEST
	case schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE.String():
		return schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNAVAILABLE_OFFLINE
	default:
		return schemav1alpha1.ManagementControllerStatus_MANAGEMENT_CONTROLLER_STATUS_UNSPECIFIED
	}
}

func (s *StateMgr) bmcActionToTrigger(action schemav1alpha1.ManagementControllerAction) string {
	switch action {
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_REBOOT:
		return schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_REBOOT.String()
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_HARD_RESET:
		return schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_COLD_RESET.String()
	case schemav1alpha1.ManagementControllerAction_MANAGEMENT_CONTROLLER_ACTION_FACTORY_RESET:
		return schemav1alpha1.ManagementControllerTransition_MANAGEMENT_CONTROLLER_TRANSITION_COLD_RESET.String()
	default:
		return ""
	}
}
