// SPDX-License-Identifier: BSD-3-Clause

// Package state provides a comprehensive state machine implementation for BMC (Baseboard Management Controller)
// systems and other applications requiring robust state management with persistence, observability, and
// concurrent access support.
//
// # Overview
//
// This package implements finite state machines (FSMs) with the following key features:
//   - Thread-safe operations with read-write mutexes
//   - State persistence with configurable callbacks
//   - Distributed tracing and metrics collection
//   - Configurable timeouts for state transitions
//   - Guard conditions and transition actions
//   - State entry/exit actions
//   - Broadcast notifications for state changes
//   - DOT graph generation for visualization
//   - Multi-state machine management
//
// # Core Concepts
//
// State Machine: A computational model consisting of a finite number of states, transitions between
// those states, and actions. At any given time, the machine is in exactly one state.
//
// State: A distinct condition or situation in which the state machine can exist. Each state can have
// optional entry and exit actions that are executed when entering or leaving the state.
//
// Transition: A change from one state to another, triggered by an event (trigger). Transitions can
// have guard conditions that must be satisfied and actions that are executed during the transition.
//
// Trigger: An event or signal that can cause a state transition. Triggers are only valid for specific
// states and their associated transitions.
//
// Guard: A boolean condition that must be true for a transition to occur. Guards provide additional
// control over when transitions are allowed.
//
// Action: Code that is executed either when entering/exiting a state or during a transition.
//
// # Basic Usage
//
// Creating a simple state machine:
//
//	config := NewConfig(
//		WithName("power-management"),
//		WithDescription("BMC power state management"),
//		WithInitialState("off"),
//		WithStates(
//			StateDefinition{
//				Name: "off",
//				Description: "System is powered off",
//				OnEntry: func(ctx context.Context) error {
//					// Perform power-off actions
//					return nil
//				},
//			},
//			StateDefinition{
//				Name: "on",
//				Description: "System is powered on",
//				OnEntry: func(ctx context.Context) error {
//					// Perform power-on actions
//					return nil
//				},
//			},
//		),
//		WithTransitions(
//			TransitionDefinition{
//				From: "off",
//				To: "on",
//				Trigger: "power_on",
//				Action: func(ctx context.Context, from, to string) error {
//					// Execute power-on sequence
//					return nil
//				},
//			},
//			TransitionDefinition{
//				From: "on",
//				To: "off",
//				Trigger: "power_off",
//				Guard: func(ctx context.Context) bool {
//					// Check if safe to power off
//					return true
//				},
//			},
//		),
//		WithPersistState(true),
//	)
//
//	sm, err := New(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Set persistence callback before starting
//	sm.SetPersistenceCallback(func(machineName, state string) error {
//		// Save state to database, file, etc.
//		return saveStateToStorage(machineName, state)
//	})
//
//	// Start the state machine
//	ctx := context.Background()
//	if err := sm.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//
//	// Trigger a state transition
//	if err := sm.Fire(ctx, "power_on", nil); err != nil {
//		log.Printf("Transition failed: %v", err)
//	}
//
// # State Persistence
//
// The package supports state persistence through configurable callbacks. When enabled,
// the current state is persisted whenever it changes:
//
//	sm.SetPersistenceCallback(func(machineName, state string) error {
//		// Save state to database, file, etc.
//		return saveStateToStorage(machineName, state)
//	})
//
// Note: Persistence callbacks must be set before starting the state machine.
//
// # State Change Notifications
//
// Applications can receive notifications when state changes occur:
//
//	sm.SetBroadcastCallback(func(machineName, previousState, currentState, trigger string) error {
//		// Notify other components, send events, etc.
//		return notifyStateChange(machineName, previousState, currentState, trigger)
//	})
//
// Note: Broadcast callbacks must be set before starting the state machine.
//
// # Multi-State Machine Management
//
// The Manager type allows managing multiple state machines:
//
//	manager := NewManager()
//	manager.AddStateMachine(powerSM)
//	manager.AddStateMachine(thermalSM)
//	manager.AddStateMachine(networkSM)
//
//	// Get a specific state machine
//	sm, err := manager.GetStateMachine("power-management")
//	if err != nil {
//		log.Printf("State machine not found: %v", err)
//	}
//
// # Observability
//
// The package provides built-in support for observability:
//
// Tracing: Distributed tracing support using OpenTelemetry provides visibility into
// state transition flows across service boundaries. Enable with WithTracing(true).
//
// Metrics: Metrics collection can be enabled with WithMetrics(true) in the configuration.
//
// Logging: Comprehensive error reporting with structured error types for different
// failure scenarios.
//
// # Thread Safety
//
// All state machine operations are thread-safe. Multiple goroutines can safely:
//   - Query the current state
//   - Check if triggers can be fired
//   - Trigger state transitions
//   - Access state machine metadata
//
// The implementation uses read-write mutexes to allow concurrent reads while ensuring
// exclusive access for state modifications.
//
// # Error Handling
//
// The package defines specific error types for different failure scenarios:
//   - Configuration errors (ErrInvalidConfig)
//   - State/transition errors (ErrInvalidState, ErrInvalidTransition, ErrInvalidTrigger)
//   - Timeout errors (ErrTransitionTimeout)
//   - Guard/action failures (ErrTransitionGuardFailed, ErrStateActionFailed, ErrTransitionActionFailed)
//   - Concurrency errors (ErrConcurrentModification)
//   - Persistence errors (ErrPersistenceFailed)
//   - Lifecycle errors (ErrStateMachineNotStarted, ErrStateMachineAlreadyStarted, ErrStateMachineStopped)
//
// # BMC Integration
//
// This package is specifically designed for BMC systems where reliable state management
// is critical for:
//   - Power management (on/off/reset states)
//   - Thermal management (normal/warning/critical states)
//   - Boot sequence management (POST/boot/ready states)
//   - Network interface management (up/down/configuring states)
//   - Sensor monitoring (active/inactive/fault states)
//   - Firmware update processes (idle/downloading/updating/verifying states)
//
// The persistence and observability features ensure that state information survives
// BMC reboots and provides visibility into system behavior for debugging and monitoring.
package state
