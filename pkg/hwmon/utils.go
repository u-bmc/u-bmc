// SPDX-License-Identifier: BSD-3-Clause

//nolint:goconst
package hwmon

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Utility functions for hwmon package operations.

// ValidateDeviceName checks if a device name is valid for hwmon.
func ValidateDeviceName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: device name cannot be empty", ErrInvalidConfig)
	}

	if len(name) > 255 {
		return fmt.Errorf("%w: device name too long (max 255 characters)", ErrInvalidConfig)
	}

	invalidChars := regexp.MustCompile(`[^a-zA-Z0-9_\-.]`)
	if invalidChars.MatchString(name) {
		return fmt.Errorf("%w: device name contains invalid characters", ErrInvalidConfig)
	}

	return nil
}

// ValidateSensorLabel checks if a sensor label is valid.
func ValidateSensorLabel(label string) error {
	if label == "" {
		return fmt.Errorf("%w: sensor label cannot be empty", ErrInvalidConfig)
	}

	if len(label) > 64 {
		return fmt.Errorf("%w: sensor label too long (max 64 characters)", ErrInvalidConfig)
	}

	return nil
}

// ValidateHwmonPath checks if a path is a valid hwmon path.
func ValidateHwmonPath(path string) error {
	if path == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("%w: path must be absolute", ErrInvalidPath)
	}

	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return fmt.Errorf("%w: path contains invalid components", ErrInvalidPath)
	}

	return nil
}

// IsHwmonDevice checks if a directory name represents a hwmon device.
func IsHwmonDevice(name string) bool {
	hwmonPattern := regexp.MustCompile(`^hwmon\d+$`)
	return hwmonPattern.MatchString(name)
}

// ExtractHwmonNumber extracts the numeric ID from a hwmon device name.
func ExtractHwmonNumber(hwmonName string) (int, error) {
	if !IsHwmonDevice(hwmonName) {
		return 0, fmt.Errorf("%w: invalid hwmon device name: %s", ErrInvalidConfig, hwmonName)
	}

	numStr := strings.TrimPrefix(hwmonName, "hwmon")
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to parse hwmon number: %w", ErrInvalidConfig, err)
	}

	return num, nil
}

// ParseSensorFilename parses a sensor filename and extracts type, index, and attribute.
func ParseSensorFilename(filename string) (SensorType, int, SensorAttribute, error) {
	patterns := map[SensorType]*regexp.Regexp{
		SensorTypeTemperature: regexp.MustCompile(`^temp(\d+)_(.+)$`),
		SensorTypeVoltage:     regexp.MustCompile(`^in(\d+)_(.+)$`),
		SensorTypeFan:         regexp.MustCompile(`^fan(\d+)_(.+)$`),
		SensorTypePower:       regexp.MustCompile(`^power(\d+)_(.+)$`),
		SensorTypeCurrent:     regexp.MustCompile(`^curr(\d+)_(.+)$`),
		SensorTypeHumidity:    regexp.MustCompile(`^humidity(\d+)_(.+)$`),
		SensorTypePressure:    regexp.MustCompile(`^pressure(\d+)_(.+)$`),
		SensorTypePWM:         regexp.MustCompile(`^pwm(\d+)(_(.+))?$`),
	}

	for sensorType, pattern := range patterns {
		matches := pattern.FindStringSubmatch(filename)
		if len(matches) >= 2 {
			index, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			attribute := AttributeInput
			if len(matches) >= 3 && matches[2] != "" {
				attribute = parseAttributeString(matches[2])
			} else if len(matches) >= 4 && matches[3] != "" {
				attribute = parseAttributeString(matches[3])
			}

			return sensorType, index, attribute, nil
		}
	}

	return SensorTypeGeneric, 0, AttributeInput, fmt.Errorf("%w: unrecognized sensor filename: %s", ErrInvalidAttribute, filename)
}

// parseAttributeString converts an attribute string to SensorAttribute.
func parseAttributeString(attr string) SensorAttribute {
	switch attr {
	case "input":
		return AttributeInput
	case "label":
		return AttributeLabel
	case "min":
		return AttributeMin
	case "max":
		return AttributeMax
	case "crit":
		return AttributeCrit
	case "alarm":
		return AttributeAlarm
	case "enable":
		return AttributeEnable
	case "target":
		return AttributeTarget
	case "fault":
		return AttributeFault
	case "beep":
		return AttributeBeep
	case "offset":
		return AttributeOffset
	case "type":
		return AttributeType
	default:
		return AttributeInput
	}
}

// BuildSensorFilename constructs a sensor filename from type, index, and attribute.
func BuildSensorFilename(sensorType SensorType, index int, attribute SensorAttribute) string {
	if sensorType == SensorTypeGeneric {
		return ""
	}

	prefix := sensorType.Prefix()
	// PWM value file is "pwmN" (no "_input" suffix)
	if sensorType == SensorTypePWM && attribute == AttributeInput {
		return fmt.Sprintf("%s%d", prefix, index)
	}

	return fmt.Sprintf("%s%d_%s", prefix, index, attribute.String())
}

// ConvertTemperature converts temperature between different units.
func ConvertTemperature(value float64, fromUnit, toUnit string) (float64, error) {
	fromUnit = strings.ToLower(fromUnit)
	toUnit = strings.ToLower(toUnit)

	if fromUnit == toUnit {
		return value, nil
	}

	var celsius float64

	switch fromUnit {
	case "c", "celsius":
		celsius = value
	case "f", "fahrenheit":
		celsius = (value - 32.0) * 5.0 / 9.0
	case "k", "kelvin":
		celsius = value - 273.15
	default:
		return 0, fmt.Errorf("%w: unsupported temperature unit: %s", ErrInvalidValue, fromUnit)
	}

	switch toUnit {
	case "c", "celsius":
		return celsius, nil
	case "f", "fahrenheit":
		return celsius*9.0/5.0 + 32.0, nil
	case "k", "kelvin":
		return celsius + 273.15, nil
	default:
		return 0, fmt.Errorf("%w: unsupported temperature unit: %s", ErrInvalidValue, toUnit)
	}
}

// FormatDuration formats a duration in a human-readable way.
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.1fμs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1000000.0)
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else { //nolint:revive
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

// IsFileReadable checks if a file exists and is readable.
func IsFileReadable(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

// IsFileWritable checks if a file exists and is writable.
func IsFileWritable(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

// SafeReadFile reads a file with error handling and validation.
func SafeReadFile(path string, maxSize int64) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: file path cannot be empty", ErrInvalidPath)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrPathNotFound, path)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("%w: %s", ErrPermissionDenied, path)
		}
		return nil, fmt.Errorf("%w: %w", ErrFileSystemError, err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("%w: path is a directory: %s", ErrInvalidPath, path)
	}

	if maxSize > 0 && info.Size() > maxSize {
		return nil, fmt.Errorf("%w: file too large (%d bytes, max %d)", ErrInvalidValue, info.Size(), maxSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("%w: %w", ErrPermissionDenied, err)
		}
		return nil, fmt.Errorf("%w: %w", ErrReadFailure, err)
	}

	return data, nil
}

// SafeWriteFile writes data to a file with error handling and validation.
func SafeWriteFile(path string, data []byte, perm os.FileMode) error {
	if path == "" {
		return fmt.Errorf("%w: file path cannot be empty", ErrInvalidPath)
	}

	if len(data) == 0 {
		return fmt.Errorf("%w: cannot write empty data", ErrInvalidValue)
	}

	err := os.WriteFile(path, data, perm)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrPathNotFound, path)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("%w: %w", ErrPermissionDenied, err)
		}
		return fmt.Errorf("%w: %w", ErrWriteFailure, err)
	}

	return nil
}

// CleanPath cleans and validates a file path.
func CleanPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	cleaned := filepath.Clean(path)

	if !filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("%w: path must be absolute", ErrInvalidPath)
	}

	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("%w: path contains parent directory references", ErrInvalidPath)
	}

	return cleaned, nil
}

// FindHwmonDevices returns a list of all hwmon device directories.
func FindHwmonDevices(basePath string) ([]string, error) {
	if err := ValidateHwmonPath(basePath); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read hwmon directory: %w", ErrDiscoveryFailure, err)
	}

	var devices []string
	for _, entry := range entries {
		if entry.IsDir() && IsHwmonDevice(entry.Name()) {
			devices = append(devices, entry.Name())
		}
	}

	return devices, nil
}

// GetDeviceName reads the device name from a hwmon device directory.
func GetDeviceName(devicePath string) (string, error) {
	nameFile := filepath.Join(devicePath, "name")
	data, err := SafeReadFile(nameFile, 256)
	if err != nil {
		return "", fmt.Errorf("failed to read device name: %w", err)
	}

	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", fmt.Errorf("%w: empty device name", ErrInvalidValue)
	}

	return name, nil
}

// WrapError wraps an error with additional context.
func WrapError(err error, operation string, path string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s failed for %s: %w", operation, path, err)
}

// RetryOperation executes an operation with retry logic.
func RetryOperation(operation func() error, maxRetries int, delay time.Duration) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delay)
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("%w: %w", ErrRetryExhausted, lastErr)
}

// ValidateValueRange checks if a value is within the specified range.
func ValidateValueRange(value float64, minVal, maxVal *float64) error {
	if minVal != nil && value < *minVal {
		return fmt.Errorf("%w: value %.3f below minimum %.3f", ErrValueOutOfRange, value, *minVal)
	}

	if maxVal != nil && value > *maxVal {
		return fmt.Errorf("%w: value %.3f above maximum %.3f", ErrValueOutOfRange, value, *maxVal)
	}

	return nil
}

// ParseFloat64 parses a string to float64 with error handling.
func ParseFloat64(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("%w: empty string", ErrValueParseFailure)
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrValueParseFailure, err)
	}

	return value, nil
}

// ParseInt64 parses a string to int64 with error handling.
func ParseInt64(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("%w: empty string", ErrValueParseFailure)
	}

	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrValueParseFailure, err)
	}

	return value, nil
}

// FormatFloat64 formats a float64 value with specified precision.
func FormatFloat64(value float64, precision int) string {
	return fmt.Sprintf("%.*f", precision, value)
}

// GetSensorTypeFromPrefix determines the sensor type from a filename prefix.
func GetSensorTypeFromPrefix(prefix string) (SensorType, error) {
	switch prefix {
	case "temp":
		return SensorTypeTemperature, nil
	case "in":
		return SensorTypeVoltage, nil
	case "fan":
		return SensorTypeFan, nil
	case "power":
		return SensorTypePower, nil
	case "curr":
		return SensorTypeCurrent, nil
	case "humidity":
		return SensorTypeHumidity, nil
	case "pressure":
		return SensorTypePressure, nil
	case "pwm":
		return SensorTypePWM, nil
	default:
		return SensorTypeGeneric, fmt.Errorf("%w: unknown sensor type prefix: %s", ErrInvalidSensorType, prefix)
	}
}

// NormalizeDevicePath ensures a device path has the correct format.
func NormalizeDevicePath(basePath, device string) (string, error) {
	if basePath == "" {
		return "", fmt.Errorf("%w: base path cannot be empty", ErrInvalidPath)
	}

	if device == "" {
		return "", fmt.Errorf("%w: device cannot be empty", ErrInvalidConfig)
	}

	cleanBase, err := CleanPath(basePath)
	if err != nil {
		return "", fmt.Errorf("invalid base path: %w", err)
	}

	if IsHwmonDevice(device) {
		return filepath.Join(cleanBase, device), nil
	}

	devices, err := FindHwmonDevices(cleanBase)
	if err != nil {
		return "", err
	}

	for _, hwmonDevice := range devices {
		devicePath := filepath.Join(cleanBase, hwmonDevice)
		name, err := GetDeviceName(devicePath)
		if err != nil {
			continue
		}
		if name == device {
			return devicePath, nil
		}
	}

	return "", fmt.Errorf("%w: device %s", ErrDeviceNotFound, device)
}
