// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

// AtomicCreateFile creates a file atomically by first writing to a temporary file
// and then renaming it to the desired filename.
func AtomicCreateFile(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	tmpfile, err := os.CreateTemp(dir, fmt.Sprintf(".%s.tmp.*", filepath.Base(filename)))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTemporaryFileCreation, err)
	}
	tmpname := tmpfile.Name()

	defer func() {
		if err != nil {
			_ = os.Remove(tmpname)
		}
	}()

	if _, err = tmpfile.Write(data); err != nil {
		_ = tmpfile.Close()
		return fmt.Errorf("%w: %w", ErrTemporaryFileWrite, err)
	}

	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("%w: %w", ErrTemporaryFileClose, err)
	}

	if err := os.Chmod(tmpname, perm); err != nil {
		return fmt.Errorf("%w: %w", ErrTemporaryFileChmod, err)
	}

	if err = unix.Renameat2(unix.AT_FDCWD, filename, unix.AT_FDCWD, tmpname, unix.RENAME_NOREPLACE); err != nil {
		if errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("%w: %s", ErrFileAlreadyExists, tmpname)
		}
		return fmt.Errorf("%w: %w", ErrAtomicRename, err)
	}

	return nil
}

// AtomicUpdateFile updates a file atomically by creating a copy, appending new content,
// and then renaming it to replace the original file.
func AtomicUpdateFile(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	tmpfile, err := os.CreateTemp(dir, fmt.Sprintf(".%s.tmp.*", filepath.Base(filename)))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTemporaryFileCreation, err)
	}
	tmpname := tmpfile.Name()

	defer func() {
		if err != nil {
			_ = os.Remove(tmpname)
		}
	}()

	src, err := os.Open(filename)
	if err == nil {
		_, err = io.Copy(tmpfile, src)
		_ = src.Close()
		if err != nil {
			_ = tmpfile.Close()
			return fmt.Errorf("%w: %w", ErrOriginalFileCopy, err)
		}
	} else if !os.IsNotExist(err) {
		_ = tmpfile.Close()
		return fmt.Errorf("%w: %w", ErrOriginalFileOpen, err)
	}

	if _, err = tmpfile.Write(data); err != nil {
		_ = tmpfile.Close()
		return fmt.Errorf("%w: %w", ErrTemporaryFileWrite, err)
	}

	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("%w: %w", ErrTemporaryFileClose, err)
	}

	if err = os.Chmod(tmpname, perm); err != nil {
		return fmt.Errorf("%w: %w", ErrTemporaryFileChmod, err)
	}

	if err = os.Rename(tmpname, filename); err != nil {
		return fmt.Errorf("%w: %w", ErrAtomicRename, err)
	}

	return nil
}
