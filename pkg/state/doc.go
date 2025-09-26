// SPDX-License-Identifier: BSD-3-Clause

// Package state provides a simplified, lightweight wrapper around the stateless finite state machine
// library for BMC (Baseboard Management Controller) systems and other applications requiring robust
// state management.
//
// # Overview
//
// This package is designed as a thin abstraction layer over the github.com/qmuntal/stateless library,
// providing BMC-specific conveniences while leveraging the underlying library's thread-safety,
// performance, and feature set. The wrapper minimizes additional locking and state duplication.
//
// Key features:
//   - Thin wrapper around proven stateless library
//   - Thread-safe operations (inherited from underlying library)
//   - State persistence callbacks
//   - State change broadcasting
//   - BMC-specific state machine builders
//   - Configurable timeouts and observability
//
// # Core Concepts
//
// State Machine: A computational model based on the stateless library's implementation.
// The underlying library handles all state management, transitions, and thread safety.
//
// State: Managed entirely by the underlying stateless library. No duplicate state tracking.
//
// Transition: Uses the stateless library's transition system with optional guards and actions.
//
// Persistence: Optional callbacks for state persistence, executed asynchronously.
//
// Broadcasting: Optional callbacks for state change notifications.
//
// # Basic Usage
//
// Creating a state machine using the simplified builder pattern:
//
//	// Create using the builder for common BMC patterns
//	sm, err := NewPowerStateMachine(
//		WithName("power-controller"),
//		WithInitialState("off"),
//		WithStates("off", "on", "transitioning"),
//		WithTransition("off", "on", "power_on"),
//		WithTransition("on", "off", "power_off"),
//		WithPersistence(func(name, state string) error {
//			return saveToDatabase(name, state)
//		}),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Fire transitions - thread-safe, handled by underlying library
//	if err := sm.Fire(ctx, "power_on"); err != nil {
//		log.Printf("Transition failed: %v", err)
//	}
//
//	// Check current state - no additional locking needed
//	currentState := sm.State()
//
// # Builder Patterns
//
// The package provides builder functions for common BMC state machine patterns:
//
//	// Generic state machine
//	sm := NewStateMachine(opts...)
//
//	// BMC power state machine
//	sm := NewPowerStateMachine(opts...)
//
//	// Thermal management state machine
//	sm := NewThermalStateMachine(opts...)
//
// # Persistence and Broadcasting
//
// State changes can be persisted and broadcast through callbacks:
//
//	sm, err := NewStateMachine(
//		WithPersistence(func(name, state string) error {
//			// Persist state asynchronously
//			return persistState(name, state)
//		}),
//		WithBroadcast(func(name, from, to, trigger string) error {
//			// Broadcast state change
//			return notifyStateChange(name, from, to, trigger)
//		}),
//	)
//
// # Thread Safety
//
// Thread safety is provided by the underlying stateless library. This wrapper adds
// minimal overhead and does not introduce additional locking mechanisms.
//
// Multiple goroutines can safely:
//   - Fire transitions
//   - Query current state
//   - Check permitted triggers
//   - Access machine metadata
//
// # Error Handling
//
// The package defines specific error types for different failure scenarios while
// also propagating errors from the underlying stateless library when appropriate.
//
// # BMC Integration
//
// This package is specifically designed for BMC systems with common patterns for:
//   - Power management (off/on/transitioning states)
//   - Thermal management (normal/warning/critical states)
//   - Boot sequence management (init/post/boot/ready states)
//   - Component health (healthy/degraded/failed states)
//
// The lightweight design ensures minimal impact on BMC performance while providing
// the reliability and features needed for critical system management.
package state
