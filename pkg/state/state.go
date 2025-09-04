// SPDX-License-Identifier: BSD-3-Clause

package state

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/qmuntal/stateless"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FSM provides a thread-safe finite state machine implementation
// with support for guards, actions, and persistence.
type FSM struct {
	config  *Config
	machine *stateless.StateMachine
	mu      sync.RWMutex
	tracer  trace.Tracer
	started bool
	stopped bool

	currentState      string
	stateActions      map[string]StateDefinition
	transitionMap     map[string]map[string]TransitionDefinition
	persistCallback   PersistenceCallback
	broadcastCallback BroadcastCallback
}

// PersistenceCallback is called when state needs to be persisted.
type PersistenceCallback func(machineName, state string) error

// BroadcastCallback is called when state changes need to be broadcast.
type BroadcastCallback func(machineName, previousState, currentState string, trigger string) error

// New creates a new state machine with the provided configuration.
func New(config *Config) (*FSM, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	sm := &FSM{
		config:        config,
		currentState:  config.InitialState,
		stateActions:  make(map[string]StateDefinition),
		transitionMap: make(map[string]map[string]TransitionDefinition),
	}

	if config.EnableTracing {
		sm.tracer = otel.Tracer("state")
	}

	sm.machine = stateless.NewStateMachine(config.InitialState)

	for _, state := range config.States {
		sm.stateActions[state.Name] = state
		sm.configureState(state)
	}

	for _, transition := range config.Transitions {
		sm.configureTransition(transition)
	}

	return sm, nil
}

// SetPersistenceCallback sets the callback for state persistence.
func (sm *FSM) SetPersistenceCallback(callback PersistenceCallback) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.started {
		return ErrStateMachineAlreadyStarted
	}

	sm.persistCallback = callback
	return nil
}

// SetBroadcastCallback sets the callback for state change broadcasts.
func (sm *FSM) SetBroadcastCallback(callback BroadcastCallback) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.started {
		return ErrStateMachineAlreadyStarted
	}

	sm.broadcastCallback = callback
	return nil
}

// Start initializes and starts the state machine.
func (sm *FSM) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.started {
		return nil
	}

	if sm.stopped {
		return ErrStateMachineStopped
	}

	sm.started = true

	if sm.config.PersistState && sm.persistCallback != nil {
		if err := sm.persistCallback(sm.config.Name, sm.currentState); err != nil {
			return fmt.Errorf("%w: %w", ErrPersistenceFailed, err)
		}
	}

	return nil
}

// Stop gracefully stops the state machine.
func (sm *FSM) Stop(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.started || sm.stopped {
		return nil
	}

	sm.stopped = true
	return nil
}

// Fire triggers a state transition with the specified trigger.
func (sm *FSM) Fire(ctx context.Context, trigger string, data map[string]any) error {
	sm.mu.Lock()

	if !sm.started {
		sm.mu.Unlock()
		return ErrStateMachineNotStarted
	}

	if sm.stopped {
		sm.mu.Unlock()
		return ErrStateMachineStopped
	}

	var span trace.Span
	if sm.tracer != nil {
		ctx, span = sm.tracer.Start(ctx, "state.Fire",
			trace.WithAttributes(
				attribute.String("state_machine.name", sm.config.Name),
				attribute.String("state.current", sm.currentState),
				attribute.String("trigger", trigger),
			))
		defer span.End()
	}

	if ok, err := sm.machine.CanFire(trigger); err != nil {
		sm.mu.Unlock()
		return fmt.Errorf("%w: trigger %s not valid in state %s: %w", ErrInvalidTrigger, trigger, sm.currentState, err)
	} else if !ok {
		sm.mu.Unlock()
		return fmt.Errorf("%w: trigger %s not valid in state %s", ErrInvalidTrigger, trigger, sm.currentState)
	}

	previousState := sm.currentState

	timeout := sm.config.StateTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	fireCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		if err := sm.machine.FireCtx(fireCtx, trigger); err != nil {
			done <- fmt.Errorf("%w: %w", ErrInvalidTransition, err)
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			if span != nil {
				span.RecordError(err)
			}
			sm.mu.Unlock()
			return err
		}
	case <-fireCtx.Done():
		if fireCtx.Err() == context.DeadlineExceeded {
			sm.mu.Unlock()
			return ErrTransitionTimeout
		}
		sm.mu.Unlock()
		return fireCtx.Err()
	}

	state, err := sm.machine.State(ctx)
	if err != nil {
		if span != nil {
			span.RecordError(err)
		}
		sm.mu.Unlock()
		return fmt.Errorf("failed to get current state: %w", err)
	}
	sm.currentState = fmt.Sprintf("%v", state)

	// Capture values and callbacks, then unlock before invoking external code.
	name := sm.config.Name
	curr := sm.currentState
	persistEnabled := sm.config.PersistState
	persistCb := sm.persistCallback
	broadcastCb := sm.broadcastCallback
	sm.mu.Unlock()

	if persistEnabled && persistCb != nil {
		if perr := persistCb(name, curr); perr != nil {
			if span != nil {
				span.RecordError(perr)
			}
			return fmt.Errorf("%w: %w", ErrPersistenceFailed, perr)
		}
	}
	if broadcastCb != nil {
		if berr := broadcastCb(name, previousState, curr, trigger); berr != nil && span != nil {
			span.RecordError(berr)
		}
	}

	if span != nil {
		span.SetAttributes(
			attribute.String("state.previous", previousState),
			attribute.String("state.new", curr),
		)
	}

	return nil
}

// CurrentState returns the current state of the state machine.
func (sm *FSM) CurrentState() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.currentState
}

// CanFire checks if the specified trigger can be fired from the current state.
func (sm *FSM) CanFire(trigger string) (bool, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.machine.CanFire(trigger)
}

// PermittedTriggers returns all triggers that can be fired from the current state.
func (sm *FSM) PermittedTriggers() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	triggers, err := sm.machine.PermittedTriggers()
	if err != nil {
		return []string{}
	}

	result := make([]string, len(triggers))
	for i, t := range triggers {
		result[i] = fmt.Sprintf("%v", t)
	}
	return result
}

// IsInState checks if the state machine is in the specified state.
func (sm *FSM) IsInState(state string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.currentState == state
}

// Name returns the name of the state machine.
func (sm *FSM) Name() string {
	return sm.config.Name
}

// Description returns the description of the state machine.
func (sm *FSM) Description() string {
	return sm.config.Description
}

// GetStateInfo returns information about a specific state.
func (sm *FSM) GetStateInfo(state string) (StateDefinition, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	def, exists := sm.stateActions[state]
	if !exists {
		return StateDefinition{}, fmt.Errorf("%w: %s", ErrInvalidState, state)
	}

	return def, nil
}

// ToGraph returns a DOT graph representation of the state machine.
func (sm *FSM) ToGraph() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.machine.ToGraph()
}

func (sm *FSM) configureState(state StateDefinition) {
	stateConfig := sm.machine.Configure(state.Name)

	if state.OnEntry != nil {
		stateConfig.OnEntry(func(ctx context.Context, args ...any) error {
			return state.OnEntry(ctx)
		})
	}

	if state.OnExit != nil {
		stateConfig.OnExit(func(ctx context.Context, args ...any) error {
			return state.OnExit(ctx)
		})
	}
}

func (sm *FSM) configureTransition(transition TransitionDefinition) {
	if sm.transitionMap[transition.From] == nil {
		sm.transitionMap[transition.From] = make(map[string]TransitionDefinition)
	}
	sm.transitionMap[transition.From][transition.Trigger] = transition

	fromCfg := sm.machine.Configure(transition.From)

	if transition.Guard != nil {
		fromCfg.PermitDynamic(transition.Trigger, func(ctx context.Context, args ...any) (any, error) {
			if transition.Guard(ctx) {
				return transition.To, nil
			}
			return nil, fmt.Errorf("guard condition failed")
		})
	} else {
		fromCfg.Permit(transition.Trigger, transition.To)
	}

	if transition.Action != nil {
		toCfg := sm.machine.Configure(transition.To)
		toCfg.OnEntryFrom(transition.Trigger, func(ctx context.Context, args ...any) error {
			return transition.Action(ctx, transition.From, transition.To)
		})
	}
}

// Manager manages multiple state machines.
type Manager struct {
	machines map[string]*FSM
	mu       sync.RWMutex
}

// NewManager creates a new state machine manager.
func NewManager() *Manager {
	return &Manager{
		machines: make(map[string]*FSM),
	}
}

// AddStateMachine adds a state machine to the manager.
func (m *Manager) AddStateMachine(sm *FSM) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sm == nil {
		return fmt.Errorf("%w: nil state machine", ErrInvalidConfig)
	}

	if _, exists := m.machines[sm.Name()]; exists {
		return fmt.Errorf("%w: %s", ErrStateMachineExists, sm.Name())
	}

	m.machines[sm.Name()] = sm
	return nil
}

// RemoveStateMachine removes a state machine from the manager.
func (m *Manager) RemoveStateMachine(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.machines[name]; !exists {
		return fmt.Errorf("%w: %s", ErrStateMachineNotFound, name)
	}

	delete(m.machines, name)
	return nil
}

// GetStateMachine retrieves a state machine by name.
func (m *Manager) GetStateMachine(name string) (*FSM, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sm, exists := m.machines[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrStateMachineNotFound, name)
	}

	return sm, nil
}

// ListStateMachines returns the names of all managed state machines.
func (m *Manager) ListStateMachines() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.machines))
	for name := range m.machines {
		names = append(names, name)
	}

	return names
}

// StopAll stops all managed state machines.
func (m *Manager) StopAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, sm := range m.machines {
		if err := sm.Stop(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
