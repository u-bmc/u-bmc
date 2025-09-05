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

func (s *StateMgr) createChassisStateMachine(chassisID int) (*state.FSM, error) {
	chassisName := fmt.Sprintf("chassis.%d", chassisID)

	config := state.NewConfig(
		state.WithName(chassisName),
		state.WithDescription(fmt.Sprintf("Chassis %d state machine", chassisID)),
		state.WithInitialState(schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String()),
		state.WithStates(
			state.StateDefinition{
				Name:        schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
				Description: "Chassis is powered off",
				OnEntry:     s.createChassisStateEntryAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF),
				OnExit:      s.createChassisStateExitAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
				Description: "Chassis is powered on",
				OnEntry:     s.createChassisStateEntryAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON),
				OnExit:      s.createChassisStateExitAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
				Description: "Chassis is transitioning between states",
				OnEntry:     s.createChassisStateEntryAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING),
				OnExit:      s.createChassisStateExitAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
				Description: "Chassis has a warning condition",
				OnEntry:     s.createChassisStateEntryAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING),
				OnExit:      s.createChassisStateExitAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
				Description: "Chassis has a critical condition",
				OnEntry:     s.createChassisStateEntryAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL),
				OnExit:      s.createChassisStateExitAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL),
			},
			state.StateDefinition{
				Name:        schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
				Description: "Chassis has failed",
				OnEntry:     s.createChassisStateEntryAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED),
				OnExit:      s.createChassisStateExitAction(chassisID, schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED),
			},
		),
		state.WithTransitions(
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
				Trigger: schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_ON.String(),
				Action:  s.createChassisTransitionAction(chassisID, schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_ON),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
				Trigger: schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF.String(),
				Action:  s.createChassisTransitionAction(chassisID, schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
				Trigger: schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_POWER_CYCLE.String(),
				Action:  s.createChassisTransitionAction(chassisID, schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_POWER_CYCLE),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
				Trigger: "CHASSIS_TRANSITION_WARNING",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
				Trigger: "CHASSIS_TRANSITION_CRITICAL",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
				Trigger: "CHASSIS_TRANSITION_COMPLETE_ON",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
				Trigger: "CHASSIS_TRANSITION_COMPLETE_OFF",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
				Trigger: "CHASSIS_TRANSITION_FAILED",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String(),
				Trigger: "CHASSIS_TRANSITION_CLEAR",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
				Trigger: "CHASSIS_TRANSITION_CRITICAL",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
				Trigger: schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF.String(),
				Action:  s.createChassisTransitionAction(chassisID, schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String(),
				Trigger: "CHASSIS_TRANSITION_WARNING",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
				Trigger: schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF.String(),
				Action:  s.createChassisTransitionAction(chassisID, schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF),
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
				Trigger: "CHASSIS_TRANSITION_FAILED",
			},
			state.TransitionDefinition{
				From:    schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String(),
				To:      schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String(),
				Trigger: schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF.String(),
				Action:  s.createChassisTransitionAction(chassisID, schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF),
			},
		),
		state.WithPersistState(s.config.PersistStateChanges),
		state.WithStateTimeout(s.config.StateTimeout),
		state.WithMetrics(s.config.EnableMetrics),
		state.WithTracing(s.config.EnableTracing),
	)

	sm, err := state.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create chassis %d state machine: %w", chassisID, err)
	}

	if err := sm.SetPersistenceCallback(s.createChassisPersistenceCallback(chassisName)); err != nil {
		return nil, fmt.Errorf("failed to set persistence callback for chassis %d: %w", chassisID, err)
	}
	if err := sm.SetBroadcastCallback(s.createChassisBroadcastCallback(chassisName)); err != nil {
		return nil, fmt.Errorf("failed to set broadcast callback for chassis %d: %w", chassisID, err)
	}

	return sm, nil
}

func (s *StateMgr) createChassisStateEntryAction(chassisID int, status schemav1alpha1.ChassisStatus) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.Info("Chassis entering state",
				"chassis_id", chassisID,
				"status", status.String())
		}
		return nil
	}
}

func (s *StateMgr) createChassisStateExitAction(chassisID int, status schemav1alpha1.ChassisStatus) state.StateAction {
	return func(ctx context.Context) error {
		if s.logger != nil {
			s.logger.Info("Chassis exiting state",
				"chassis_id", chassisID,
				"status", status.String())
		}
		return nil
	}
}

func (s *StateMgr) createChassisTransitionAction(chassisID int, transition schemav1alpha1.ChassisTransition) state.TransitionAction {
	return func(ctx context.Context, from, to string) error {
		if s.logger != nil {
			s.logger.Info("Chassis state transition",
				"chassis_id", chassisID,
				"from", from,
				"to", to,
				"transition", transition.String())
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

func (s *StateMgr) createChassisBroadcastCallback(componentName string) state.BroadcastCallback {
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

func (s *StateMgr) handleChassisStateRequest(req micro.Request) {
	ctx := telemetry.GetCtxFromReq(req)
	if s.tracer != nil {
		var span trace.Span
		_, span = s.tracer.Start(ctx, "statemgr.handleChassisStateRequest")
		defer span.End()
		span.SetAttributes(attribute.String("subject", req.Subject()))
	}

	parts := strings.Split(req.Subject(), ".")
	if len(parts) < 4 {
		s.respondWithError(req, ErrInvalidRequest, "invalid subject format")
		return
	}

	chassisIDStr := parts[2]
	operation := parts[3]

	chassisID, err := strconv.Atoi(chassisIDStr)
	if err != nil {
		s.respondWithError(req, ErrInvalidComponentID, fmt.Sprintf("invalid chassis ID: %s", chassisIDStr))
		return
	}

	chassisName := fmt.Sprintf("chassis.%d", chassisID)

	switch operation {
	case operationState:
		s.handleGetChassisState(req, chassisName)
	case operationTransition:
		s.handleChassisTransition(req, chassisName)
	case operationControl:
		s.handleChassisControl(req, chassisName)
	case operationInfo:
		s.handleGetChassisInfo(req, chassisName)
	default:
		s.respondWithError(req, ErrInvalidRequest, fmt.Sprintf("unknown operation: %s", operation))
	}
}

func (s *StateMgr) handleGetChassisState(req micro.Request, chassisName string) {
	sm, exists := s.stateMachines[chassisName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("chassis %s not found", chassisName))
		return
	}

	currentState := sm.CurrentState()
	statusEnum := s.chassisStateStringToEnum(currentState)

	response := &schemav1alpha1.GetChassisResponse{
		Chassis: []*schemav1alpha1.Chassis{
			{
				Name:   chassisName,
				Status: &statusEnum,
			},
		},
	}

	s.respondWithProtobuf(req, response)
}

func (s *StateMgr) handleChassisTransition(req micro.Request, chassisName string) {
	ctx := telemetry.GetCtxFromReq(req)

	var request schemav1alpha1.ChassisChangeStateRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		s.respondWithError(req, ErrUnmarshalingFailed, err.Error())
		return
	}

	sm, exists := s.stateMachines[chassisName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("chassis %s not found", chassisName))
		return
	}

	trigger := s.chassisTransitionToTrigger(request.Transition)
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
	statusEnum := s.chassisStateStringToEnum(currentState)

	response := &schemav1alpha1.ChassisChangeStateResponse{
		Success:      true,
		Status:       &statusEnum,
		TransitionId: &[]string{fmt.Sprintf("%s-%d", chassisName, time.Now().UnixNano())}[0],
		Metadata:     request.Metadata,
	}

	s.respondWithProtobuf(req, response)
}

func (s *StateMgr) handleChassisControl(req micro.Request, chassisName string) {
	ctx := telemetry.GetCtxFromReq(req)

	var request schemav1alpha1.ChassisControlRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		s.respondWithError(req, ErrUnmarshalingFailed, err.Error())
		return
	}

	sm, exists := s.stateMachines[chassisName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("chassis %s not found", chassisName))
		return
	}

	trigger := s.chassisControlActionToTrigger(request.Action)
	if trigger == "" {
		s.respondWithError(req, ErrInvalidRequest, fmt.Sprintf("unsupported action: %v", request.Action))
		return
	}

	data := map[string]interface{}{
		"force": request.GetForce(),
	}

	for k, v := range request.Parameters {
		data[k] = v
	}

	err := sm.Fire(ctx, trigger, data)
	if err != nil {
		s.respondWithError(req, ErrChassisControlFailed, err.Error())
		return
	}

	currentState := sm.CurrentState()
	statusEnum := s.chassisStateStringToEnum(currentState)

	response := &schemav1alpha1.ChassisControlResponse{
		Success: true,
		Status:  &statusEnum,
	}

	s.respondWithProtobuf(req, response)
}

func (s *StateMgr) handleGetChassisInfo(req micro.Request, chassisName string) {
	sm, exists := s.stateMachines[chassisName]
	if !exists {
		s.respondWithError(req, ErrComponentNotFound, fmt.Sprintf("chassis %s not found", chassisName))
		return
	}

	currentState := sm.CurrentState()
	statusEnum := s.chassisStateStringToEnum(currentState)
	triggers := sm.PermittedTriggers()

	response := &schemav1alpha1.Chassis{
		Name:   chassisName,
		Status: &statusEnum,
		Metadata: map[string]string{
			"permitted_triggers": strings.Join(triggers, ","),
			"state_machine":      "active",
		},
	}

	s.respondWithProtobuf(req, response)
}

func (s *StateMgr) chassisStateStringToEnum(stateName string) schemav1alpha1.ChassisStatus {
	switch stateName {
	case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF.String():
		return schemav1alpha1.ChassisStatus_CHASSIS_STATUS_OFF
	case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON.String():
		return schemav1alpha1.ChassisStatus_CHASSIS_STATUS_ON
	case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING.String():
		return schemav1alpha1.ChassisStatus_CHASSIS_STATUS_TRANSITIONING
	case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING.String():
		return schemav1alpha1.ChassisStatus_CHASSIS_STATUS_WARNING
	case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL.String():
		return schemav1alpha1.ChassisStatus_CHASSIS_STATUS_CRITICAL
	case schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED.String():
		return schemav1alpha1.ChassisStatus_CHASSIS_STATUS_FAILED
	default:
		return schemav1alpha1.ChassisStatus_CHASSIS_STATUS_UNSPECIFIED
	}
}

func (s *StateMgr) chassisTransitionToTrigger(transition schemav1alpha1.ChassisTransition) string {
	switch transition {
	case schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_ON:
		return schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_ON.String()
	case schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF:
		return schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF.String()
	case schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_POWER_CYCLE:
		return schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_POWER_CYCLE.String()
	default:
		return ""
	}
}

func (s *StateMgr) chassisControlActionToTrigger(action schemav1alpha1.ChassisControlAction) string {
	switch action {
	case schemav1alpha1.ChassisControlAction_CHASSIS_CONTROL_ACTION_POWER_ON:
		return schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_ON.String()
	case schemav1alpha1.ChassisControlAction_CHASSIS_CONTROL_ACTION_POWER_OFF:
		return schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_OFF.String()
	case schemav1alpha1.ChassisControlAction_CHASSIS_CONTROL_ACTION_POWER_CYCLE:
		return schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_POWER_CYCLE.String()
	case schemav1alpha1.ChassisControlAction_CHASSIS_CONTROL_ACTION_RESET:
		return schemav1alpha1.ChassisTransition_CHASSIS_TRANSITION_POWER_CYCLE.String()
	default:
		return ""
	}
}
