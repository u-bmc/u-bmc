// SPDX-License-Identifier: BSD-3-Clause

package hwmon

import "errors"

var (
	// ErrFileNotFound indicates that the specified hwmon file does not exist.
	ErrFileNotFound = errors.New("hwmon file not found")
	// ErrPermissionDenied indicates that access to the hwmon file was denied.
	ErrPermissionDenied = errors.New("permission denied accessing hwmon file")
	// ErrInvalidValue indicates that the value read from or written to hwmon is invalid.
	ErrInvalidValue = errors.New("invalid hwmon value")
	// ErrDeviceNotFound indicates that the specified hwmon device was not found.
	ErrDeviceNotFound = errors.New("hwmon device not found")
	// ErrReadFailure indicates that reading from hwmon failed.
	ErrReadFailure = errors.New("hwmon read failure")
	// ErrWriteFailure indicates that writing to hwmon failed.
	ErrWriteFailure = errors.New("hwmon write failure")
	// ErrInvalidPath indicates that the provided hwmon path is invalid.
	ErrInvalidPath = errors.New("invalid hwmon path")
	// ErrOperationTimeout indicates that the hwmon operation timed out.
	ErrOperationTimeout = errors.New("hwmon operation timeout")
)
