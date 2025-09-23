// SPDX-License-Identifier: BSD-3-Clause

package id

import "errors"

var (
	// ErrDirectoryCreation indicates a failure to create the directory for storing persistent IDs.
	ErrDirectoryCreation = errors.New("failed to create directory for persistent ID storage")
	// ErrFileCreation indicates a failure to create the persistent ID file.
	ErrFileCreation = errors.New("failed to create persistent ID file")
	// ErrFileRead indicates a failure to read the persistent ID file.
	ErrFileRead = errors.New("failed to read persistent ID file")
	// ErrFileWrite indicates a failure to write the persistent ID file.
	ErrFileWrite = errors.New("failed to write persistent ID file")
	// ErrFileUpdate indicates a failure to update the persistent ID file.
	ErrFileUpdate = errors.New("failed to update persistent ID file")
	// ErrInvalidUUID indicates that the content of a persistent ID file is not a valid UUID.
	ErrInvalidUUID = errors.New("invalid UUID format in persistent ID file")
	// ErrInvalidPath indicates that an invalid file path was provided for persistent ID storage.
	ErrInvalidPath = errors.New("invalid path for persistent ID storage")
	// ErrEmptyFilename indicates that an empty filename was provided for persistent ID storage.
	ErrEmptyFilename = errors.New("filename for persistent ID cannot be empty")
	// ErrFilePermissions indicates a failure to set proper permissions on the persistent ID file.
	ErrFilePermissions = errors.New("failed to set permissions on persistent ID file")
	// ErrFileStat indicates a failure to get file statistics for the persistent ID file.
	ErrFileStat = errors.New("failed to get file statistics for persistent ID file")
)
