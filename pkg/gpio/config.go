// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package gpio

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Direction represents the GPIO line direction.
type Direction int

const (
	// DirectionInput configures the GPIO line as an input.
	DirectionInput Direction = iota
	// DirectionOutput configures the GPIO line as an output.
	DirectionOutput
)

// String returns the string representation of the Direction.
func (d Direction) String() string {
	switch d {
	case DirectionInput:
		return "Input"
	case DirectionOutput:
		return "Output"
	default:
		return fmt.Sprintf("Direction(%d)", d)
	}
}

// Bias represents the GPIO line bias setting.
type Bias int

const (
	// BiasDisabled disables internal pull-up/pull-down resistors.
	BiasDisabled Bias = iota
	// BiasPullUp enables internal pull-up resistor.
	BiasPullUp
	// BiasPullDown enables internal pull-down resistor.
	BiasPullDown
)

// String returns the string representation of the Bias.
func (b Bias) String() string {
	switch b {
	case BiasDisabled:
		return "Disabled"
	case BiasPullUp:
		return "Pull-Up"
	case BiasPullDown:
		return "Pull-Down"
	default:
		return fmt.Sprintf("Bias(%d)", b)
	}
}

// Edge represents GPIO edge detection settings.
type Edge int

const (
	// EdgeNone disables edge detection.
	EdgeNone Edge = iota
	// EdgeRising enables detection of rising edges.
	EdgeRising
	// EdgeFalling enables detection of falling edges.
	EdgeFalling
	// EdgeBoth enables detection of both rising and falling edges.
	EdgeBoth
)

// String returns the string representation of the Edge.
func (e Edge) String() string {
	switch e {
	case EdgeNone:
		return "None"
	case EdgeRising:
		return "Rising"
	case EdgeFalling:
		return "Falling"
	case EdgeBoth:
		return "Both"
	default:
		return fmt.Sprintf("Edge(%d)", e)
	}
}

// Drive represents the GPIO drive type.
type Drive int

const (
	// DrivePushPull configures the line for push-pull output.
	DrivePushPull Drive = iota
	// DriveOpenDrain configures the line for open-drain output.
	DriveOpenDrain
	// DriveOpenSource configures the line for open-source output.
	DriveOpenSource
)

// String returns the string representation of the Drive.
func (d Drive) String() string {
	switch d {
	case DrivePushPull:
		return "Push-Pull"
	case DriveOpenDrain:
		return "Open-Drain"
	case DriveOpenSource:
		return "Open-Source"
	default:
		return fmt.Sprintf("Drive(%d)", d)
	}
}

// ActiveState represents whether the line is active high or low.
type ActiveState int

const (
	// ActiveHigh means logical high is represented by high voltage.
	ActiveHigh ActiveState = iota
	// ActiveLow means logical high is represented by low voltage.
	ActiveLow
)

// String returns the string representation of the ActiveState.
func (a ActiveState) String() string {
	switch a {
	case ActiveHigh:
		return "Active-High"
	case ActiveLow:
		return "Active-Low"
	default:
		return fmt.Sprintf("ActiveState(%d)", a)
	}
}

// LineConfig holds configuration for a single GPIO line.
// When used in line-specific configurations, all fields except Consumer, DebouncePeriod,
// and EventBufferSize will override defaults even if zero-values.
// Consumer (when empty), DebouncePeriod (when zero), and EventBufferSize (when zero) inherit from defaults.
type LineConfig struct {
	// Direction specifies whether the line is an input or output
	Direction Direction
	// InitialValue is the initial value for output lines (0 or 1)
	InitialValue int
	// Bias configures internal pull-up/pull-down resistors
	Bias Bias
	// Edge configures edge detection for input lines
	Edge Edge
	// Drive configures the output drive type
	Drive Drive
	// ActiveState configures active high/low behavior
	ActiveState ActiveState
	// DebouncePeriod configures input debouncing (hardware dependent)
	DebouncePeriod time.Duration
	// Consumer is a string identifying the consumer of this line
	Consumer string
	// EventBufferSize configures the size of event buffers for edge detection
	EventBufferSize int
}

// Config holds the configuration for GPIO operations.
type Config struct {
	// ChipPath is the path to the GPIO chip device (e.g., "/dev/gpiochip0")
	ChipPath string
	// Lines maps line names/labels to their configuration
	Lines map[string]LineConfig
	// LineNumbers maps line numbers to their configuration
	LineNumbers map[int]LineConfig
	// DefaultConfig provides default settings for unconfigured options
	DefaultConfig LineConfig
	// Timeout is the default timeout for GPIO operations
	Timeout time.Duration
	// EventBufferSize configures the size of event buffers for edge detection
	EventBufferSize int
}

// Option represents a configuration option for GPIO operations.
type Option interface {
	apply(*Config)
}

type chipPathOption struct {
	chipPath string
}

func (o *chipPathOption) apply(c *Config) {
	c.ChipPath = o.chipPath
}

// WithChip sets the GPIO chip path.
func WithChip(chipPath string) Option {
	return &chipPathOption{
		chipPath: chipPath,
	}
}

type linesOption struct {
	lines map[string]LineConfig
}

func (o *linesOption) apply(c *Config) {
	if c.Lines == nil {
		c.Lines = make(map[string]LineConfig)
	}
	for name, config := range o.lines {
		c.Lines[name] = config
	}
}

// WithLines sets the configuration for multiple named GPIO lines.
func WithLines(lines map[string]LineConfig) Option {
	return &linesOption{
		lines: lines,
	}
}

type lineNumbersOption struct {
	lineNumbers map[int]LineConfig
}

func (o *lineNumbersOption) apply(c *Config) {
	if c.LineNumbers == nil {
		c.LineNumbers = make(map[int]LineConfig)
	}
	for number, config := range o.lineNumbers {
		c.LineNumbers[number] = config
	}
}

// WithLineNumbers sets the configuration for multiple GPIO lines by number.
func WithLineNumbers(lineNumbers map[int]LineConfig) Option {
	return &lineNumbersOption{
		lineNumbers: lineNumbers,
	}
}

type directionOption struct {
	direction Direction
}

func (o *directionOption) apply(c *Config) {
	c.DefaultConfig.Direction = o.direction
}

// WithDirection sets the default direction for GPIO lines.
func WithDirection(direction Direction) Option {
	return &directionOption{
		direction: direction,
	}
}

type initialValueOption struct {
	value int
}

func (o *initialValueOption) apply(c *Config) {
	c.DefaultConfig.InitialValue = o.value
}

// WithInitialValue sets the default initial value for output GPIO lines.
func WithInitialValue(value int) Option {
	return &initialValueOption{
		value: value,
	}
}

type biasOption struct {
	bias Bias
}

func (o *biasOption) apply(c *Config) {
	c.DefaultConfig.Bias = o.bias
}

// WithBias sets the default bias setting for GPIO lines.
func WithBias(bias Bias) Option {
	return &biasOption{
		bias: bias,
	}
}

type edgeOption struct {
	edge Edge
}

func (o *edgeOption) apply(c *Config) {
	c.DefaultConfig.Edge = o.edge
}

// WithEdge sets the default edge detection setting for input GPIO lines.
func WithEdge(edge Edge) Option {
	return &edgeOption{
		edge: edge,
	}
}

type driveOption struct {
	drive Drive
}

func (o *driveOption) apply(c *Config) {
	c.DefaultConfig.Drive = o.drive
}

// WithDrive sets the default drive type for output GPIO lines.
func WithDrive(drive Drive) Option {
	return &driveOption{
		drive: drive,
	}
}

type activeStateOption struct {
	activeState ActiveState
}

func (o *activeStateOption) apply(c *Config) {
	c.DefaultConfig.ActiveState = o.activeState
}

// WithActiveState sets the default active state for GPIO lines.
func WithActiveState(activeState ActiveState) Option {
	return &activeStateOption{
		activeState: activeState,
	}
}

type debouncePeriodOption struct {
	period time.Duration
}

func (o *debouncePeriodOption) apply(c *Config) {
	c.DefaultConfig.DebouncePeriod = o.period
}

// WithDebouncePeriod sets the default debounce period for input GPIO lines.
func WithDebouncePeriod(period time.Duration) Option {
	return &debouncePeriodOption{
		period: period,
	}
}

type consumerOption struct {
	consumer string
}

func (o *consumerOption) apply(c *Config) {
	c.DefaultConfig.Consumer = o.consumer
}

// WithConsumer sets the default consumer string for GPIO lines.
func WithConsumer(consumer string) Option {
	return &consumerOption{
		consumer: consumer,
	}
}

type timeoutOption struct {
	timeout time.Duration
}

func (o *timeoutOption) apply(c *Config) {
	c.Timeout = o.timeout
}

// WithTimeout sets the default timeout for GPIO operations.
func WithTimeout(timeout time.Duration) Option {
	return &timeoutOption{
		timeout: timeout,
	}
}

type eventBufferSizeOption struct {
	size int
}

func (o *eventBufferSizeOption) apply(c *Config) {
	c.EventBufferSize = o.size
	// Keep line-level defaults in sync so single-line requests and merges inherit it.
	c.DefaultConfig.EventBufferSize = o.size
}

// WithEventBufferSize sets the event buffer size for edge detection.
func WithEventBufferSize(size int) Option {
	return &eventBufferSizeOption{
		size: size,
	}
}

// NewConfig creates a new Config with sane defaults and applies the provided options.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		ChipPath:    "/dev/gpiochip0",
		Lines:       make(map[string]LineConfig),
		LineNumbers: make(map[int]LineConfig),
		DefaultConfig: LineConfig{
			Direction:       DirectionOutput,
			InitialValue:    0,
			Bias:            BiasDisabled,
			Edge:            EdgeNone,
			Drive:           DrivePushPull,
			ActiveState:     ActiveHigh,
			DebouncePeriod:  0,
			Consumer:        "u-bmc",
			EventBufferSize: 16,
		},
		Timeout:         5 * time.Second,
		EventBufferSize: 16,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

// NewLineConfig creates a new LineConfig with sane defaults and applies the provided options.
func NewLineConfig(opts ...Option) LineConfig {
	cfg := NewConfig(opts...)
	return cfg.DefaultConfig
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	if c.ChipPath == "" {
		return fmt.Errorf("%w: chip path cannot be empty", ErrInvalidConfiguration)
	}

	if !strings.HasPrefix(c.ChipPath, "/dev/gpiochip") {
		return fmt.Errorf("%w: chip path must start with '/dev/gpiochip'", ErrInvalidChipPath)
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("%w: timeout must be positive", ErrInvalidTimeout)
	}

	if c.EventBufferSize <= 0 {
		return fmt.Errorf("%w: event buffer size must be positive", ErrInvalidConfiguration)
	}

	// Validate line configurations
	for name, lineConfig := range c.Lines {
		if err := c.validateLineConfig(name, lineConfig); err != nil {
			return err
		}
	}

	for number, lineConfig := range c.LineNumbers {
		if err := c.validateLineConfig(fmt.Sprintf("line_%d", number), lineConfig); err != nil {
			return err
		}
	}

	return c.validateLineConfig("default", c.DefaultConfig)
}

// validateLineConfig validates a single line configuration.
func (c *Config) validateLineConfig(name string, lineConfig LineConfig) error {
	if lineConfig.InitialValue < 0 || lineConfig.InitialValue > 1 {
		return fmt.Errorf("%w: initial value for line '%s' must be 0 or 1", ErrInvalidValue, name)
	}

	if lineConfig.Direction == DirectionOutput && lineConfig.Edge != EdgeNone {
		return fmt.Errorf("%w: output line '%s' cannot have edge detection", ErrConfigurationConflict, name)
	}

	if lineConfig.Direction == DirectionInput && lineConfig.Drive != DrivePushPull {
		return fmt.Errorf("%w: input line '%s' cannot have custom drive setting", ErrConfigurationConflict, name)
	}

	if lineConfig.DebouncePeriod < 0 {
		return fmt.Errorf("%w: debounce period for line '%s' cannot be negative", ErrInvalidConfiguration, name)
	}

	if lineConfig.EventBufferSize < 0 {
		return fmt.Errorf("%w: event buffer size for line '%s' cannot be negative", ErrInvalidConfiguration, name)
	}

	return nil
}

// GetLineConfig returns the effective configuration for a named line.
// It merges the line-specific config with the default config.
func (c *Config) GetLineConfig(name string) LineConfig {
	if lineConfig, exists := c.Lines[name]; exists {
		return c.mergeWithDefault(lineConfig)
	}
	return c.DefaultConfig
}

// GetLineNumberConfig returns the effective configuration for a numbered line.
// It merges the line-specific config with the default config.
func (c *Config) GetLineNumberConfig(number int) LineConfig {
	if lineConfig, exists := c.LineNumbers[number]; exists {
		return c.mergeWithDefault(lineConfig)
	}
	return c.DefaultConfig
}

// mergeWithDefault merges a line config with the default config.
// Line-level values fully override defaults for Direction, InitialValue, Bias, Edge, Drive, and ActiveState.
// Only Consumer and DebouncePeriod skip zero-values and inherit from defaults when unset.
func (c *Config) mergeWithDefault(lineConfig LineConfig) LineConfig {
	result := c.DefaultConfig

	// Only override non-zero/non-default values for these fields
	if lineConfig.Consumer != "" {
		result.Consumer = lineConfig.Consumer
	}
	if lineConfig.DebouncePeriod != 0 {
		result.DebouncePeriod = lineConfig.DebouncePeriod
	}
	if lineConfig.EventBufferSize != 0 {
		result.EventBufferSize = lineConfig.EventBufferSize
	}

	// Always use explicit values from lineConfig for enums and other fields
	result.Direction = lineConfig.Direction
	result.InitialValue = lineConfig.InitialValue
	result.Bias = lineConfig.Bias
	result.Edge = lineConfig.Edge
	result.Drive = lineConfig.Drive
	result.ActiveState = lineConfig.ActiveState

	return result
}

// GetAllLineNames returns all configured line names.
func (c *Config) GetAllLineNames() []string {
	names := make([]string, 0, len(c.Lines))
	for name := range c.Lines {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetAllLineNumbers returns all configured line numbers.
func (c *Config) GetAllLineNumbers() []int {
	numbers := make([]int, 0, len(c.LineNumbers))
	for number := range c.LineNumbers {
		numbers = append(numbers, number)
	}
	sort.Ints(numbers)
	return numbers
}
