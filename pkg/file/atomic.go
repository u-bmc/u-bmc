// SPDX-License-Identifier: BSD-3-Clause

package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AtomicCreateFile creates a file atomically by first writing to a temporary file
// and then renaming it to the desired filename.
func AtomicCreateFile(filename string, data []byte, perm os.FileMode) error {
	if stat, err := os.Stat(filename); err == nil && stat.Mode().IsRegular() {
		return fmt.Errorf("%w: %s", os.ErrExist, filename)
	}

	dir := filepath.Dir(filename)
	tmpfile, err := os.CreateTemp(dir, fmt.Sprintf(".%s.tmp.*", filepath.Base(filename)))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpname := tmpfile.Name()

	defer func() {
		if err != nil {
			os.Remove(tmpname)
		}
	}()

	if _, err = tmpfile.Write(data); err != nil {
		tmpfile.Close()
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err = tmpfile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err = os.Chmod(tmpname, perm); err != nil {
		return fmt.Errorf("failed to chmod temporary file: %w", err)
	}

	if err = os.Rename(tmpname, filename); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// AtomicUpdateFile updates a file atomically by creating a copy, appending new content,
// and then renaming it to replace the original file.
func AtomicUpdateFile(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	tmpfile, err := os.CreateTemp(dir, fmt.Sprintf(".%s.tmp.*", filepath.Base(filename)))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpname := tmpfile.Name()

	defer func() {
		if err != nil {
			os.Remove(tmpname)
		}
	}()

	src, err := os.Open(filename)
	if err == nil {
		_, err = io.Copy(tmpfile, src)
		src.Close()
		if err != nil {
			tmpfile.Close()
			return fmt.Errorf("failed to copy original file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		tmpfile.Close()
		return fmt.Errorf("failed to open original file: %w", err)
	}

	if _, err = tmpfile.Write(data); err != nil {
		tmpfile.Close()
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err = tmpfile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err = os.Chmod(tmpname, perm); err != nil {
		return fmt.Errorf("failed to chmod temporary file: %w", err)
	}

	if err = os.Rename(tmpname, filename); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}
