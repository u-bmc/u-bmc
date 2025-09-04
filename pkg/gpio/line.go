// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package gpio

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LineGroup represents a group of GPIO lines that can be operated on together.
type LineGroup struct {
	lines map[string]*Line
	mu    sync.RWMutex
}

// NewLineGroup creates a new line group.
func NewLineGroup() *LineGroup {
	return &LineGroup{
		lines: make(map[string]*Line),
	}
}

// Add adds a line to the group.
func (lg *LineGroup) Add(name string, line *Line) {
	lg.mu.Lock()
	defer lg.mu.Unlock()
	if line != nil {
		lg.lines[name] = line
	}
}

// Remove removes a line from the group.
func (lg *LineGroup) Remove(name string) {
	lg.mu.Lock()
	defer lg.mu.Unlock()
	delete(lg.lines, name)
}

// Get gets a line from the group.
func (lg *LineGroup) Get(name string) (*Line, bool) {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	line, exists := lg.lines[name]
	return line, exists
}

// SetValues sets values for multiple output lines simultaneously.
func (lg *LineGroup) SetValues(values map[string]int) error {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	var firstError error
	for name, value := range values {
		if line, exists := lg.lines[name]; exists {
			if err := line.SetValue(value); err != nil && firstError == nil {
				firstError = err
			}
		}
	}
	return firstError
}

// GetValues gets values from multiple lines simultaneously.
func (lg *LineGroup) GetValues() (map[string]int, error) {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	values := make(map[string]int)
	var firstError error

	for name, line := range lg.lines {
		if value, err := line.GetValue(); err != nil {
			if firstError == nil {
				firstError = err
			}
		} else {
			values[name] = value
		}
	}

	return values, firstError
}

// Close closes all lines in the group.
func (lg *LineGroup) Close() error {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	var firstError error
	for _, line := range lg.lines {
		if err := line.Close(); err != nil && firstError == nil {
			firstError = err
		}
	}
	lg.lines = make(map[string]*Line)
	return firstError
}

// BlinkPattern represents a blinking pattern for LEDs.
type BlinkPattern struct {
	// OnDuration is how long the LED stays on
	OnDuration time.Duration
	// OffDuration is how long the LED stays off
	OffDuration time.Duration
	// Cycles is the number of blink cycles (0 = infinite)
	Cycles int
}

// Blink performs a blinking pattern on an output line.
func (l *Line) Blink(ctx context.Context, pattern BlinkPattern) error {
	if l.config.Direction != DirectionOutput {
		return fmt.Errorf("%w: cannot blink input line", ErrInvalidDirection)
	}

	if pattern.OnDuration <= 0 || pattern.OffDuration <= 0 {
		return fmt.Errorf("%w: blink durations must be > 0", ErrInvalidConfiguration)
	}

	cycles := pattern.Cycles
	if cycles == 0 {
		cycles = -1 // Infinite
	}

	for cycles != 0 {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
		default:
		}

		// Turn on
		if err := l.SetValue(1); err != nil {
			return fmt.Errorf("failed to turn on during blink: %w", err)
		}

		select {
		case <-time.After(pattern.OnDuration):
		case <-ctx.Done():
			_ = l.SetValue(0) // Try to turn off before returning
			return fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
		}

		// Turn off
		if err := l.SetValue(0); err != nil {
			return fmt.Errorf("failed to turn off during blink: %w", err)
		}

		select {
		case <-time.After(pattern.OffDuration):
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
		}

		if cycles > 0 {
			cycles--
		}
	}

	return nil
}

// Pulse performs a single pulse (high-low transition) with specified duration.
func (l *Line) Pulse(duration time.Duration) error {
	return l.Toggle(duration)
}

// PulseCtx performs a single pulse with context support.
func (l *Line) PulseCtx(ctx context.Context, duration time.Duration) error {
	return l.ToggleCtx(ctx, duration)
}

// Hold sets the line high for a specified duration, then sets it low.
func (l *Line) Hold(ctx context.Context, duration time.Duration) error {
	if err := l.SetValue(1); err != nil {
		return fmt.Errorf("failed to set high: %w", err)
	}

	select {
	case <-time.After(duration):
	case <-ctx.Done():
		_ = l.SetValue(0) // Try to turn off before returning
		return fmt.Errorf("%w: %w", ErrOperationCanceled, ctx.Err())
	}

	if err := l.SetValue(0); err != nil {
		return fmt.Errorf("failed to set low: %w", err)
	}

	return nil
}

// LineMonitor provides monitoring capabilities for GPIO lines.
type LineMonitor struct {
	line     *Line
	callback func(Event)
	stop     chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewLineMonitor creates a new line monitor for the given line.
func NewLineMonitor(line *Line, callback func(Event)) *LineMonitor {
	return &LineMonitor{
		line:     line,
		callback: callback,
		stop:     make(chan struct{}),
	}
}

// Start starts monitoring the line for events.
func (lm *LineMonitor) Start() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.running {
		return fmt.Errorf("monitor already running")
	}

	// Allow restart after Stop.
	if lm.stop == nil {
		lm.stop = make(chan struct{})
	} else {
		select {
		case <-lm.stop:
			lm.stop = make(chan struct{})
		default:
		}
	}

	if lm.line.config.Direction != DirectionInput || lm.line.config.Edge == EdgeNone {
		return fmt.Errorf("%w: line must be configured for input with edge detection", ErrInvalidConfiguration)
	}

	lm.running = true
	go lm.monitorLoop()
	return nil
}

// Stop stops monitoring the line.
func (lm *LineMonitor) Stop() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if !lm.running {
		return
	}

	close(lm.stop)
	lm.running = false
}

// IsRunning returns whether the monitor is currently running.
func (lm *LineMonitor) IsRunning() bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	return lm.running
}

// monitorLoop is the main monitoring loop.
func (lm *LineMonitor) monitorLoop() {
	defer func() {
		lm.mu.Lock()
		lm.running = false
		lm.mu.Unlock()
	}()

	events := lm.line.Events()
	if events == nil {
		return
	}

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return // Channel closed
			}
			if lm.callback != nil {
				lm.callback(event)
			}
		case <-lm.stop:
			return
		}
	}
}

// LineState tracks the state of a GPIO line over time.
type LineState struct {
	line         *Line
	currentValue int
	lastChanged  time.Time
	changeCount  uint64
	mu           sync.RWMutex
}

// NewLineState creates a new line state tracker.
func NewLineState(line *Line) (*LineState, error) {
	value, err := line.GetValue()
	if err != nil {
		return nil, err
	}

	return &LineState{
		line:         line,
		currentValue: value,
		lastChanged:  time.Now(),
	}, nil
}

// Update reads the current line value and updates the state.
func (ls *LineState) Update() error {
	value, err := ls.line.GetValue()
	if err != nil {
		return err
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	if value != ls.currentValue {
		ls.currentValue = value
		ls.lastChanged = time.Now()
		ls.changeCount++
	}

	return nil
}

// GetState returns the current state information: current value, time of last change, and total change count.
func (ls *LineState) GetState() (int, time.Time, uint64) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.currentValue, ls.lastChanged, ls.changeCount
}

// TimeSinceLastChange returns the duration since the last state change.
func (ls *LineState) TimeSinceLastChange() time.Duration {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return time.Since(ls.lastChanged)
}

// PowerControlHelper provides common power control operations for BMC systems.
type PowerControlHelper struct {
	powerButton *Line
	resetButton *Line
	powerLed    *Line
	powerStatus *Line
}

// NewPowerControlHelper creates a new power control helper.
func NewPowerControlHelper(powerButton, resetButton, powerLed, powerStatus *Line) *PowerControlHelper {
	return &PowerControlHelper{
		powerButton: powerButton,
		resetButton: resetButton,
		powerLed:    powerLed,
		powerStatus: powerStatus,
	}
}

// PowerOn performs a momentary power button press.
func (pch *PowerControlHelper) PowerOn(ctx context.Context) error {
	if pch.powerButton == nil {
		return fmt.Errorf("power button not configured")
	}
	return pch.powerButton.PulseCtx(ctx, 200*time.Millisecond)
}

// PowerOff performs a momentary power button press (soft shutdown).
func (pch *PowerControlHelper) PowerOff(ctx context.Context) error {
	if pch.powerButton == nil {
		return fmt.Errorf("power button not configured")
	}
	return pch.powerButton.PulseCtx(ctx, 200*time.Millisecond)
}

// ForceOff performs a long power button press (hard shutdown).
func (pch *PowerControlHelper) ForceOff(ctx context.Context) error {
	if pch.powerButton == nil {
		return fmt.Errorf("power button not configured")
	}
	return pch.powerButton.Hold(ctx, 4*time.Second)
}

// Reset performs a reset button press.
func (pch *PowerControlHelper) Reset(ctx context.Context) error {
	if pch.resetButton == nil {
		return fmt.Errorf("reset button not configured")
	}
	return pch.resetButton.PulseCtx(ctx, 200*time.Millisecond)
}

// GetPowerStatus reads the current power status.
func (pch *PowerControlHelper) GetPowerStatus() (bool, error) {
	if pch.powerStatus == nil {
		return false, fmt.Errorf("power status not configured")
	}
	value, err := pch.powerStatus.GetValue()
	if err != nil {
		return false, err
	}
	return value == 1, nil
}

// TurnOnPowerLed turns on the power LED.
func (pch *PowerControlHelper) TurnOnPowerLed() error {
	if pch.powerLed == nil {
		return fmt.Errorf("power LED not configured")
	}
	return pch.powerLed.SetValue(1)
}

// TurnOffPowerLed turns off the power LED.
func (pch *PowerControlHelper) TurnOffPowerLed() error {
	if pch.powerLed == nil {
		return fmt.Errorf("power LED not configured")
	}
	return pch.powerLed.SetValue(0)
}

// BlinkPowerLed blinks the power LED with the specified pattern.
func (pch *PowerControlHelper) BlinkPowerLed(ctx context.Context, pattern BlinkPattern) error {
	if pch.powerLed == nil {
		return fmt.Errorf("power LED not configured")
	}
	return pch.powerLed.Blink(ctx, pattern)
}

// CommonBlinkPatterns returns commonly used blink patterns for BMC use cases.
// Each helper avoids global state and can be inlined where needed.

// SlowBlink returns a slow attention pattern.
func SlowBlink() BlinkPattern {
	return BlinkPattern{
		OnDuration:  500 * time.Millisecond,
		OffDuration: 1500 * time.Millisecond,
		Cycles:      0, // Infinite.
	}
}

// FastBlink returns a rapid attention pattern.
func FastBlink() BlinkPattern {
	return BlinkPattern{
		OnDuration:  100 * time.Millisecond,
		OffDuration: 100 * time.Millisecond,
		Cycles:      0, // Infinite.
	}
}

// BootBlink returns a startup indication pattern.
func BootBlink() BlinkPattern {
	return BlinkPattern{
		OnDuration:  200 * time.Millisecond,
		OffDuration: 200 * time.Millisecond,
		Cycles:      5,
	}
}

// ErrorBlink returns an error indication pattern.
func ErrorBlink() BlinkPattern {
	return BlinkPattern{
		OnDuration:  50 * time.Millisecond,
		OffDuration: 150 * time.Millisecond,
		Cycles:      10,
	}
}
