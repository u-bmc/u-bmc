// SPDX-License-Identifier: BSD-3-Clause

package state

import "errors"

var (
	// ErrInvalidConfig indicates that the state machine configuration is invalid.
	ErrInvalidConfig = errors.New("invalid state machine configuration")
	// ErrStateMachineNotFound indicates that the requested state machine does not exist.
	ErrStateMachineNotFound = errors.New("state machine not found")
	// ErrStateMachineExists indicates that a state machine with the same name already exists.
	ErrStateMachineExists = errors.New("state machine already exists")
	// ErrInvalidState indicates that the specified state is not valid for the state machine.
	ErrInvalidState = errors.New("invalid state")
	// ErrInvalidTrigger indicates that the specified trigger is not valid for the current state.
	ErrInvalidTrigger = errors.New("invalid trigger")
	// ErrInvalidTransition indicates that the requested state transition is not allowed.
	ErrInvalidTransition = errors.New("invalid state transition")
	// ErrTransitionTimeout indicates that a state transition exceeded the configured timeout.
	ErrTransitionTimeout = errors.New("state transition timeout")
	// ErrTransitionGuardFailed indicates that a transition guard condition was not met.
	ErrTransitionGuardFailed = errors.New("transition guard condition failed")
	// ErrStateActionFailed indicates that a state entry or exit action failed.
	ErrStateActionFailed = errors.New("state action failed")
	// ErrTransitionActionFailed indicates that a transition action failed.
	ErrTransitionActionFailed = errors.New("transition action failed")
	// ErrStateMachineLocked indicates that the state machine is locked and cannot be modified.
	ErrStateMachineLocked = errors.New("state machine is locked")
	// ErrConcurrentModification indicates that a concurrent modification was attempted.
	ErrConcurrentModification = errors.New("concurrent modification detected")
	// ErrPersistenceFailed indicates that persisting the state failed.
	ErrPersistenceFailed = errors.New("failed to persist state")
	// ErrNilContext indicates that a nil context was provided.
	ErrNilContext = errors.New("context cannot be nil")
	// ErrNilCallback indicates that a nil callback was provided.
	ErrNilCallback = errors.New("callback cannot be nil")
	// ErrAlreadyInState indicates that the state machine is already in the requested state.
	ErrAlreadyInState = errors.New("already in requested state")
	// ErrStateMachineNotStarted indicates that the state machine has not been started.
	ErrStateMachineNotStarted = errors.New("state machine not started")
	// ErrStateMachineAlreadyStarted indicates that the state machine has already been started.
	ErrStateMachineAlreadyStarted = errors.New("state machine already started")
	// ErrStateMachineStopped indicates that the state machine has been stopped.
	ErrStateMachineStopped = errors.New("state machine stopped")
)
