// SPDX-License-Identifier: BSD-3-Clause

package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/qmuntal/stateless"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Machine represents a finite state machine wrapper around the stateless library.
type Machine struct {
	name        string
	description string
	machine     *stateless.StateMachine
	tracer      trace.Tracer

	// Callbacks
	persistenceCallback PersistenceCallback
	broadcastCallback   BroadcastCallback
	stateEntryCallback  EntryCallback
	stateExitCallback   ExitCallback

	// Configuration
	timeout time.Duration

	// Lifecycle management
	started bool
	stopped bool
	mu      sync.RWMutex
}

// New creates a new state machine with the provided configuration.
func New(config *Config) (*Machine, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	m := &Machine{
		name:                config.Name,
		description:         config.Description,
		timeout:             config.StateTimeout,
		persistenceCallback: config.PersistenceCallback,
		broadcastCallback:   config.BroadcastCallback,
		stateEntryCallback:  config.OnStateEntry,
		stateExitCallback:   config.OnStateExit,
		tracer:              otel.Tracer("state"),
	}

	// Create the underlying stateless machine
	m.machine = stateless.NewStateMachine(config.InitialState)

	// Configure states with entry/exit actions
	for _, state := range config.States {
		stateConfig := m.machine.Configure(state)

		if m.stateEntryCallback != nil {
			stateConfig.OnEntry(m.createEntryAction(state))
		}

		if m.stateExitCallback != nil {
			stateConfig.OnExit(m.createExitAction(state))
		}
	}

	// Configure transitions
	for _, transition := range config.Transitions {
		if err := m.configureTransition(transition); err != nil {
			return nil, fmt.Errorf("failed to configure transition %s->%s: %w",
				transition.From, transition.To, err)
		}
	}

	return m, nil
}

// configureTransition sets up a transition in the underlying state machine.
func (m *Machine) configureTransition(t Transition) error {
	fromConfig := m.machine.Configure(t.From)

	if t.Guard != nil {
		// Use PermitDynamic for guarded transitions
		fromConfig.PermitDynamic(t.Trigger, func(ctx context.Context, args ...any) (any, error) {
			if t.Guard() {
				return t.To, nil
			}
			return nil, fmt.Errorf("guard condition failed for transition %s->%s", t.From, t.To)
		})
	} else {
		// Simple permit for unguarded transitions
		fromConfig.Permit(t.Trigger, t.To)
	}

	// Configure transition action if present
	if t.Action != nil {
		toConfig := m.machine.Configure(t.To)
		toConfig.OnEntryFrom(t.Trigger, func(ctx context.Context, args ...any) error {
			if err := t.Action(t.From, t.To, t.Trigger); err != nil {
				return fmt.Errorf("transition action failed: %w", err)
			}
			return nil
		})
	}

	return nil
}

// createEntryAction creates a state entry action that calls the configured callback.
func (m *Machine) createEntryAction(state string) func(context.Context, ...any) error {
	return func(ctx context.Context, args ...any) error {
		if m.stateEntryCallback != nil {
			if err := m.stateEntryCallback(ctx, m.name, state); err != nil {
				return fmt.Errorf("state entry callback failed: %w", err)
			}
		}
		return nil
	}
}

// createExitAction creates a state exit action that calls the configured callback.
func (m *Machine) createExitAction(state string) func(context.Context, ...any) error {
	return func(ctx context.Context, args ...any) error {
		if m.stateExitCallback != nil {
			if err := m.stateExitCallback(ctx, m.name, state); err != nil {
				return fmt.Errorf("state exit callback failed: %w", err)
			}
		}
		return nil
	}
}

// Start initializes the state machine. This is primarily for lifecycle management
// and callback setup.
func (m *Machine) Start(ctx context.Context) error {
	// Capture required fields under lock
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return ErrStateMachineAlreadyStarted
	}

	if m.stopped {
		m.mu.Unlock()
		return ErrStateMachineStopped
	}

	// Set started to true and capture callback reference
	m.started = true
	persistenceCallback := m.persistenceCallback
	machineName := m.name
	m.mu.Unlock()

	// Persist initial state if persistence is enabled - done outside of lock
	if persistenceCallback != nil {
		currentState := m.State(ctx)
		if err := persistenceCallback(ctx, machineName, currentState); err != nil {
			// Roll back started state on error
			m.mu.Lock()
			m.started = false
			m.mu.Unlock()
			return fmt.Errorf("%w: %w", ErrPersistenceFailed, err)
		}
	}

	return nil
}

// Stop gracefully stops the state machine.
func (m *Machine) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started || m.stopped {
		return nil
	}

	m.stopped = true
	return nil
}

// Fire triggers a state transition. This is the primary method for state changes.
func (m *Machine) Fire(ctx context.Context, trigger string) error {
	m.mu.RLock()
	started := m.started
	stopped := m.stopped
	m.mu.RUnlock()

	if !started {
		return ErrStateMachineNotStarted
	}

	if stopped {
		return ErrStateMachineStopped
	}

	var span trace.Span
	if m.tracer != nil {
		ctx, span = m.tracer.Start(ctx, "state.Fire",
			trace.WithAttributes(
				attribute.String("machine.name", m.name),
				attribute.String("trigger", trigger),
				attribute.String("current_state", m.State(ctx)),
			))
		defer span.End()
	}

	// Check if trigger can be fired
	canFire, err := m.machine.CanFireCtx(ctx, trigger)
	if err != nil {
		if span != nil {
			span.RecordError(err)
		}
		return fmt.Errorf("failed to check if trigger can be fired: %w", err)
	}

	if !canFire {
		err := fmt.Errorf("trigger %s cannot be fired from current state %s", trigger, m.State(ctx))
		if span != nil {
			span.RecordError(err)
		}
		return err
	}

	previousState := m.State(ctx)

	// Fire the transition with timeout
	fireCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- m.machine.FireCtx(fireCtx, trigger)
	}()

	select {
	case err := <-done:
		if err != nil {
			if span != nil {
				span.RecordError(err)
			}
			return fmt.Errorf("state transition failed: %w", err)
		}
	case <-fireCtx.Done():
		if fireCtx.Err() == context.DeadlineExceeded {
			return ErrTransitionTimeout
		}
		return fireCtx.Err()
	}

	currentState := m.State(ctx)

	// Handle persistence and broadcasting asynchronously to avoid blocking
	go m.handlePostTransition(ctx, previousState, currentState, trigger)

	if span != nil {
		span.SetAttributes(
			attribute.String("previous_state", previousState),
			attribute.String("new_state", currentState),
		)
	}

	return nil
}

// handlePostTransition handles persistence and broadcasting after a successful transition.
func (m *Machine) handlePostTransition(ctx context.Context, previousState, currentState, trigger string) {
	var span trace.Span
	if m.tracer != nil {
		tctx, tspan := m.tracer.Start(ctx, "state.postTransition",
			trace.WithAttributes(
				attribute.String("machine.name", m.name),
				attribute.String("previous_state", previousState),
				attribute.String("current_state", currentState),
				attribute.String("trigger", trigger),
			))
		span = tspan
		ctx = tctx
		defer span.End()
	}

	// Handle persistence
	if m.persistenceCallback != nil {
		if err := m.persistenceCallback(ctx, m.name, currentState); err != nil {
			if span != nil {
				span.RecordError(fmt.Errorf("%w: %w", ErrPersistenceFailed, err))
			}
		}
	}

	// Handle broadcasting
	if m.broadcastCallback != nil {
		if err := m.broadcastCallback(ctx, m.name, previousState, currentState, trigger); err != nil {
			if span != nil {
				span.RecordError(fmt.Errorf("%w: %w", ErrBroadcastFailed, err))
			}
		}
	}
}

// State returns the current state of the machine.
// This leverages the underlying library's thread-safe state access.
func (m *Machine) State(ctx context.Context) string {
	state, err := m.machine.State(ctx)
	if err != nil {
		// This should rarely happen with the stateless library
		return "unknown"
	}
	return fmt.Sprintf("%v", state)
}

// CanFire checks if the specified trigger can be fired from the current state.
func (m *Machine) CanFire(trigger string) bool {
	canFire, err := m.machine.CanFire(trigger)
	if err != nil {
		return false
	}
	return canFire
}

// PermittedTriggers returns all triggers that can be fired from the current state.
func (m *Machine) PermittedTriggers() []string {
	triggers, err := m.machine.PermittedTriggers()
	if err != nil {
		return []string{}
	}

	result := make([]string, len(triggers))
	for i, t := range triggers {
		result[i] = fmt.Sprintf("%v", t)
	}
	return result
}

// IsInState checks if the machine is currently in the specified state.
func (m *Machine) IsInState(state string) bool {
	isInState, err := m.machine.IsInStateCtx(context.Background(), state)
	if err != nil {
		return false
	}
	return isInState
}

// Name returns the name of the state machine.
func (m *Machine) Name() string {
	return m.name
}

// Description returns the description of the state machine.
func (m *Machine) Description() string {
	return m.description
}

// ToGraph returns a DOT graph representation of the state machine.
func (m *Machine) ToGraph() string {
	return m.machine.ToGraph()
}

// Manager manages multiple state machines with minimal overhead.
type Manager struct {
	machines map[string]*Machine
	mu       sync.RWMutex
}

// NewManager creates a new state machine manager.
func NewManager() *Manager {
	return &Manager{
		machines: make(map[string]*Machine),
	}
}

// Add adds a state machine to the manager.
func (m *Manager) Add(machine *Machine) error {
	if machine == nil {
		return ErrInvalidConfig
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.machines[machine.Name()]; exists {
		return fmt.Errorf("%w: %s", ErrStateMachineExists, machine.Name())
	}

	m.machines[machine.Name()] = machine
	return nil
}

// Remove removes a state machine from the manager.
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.machines[name]; !exists {
		return fmt.Errorf("%w: %s", ErrStateMachineNotFound, name)
	}

	delete(m.machines, name)
	return nil
}

// Get retrieves a state machine by name.
func (m *Manager) Get(name string) (*Machine, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	machine, exists := m.machines[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrStateMachineNotFound, name)
	}

	return machine, nil
}

// List returns the names of all managed state machines.
func (m *Manager) List() []string {
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
	m.mu.RLock()
	machines := make([]*Machine, 0, len(m.machines))
	for _, machine := range m.machines {
		machines = append(machines, machine)
	}
	m.mu.RUnlock()

	var errs []error
	for _, machine := range machines {
		if err := machine.Stop(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to stop some machines: %v", errs)
	}

	return nil
}
