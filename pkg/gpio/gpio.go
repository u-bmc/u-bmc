// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package gpio

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// Event represents a GPIO event (edge detection).
type Event struct {
	// Line is the line number that generated the event
	Line int
	// Timestamp is when the event occurred
	Timestamp time.Time
	// Type indicates the type of edge that triggered the event
	Type Edge
	// Value is the line value when the event occurred
	Value int
}

// Line represents a single GPIO line.
type Line struct {
	// chip is the GPIO chip this line belongs to
	chip string
	// number is the line number within the chip
	number int
	// name is the human-readable name/label of the line
	name string
	// config holds the line configuration
	config LineConfig
	// line is the underlying gpio-cdev line
	line *gpiocdev.Line
	// events is the channel for GPIO events (if edge detection is enabled)
	events chan Event
	// mu protects concurrent access to this line
	mu sync.RWMutex
	// closed indicates if the line has been closed
	closed bool
	// manager is a reference back to the manager
	manager *Manager
}

// Manager manages GPIO chips and lines.
type Manager struct {
	// lines maps line identifiers to Line instances
	lines map[string]*Line
	// chips maps chip paths to chip info
	chips map[string]*gpiocdev.Chip
	// mu protects concurrent access to the manager
	mu sync.RWMutex
	// closed indicates if the manager has been closed
	closed bool
}

// NewManager creates a new GPIO manager.
func NewManager() *Manager {
	return &Manager{
		lines: make(map[string]*Line),
		chips: make(map[string]*gpiocdev.Chip),
	}
}

// RequestLine requests a single GPIO line by name/label.
func (m *Manager) RequestLine(chipPath, lineName string, opts ...Option) (*Line, error) {
	config := NewConfig(
		WithChip(chipPath),
		WithConsumer("u-bmc"),
	)
	for _, opt := range opts {
		opt.apply(config)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	return m.requestLineByName(chipPath, lineName, config.DefaultConfig)
}

// RequestLineByNumber requests a single GPIO line by number.
func (m *Manager) RequestLineByNumber(chipPath string, lineNumber int, opts ...Option) (*Line, error) {
	config := NewConfig(
		WithChip(chipPath),
		WithConsumer("u-bmc"),
	)
	for _, opt := range opts {
		opt.apply(config)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	return m.requestLineByNumber(chipPath, lineNumber, config.DefaultConfig)
}

// RequestLines requests multiple GPIO lines according to the provided configuration.
func (m *Manager) RequestLines(config *Config) (map[string]*Line, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfiguration, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, ErrManagerClosed
	}

	lines := make(map[string]*Line)

	// Request lines by name
	for name := range config.Lines {
		lineConfig := config.GetLineConfig(name)
		line, err := m.requestLineByNameLocked(config.ChipPath, name, lineConfig)
		if err != nil {
			// Clean up any successfully requested lines
			for _, l := range lines {
				_ = l.Close()
			}
			return nil, err
		}
		lines[name] = line
	}

	// Request lines by number
	for number := range config.LineNumbers {
		lineConfig := config.GetLineNumberConfig(number)
		line, err := m.requestLineByNumberLocked(config.ChipPath, number, lineConfig)
		if err != nil {
			// Clean up any successfully requested lines
			for _, l := range lines {
				_ = l.Close()
			}
			return nil, err
		}
		key := fmt.Sprintf("line_%d", number)
		lines[key] = line
	}

	return lines, nil
}

// GetLine returns a previously requested line by its identifier.
func (m *Manager) GetLine(identifier string) (*Line, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	line, exists := m.lines[identifier]
	return line, exists
}

// GetAllLines returns all currently managed lines.
func (m *Manager) GetAllLines() map[string]*Line {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Line, len(m.lines))
	for id, line := range m.lines {
		result[id] = line
	}
	return result
}

// Close closes all managed lines and chips.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	var lastErr error

	// Close all lines
	for _, line := range m.lines {
		if err := line.close(); err != nil {
			lastErr = err
		}
	}

	// Close all chips
	for _, chip := range m.chips {
		if err := chip.Close(); err != nil {
			lastErr = err
		}
	}

	m.lines = make(map[string]*Line)
	m.chips = make(map[string]*gpiocdev.Chip)
	m.closed = true

	return lastErr
}

// requestLineByName requests a line by name with proper locking.
func (m *Manager) requestLineByName(chipPath, lineName string, config LineConfig) (*Line, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requestLineByNameLocked(chipPath, lineName, config)
}

// requestLineByNameLocked requests a line by name (caller must hold lock).
func (m *Manager) requestLineByNameLocked(chipPath, lineName string, config LineConfig) (*Line, error) {
	if m.closed {
		return nil, ErrManagerClosed
	}

	identifier := fmt.Sprintf("%s:%s", chipPath, lineName)
	if existing, exists := m.lines[identifier]; exists {
		existing.mu.RLock()
		isClosed := existing.closed
		existing.mu.RUnlock()
		if !isClosed {
			return nil, ErrLineAlreadyRequested
		}
		delete(m.lines, identifier)
	}

	chip, err := m.getOrOpenChip(chipPath)
	if err != nil {
		return nil, err
	}

	// Find line number by name
	lineNumber := -1
	for i := 0; i < chip.Lines(); i++ {
		info, err := chip.LineInfo(i)
		if err != nil {
			continue
		}
		if info.Name == lineName {
			lineNumber = i
			break
		}
	}

	if lineNumber == -1 {
		return nil, fmt.Errorf("%w: line '%s' not found in chip '%s'", ErrLineNotFound, lineName, chipPath)
	}

	return m.createLine(chipPath, lineNumber, lineName, config, identifier)
}

// requestLineByNumber requests a line by number with proper locking.
func (m *Manager) requestLineByNumber(chipPath string, lineNumber int, config LineConfig) (*Line, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requestLineByNumberLocked(chipPath, lineNumber, config)
}

// requestLineByNumberLocked requests a line by number (caller must hold lock).
func (m *Manager) requestLineByNumberLocked(chipPath string, lineNumber int, config LineConfig) (*Line, error) {
	if m.closed {
		return nil, ErrManagerClosed
	}

	identifier := fmt.Sprintf("%s:%d", chipPath, lineNumber)
	if existing, exists := m.lines[identifier]; exists {
		existing.mu.RLock()
		isClosed := existing.closed
		existing.mu.RUnlock()
		if !isClosed {
			return nil, ErrLineAlreadyRequested
		}
		delete(m.lines, identifier)
	}

	chip, err := m.getOrOpenChip(chipPath)
	if err != nil {
		return nil, err
	}

	if lineNumber < 0 || lineNumber >= chip.Lines() {
		return nil, fmt.Errorf("%w: line number %d not valid for chip '%s'", ErrInvalidLineNumber, lineNumber, chipPath)
	}

	// Get line name for reference
	lineName := fmt.Sprintf("line_%d", lineNumber)
	if info, err := chip.LineInfo(lineNumber); err == nil && info.Name != "" {
		lineName = info.Name
	}

	return m.createLine(chipPath, lineNumber, lineName, config, identifier)
}

// getOrOpenChip gets an existing chip or opens a new one.
func (m *Manager) getOrOpenChip(chipPath string) (*gpiocdev.Chip, error) {
	if chip, exists := m.chips[chipPath]; exists {
		return chip, nil
	}

	chip, err := gpiocdev.NewChip(chipPath)
	if err != nil {
		// Prefer access denied when applicable
		if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
			return nil, fmt.Errorf("%w: failed to open chip '%s': %w", ErrChipAccessDenied, chipPath, err)
		}
		return nil, fmt.Errorf("%w: failed to open chip '%s': %w", ErrChipNotFound, chipPath, err)
	}

	m.chips[chipPath] = chip
	return chip, nil
}

// createLine creates and configures a new Line instance.
func (m *Manager) createLine(chipPath string, lineNumber int, lineName string, config LineConfig, identifier string) (*Line, error) {
	// Build gpio-cdev line request options
	var opts []gpiocdev.LineReqOption

	// Set consumer
	if config.Consumer != "" {
		opts = append(opts, gpiocdev.WithConsumer(config.Consumer))
	}

	// Set direction and initial value
	if config.Direction == DirectionOutput {
		opts = append(opts, gpiocdev.AsOutput(config.InitialValue))
	} else {
		opts = append(opts, gpiocdev.AsInput)
	}

	// Set bias
	switch config.Bias {
	case BiasPullUp:
		opts = append(opts, gpiocdev.WithPullUp)
	case BiasPullDown:
		opts = append(opts, gpiocdev.WithPullDown)
	case BiasDisabled:
		opts = append(opts, gpiocdev.WithBiasDisabled)
	}

	// Set edge detection
	if config.Direction == DirectionInput && config.Edge != EdgeNone {
		switch config.Edge {
		case EdgeRising:
			opts = append(opts, gpiocdev.WithRisingEdge)
		case EdgeFalling:
			opts = append(opts, gpiocdev.WithFallingEdge)
		case EdgeBoth:
			opts = append(opts, gpiocdev.WithBothEdges)
		}
	}

	// Set drive type for outputs
	if config.Direction == DirectionOutput {
		switch config.Drive {
		case DriveOpenDrain:
			opts = append(opts, gpiocdev.AsOpenDrain)
		case DriveOpenSource:
			opts = append(opts, gpiocdev.AsOpenSource)
		}
	}

	// Set active state
	if config.ActiveState == ActiveLow {
		opts = append(opts, gpiocdev.AsActiveLow)
	}

	// Set debounce period (if supported)
	if config.DebouncePeriod > 0 {
		opts = append(opts, gpiocdev.WithDebounce(config.DebouncePeriod))
	}

	// If edge detection is enabled, prepare handler before requesting
	var events chan Event
	if config.Direction == DirectionInput && config.Edge != EdgeNone {
		buf := config.EventBufferSize
		if buf <= 0 {
			buf = 16
		}
		events = make(chan Event, buf)
		eventHandler := func(evt gpiocdev.LineEvent) {
			e := Event{
				Line:      lineNumber,
				Timestamp: time.Unix(0, int64(evt.Timestamp)),
			}
			switch evt.Type {
			case gpiocdev.LineEventRisingEdge:
				e.Type, e.Value = EdgeRising, 1
			case gpiocdev.LineEventFallingEdge:
				e.Type, e.Value = EdgeFalling, 0
			}
			select {
			case events <- e:
			default:
			}
		}
		opts = append(opts, gpiocdev.WithEventHandler(eventHandler))
	}

	// Request the line once
	line, err := gpiocdev.RequestLine(chipPath, lineNumber, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to request line %d from chip '%s': %w", ErrLineNotFound, lineNumber, chipPath, err)
	}

	// Create Line instance
	gpioLine := &Line{
		chip:    chipPath,
		number:  lineNumber,
		name:    lineName,
		config:  config,
		line:    line,
		manager: m,
	}

	// Attach events channel if configured
	if events != nil {
		gpioLine.events = events
	}

	m.lines[identifier] = gpioLine
	return gpioLine, nil
}

// SetValue sets the value of an output GPIO line.
func (l *Line) SetValue(value int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return ErrLineClosed
	}

	if l.config.Direction != DirectionOutput {
		return fmt.Errorf("%w: cannot set value on input line", ErrInvalidDirection)
	}

	if value < 0 || value > 1 {
		return fmt.Errorf("%w: value must be 0 or 1", ErrInvalidValue)
	}

	if err := l.line.SetValue(value); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteOperation, err)
	}

	return nil
}

// GetValue gets the current value of a GPIO line.
func (l *Line) GetValue() (int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.closed {
		return 0, ErrLineClosed
	}

	value, err := l.line.Value()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrReadOperation, err)
	}

	return value, nil
}

// Toggle performs a toggle operation on an output GPIO line.
// It sets the line high, waits for the specified duration, then sets it low.
func (l *Line) Toggle(duration time.Duration) error {
	if err := l.SetValue(1); err != nil {
		return fmt.Errorf("%w: failed to set high: %w", ErrToggleOperation, err)
	}

	time.Sleep(duration)

	if err := l.SetValue(0); err != nil {
		return fmt.Errorf("%w: failed to set low: %w", ErrToggleOperation, err)
	}

	return nil
}

// ToggleCtx performs a toggle operation with context support.
func (l *Line) ToggleCtx(ctx context.Context, duration time.Duration) error {
	if err := l.SetValue(1); err != nil {
		return fmt.Errorf("%w: failed to set high: %w", ErrToggleOperation, err)
	}

	select {
	case <-time.After(duration):
		// Duration elapsed normally
	case <-ctx.Done():
		// Context canceled, still try to set low
		_ = l.SetValue(0)
		return fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
	}

	if err := l.SetValue(0); err != nil {
		return fmt.Errorf("%w: failed to set low: %w", ErrToggleOperation, err)
	}

	return nil
}

// Events returns the event channel for this line (if edge detection is enabled).
func (l *Line) Events() <-chan Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.events
}

// Info returns information about this GPIO line.
func (l *Line) Info() (string, int, LineConfig) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.name, l.number, l.config
}

// Close closes the GPIO line and releases its resources.
func (l *Line) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}

// close closes the line (caller must hold lock).
func (l *Line) close() error {
	if l.closed {
		return nil
	}

	var err error
	if l.line != nil {
		err = l.line.Close()
	}

	if l.events != nil {
		close(l.events)
	}

	l.closed = true
	return err
}

// CreateToggleAction creates a stateless action function for GPIO toggle operations.
// This is compatible with state machines and other callback-based systems.
func CreateToggleAction(manager *Manager, chipPath, lineName string, duration time.Duration, opts ...Option) func(context.Context, ...any) error {
	return func(ctx context.Context, args ...any) error {
		line, err := manager.RequestLine(chipPath, lineName, opts...)
		if err != nil {
			return err
		}
		defer line.Close()

		return line.ToggleCtx(ctx, duration)
	}
}

// CreateToggleActionByNumber creates a stateless action function for GPIO toggle operations using line numbers.
func CreateToggleActionByNumber(manager *Manager, chipPath string, lineNumber int, duration time.Duration, opts ...Option) func(context.Context, ...any) error {
	return func(ctx context.Context, args ...any) error {
		line, err := manager.RequestLineByNumber(chipPath, lineNumber, opts...)
		if err != nil {
			return err
		}
		defer line.Close()

		return line.ToggleCtx(ctx, duration)
	}
}
