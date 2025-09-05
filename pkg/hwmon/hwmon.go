// SPDX-License-Identifier: BSD-3-Clause

package hwmon

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Sensor represents a hardware monitoring sensor with configuration and state management.
type Sensor struct {
	config      *Config
	discoverer  *Discoverer
	sensorPath  string
	devicePath  string
	sensorInfo  *SensorInfo
	cache       *sensorCache
	mu          sync.RWMutex
	initialized bool
	lastError   error
	lastErrorAt time.Time
}

// sensorCache holds cached sensor values with TTL support.
type sensorCache struct {
	value     Value
	timestamp time.Time
	ttl       time.Duration
	mu        sync.RWMutex
}

// NewSensor creates a new sensor instance with the provided configuration.
func NewSensor(config *Config) (*Sensor, error) {
	if config == nil {
		return nil, fmt.Errorf("%w: config cannot be nil", ErrInvalidConfig)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}

	sensor := &Sensor{
		config:     config.Clone(),
		discoverer: NewDiscoverer(WithDiscoveryPath(config.BasePath)),
	}

	if config.CachingEnabled {
		sensor.cache = &sensorCache{
			ttl: config.CacheTTL,
		}
	}

	return sensor, nil
}

// Initialize discovers and validates the sensor path.
func (s *Sensor) Initialize(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	if s.config.CustomPath != "" {
		s.sensorPath = s.config.CustomPath
		s.devicePath = filepath.Dir(s.config.CustomPath)
	} else {
		if err := s.discoverSensorPath(ctx); err != nil {
			s.lastError = err
			s.lastErrorAt = time.Now()
			return err
		}
	}

	if err := s.validateSensorPath(); err != nil {
		s.lastError = err
		s.lastErrorAt = time.Now()
		return err
	}

	s.initialized = true
	s.lastError = nil
	return nil
}

// ReadValue reads the current sensor value.
func (s *Sensor) ReadValue(ctx context.Context) (Value, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	if !s.initialized {
		if err := s.Initialize(ctx); err != nil {
			return nil, err
		}
	}

	if s.config.CachingEnabled && s.cache != nil {
		if value := s.getCachedValue(); value != nil {
			return value, nil
		}
	}

	value, err := s.readValueWithRetry(ctx)
	if err != nil {
		s.mu.Lock()
		s.lastError = err
		s.lastErrorAt = time.Now()
		s.mu.Unlock()
		return nil, err
	}

	if s.config.ValidationEnabled {
		if err := s.validateValue(value); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrValidationFailed, err)
		}
	}

	if s.config.CachingEnabled && s.cache != nil {
		s.setCachedValue(value)
	}

	s.mu.Lock()
	s.lastError = nil
	s.mu.Unlock()

	return value, nil
}

// WriteValue writes a value to the sensor (if writable).
func (s *Sensor) WriteValue(ctx context.Context, value Value) error {
	if ctx == nil {
		return ErrNilContext
	}

	if !s.initialized {
		if err := s.Initialize(ctx); err != nil {
			return err
		}
	}

	if !s.config.Writable {
		return fmt.Errorf("%w: sensor is configured as read-only", ErrReadOnlySensor)
	}

	if s.config.ValidationEnabled {
		if err := s.validateValue(value); err != nil {
			return fmt.Errorf("%w: %w", ErrValidationFailed, err)
		}
	}

	err := s.writeValueWithRetry(ctx, value)
	if err != nil {
		s.mu.Lock()
		s.lastError = err
		s.lastErrorAt = time.Now()
		s.mu.Unlock()
		return err
	}

	if s.config.CachingEnabled && s.cache != nil {
		s.cache.mu.Lock()
		s.cache.value = nil
		s.cache.mu.Unlock()
	}

	s.mu.Lock()
	s.lastError = nil
	s.mu.Unlock()

	return nil
}

// ReadRawValue reads the raw string value from sysfs.
func (s *Sensor) ReadRawValue(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrNilContext
	}

	if !s.initialized {
		if err := s.Initialize(ctx); err != nil {
			return "", err
		}
	}

	return s.readRawValueWithRetry(ctx)
}

// WriteRawValue writes a raw string value to sysfs.
func (s *Sensor) WriteRawValue(ctx context.Context, value string) error {
	if ctx == nil {
		return ErrNilContext
	}

	if !s.initialized {
		if err := s.Initialize(ctx); err != nil {
			return err
		}
	}

	if !s.config.Writable {
		return fmt.Errorf("%w: sensor is configured as read-only", ErrReadOnlySensor)
	}

	return s.writeRawValueWithRetry(ctx, value)
}

// IsAvailable checks if the sensor is currently available.
func (s *Sensor) IsAvailable(ctx context.Context) bool {
	if !s.initialized {
		return s.Initialize(ctx) == nil
	}

	s.mu.RLock()
	path := s.sensorPath
	s.mu.RUnlock()

	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	return err == nil
}

// GetInfo returns sensor information.
func (s *Sensor) GetInfo() *SensorInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sensorInfo
}

// GetPath returns the sensor sysfs path.
func (s *Sensor) GetPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sensorPath
}

// GetDevicePath returns the device sysfs path.
func (s *Sensor) GetDevicePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.devicePath
}

// GetConfig returns a copy of the sensor configuration.
func (s *Sensor) GetConfig() *Config {
	return s.config.Clone()
}

// IsWritable returns true if the sensor supports writing.
func (s *Sensor) IsWritable() bool {
	return s.config.Writable
}

// Type returns the sensor type.
func (s *Sensor) Type() SensorType {
	return s.config.SensorType
}

// Label returns the sensor label.
func (s *Sensor) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sensorInfo != nil && s.sensorInfo.Label != "" {
		return s.sensorInfo.Label
	}
	return s.config.SensorLabel
}

// Name returns the sensor name.
func (s *Sensor) Name() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sensorInfo != nil {
		return s.sensorInfo.Name
	}
	return fmt.Sprintf("%s%d", s.config.SensorType.Prefix(), s.config.SensorIndex)
}

// LastError returns the last error that occurred.
func (s *Sensor) LastError() (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastErrorAt, s.lastError
}

// discoverSensorPath discovers the sensor path using the configured parameters.
func (s *Sensor) discoverSensorPath(ctx context.Context) error {
	device, err := s.discoverer.FindDevice(ctx, s.config.Device)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDeviceNotFound, err)
	}

	s.devicePath = device.Path

	var sensorInfo *SensorInfo
	if s.config.UseIndex || s.config.SensorLabel == "" {
		sensorInfo, err = device.GetSensorByTypeAndIndex(ctx, s.config.SensorType, s.config.SensorIndex)
	} else {
		sensorInfo, err = device.GetSensorByLabel(ctx, s.config.SensorLabel)
	}

	if err != nil {
		return fmt.Errorf("%w: %w", ErrSensorNotFound, err)
	}

	attributePath, err := sensorInfo.GetAttributePath(s.config.Attribute)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAttributeNotSupported, err)
	}

	s.sensorPath = attributePath
	s.sensorInfo = sensorInfo
	return nil
}

// validateSensorPath validates that the sensor path exists and is accessible.
func (s *Sensor) validateSensorPath() error {
	if s.sensorPath == "" {
		return fmt.Errorf("%w: sensor path is empty", ErrInvalidPath)
	}

	info, err := os.Stat(s.sensorPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrPathNotFound, s.sensorPath)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("%w: %s", ErrPermissionDenied, s.sensorPath)
		}
		return fmt.Errorf("%w: %w", ErrFileSystemError, err)
	}

	if info.IsDir() {
		return fmt.Errorf("%w: path is a directory: %s", ErrInvalidPath, s.sensorPath)
	}

	if s.config.Writable {
		file, err := os.OpenFile(s.sensorPath, os.O_WRONLY, 0)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%w: cannot write to %s", ErrPermissionDenied, s.sensorPath)
			}
			return fmt.Errorf("%w: cannot open for writing: %s", ErrFileSystemError, s.sensorPath)
		}
		_ = file.Close()
	}

	return nil
}

// readValueWithRetry reads a sensor value with retry logic.
func (s *Sensor) readValueWithRetry(ctx context.Context) (Value, error) {
	var lastErr error

	for attempt := 0; attempt <= s.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
			case <-time.After(s.config.RetryDelay):
			}
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)

		rawValue, err := s.readRawValue(timeoutCtx)
		cancel()

		if err != nil {
			lastErr = err
			continue
		}

		value, err := ParseValue(rawValue, s.config.SensorType)
		if err != nil {
			lastErr = fmt.Errorf("%w: %w", ErrValueParseFailure, err)
			continue
		}

		return value, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrRetryExhausted, lastErr)
	}

	return nil, fmt.Errorf("%w: unknown error after %d attempts", ErrRetryExhausted, s.config.RetryCount+1)
}

// writeValueWithRetry writes a sensor value with retry logic.
func (s *Sensor) writeValueWithRetry(ctx context.Context, value Value) error {
	rawValue := FormatValue(value)
	return s.writeRawValueWithRetry(ctx, rawValue)
}

// readRawValueWithRetry reads a raw sensor value with retry logic.
func (s *Sensor) readRawValueWithRetry(ctx context.Context) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= s.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
			case <-time.After(s.config.RetryDelay):
			}
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
		value, err := s.readRawValue(timeoutCtx)
		cancel()

		if err == nil {
			return value, nil
		}

		lastErr = err
	}

	return "", fmt.Errorf("%w: %w", ErrRetryExhausted, lastErr)
}

// writeRawValueWithRetry writes a raw sensor value with retry logic.
func (s *Sensor) writeRawValueWithRetry(ctx context.Context, value string) error {
	var lastErr error

	for attempt := 0; attempt <= s.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
			case <-time.After(s.config.RetryDelay):
			}
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
		err := s.writeRawValue(timeoutCtx, value)
		cancel()

		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("%w: %w", ErrRetryExhausted, lastErr)
}

// readRawValue reads the raw value from the sensor file.
func (s *Sensor) readRawValue(ctx context.Context) (string, error) {
	done := make(chan struct{})
	var result string
	var err error

	go func() {
		defer close(done)

		data, readErr := os.ReadFile(s.sensorPath)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				err = fmt.Errorf("%w: sensor file not found", ErrDeviceUnavailable)
			} else if os.IsPermission(readErr) {
				err = fmt.Errorf("%w: %w", ErrPermissionDenied, readErr)
			} else {
				err = fmt.Errorf("%w: %w", ErrReadFailure, readErr)
			}
			return
		}

		result = strings.TrimSpace(string(data))
		if result == "" {
			err = fmt.Errorf("%w: empty value read from sensor", ErrInvalidValue)
			return
		}
	}()

	select {
	case <-done:
		return result, err
	case <-ctx.Done():
		return "", fmt.Errorf("%w: %w", ErrReadTimeout, ctx.Err())
	}
}

// writeRawValue writes a raw value to the sensor file.
func (s *Sensor) writeRawValue(ctx context.Context, value string) error {
	done := make(chan error, 1)

	go func() {
		err := os.WriteFile(s.sensorPath, []byte(value), 0o600)
		if err != nil {
			if os.IsNotExist(err) {
				done <- fmt.Errorf("%w: sensor file not found", ErrDeviceUnavailable)
			} else if os.IsPermission(err) {
				done <- fmt.Errorf("%w: %w", ErrPermissionDenied, err)
			} else {
				done <- fmt.Errorf("%w: %w", ErrWriteFailure, err)
			}
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%w: %w", ErrWriteTimeout, ctx.Err())
	}
}

// validateValue validates a sensor value against configured constraints.
func (s *Sensor) validateValue(value Value) error {
	if !value.IsValid() {
		return fmt.Errorf("%w: value failed basic validation", ErrInvalidValue)
	}

	if s.config.HasValueRange() {
		floatValue := value.Float()

		if s.config.MinValue != nil && floatValue < *s.config.MinValue {
			return fmt.Errorf("%w: value %.3f below minimum %.3f", ErrValueOutOfRange, floatValue, *s.config.MinValue)
		}

		if s.config.MaxValue != nil && floatValue > *s.config.MaxValue {
			return fmt.Errorf("%w: value %.3f above maximum %.3f", ErrValueOutOfRange, floatValue, *s.config.MaxValue)
		}
	}

	return nil
}

// getCachedValue retrieves a cached value if it's still valid.
func (s *Sensor) getCachedValue() Value {
	if s.cache == nil {
		return nil
	}

	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()

	if s.cache.value == nil {
		return nil
	}

	if time.Since(s.cache.timestamp) > s.cache.ttl {
		return nil
	}

	return s.cache.value
}

// setCachedValue stores a value in the cache.
func (s *Sensor) setCachedValue(value Value) {
	if s.cache == nil {
		return
	}

	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	s.cache.value = value
	s.cache.timestamp = time.Now()
}

// String returns a string representation of the sensor.
func (s *Sensor) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	label := s.Label()
	if label == "" {
		label = s.Name()
	}

	return fmt.Sprintf("Sensor{device=%s, label=%s, type=%s, path=%s}",
		s.config.Device, label, s.config.SensorType.String(), s.sensorPath)
}

// SensorManager manages multiple sensors and provides batch operations.
type SensorManager struct {
	sensors map[string]*Sensor
	mu      sync.RWMutex
}

// NewSensorManager creates a new sensor manager.
func NewSensorManager() *SensorManager {
	return &SensorManager{
		sensors: make(map[string]*Sensor),
	}
}

// AddSensor adds a sensor to the manager.
func (sm *SensorManager) AddSensor(name string, sensor *Sensor) error {
	if name == "" {
		return fmt.Errorf("%w: sensor name cannot be empty", ErrInvalidConfig)
	}

	if sensor == nil {
		return fmt.Errorf("%w: sensor cannot be nil", ErrInvalidConfig)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sensors[name]; exists {
		return fmt.Errorf("sensor %s already exists", name)
	}

	sm.sensors[name] = sensor
	return nil
}

// RemoveSensor removes a sensor from the manager.
func (sm *SensorManager) RemoveSensor(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sensors[name]; !exists {
		return fmt.Errorf("%w: sensor %s", ErrSensorNotFound, name)
	}

	delete(sm.sensors, name)
	return nil
}

// GetSensor retrieves a sensor by name.
func (sm *SensorManager) GetSensor(name string) (*Sensor, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sensor, exists := sm.sensors[name]
	if !exists {
		return nil, fmt.Errorf("%w: sensor %s", ErrSensorNotFound, name)
	}

	return sensor, nil
}

// ListSensors returns all sensor names.
func (sm *SensorManager) ListSensors() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.sensors))
	for name := range sm.sensors {
		names = append(names, name)
	}

	return names
}

// ReadAllSensors reads values from all sensors.
func (sm *SensorManager) ReadAllSensors(ctx context.Context) (map[string]Value, map[string]error) {
	sm.mu.RLock()
	sensors := make(map[string]*Sensor, len(sm.sensors))
	maps.Copy(sensors, sm.sensors)
	sm.mu.RUnlock()

	values := make(map[string]Value)
	errors := make(map[string]error)

	for name, sensor := range sensors {
		value, err := sensor.ReadValue(ctx)
		if err != nil {
			errors[name] = err
		} else {
			values[name] = value
		}
	}

	return values, errors
}

// InitializeAllSensors initializes all sensors in the manager.
func (sm *SensorManager) InitializeAllSensors(ctx context.Context) map[string]error {
	sm.mu.RLock()
	sensors := make(map[string]*Sensor, len(sm.sensors))
	maps.Copy(sensors, sm.sensors)
	sm.mu.RUnlock()

	errors := make(map[string]error)

	for name, sensor := range sensors {
		if err := sensor.Initialize(ctx); err != nil {
			errors[name] = err
		}
	}

	return errors
}

// OfNameAndLabel is a convenience function that creates a sensor
// for the specified device name and sensor label (for backward compatibility).
func OfNameAndLabel(name string, label string) (*Sensor, error) {
	config := NewConfig(
		WithDevice(name),
		WithSensorLabel(label),
		WithSensorType(SensorTypeTemperature),
		WithAttribute(AttributeInput),
	)

	sensor, err := NewSensor(config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sensor.Initialize(ctx); err != nil {
		return nil, err
	}

	return sensor, nil
}

// ReadWithCtx reads a value from a file path with context support (for backward compatibility).
func ReadWithCtx(ctx context.Context, path string) (string, error) {
	if ctx == nil {
		return "", ErrNilContext
	}

	done := make(chan struct{})
	var result string
	var err error

	go func() {
		defer close(done)
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			err = readErr
			return
		}
		result = strings.TrimSpace(string(data))
	}()

	select {
	case <-done:
		return result, err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
