// SPDX-License-Identifier: BSD-3-Clause

package hwmon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultHwmonPath is the default path to hwmon devices in sysfs.
	DefaultHwmonPath = "/sys/class/hwmon"
)

// ReadInt reads an integer value from the specified hwmon file path.
func ReadInt(path string) (int, error) {
	return ReadIntCtx(context.Background(), path)
}

// ReadIntCtx reads an integer value from the specified hwmon file path with context support.
func ReadIntCtx(ctx context.Context, path string) (int, error) {
	if path == "" {
		return 0, fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	done := make(chan struct {
		value int
		err   error
	}, 1)

	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			done <- struct {
				value int
				err   error
			}{0, mapFileError(err, path)}
			return
		}

		value, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			done <- struct {
				value int
				err   error
			}{0, fmt.Errorf("%w: failed to parse integer from %s: %v", ErrInvalidValue, path, err)}
			return
		}

		done <- struct {
			value int
			err   error
		}{value, nil}
	}()

	select {
	case result := <-done:
		return result.value, result.err
	case <-ctx.Done():
		return 0, fmt.Errorf("%w: %v", ErrOperationTimeout, ctx.Err())
	}
}

// WriteInt writes an integer value to the specified hwmon file path.
func WriteInt(path string, value int) error {
	return WriteIntCtx(context.Background(), path, value)
}

// WriteIntCtx writes an integer value to the specified hwmon file path with context support.
func WriteIntCtx(ctx context.Context, path string, value int) error {
	if path == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	done := make(chan error, 1)

	go func() {
		data := strconv.Itoa(value)
		err := os.WriteFile(path, []byte(data), 0644)
		if err != nil {
			done <- mapFileError(err, path)
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrOperationTimeout, ctx.Err())
	}
}

// ReadString reads a string value from the specified hwmon file path.
func ReadString(path string) (string, error) {
	return ReadStringCtx(context.Background(), path)
}

// ReadStringCtx reads a string value from the specified hwmon file path with context support.
func ReadStringCtx(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	done := make(chan struct {
		value string
		err   error
	}, 1)

	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			done <- struct {
				value string
				err   error
			}{"", mapFileError(err, path)}
			return
		}

		value := strings.TrimSpace(string(data))
		done <- struct {
			value string
			err   error
		}{value, nil}
	}()

	select {
	case result := <-done:
		return result.value, result.err
	case <-ctx.Done():
		return "", fmt.Errorf("%w: %v", ErrOperationTimeout, ctx.Err())
	}
}

// WriteString writes a string value to the specified hwmon file path.
func WriteString(path, value string) error {
	return WriteStringCtx(context.Background(), path, value)
}

// WriteStringCtx writes a string value to the specified hwmon file path with context support.
func WriteStringCtx(ctx context.Context, path, value string) error {
	if path == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	done := make(chan error, 1)

	go func() {
		err := os.WriteFile(path, []byte(value), 0644)
		if err != nil {
			done <- mapFileError(err, path)
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrOperationTimeout, ctx.Err())
	}
}

// ListDevices returns a list of all hwmon device paths.
func ListDevices() ([]string, error) {
	return ListDevicesCtx(context.Background())
}

// ListDevicesCtx returns a list of all hwmon device paths with context support.
func ListDevicesCtx(ctx context.Context) ([]string, error) {
	return ListDevicesInPathCtx(ctx, DefaultHwmonPath)
}

// ListDevicesInPath returns a list of hwmon device paths in the specified directory.
func ListDevicesInPath(hwmonPath string) ([]string, error) {
	return ListDevicesInPathCtx(context.Background(), hwmonPath)
}

// ListDevicesInPathCtx returns a list of hwmon device paths in the specified directory with context support.
func ListDevicesInPathCtx(ctx context.Context, hwmonPath string) ([]string, error) {
	if hwmonPath == "" {
		return nil, fmt.Errorf("%w: hwmon path cannot be empty", ErrInvalidPath)
	}

	done := make(chan struct {
		devices []string
		err     error
	}, 1)

	go func() {
		entries, err := os.ReadDir(hwmonPath)
		if err != nil {
			done <- struct {
				devices []string
				err     error
			}{nil, mapFileError(err, hwmonPath)}
			return
		}

		var devices []string
		hwmonPattern := regexp.MustCompile(`^hwmon\d+$`)

		for _, entry := range entries {
			if entry.IsDir() && hwmonPattern.MatchString(entry.Name()) {
				devicePath := filepath.Join(hwmonPath, entry.Name())
				devices = append(devices, devicePath)
			}
		}

		done <- struct {
			devices []string
			err     error
		}{devices, nil}
	}()

	select {
	case result := <-done:
		return result.devices, result.err
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %v", ErrOperationTimeout, ctx.Err())
	}
}

// FindDeviceByName finds a hwmon device by its name attribute.
func FindDeviceByName(deviceName string) (string, error) {
	return FindDeviceByNameCtx(context.Background(), deviceName)
}

// FindDeviceByNameCtx finds a hwmon device by its name attribute with context support.
func FindDeviceByNameCtx(ctx context.Context, deviceName string) (string, error) {
	return FindDeviceByNameInPathCtx(ctx, DefaultHwmonPath, deviceName)
}

// FindDeviceByNameInPath finds a hwmon device by its name attribute in the specified directory.
func FindDeviceByNameInPath(hwmonPath, deviceName string) (string, error) {
	return FindDeviceByNameInPathCtx(context.Background(), hwmonPath, deviceName)
}

// FindDeviceByNameInPathCtx finds a hwmon device by its name attribute in the specified directory with context support.
func FindDeviceByNameInPathCtx(ctx context.Context, hwmonPath, deviceName string) (string, error) {
	if deviceName == "" {
		return "", fmt.Errorf("%w: device name cannot be empty", ErrInvalidPath)
	}

	devices, err := ListDevicesInPathCtx(ctx, hwmonPath)
	if err != nil {
		return "", err
	}

	for _, device := range devices {
		nameFile := filepath.Join(device, "name")
		name, err := ReadStringCtx(ctx, nameFile)
		if err != nil {
			continue // Skip devices where we can't read the name
		}

		if name == deviceName {
			return device, nil
		}
	}

	return "", fmt.Errorf("%w: device with name '%s'", ErrDeviceNotFound, deviceName)
}

// ListAttributes returns a list of attribute files in the specified device directory that match the pattern.
func ListAttributes(devicePath, pattern string) ([]string, error) {
	return ListAttributesCtx(context.Background(), devicePath, pattern)
}

// ListAttributesCtx returns a list of attribute files in the specified device directory that match the pattern with context support.
func ListAttributesCtx(ctx context.Context, devicePath, pattern string) ([]string, error) {
	if devicePath == "" {
		return nil, fmt.Errorf("%w: device path cannot be empty", ErrInvalidPath)
	}

	done := make(chan struct {
		attributes []string
		err        error
	}, 1)

	go func() {
		entries, err := os.ReadDir(devicePath)
		if err != nil {
			done <- struct {
				attributes []string
				err        error
			}{nil, mapFileError(err, devicePath)}
			return
		}

		var attributes []string
		var regex *regexp.Regexp

		if pattern != "" {
			regex, err = regexp.Compile(pattern)
			if err != nil {
				done <- struct {
					attributes []string
					err        error
				}{nil, fmt.Errorf("%w: invalid pattern '%s': %v", ErrInvalidValue, pattern, err)}
				return
			}
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				if regex == nil || regex.MatchString(entry.Name()) {
					attributes = append(attributes, entry.Name())
				}
			}
		}

		done <- struct {
			attributes []string
			err        error
		}{attributes, nil}
	}()

	select {
	case result := <-done:
		return result.attributes, result.err
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %v", ErrOperationTimeout, ctx.Err())
	}
}

// FileExists checks if a hwmon file exists.
func FileExists(path string) bool {
	return FileExistsCtx(context.Background(), path)
}

// FileExistsCtx checks if a hwmon file exists with context support.
func FileExistsCtx(ctx context.Context, path string) bool {
	if path == "" {
		return false
	}

	done := make(chan bool, 1)

	go func() {
		_, err := os.Stat(path)
		done <- err == nil
	}()

	select {
	case exists := <-done:
		return exists
	case <-ctx.Done():
		return false
	}
}

// IsWritable checks if a hwmon file is writable.
func IsWritable(path string) bool {
	return IsWritableCtx(context.Background(), path)
}

// IsWritableCtx checks if a hwmon file is writable with context support.
func IsWritableCtx(ctx context.Context, path string) bool {
	if path == "" {
		return false
	}

	done := make(chan bool, 1)

	go func() {
		// Try to open the file for writing to check if it's writable
		file, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err == nil {
			file.Close()
			done <- true
		} else {
			done <- false
		}
	}()

	select {
	case writable := <-done:
		return writable
	case <-ctx.Done():
		return false
	}
}

// WaitForDevice waits for a hwmon device to appear with the specified name.
func WaitForDevice(deviceName string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WaitForDeviceCtx(ctx, deviceName)
}

// WaitForDeviceCtx waits for a hwmon device to appear with the specified name with context support.
func WaitForDeviceCtx(ctx context.Context, deviceName string) (string, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("%w: %v", ErrOperationTimeout, ctx.Err())
		case <-ticker.C:
			device, err := FindDeviceByNameCtx(ctx, deviceName)
			if err == nil {
				return device, nil
			}
			// Continue waiting if device not found
		}
	}
}

// mapFileError maps OS file errors to hwmon package errors.
func mapFileError(err error, path string) error {
	if err == nil {
		return nil
	}

	switch {
	case os.IsNotExist(err):
		return fmt.Errorf("%w: %s", ErrFileNotFound, path)
	case os.IsPermission(err):
		return fmt.Errorf("%w: %s", ErrPermissionDenied, path)
	default:
		if strings.Contains(err.Error(), "read") {
			return fmt.Errorf("%w: %s: %v", ErrReadFailure, path, err)
		}
		if strings.Contains(err.Error(), "write") {
			return fmt.Errorf("%w: %s: %v", ErrWriteFailure, path, err)
		}
		return fmt.Errorf("%w: %s: %v", ErrReadFailure, path, err)
	}
}
