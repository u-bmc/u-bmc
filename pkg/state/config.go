// SPDX-License-Identifier: BSD-3-Clause

package state

import (
	"context"
	"fmt"
	"time"
)

// Config holds the configuration for a state machine wrapper.
type Config struct {
	// Name is the unique identifier for the state machine
	Name string
	// Description provides human-readable information about the state machine
	Description string
	// InitialState is the starting state of the machine
	InitialState string
	// States defines all possible states (simplified as string slice)
	States []string
	// Transitions defines allowed transitions including from/to states, triggers, and optional guard and action handlers
	Transitions []Transition
	// StateTimeout is the maximum time a state transition can take
	StateTimeout time.Duration
	// PersistenceCallback is called when state changes need to be persisted
	PersistenceCallback PersistenceCallback
	// BroadcastCallback is called when state changes need to be broadcast
	BroadcastCallback BroadcastCallback
	// OnStateEntry is called when entering any state
	OnStateEntry EntryCallback
	// OnStateExit is called when exiting any state
	OnStateExit ExitCallback
}

// Transition represents a simple state transition.
type Transition struct {
	From    string
	To      string
	Trigger string
	Guard   GuardFunc
	Action  ActionFunc
}

// PersistenceCallback is called when state needs to be persisted.
type PersistenceCallback func(ctx context.Context, machineName, state string) error

// BroadcastCallback is called when state changes need to be broadcast.
type BroadcastCallback func(ctx context.Context, machineName, previousState, currentState, trigger string) error

// EntryCallback is called when entering a state.
type EntryCallback func(ctx context.Context, machineName, state string) error

// ExitCallback is called when exiting a state.
type ExitCallback func(ctx context.Context, machineName, state string) error

// GuardFunc determines if a transition is allowed.
type GuardFunc func() bool

// ActionFunc is executed during a transition.
type ActionFunc func(from, to, trigger string) error

// Option represents a configuration option for the state machine.
type Option interface {
	apply(*Config)
}

type nameOption struct {
	name string
}

func (o *nameOption) apply(c *Config) {
	c.Name = o.name
}

// WithName sets the name of the state machine.
func WithName(name string) Option {
	return &nameOption{name: name}
}

type descriptionOption struct {
	description string
}

func (o *descriptionOption) apply(c *Config) {
	c.Description = o.description
}

// WithDescription sets the description of the state machine.
func WithDescription(description string) Option {
	return &descriptionOption{description: description}
}

type initialStateOption struct {
	state string
}

func (o *initialStateOption) apply(c *Config) {
	c.InitialState = o.state
}

// WithInitialState sets the initial state of the state machine.
func WithInitialState(state string) Option {
	return &initialStateOption{state: state}
}

type statesOption struct {
	states []string
}

func (o *statesOption) apply(c *Config) {
	c.States = append([]string(nil), o.states...)
}

// WithStates sets the available states for the state machine.
func WithStates(states ...string) Option {
	return &statesOption{states: states}
}

type transitionOption struct {
	transition Transition
}

func (o *transitionOption) apply(c *Config) {
	c.Transitions = append(c.Transitions, o.transition)
}

// WithTransition adds a transition to the state machine.
func WithTransition(from, to, trigger string) Option {
	return &transitionOption{
		transition: Transition{
			From:    from,
			To:      to,
			Trigger: trigger,
		},
	}
}

// WithGuardedTransition adds a transition with a guard condition.
func WithGuardedTransition(from, to, trigger string, guard GuardFunc) Option {
	return &transitionOption{
		transition: Transition{
			From:    from,
			To:      to,
			Trigger: trigger,
			Guard:   guard,
		},
	}
}

// WithActionTransition adds a transition with an action.
func WithActionTransition(from, to, trigger string, action ActionFunc) Option {
	return &transitionOption{
		transition: Transition{
			From:    from,
			To:      to,
			Trigger: trigger,
			Action:  action,
		},
	}
}

// WithCompleteTransition adds a transition with both guard and action.
func WithCompleteTransition(from, to, trigger string, guard GuardFunc, action ActionFunc) Option {
	return &transitionOption{
		transition: Transition{
			From:    from,
			To:      to,
			Trigger: trigger,
			Guard:   guard,
			Action:  action,
		},
	}
}

type stateTimeoutOption struct {
	timeout time.Duration
}

func (o *stateTimeoutOption) apply(c *Config) {
	c.StateTimeout = o.timeout
}

// WithStateTimeout sets the maximum duration for state transitions.
func WithStateTimeout(timeout time.Duration) Option {
	return &stateTimeoutOption{timeout: timeout}
}

type persistenceOption struct {
	callback PersistenceCallback
}

func (o *persistenceOption) apply(c *Config) {
	c.PersistenceCallback = o.callback
}

// WithPersistence sets the persistence callback.
func WithPersistence(callback PersistenceCallback) Option {
	return &persistenceOption{callback: callback}
}

type broadcastOption struct {
	callback BroadcastCallback
}

func (o *broadcastOption) apply(c *Config) {
	c.BroadcastCallback = o.callback
}

// WithBroadcast sets the broadcast callback.
func WithBroadcast(callback BroadcastCallback) Option {
	return &broadcastOption{callback: callback}
}

type stateEntryOption struct {
	callback EntryCallback
}

func (o *stateEntryOption) apply(c *Config) {
	c.OnStateEntry = o.callback
}

// WithStateEntry sets the state entry callback.
func WithStateEntry(callback EntryCallback) Option {
	return &stateEntryOption{callback: callback}
}

type stateExitOption struct {
	callback ExitCallback
}

func (o *stateExitOption) apply(c *Config) {
	c.OnStateExit = o.callback
}

// WithStateExit sets the state exit callback.
func WithStateExit(callback ExitCallback) Option {
	return &stateExitOption{callback: callback}
}

// NewConfig creates a new state machine configuration with the provided options.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Name:         "",
		Description:  "",
		InitialState: "",
		States:       []string{},
		Transitions:  []Transition{},
		StateTimeout: 30 * time.Second,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidConfig)
	}

	if c.InitialState == "" {
		return fmt.Errorf("%w: initial state cannot be empty", ErrInvalidConfig)
	}

	if len(c.States) == 0 {
		return fmt.Errorf("%w: at least one state must be defined", ErrInvalidConfig)
	}

	// Check if initial state exists in states list
	initialStateFound := false
	stateNames := make(map[string]bool)
	for _, state := range c.States {
		if state == "" {
			return fmt.Errorf("%w: state name cannot be empty", ErrInvalidConfig)
		}
		if stateNames[state] {
			return fmt.Errorf("%w: duplicate state name: %s", ErrInvalidConfig, state)
		}
		stateNames[state] = true
		if state == c.InitialState {
			initialStateFound = true
		}
	}

	if !initialStateFound {
		return fmt.Errorf("%w: initial state %s not found in states list", ErrInvalidConfig, c.InitialState)
	}

	// Validate transitions
	for _, transition := range c.Transitions {
		if transition.From == "" || transition.To == "" {
			return fmt.Errorf("%w: transition from and to states cannot be empty", ErrInvalidConfig)
		}
		if transition.Trigger == "" {
			return fmt.Errorf("%w: transition trigger cannot be empty", ErrInvalidConfig)
		}
		if !stateNames[transition.From] {
			return fmt.Errorf("%w: transition from state %s not found", ErrInvalidConfig, transition.From)
		}
		if !stateNames[transition.To] {
			return fmt.Errorf("%w: transition to state %s not found", ErrInvalidConfig, transition.To)
		}
	}

	if c.StateTimeout <= 0 {
		return fmt.Errorf("%w: state timeout must be positive", ErrInvalidConfig)
	}

	return nil
}
