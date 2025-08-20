// SPDX-License-Identifier: BSD-3-Clause

package id

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/u-bmc/u-bmc/pkg/file"
)

func NewID() string {
	return uuid.New().String()
}

func GetOrCreatePersistentID(name, path string) (string, error) {
	fullPath := filepath.Join(path, name)

	var id string
	if _, err := os.Stat(fullPath); err != nil && !os.IsNotExist(err) {
		return "", err
	} else if os.IsNotExist(err) {
		uuid := uuid.New()

		if err := file.AtomicCreateFile(filepath.Join(path, name), []byte(uuid.String()), os.ModePerm); err != nil && err != os.ErrExist {
			return "", err
		}

		id = uuid.String()
	} else {
		b, err := os.ReadFile(fullPath)
		if err != nil {
			return "", nil
		}

		uuid, err := uuid.ParseBytes(b)
		if err != nil {
			return "", nil
		}

		id = uuid.String()
	}

	return id, nil
}

func UpdatePersistentID(name, path string) (string, error) {
	id := uuid.New()

	if err := file.AtomicUpdateFile(filepath.Join(path, name), []byte(id.String()), os.ModePerm); err != nil {
		return "", err
	}

	return id.String(), nil
}
