// SPDX-License-Identifier: BSD-3-Clause

package file

import "errors"

var (
	// ErrTemporaryFileCreation indicates a failure to create a temporary file.
	ErrTemporaryFileCreation = errors.New("failed to create temporary file")
	// ErrTemporaryFileWrite indicates a failure to write data to a temporary file.
	ErrTemporaryFileWrite = errors.New("failed to write to temporary file")
	// ErrTemporaryFileClose indicates a failure to close a temporary file.
	ErrTemporaryFileClose = errors.New("failed to close temporary file")
	// ErrTemporaryFileChmod indicates a failure to set permissions on a temporary file.
	ErrTemporaryFileChmod = errors.New("failed to set permissions on temporary file")
	// ErrAtomicRename indicates a failure to atomically rename a temporary file to its final location.
	ErrAtomicRename = errors.New("failed to atomically rename temporary file")
	// ErrOriginalFileRead indicates a failure to read the original file during an update operation.
	ErrOriginalFileRead = errors.New("failed to read original file")
	// ErrOriginalFileOpen indicates a failure to open the original file during an update operation.
	ErrOriginalFileOpen = errors.New("failed to open original file")
	// ErrOriginalFileCopy indicates a failure to copy content from the original file to the temporary file.
	ErrOriginalFileCopy = errors.New("failed to copy original file content")
	// ErrFileAlreadyExists indicates that a file already exists when attempting atomic creation.
	ErrFileAlreadyExists = errors.New("file already exists")
	// ErrInvalidFileMode indicates that an invalid file mode was provided.
	ErrInvalidFileMode = errors.New("invalid file mode")
	// ErrEmptyFilename indicates that an empty filename was provided.
	ErrEmptyFilename = errors.New("filename cannot be empty")
	// ErrInvalidPath indicates that an invalid file path was provided.
	ErrInvalidPath = errors.New("invalid file path")
)
