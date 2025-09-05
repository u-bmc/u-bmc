// SPDX-License-Identifier: BSD-3-Clause

package hwmon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Device represents a hwmon device with its metadata and capabilities.
type Device struct {
	Name     string
	Path     string
	HwmonID  string
	Sensors  map[string]*SensorInfo
	mu       sync.RWMutex
	lastScan time.Time
}

// SensorInfo contains metadata about a discovered sensor.
type SensorInfo struct {
	Name       string
	Label      string
	Index      int
	Type       SensorType
	Attributes map[SensorAttribute]string
	Writable   bool
	DevicePath string
}

// Discoverer handles discovery of hwmon devices and sensors.
type Discoverer struct {
	basePath      string
	timeout       time.Duration
	cacheEnabled  bool
	cacheTTL      time.Duration
	deviceCache   map[string]*Device
	lastDiscovery time.Time
	mu            sync.RWMutex
}

// DiscoveryConfig holds configuration for the discoverer.
type DiscoveryConfig struct {
	BasePath     string
	Timeout      time.Duration
	CacheEnabled bool
	CacheTTL     time.Duration
}

// DiscoveryOption represents a configuration option for the discoverer.
type DiscoveryOption interface {
	apply(*DiscoveryConfig)
}

type discoveryBasePathOption struct {
	path string
}

func (o *discoveryBasePathOption) apply(c *DiscoveryConfig) {
	c.BasePath = o.path
}

// WithDiscoveryPath sets the base hwmon path for discovery.
func WithDiscoveryPath(path string) DiscoveryOption {
	return &discoveryBasePathOption{path: path}
}

type discoveryTimeoutOption struct {
	timeout time.Duration
}

func (o *discoveryTimeoutOption) apply(c *DiscoveryConfig) {
	c.Timeout = o.timeout
}

// WithDiscoveryTimeout sets the timeout for discovery operations.
func WithDiscoveryTimeout(timeout time.Duration) DiscoveryOption {
	return &discoveryTimeoutOption{timeout: timeout}
}

type discoveryCacheOption struct {
	enabled bool
	ttl     time.Duration
}

func (o *discoveryCacheOption) apply(c *DiscoveryConfig) {
	c.CacheEnabled = o.enabled
	c.CacheTTL = o.ttl
}

// WithDiscoveryCache enables discovery result caching with TTL.
func WithDiscoveryCache(enabled bool, ttl time.Duration) DiscoveryOption {
	return &discoveryCacheOption{enabled: enabled, ttl: ttl}
}

// NewDiscoverer creates a new hwmon discoverer with the specified options.
func NewDiscoverer(opts ...DiscoveryOption) *Discoverer {
	cfg := &DiscoveryConfig{
		BasePath:     "/sys/class/hwmon",
		Timeout:      10 * time.Second,
		CacheEnabled: true,
		CacheTTL:     30 * time.Second,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &Discoverer{
		basePath:     cfg.BasePath,
		timeout:      cfg.Timeout,
		cacheEnabled: cfg.CacheEnabled,
		cacheTTL:     cfg.CacheTTL,
		deviceCache:  make(map[string]*Device),
	}
}

// DiscoverDevices discovers all available hwmon devices.
func (d *Discoverer) DiscoverDevices(ctx context.Context) ([]*Device, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cacheEnabled && time.Since(d.lastDiscovery) < d.cacheTTL {
		devices := make([]*Device, 0, len(d.deviceCache))
		for _, device := range d.deviceCache {
			devices = append(devices, device)
		}
		return devices, nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	entries, err := os.ReadDir(d.basePath)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read hwmon directory: %w", ErrDiscoveryFailure, err)
	}

	devices := make([]*Device, 0, len(entries))
	deviceMap := make(map[string]*Device)

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "hwmon") {
			continue
		}

		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("%w: %w", ErrReadTimeout, timeoutCtx.Err())
		default:
		}

		device, err := d.discoverDevice(timeoutCtx, entry.Name())
		if err != nil {
			continue
		}

		devices = append(devices, device)
		deviceMap[device.Name] = device
	}

	if d.cacheEnabled {
		d.deviceCache = deviceMap
		d.lastDiscovery = time.Now()
	}

	sort.Slice(devices, func(i, j int) bool {
		ii, _ := ExtractHwmonNumber(devices[i].HwmonID)
		ij, _ := ExtractHwmonNumber(devices[j].HwmonID)
		if ii == ij {
			return devices[i].HwmonID < devices[j].HwmonID
		}
		return ii < ij
	})

	return devices, nil
}

// FindDevice finds a specific hwmon device by name.
func (d *Discoverer) FindDevice(ctx context.Context, deviceName string) (*Device, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	if deviceName == "" {
		return nil, fmt.Errorf("%w: device name cannot be empty", ErrInvalidConfig)
	}

	d.mu.RLock()
	if d.cacheEnabled && time.Since(d.lastDiscovery) < d.cacheTTL {
		if device, exists := d.deviceCache[deviceName]; exists {
			d.mu.RUnlock()
			return device, nil
		}
	}
	d.mu.RUnlock()

	devices, err := d.DiscoverDevices(ctx)
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		if device.Name == deviceName {
			return device, nil
		}
	}

	return nil, fmt.Errorf("%w: device %s", ErrDeviceNotFound, deviceName)
}

// DiscoverSensors discovers all sensors of a specific type across all devices.
func (d *Discoverer) DiscoverSensors(ctx context.Context, sensorType SensorType) ([]*SensorInfo, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	devices, err := d.DiscoverDevices(ctx)
	if err != nil {
		return nil, err
	}

	var sensors []*SensorInfo
	for _, device := range devices {
		deviceSensors, err := device.GetSensorsByType(ctx, sensorType)
		if err != nil {
			continue
		}
		sensors = append(sensors, deviceSensors...)
	}

	return sensors, nil
}

// DiscoverSensorsByLabel discovers sensors with specific labels across all devices.
func (d *Discoverer) DiscoverSensorsByLabel(ctx context.Context, label string) ([]*SensorInfo, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	if label == "" {
		return nil, fmt.Errorf("%w: sensor label cannot be empty", ErrInvalidConfig)
	}

	devices, err := d.DiscoverDevices(ctx)
	if err != nil {
		return nil, err
	}

	sensors := make([]*SensorInfo, 0, len(devices))
	for _, device := range devices {
		sensor, err := device.GetSensorByLabel(ctx, label)
		if err != nil {
			continue
		}
		sensors = append(sensors, sensor)
	}

	return sensors, nil
}

// discoverDevice discovers a single hwmon device and its sensors.
func (d *Discoverer) discoverDevice(ctx context.Context, hwmonID string) (*Device, error) {
	devicePath := filepath.Join(d.basePath, hwmonID)

	nameFile := filepath.Join(devicePath, "name")
	nameBytes, err := os.ReadFile(nameFile)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read device name: %w", ErrDiscoveryFailure, err)
	}

	deviceName := strings.TrimSpace(string(nameBytes))
	if deviceName == "" {
		return nil, fmt.Errorf("%w: empty device name", ErrDiscoveryFailure)
	}

	device := &Device{
		Name:     deviceName,
		Path:     devicePath,
		HwmonID:  hwmonID,
		Sensors:  make(map[string]*SensorInfo),
		lastScan: time.Now(),
	}

	if err := device.scanSensors(ctx); err != nil {
		return nil, fmt.Errorf("failed to scan sensors for device %s: %w", deviceName, err)
	}

	return device, nil
}

// GetSensors returns all sensors for the device.
func (d *Device) GetSensors(ctx context.Context) ([]*SensorInfo, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	sensors := make([]*SensorInfo, 0, len(d.Sensors))
	for _, sensor := range d.Sensors {
		sensors = append(sensors, sensor)
	}

	return sensors, nil
}

// GetSensorsByType returns all sensors of a specific type for the device.
func (d *Device) GetSensorsByType(ctx context.Context, sensorType SensorType) ([]*SensorInfo, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	var sensors []*SensorInfo
	for _, sensor := range d.Sensors {
		if sensor.Type == sensorType {
			sensors = append(sensors, sensor)
		}
	}

	return sensors, nil
}

// GetSensorByLabel finds a sensor by its label.
func (d *Device) GetSensorByLabel(ctx context.Context, label string) (*SensorInfo, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	if label == "" {
		return nil, fmt.Errorf("%w: sensor label cannot be empty", ErrInvalidConfig)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, sensor := range d.Sensors {
		if sensor.Label == label {
			return sensor, nil
		}
	}

	return nil, fmt.Errorf("%w: sensor with label %s", ErrSensorNotFound, label)
}

// GetSensorByTypeAndIndex finds a sensor by type and index.
func (d *Device) GetSensorByTypeAndIndex(ctx context.Context, sensorType SensorType, index int) (*SensorInfo, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	if index < 1 {
		return nil, fmt.Errorf("%w: sensor index must be positive", ErrInvalidSensorIndex)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, sensor := range d.Sensors {
		if sensor.Type == sensorType && sensor.Index == index {
			return sensor, nil
		}
	}

	return nil, fmt.Errorf("%w: sensor %s%d", ErrSensorNotFound, sensorType.Prefix(), index)
}

// scanSensors scans the device directory for available sensors.
func (d *Device) scanSensors(ctx context.Context) error {
	entries, err := os.ReadDir(d.Path)
	if err != nil {
		return fmt.Errorf("%w: failed to read device directory: %w", ErrDiscoveryFailure, err)
	}

	sensorMap := make(map[string]*SensorInfo)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
		default:
		}

		fileName := entry.Name()
		sensorInfo := d.parseSensorFile(fileName)
		if sensorInfo == nil {
			continue
		}

		sensorInfo.DevicePath = d.Path
		key := fmt.Sprintf("%s%d", sensorInfo.Type.Prefix(), sensorInfo.Index)

		existing, exists := sensorMap[key]
		if !exists {
			existing = &SensorInfo{
				Name:       key,
				Index:      sensorInfo.Index,
				Type:       sensorInfo.Type,
				Attributes: make(map[SensorAttribute]string),
				DevicePath: d.Path,
			}
			sensorMap[key] = existing
		}

		existing.Attributes[sensorInfo.getAttributeFromFile(fileName)] = filepath.Join(d.Path, fileName)

		attr := sensorInfo.getAttributeFromFile(fileName)
		fullPath := filepath.Join(d.Path, fileName)
		existing.Attributes[attr] = fullPath

		if attr == AttributeLabel {
			labelBytes, err := os.ReadFile(filepath.Join(d.Path, fileName))
			if err == nil {
				existing.Label = strings.TrimSpace(string(labelBytes))
			}
		}

		// Mark writable if attribute is writable OR it's the PWM value file ("pwmN").
		if IsFileWritable(fullPath) && (attr.IsWritable() || (sensorInfo.Type == SensorTypePWM && attr == AttributeInput && !strings.Contains(fileName, "_"))) {
			existing.Writable = true
		}
	}

	d.Sensors = sensorMap
	return nil
}

// parseSensorFile parses a sensor filename and returns sensor information.
func (d *Device) parseSensorFile(fileName string) *SensorInfo {
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
		matches := pattern.FindStringSubmatch(fileName)
		if len(matches) >= 2 {
			index, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			return &SensorInfo{
				Index: index,
				Type:  sensorType,
			}
		}
	}

	return nil
}

// getAttributeFromFile determines the sensor attribute from a filename.
func (s *SensorInfo) getAttributeFromFile(fileName string) SensorAttribute {
	if strings.HasSuffix(fileName, "_input") {
		return AttributeInput
	} else if strings.HasSuffix(fileName, "_label") {
		return AttributeLabel
	} else if strings.HasSuffix(fileName, "_min") {
		return AttributeMin
	} else if strings.HasSuffix(fileName, "_max") {
		return AttributeMax
	} else if strings.HasSuffix(fileName, "_crit") {
		return AttributeCrit
	} else if strings.HasSuffix(fileName, "_alarm") {
		return AttributeAlarm
	} else if strings.HasSuffix(fileName, "_enable") {
		return AttributeEnable
	} else if strings.HasSuffix(fileName, "_target") {
		return AttributeTarget
	} else if strings.HasSuffix(fileName, "_fault") {
		return AttributeFault
	} else if strings.HasSuffix(fileName, "_beep") {
		return AttributeBeep
	} else if strings.HasSuffix(fileName, "_offset") {
		return AttributeOffset
	} else if strings.HasSuffix(fileName, "_type") {
		return AttributeType
	}

	return AttributeInput
}

// HasAttribute checks if the sensor supports a specific attribute.
func (s *SensorInfo) HasAttribute(attr SensorAttribute) bool {
	_, exists := s.Attributes[attr]
	return exists
}

// GetAttributePath returns the sysfs path for a specific attribute.
func (s *SensorInfo) GetAttributePath(attr SensorAttribute) (string, error) {
	path, exists := s.Attributes[attr]
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrAttributeNotSupported, attr.String())
	}
	return path, nil
}

// String returns a string representation of the sensor info.
func (s *SensorInfo) String() string {
	label := s.Label
	if label == "" {
		label = s.Name
	}
	return fmt.Sprintf("%s (%s%d)", label, s.Type.Prefix(), s.Index)
}

// String returns a string representation of the device.
func (d *Device) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("Device{name=%s, hwmon=%s, sensors=%d}", d.Name, d.HwmonID, len(d.Sensors))
}
