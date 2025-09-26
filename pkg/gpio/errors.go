// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package gpio

import "errors"

var (
	// ErrChipNotFound indicates that the specified GPIO chip could not be found.
	ErrChipNotFound = errors.New("GPIO chip not found")

	// ErrLineNotFound indicates that the specified GPIO line could not be found.
	ErrLineNotFound = errors.New("GPIO line not found")

	// ErrPermissionDenied indicates insufficient permissions for the GPIO operation.
	ErrPermissionDenied = errors.New("permission denied for GPIO operation")

	// ErrInvalidValue indicates that an invalid value was provided for a GPIO operation.
	ErrInvalidValue = errors.New("invalid GPIO value")

	// ErrInvalidDuration indicates that an invalid duration was provided.
	ErrInvalidDuration = errors.New("invalid duration")

	// ErrOperationFailed indicates that a GPIO operation failed.
	ErrOperationFailed = errors.New("GPIO operation failed")

	// ErrLineClosed indicates that an operation was attempted on a closed GPIO line.
	ErrLineClosed = errors.New("GPIO line is closed")
)
