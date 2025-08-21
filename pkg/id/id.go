// SPDX-License-Identifier: BSD-3-Clause

package id

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/u-bmc/u-bmc/pkg/file"
)

// NewID generates and returns a new UUID as a string.
func NewID() string {
	return uuid.New().String()
}

// GetOrCreatePersistentID retrieves an existing UUID from a file or creates a new one if the file doesn't exist.
// It takes a filename and directory path, and returns the UUID string and any error encountered.
// If the file exists, it reads and parses the UUID from the file.
// If the file doesn't exist, it generates a new UUID and atomically writes it to the file.
func GetOrCreatePersistentID(name, path string) (string, error) {
	fullPath := filepath.Join(path, name)

	var idstr string
	if _, err := os.Stat(fullPath); err != nil && !os.IsNotExist(err) {
		return "", err
	} else if os.IsNotExist(err) {
		id := uuid.New()

		if err := file.AtomicCreateFile(fullPath, []byte(id.String()), os.ModePerm); err != nil && !os.IsExist(err) {
			return "", err
		}

		idstr = id.String()
	} else {
		b, err := os.ReadFile(fullPath)
		if err != nil {
			return "", err
		}

		id, err := uuid.ParseBytes(b)
		if err != nil {
			return "", err
		}

		idstr = id.String()
	}

	return idstr, nil
}

// UpdatePersistentID generates a new UUID and atomically updates the specified file with the new value.
// It takes a filename and directory path, and returns the new UUID string and any error encountered.
// This function will overwrite any existing UUID in the file.
func UpdatePersistentID(name, path string) (string, error) {
	id := uuid.New()

	if err := file.AtomicUpdateFile(filepath.Join(path, name), []byte(id.String()), os.ModePerm); err != nil {
		return "", err
	}

	return id.String(), nil
}
