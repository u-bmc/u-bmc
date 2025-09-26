// SPDX-License-Identifier: BSD-3-Clause

// Package statemgr provides state management for BMC components including hosts, chassis, and management controllers.
//
// The state manager service orchestrates component state transitions and coordinates with other services
// to ensure consistent system behavior. It maintains finite state machines for each component and handles
// state persistence, event broadcasting, and integration with power and LED management services.
//
// # Architecture Overview
//
// The statemgr service acts as the central coordinator for component states:
//  1. API clients send state change requests to statemgr
//  2. statemgr validates transitions using finite state machines
//  3. statemgr delegates physical operations to powermgr via NATS IPC
//  4. statemgr coordinates visual feedback via ledmgr
//  5. statemgr persists state changes to JetStream (optional)
//  6. statemgr broadcasts state events for monitoring
//
// # Component State Machines
//
// Each component type has its own state machine with defined states and transitions:
//
// Host States:
//   - OFF: Host is powered off
//   - ON: Host is powered on and operational
//   - TRANSITIONING: Host is changing power state
//   - QUIESCED: Host is in a suspended state
//   - DIAGNOSTIC: Host is in diagnostic mode
//   - ERROR: Host has encountered an error
//
// Chassis States:
//   - OFF: Chassis is powered off
//   - ON: Chassis is powered on
//   - TRANSITIONING: Chassis is changing state
//   - WARNING: Chassis has warning conditions
//   - CRITICAL: Chassis has critical conditions
//   - FAILED: Chassis has failed
//
// Management Controller States:
//   - NOT_READY: BMC is not ready for operations
//   - READY: BMC is ready and operational
//   - DISABLED: BMC is disabled
//   - ERROR: BMC has encountered an error
//   - QUIESCED: BMC is in a suspended state
//   - DIAGNOSTIC: BMC is in diagnostic mode
//
// # State Transitions
//
// State transitions are triggered by API actions and internal events:
//
//	// API-triggered transitions
//	HOST_ACTION_ON: OFF -> TRANSITIONING -> ON
//	HOST_ACTION_OFF: ON -> TRANSITIONING -> OFF
//	HOST_ACTION_REBOOT: ON -> TRANSITIONING -> ON
//	HOST_ACTION_FORCE_OFF: ERROR -> OFF
//	HOST_ACTION_FORCE_RESTART: ERROR -> TRANSITIONING -> ON
//
//	// Internal transitions
//	TRANSITION_COMPLETE_ON: TRANSITIONING -> ON
//	TRANSITION_COMPLETE_OFF: TRANSITIONING -> OFF
//	TRANSITION_ERROR: TRANSITIONING -> ERROR
//	TRANSITION_TIMEOUT: TRANSITIONING -> QUIESCED
//	TRANSITION_RESUME: QUIESCED -> ON
//
// # Basic Usage
//
// The state manager is typically used as a service in BMC systems:
//
//	statemgr := statemgr.New(
//		statemgr.WithServiceName("statemgr"),
//		statemgr.WithHostManagement(true),
//		statemgr.WithChassisManagement(true),
//		statemgr.WithBMCManagement(true),
//		statemgr.WithPersistStateChanges(true),
//		statemgr.WithBroadcastStateChanges(true),
//	)
//
//	// Run as part of BMC service framework
//	if err := statemgr.Run(ctx, ipcConn); err != nil {
//		log.Fatal(err)
//	}
//
// # IPC Communication
//
// The service exposes NATS-based endpoints for state management:
//
//	// State queries
//	statemgr.host.{id}.state -> GetHostResponse
//	statemgr.chassis.{id}.state -> GetChassisResponse
//	statemgr.bmc.{id}.state -> GetManagementControllerResponse
//
//	// State control
//	statemgr.host.{id}.control -> ChangeHostStateRequest/Response
//	statemgr.chassis.{id}.control -> ChangeChassisStateRequest/Response
//	statemgr.bmc.{id}.control -> ChangeManagementControllerStateRequest/Response
//
//	// Component information
//	statemgr.host.{id}.info -> Host
//	statemgr.chassis.{id}.info -> Chassis
//	statemgr.bmc.{id}.info -> ManagementController
//
// # JetStream Persistence
//
// State changes can be persisted to JetStream for recovery after power loss:
//
//	statemgr := statemgr.New(
//		statemgr.WithPersistStateChanges(true),
//		statemgr.WithStreamName("STATEMGR"),
//		statemgr.WithStreamSubjects("statemgr.state.>", "statemgr.event.>"),
//		statemgr.WithStreamRetention(0), // Keep forever
//	)
//
// State events are published to subjects like:
//   - statemgr.state.host.0
//   - statemgr.state.chassis.0
//   - statemgr.state.bmc.0
//
// # Event Broadcasting
//
// State transitions are broadcast for monitoring and integration:
//
//	statemgr := statemgr.New(
//		statemgr.WithBroadcastStateChanges(true),
//	)
//
// Transition events are published to subjects like:
//   - statemgr.event.host.0.transition
//   - statemgr.event.chassis.0.transition
//   - statemgr.event.bmc.0.transition
//
// # Integration with Power Manager
//
// The state manager coordinates with powermgr for physical operations:
//
//  1. Client requests host power on via statemgr API
//  2. statemgr validates OFF -> TRANSITIONING transition
//  3. statemgr sends power on request to powermgr
//  4. powermgr executes physical power operation
//  5. powermgr responds with success/failure
//  6. statemgr transitions to ON or ERROR state
//  7. statemgr persists and broadcasts state change
//
// Subject patterns for power control:
//   - powermgr.host.{id}.action
//   - powermgr.chassis.{id}.action
//   - powermgr.bmc.{id}.action
//
// # Integration with LED Manager
//
// The state manager coordinates with ledmgr for visual feedback:
//
//  1. Host transitions to ON state
//  2. statemgr sends LED control request to ledmgr
//  3. ledmgr sets power LED to solid on
//  4. Operator sees visual confirmation
//
// Subject patterns for LED control:
//   - ledmgr.host.{id}.control
//   - ledmgr.chassis.{id}.control
//   - ledmgr.bmc.{id}.control
//
// # State Machine Configuration
//
// State machines are configured with callbacks and timeouts:
//
//	sm, err := state.NewStateMachine(
//		state.WithName("host.0"),
//		state.WithInitialState("HOST_STATUS_OFF"),
//		state.WithStates("HOST_STATUS_OFF", "HOST_STATUS_ON", "HOST_STATUS_TRANSITIONING"),
//		state.WithActionTransition("HOST_STATUS_OFF", "HOST_STATUS_TRANSITIONING", "HOST_ACTION_ON", actionFunc),
//		state.WithTransition("HOST_STATUS_TRANSITIONING", "HOST_STATUS_ON", "TRANSITION_COMPLETE_ON"),
//		state.WithStateTimeout(30 * time.Second),
//		state.WithStateEntry(entryCallback),
//		state.WithStateExit(exitCallback),
//		state.WithPersistence(persistenceCallback),
//		state.WithBroadcast(broadcastCallback),
//	)
//
// # Error Handling
//
// The package provides specific error types for state operations:
//
//	err := sm.Fire(ctx, "HOST_ACTION_ON")
//	if err != nil {
//		switch {
//		case errors.Is(err, statemgr.ErrInvalidStateTransition):
//			log.Error("Invalid state transition requested")
//		case errors.Is(err, statemgr.ErrStateTransitionTimeout):
//			log.Error("State transition timed out")
//		case errors.Is(err, statemgr.ErrComponentNotFound):
//			log.Error("Component not found")
//		case errors.Is(err, statemgr.ErrStatePersistenceFailed):
//			log.Error("Failed to persist state change")
//		default:
//			log.Errorf("Unexpected error: %v", err)
//		}
//	}
//
// # State Recovery
//
// When persistence is enabled, states can be recovered after restart:
//
//  1. Service starts and connects to JetStream
//  2. State machines are created with initial states
//  3. Persisted state events are replayed
//  4. State machines are restored to last known states
//  5. Normal operation resumes
//
// # Metrics and Observability
//
// The service provides comprehensive metrics:
//
//   - statemgr_transitions_total: Total number of state transitions
//   - statemgr_transition_duration_seconds: Duration of state transitions
//   - statemgr_transition_failures_total: Total number of failed transitions
//   - statemgr_current_state: Current state of components (encoded as integer)
//
// All metrics are labeled with component name, operation type, and status.
//
// # Resource Management
//
// Always ensure proper cleanup of state manager resources:
//
//	statemgr := statemgr.New(config...)
//	defer statemgr.Close()
//
//	// State machines are automatically stopped
//	// JetStream streams are preserved
//	// IPC connections are cleaned up on context cancellation
//
// # Thread Safety
//
// The StateMgr service is thread-safe and can handle concurrent operations.
// Individual state machines serialize operations per component.
// Multiple components can transition simultaneously.
//
// # Platform Considerations
//
// This package is designed for BMC systems requiring state management:
//
// Supported Scenarios:
//   - Server power management
//   - Chassis monitoring and control
//   - BMC lifecycle management
//   - Multi-host systems
//   - High-availability configurations
//   - Remote management interfaces
//
// Requirements:
//   - NATS server for IPC communication
//   - JetStream for persistence (optional)
//   - Integration with powermgr service
//   - Integration with ledmgr service (optional)
//
// # Performance Considerations
//
// State operations have different performance characteristics:
//
//   - State queries: ~1-10ms (in-memory)
//   - State transitions: ~10-100ms (validation + callbacks)
//   - Persistence: ~10-50ms (JetStream write)
//   - Broadcasting: ~1-10ms (NATS publish)
//   - Power coordination: ~100ms-30s (depends on hardware)
//
// The service is optimized for:
//   - Low latency state queries
//   - Reliable state persistence
//   - Consistent state coordination
//   - Comprehensive event broadcasting
//   - Graceful error handling and recovery
package statemgr
