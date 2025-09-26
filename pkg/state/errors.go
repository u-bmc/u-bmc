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

	// ErrPersistenceFailed indicates that persisting the state failed.
	ErrPersistenceFailed = errors.New("failed to persist state")

	// ErrBroadcastFailed indicates that broadcasting a state change failed.
	ErrBroadcastFailed = errors.New("failed to broadcast state change")

	// ErrNilCallback indicates that a nil callback was provided.
	ErrNilCallback = errors.New("callback cannot be nil")

	// ErrTransitionTimeout indicates that a state transition exceeded the configured timeout.
	ErrTransitionTimeout = errors.New("state transition timeout")

	// ErrStateMachineNotStarted indicates that the state machine has not been started.
	ErrStateMachineNotStarted = errors.New("state machine not started")

	// ErrStateMachineAlreadyStarted indicates that the state machine has already been started.
	ErrStateMachineAlreadyStarted = errors.New("state machine already started")

	// ErrStateMachineStopped indicates that the state machine has been stopped.
	ErrStateMachineStopped = errors.New("state machine stopped")
)
