// SPDX-License-Identifier: BSD-3-Clause

package mount

import "errors"

var (
	// ErrMountPointCreation indicates failure to create a mount point directory.
	ErrMountPointCreation = errors.New("failed to create mount point")
	// ErrMountFailed indicates the mount operation failed.
	ErrMountFailed = errors.New("mount operation failed")
	// ErrMountVerification indicates the mounted filesystem doesn't match expected specification.
	ErrMountVerification = errors.New("mount verification failed")
	// ErrProcMountsRead indicates failure to read /proc/mounts.
	ErrProcMountsRead = errors.New("failed to read /proc/mounts")
	// ErrInvalidMountSpec indicates an invalid mount specification was provided.
	ErrInvalidMountSpec = errors.New("invalid mount specification")
	// ErrMountPointNotDirectory indicates the mount point exists but is not a directory.
	ErrMountPointNotDirectory = errors.New("mount point is not a directory")
	// ErrInsufficientPrivileges indicates the operation requires elevated privileges.
	ErrInsufficientPrivileges = errors.New("insufficient privileges for mount operation")
	// ErrFilesystemNotSupported indicates the filesystem type is not supported by the kernel.
	ErrFilesystemNotSupported = errors.New("filesystem type not supported")
	// ErrDeviceBusy indicates the mount point or device is busy and cannot be mounted.
	ErrDeviceBusy = errors.New("device or mount point is busy")
	// ErrMountAlreadyExists indicates the filesystem is already mounted at the specified location.
	ErrMountAlreadyExists = errors.New("filesystem already mounted")
)
