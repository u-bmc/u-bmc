// SPDX-License-Identifier: BSD-3-Clause

package sensormon

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	v1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/hwmon"
)

// HwmonBackend implements sensor reading from hwmon files.
type HwmonBackend struct {
	config      *HwmonSensorConfig
	devicePath  string
	attrPath    string
	scaleFactor int
	mu          sync.RWMutex
}

// NewHwmonBackend creates a new hwmon backend instance.
func NewHwmonBackend(config interface{}) (SensorBackend, error) {
	hwmonConfig, ok := config.(*HwmonSensorConfig)
	if !ok {
		return nil, fmt.Errorf("invalid hwmon config type")
	}

	backend := &HwmonBackend{
		config:      hwmonConfig,
		scaleFactor: hwmonConfig.ScaleFactor,
	}

	if backend.scaleFactor == 0 {
		backend.scaleFactor = 1
	}

	return backend, nil
}

// Configure sets up the hwmon backend with device discovery.
func (h *HwmonBackend) Configure(config interface{}) error {
	hwmonConfig, ok := config.(*HwmonSensorConfig)
	if !ok {
		return fmt.Errorf("invalid hwmon config type")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.config = hwmonConfig

	// If device path is provided directly, use it
	if hwmonConfig.DevicePath != "" {
		h.devicePath = hwmonConfig.DevicePath
		h.attrPath = filepath.Join(h.devicePath, hwmonConfig.AttributeName)
		return h.validatePaths()
	}

	// Otherwise, try to discover the device using the match pattern
	if hwmonConfig.MatchPattern != "" {
		devicePath, err := h.discoverDevice(hwmonConfig.MatchPattern, hwmonConfig.RequiredFiles)
		if err != nil {
			return fmt.Errorf("failed to discover hwmon device: %w", err)
		}
		h.devicePath = devicePath
		h.attrPath = filepath.Join(h.devicePath, hwmonConfig.AttributeName)
		return h.validatePaths()
	}

	return fmt.Errorf("hwmon config must specify either device_path or match_pattern")
}

// discoverDevice discovers hwmon device by pattern matching.
func (h *HwmonBackend) discoverDevice(pattern string, requiredFiles []string) (string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid match pattern: %w", err)
	}

	hwmonPath := "/sys/class/hwmon"
	entries, err := os.ReadDir(hwmonPath)
	if err != nil {
		return "", fmt.Errorf("failed to read hwmon directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		devicePath := filepath.Join(hwmonPath, entry.Name())

		// Check if device name matches pattern
		namePath := filepath.Join(devicePath, "name")
		if nameData, err := os.ReadFile(namePath); err == nil {
			deviceName := string(nameData)
			if regex.MatchString(deviceName) {
				// Check if all required files exist
				if h.checkRequiredFiles(devicePath, requiredFiles) {
					return devicePath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no matching hwmon device found for pattern: %s", pattern)
}

// checkRequiredFiles verifies that all required files exist in the device path.
func (h *HwmonBackend) checkRequiredFiles(devicePath string, requiredFiles []string) bool {
	for _, file := range requiredFiles {
		filePath := filepath.Join(devicePath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// validatePaths ensures the configured paths exist and are accessible.
func (h *HwmonBackend) validatePaths() error {
	if _, err := os.Stat(h.devicePath); os.IsNotExist(err) {
		return fmt.Errorf("hwmon device path does not exist: %s", h.devicePath)
	}

	if _, err := os.Stat(h.attrPath); os.IsNotExist(err) {
		return fmt.Errorf("hwmon attribute path does not exist: %s", h.attrPath)
	}

	return nil
}

// ReadValue reads the current sensor value from hwmon.
func (h *HwmonBackend) ReadValue() (interface{}, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.attrPath == "" {
		return nil, fmt.Errorf("hwmon backend not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rawValue, err := hwmon.ReadIntCtx(ctx, h.attrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read hwmon value: %w", err)
	}

	// Apply scaling factor
	value := float64(rawValue) / float64(h.scaleFactor)
	return value, nil
}

// GetStatus returns the current sensor status.
func (h *HwmonBackend) GetStatus() (v1alpha1.SensorStatus, error) {
	_, err := h.ReadValue()
	if err != nil {
		return v1alpha1.SensorStatus_SENSOR_STATUS_ERROR, err
	}
	return v1alpha1.SensorStatus_SENSOR_STATUS_ENABLED, nil
}

// Close cleans up the hwmon backend.
func (h *HwmonBackend) Close() error {
	// Nothing to close for hwmon
	return nil
}

// GPIOBackend implements sensor reading from GPIO.
type GPIOBackend struct {
	config *GPIOSensorConfig
	mu     sync.RWMutex
}

// NewGPIOBackend creates a new GPIO backend instance.
func NewGPIOBackend(config interface{}) (SensorBackend, error) {
	gpioConfig, ok := config.(*GPIOSensorConfig)
	if !ok {
		return nil, fmt.Errorf("invalid gpio config type")
	}

	backend := &GPIOBackend{
		config: gpioConfig,
	}

	return backend, nil
}

// Configure sets up the GPIO backend.
func (g *GPIOBackend) Configure(config interface{}) error {
	gpioConfig, ok := config.(*GPIOSensorConfig)
	if !ok {
		return fmt.Errorf("invalid gpio config type")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.config = gpioConfig

	// Validate GPIO chip path
	if _, err := os.Stat(gpioConfig.ChipPath); os.IsNotExist(err) {
		return fmt.Errorf("gpio chip path does not exist: %s", gpioConfig.ChipPath)
	}

	return nil
}

// ReadValue reads the current sensor value from GPIO.
func (g *GPIOBackend) ReadValue() (interface{}, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.config == nil {
		return nil, fmt.Errorf("gpio backend not configured")
	}

	// For now, return a placeholder discrete reading
	// In a real implementation, this would use the GPIO library
	state := "active"
	if g.config.ValueMapping != nil {
		if mappedState, exists := g.config.ValueMapping["1"]; exists {
			state = mappedState
		}
	}

	return state, nil
}

// GetStatus returns the current sensor status.
func (g *GPIOBackend) GetStatus() (v1alpha1.SensorStatus, error) {
	_, err := g.ReadValue()
	if err != nil {
		return v1alpha1.SensorStatus_SENSOR_STATUS_ERROR, err
	}
	return v1alpha1.SensorStatus_SENSOR_STATUS_ENABLED, nil
}

// Close cleans up the GPIO backend.
func (g *GPIOBackend) Close() error {
	// Nothing to close for GPIO
	return nil
}

// MockBackend implements a mock sensor for testing.
type MockBackend struct {
	config    *MockSensorConfig
	startTime time.Time
	stepCount int
	mu        sync.RWMutex
	rand      *rand.Rand
}

// NewMockBackend creates a new mock backend instance.
func NewMockBackend(config interface{}) (SensorBackend, error) {
	mockConfig, ok := config.(*MockSensorConfig)
	if !ok {
		return nil, fmt.Errorf("invalid mock config type")
	}

	backend := &MockBackend{
		config:    mockConfig,
		startTime: time.Now(),
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return backend, nil
}

// Configure sets up the mock backend.
func (m *MockBackend) Configure(config interface{}) error {
	mockConfig, ok := config.(*MockSensorConfig)
	if !ok {
		return fmt.Errorf("invalid mock config type")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = mockConfig
	m.startTime = time.Now()
	m.stepCount = 0

	return nil
}

// ReadValue generates a mock sensor value based on the configured behavior.
func (m *MockBackend) ReadValue() (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return nil, fmt.Errorf("mock backend not configured")
	}

	// Simulate failure if configured
	if m.config.FailureRate > 0 && m.rand.Float64() < m.config.FailureRate {
		return nil, fmt.Errorf("simulated sensor read failure")
	}

	var value float64

	switch m.config.Behavior {
	case MockBehaviorFixed:
		value = m.config.BaseValue

	case MockBehaviorRandomize:
		variance := m.config.Variance
		if variance == 0 {
			variance = m.config.BaseValue * 0.1 // Default 10% variance
		}
		value = m.config.BaseValue + (m.rand.Float64()-0.5)*2*variance

	case MockBehaviorSine:
		period := m.config.Period
		if period == 0 {
			period = 60 * time.Second // Default 1 minute period
		}
		elapsed := time.Since(m.startTime).Seconds()
		periodSeconds := period.Seconds()
		amplitude := m.config.Variance
		if amplitude == 0 {
			amplitude = m.config.BaseValue * 0.2 // Default 20% amplitude
		}
		value = m.config.BaseValue + amplitude*math.Sin(2*math.Pi*elapsed/periodSeconds)

	case MockBehaviorStep:
		stepSize := m.config.StepSize
		if stepSize == 0 {
			stepSize = m.config.BaseValue * 0.05 // Default 5% step
		}
		value = m.config.BaseValue + float64(m.stepCount)*stepSize
		m.stepCount++

		// Reset step count if we hit boundaries
		if m.config.MaxValue > 0 && value > m.config.MaxValue {
			m.stepCount = 0
		}
		if m.config.MinValue > 0 && value < m.config.MinValue {
			m.stepCount = 0
		}

	default:
		value = m.config.BaseValue
	}

	// Apply min/max constraints
	if m.config.MinValue > 0 && value < m.config.MinValue {
		value = m.config.MinValue
	}
	if m.config.MaxValue > 0 && value > m.config.MaxValue {
		value = m.config.MaxValue
	}

	return value, nil
}

// GetStatus returns the current sensor status.
func (m *MockBackend) GetStatus() (v1alpha1.SensorStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return v1alpha1.SensorStatus_SENSOR_STATUS_ERROR, fmt.Errorf("mock backend not configured")
	}

	// Simulate status changes for testing
	if m.config.FailureRate > 0 && m.rand.Float64() < m.config.FailureRate*0.1 {
		return v1alpha1.SensorStatus_SENSOR_STATUS_ERROR, nil
	}

	return v1alpha1.SensorStatus_SENSOR_STATUS_ENABLED, nil
}

// Close cleans up the mock backend.
func (m *MockBackend) Close() error {
	// Nothing to close for mock
	return nil
}

// Backend factory functions
var defaultBackendFactories = map[string]SensorBackendFactory{
	string(BackendTypeHwmon): NewHwmonBackend,
	string(BackendTypeGPIO):  NewGPIOBackend,
	string(BackendTypeMock):  NewMockBackend,
}

// GetBackendFactory returns a backend factory by name.
func GetBackendFactory(name string) (SensorBackendFactory, bool) {
	factory, exists := defaultBackendFactories[name]
	return factory, exists
}

// RegisterBackendFactory registers a custom backend factory.
func RegisterBackendFactory(name string, factory SensorBackendFactory) {
	defaultBackendFactories[name] = factory
}

// Helper functions for creating backend configurations

// NewHwmonSensorConfig creates a new hwmon sensor configuration.
func NewHwmonSensorConfig(devicePath, attributeName string) *HwmonSensorConfig {
	return &HwmonSensorConfig{
		DevicePath:    devicePath,
		AttributeName: attributeName,
		ScaleFactor:   1000, // Default scaling for most hwmon sensors
	}
}

// NewHwmonSensorConfigWithPattern creates a new hwmon sensor configuration with pattern matching.
func NewHwmonSensorConfigWithPattern(pattern, attributeName string, requiredFiles ...string) *HwmonSensorConfig {
	return &HwmonSensorConfig{
		MatchPattern:  pattern,
		AttributeName: attributeName,
		RequiredFiles: requiredFiles,
		ScaleFactor:   1000,
	}
}

// NewGPIOSensorConfig creates a new GPIO sensor configuration.
func NewGPIOSensorConfig(chipPath string, line int, activeState string) *GPIOSensorConfig {
	return &GPIOSensorConfig{
		ChipPath:    chipPath,
		Line:        line,
		ActiveState: activeState,
	}
}

// NewMockSensorConfig creates a new mock sensor configuration.
func NewMockSensorConfig(behavior MockSensorBehavior, baseValue float64) *MockSensorConfig {
	return &MockSensorConfig{
		Behavior:  behavior,
		BaseValue: baseValue,
		Variance:  baseValue * 0.1, // Default 10% variance
	}
}

// NewMockTemperatureSensor creates a mock temperature sensor configuration.
func NewMockTemperatureSensor(baseTemp float64) *MockSensorConfig {
	return &MockSensorConfig{
		Behavior:  MockBehaviorRandomize,
		BaseValue: baseTemp,
		Variance:  5.0,           // ±5°C variance
		MinValue:  baseTemp - 10, // Minimum temperature
		MaxValue:  baseTemp + 15, // Maximum temperature
	}
}

// NewMockFanSensor creates a mock fan sensor configuration.
func NewMockFanSensor(baseRPM float64) *MockSensorConfig {
	return &MockSensorConfig{
		Behavior:  MockBehaviorSine,
		BaseValue: baseRPM,
		Variance:  baseRPM * 0.1,    // ±10% variance
		Period:    30 * time.Second, // 30 second period
		MinValue:  baseRPM * 0.5,    // Minimum 50% of base
		MaxValue:  baseRPM * 1.2,    // Maximum 120% of base
	}
}

// NewMockVoltageSensor creates a mock voltage sensor configuration.
func NewMockVoltageSensor(baseVoltage float64) *MockSensorConfig {
	return &MockSensorConfig{
		Behavior:  MockBehaviorRandomize,
		BaseValue: baseVoltage,
		Variance:  baseVoltage * 0.02, // ±2% variance
		MinValue:  baseVoltage * 0.9,  // Minimum 90% of base
		MaxValue:  baseVoltage * 1.1,  // Maximum 110% of base
	}
}

// NewMockPowerSensor creates a mock power sensor configuration.
func NewMockPowerSensor(basePower float64) *MockSensorConfig {
	return &MockSensorConfig{
		Behavior:  MockBehaviorStep,
		BaseValue: basePower,
		StepSize:  basePower * 0.05, // 5% steps
		Period:    10 * time.Second, // Change every 10 seconds
		MinValue:  basePower * 0.3,  // Minimum 30% of base
		MaxValue:  basePower * 1.5,  // Maximum 150% of base
	}
}
