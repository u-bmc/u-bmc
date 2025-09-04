// SPDX-License-Identifier: BSD-3-Clause

package state

import (
	"context"
	"fmt"
	"time"
)

// Config holds the configuration for a state machine.
type Config struct {
	// Name is the unique identifier for the state machine
	Name string
	// Description provides human-readable information about the state machine
	Description string
	// InitialState is the starting state of the machine
	InitialState string
	// States defines all possible states
	States []StateDefinition
	// Transitions defines all allowed state transitions
	Transitions []TransitionDefinition
	// PersistState enables state persistence
	PersistState bool
	// StateTimeout is the maximum time a state transition can take
	StateTimeout time.Duration
	// EnableMetrics enables state transition metrics collection
	EnableMetrics bool
	// EnableTracing enables distributed tracing for state transitions
	EnableTracing bool
}

// StateDefinition defines a single state in the state machine.
type StateDefinition struct { //nolint:revive // keeping struct name for clarity considering we have TransitionDefinition as well
	// Name is the unique identifier for the state
	Name string
	// Description provides human-readable information about the state
	Description string
	// OnEntry is called when entering this state
	OnEntry StateAction
	// OnExit is called when leaving this state
	OnExit StateAction
}

// TransitionDefinition defines a valid state transition.
type TransitionDefinition struct {
	// From is the source state
	From string
	// To is the destination state
	To string
	// Trigger is the event that causes the transition
	Trigger string
	// Guard is an optional condition that must be true for the transition
	Guard TransitionGuard
	// Action is executed during the transition
	Action TransitionAction
}

// StateAction is a function executed when entering or exiting a state.
type StateAction func(ctx context.Context) error //nolint:revive // keeping signature for clarity considering we have TransitionAction as well

// TransitionAction is a function executed during a state transition.
type TransitionAction func(ctx context.Context, from, to string) error

// TransitionGuard is a function that determines if a transition is allowed.
type TransitionGuard func(ctx context.Context) bool

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
	states []StateDefinition
}

func (o *statesOption) apply(c *Config) {
	c.States = append([]StateDefinition(nil), o.states...)
}

// WithStates sets the available states for the state machine.
func WithStates(states ...StateDefinition) Option {
	return &statesOption{states: states}
}

type transitionsOption struct {
	transitions []TransitionDefinition
}

func (o *transitionsOption) apply(c *Config) {
	c.Transitions = append([]TransitionDefinition(nil), o.transitions...)
}

// WithTransitions sets the allowed transitions for the state machine.
func WithTransitions(transitions ...TransitionDefinition) Option {
	return &transitionsOption{transitions: transitions}
}

type persistStateOption struct {
	persist bool
}

func (o *persistStateOption) apply(c *Config) {
	c.PersistState = o.persist
}

// WithPersistState enables or disables state persistence.
func WithPersistState(persist bool) Option {
	return &persistStateOption{persist: persist}
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

type enableMetricsOption struct {
	enable bool
}

func (o *enableMetricsOption) apply(c *Config) {
	c.EnableMetrics = o.enable
}

// WithMetrics enables or disables metrics collection.
func WithMetrics(enable bool) Option {
	return &enableMetricsOption{enable: enable}
}

type enableTracingOption struct {
	enable bool
}

func (o *enableTracingOption) apply(c *Config) {
	c.EnableTracing = o.enable
}

// WithTracing enables or disables distributed tracing.
func WithTracing(enable bool) Option {
	return &enableTracingOption{enable: enable}
}

// NewConfig creates a new state machine configuration with the provided options.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Name:          "",
		Description:   "",
		InitialState:  "unspecified",
		States:        []StateDefinition{},
		Transitions:   []TransitionDefinition{},
		PersistState:  false,
		StateTimeout:  30 * time.Second,
		EnableMetrics: true,
		EnableTracing: true,
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
		if state.Name == "" {
			return fmt.Errorf("%w: state name cannot be empty", ErrInvalidConfig)
		}
		if stateNames[state.Name] {
			return fmt.Errorf("%w: duplicate state name: %s", ErrInvalidConfig, state.Name)
		}
		stateNames[state.Name] = true
		if state.Name == c.InitialState {
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
