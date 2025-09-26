// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package gpio

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// RequestLine requests a single GPIO line with the specified configuration options.
// Returns a *gpiocdev.Line that can be used directly with the underlying library.
func RequestLine(chip, lineName string, opts ...Option) (*gpiocdev.Line, error) {
	if chip == "" {
		return nil, fmt.Errorf("%w: chip path cannot be empty", ErrOperationFailed)
	}
	if lineName == "" {
		return nil, fmt.Errorf("%w: line name cannot be empty", ErrOperationFailed)
	}

	if err := gpiocdev.IsChip(chip); err != nil {
		return nil, mapGpiocdevError(err, fmt.Sprintf("invalid chip path '%s'", chip))
	}

	foundChip, offset, err := gpiocdev.FindLine(lineName)
	if err != nil {
		return nil, mapGpiocdevError(err, fmt.Sprintf("failed to find line '%s'", lineName))
	}
	// Normalize device identifiers (path vs basename) before comparing.
	if filepath.Base(foundChip) != filepath.Base(chip) {
		return nil, fmt.Errorf("%w: line '%s' not found on chip '%s'", ErrLineNotFound, lineName, chip)
	}

	// Default consumer, allow caller to override by placing their option last.
	defaultOpts := []gpiocdev.LineReqOption{gpiocdev.WithConsumer("u-bmc")}
	gpiocdevOpts := append(defaultOpts, convertOptions(opts)...)

	line, err := gpiocdev.RequestLine(chip, offset, gpiocdevOpts...)
	if err != nil {
		return nil, mapGpiocdevError(err, fmt.Sprintf("failed to request line '%s' from chip '%s'", lineName, chip))
	}

	return line, nil
}

// RequestLineByNumber requests a single GPIO line by its number with the specified configuration options.
// Returns a *gpiocdev.Line that can be used directly with the underlying library.
func RequestLineByNumber(chip string, lineNumber int, opts ...Option) (*gpiocdev.Line, error) {
	if chip == "" {
		return nil, fmt.Errorf("%w: chip path cannot be empty", ErrOperationFailed)
	}
	if lineNumber < 0 {
		return nil, fmt.Errorf("%w: line number cannot be negative", ErrInvalidValue)
	}

	defaultOpts := []gpiocdev.LineReqOption{gpiocdev.WithConsumer("u-bmc")}
	gpiocdevOpts := append(defaultOpts, convertOptions(opts)...)

	line, err := gpiocdev.RequestLine(chip, lineNumber, gpiocdevOpts...)
	if err != nil {
		return nil, mapGpiocdevError(err, fmt.Sprintf("failed to request line %d from chip '%s'", lineNumber, chip))
	}

	return line, nil
}

// RequestLines requests multiple GPIO lines at once.
// Accepts a slice of line offsets and configuration options.
// Returns a *gpiocdev.Lines that can be used directly with the underlying library.
func RequestLines(chip string, lineOffsets []int, opts ...Option) (*gpiocdev.Lines, error) {
	if chip == "" {
		return nil, fmt.Errorf("%w: chip path cannot be empty", ErrOperationFailed)
	}
	if len(lineOffsets) == 0 {
		return nil, fmt.Errorf("%w: at least one line offset must be specified", ErrInvalidValue)
	}

	for _, offset := range lineOffsets {
		if offset < 0 {
			return nil, fmt.Errorf("%w: line offset cannot be negative: %d", ErrInvalidValue, offset)
		}
	}

	defaultOpts := []gpiocdev.LineReqOption{gpiocdev.WithConsumer("u-bmc")}
	gpiocdevOpts := append(defaultOpts, convertOptions(opts)...)

	lines, err := gpiocdev.RequestLines(chip, lineOffsets, gpiocdevOpts...)
	if err != nil {
		return nil, mapGpiocdevError(err, fmt.Sprintf("failed to request lines from chip '%s'", chip))
	}

	return lines, nil
}

// ToggleGPIO performs a toggle operation on a GPIO line for the specified duration.
// This is equivalent to: set high, wait for duration, set low.
// The line is automatically closed after the operation.
func ToggleGPIO(chip, lineName string, duration time.Duration, opts ...Option) error {
	if duration <= 0 {
		return fmt.Errorf("%w: duration must be positive", ErrInvalidDuration)
	}

	line, err := RequestLine(chip, lineName, append(opts, AsOutput())...)
	if err != nil {
		return err
	}
	defer line.Close()

	if err := line.SetValue(1); err != nil {
		return fmt.Errorf("%w: failed to set GPIO high: %w", ErrOperationFailed, err)
	}

	time.Sleep(duration)

	if err := line.SetValue(0); err != nil {
		return fmt.Errorf("%w: failed to set GPIO low: %w", ErrOperationFailed, err)
	}

	return nil
}

// ToggleGPIOByNumber performs a toggle operation on a GPIO line by number for the specified duration.
// This is equivalent to: set high, wait for duration, set low.
// The line is automatically closed after the operation.
func ToggleGPIOByNumber(chip string, lineNumber int, duration time.Duration, opts ...Option) error {
	if duration <= 0 {
		return fmt.Errorf("%w: duration must be positive", ErrInvalidDuration)
	}

	line, err := RequestLineByNumber(chip, lineNumber, append(opts, AsOutput())...)
	if err != nil {
		return err
	}
	defer line.Close()

	if err := line.SetValue(1); err != nil {
		return fmt.Errorf("%w: failed to set GPIO high: %w", ErrOperationFailed, err)
	}

	time.Sleep(duration)

	if err := line.SetValue(0); err != nil {
		return fmt.Errorf("%w: failed to set GPIO low: %w", ErrOperationFailed, err)
	}

	return nil
}

// ToggleGPIOCtx performs a toggle operation with context support.
// The context can be used to cancel the operation during the wait period.
func ToggleGPIOCtx(ctx context.Context, chip, lineName string, duration time.Duration, opts ...Option) error {
	if duration <= 0 {
		return fmt.Errorf("%w: duration must be positive", ErrInvalidDuration)
	}

	line, err := RequestLine(chip, lineName, append(opts, AsOutput())...)
	if err != nil {
		return err
	}
	defer line.Close()

	if err := line.SetValue(1); err != nil {
		return fmt.Errorf("%w: failed to set GPIO high: %w", ErrOperationFailed, err)
	}

	select {
	case <-time.After(duration):
	case <-ctx.Done():
		_ = line.SetValue(0)
		return ctx.Err()
	}

	if err := line.SetValue(0); err != nil {
		return fmt.Errorf("%w: failed to set GPIO low: %w", ErrOperationFailed, err)
	}

	return nil
}

// SetGPIO sets a GPIO line to the specified value.
// The line is automatically closed after the operation.
func SetGPIO(chip, lineName string, value int, opts ...Option) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("%w: value must be 0 or 1", ErrInvalidValue)
	}

	line, err := RequestLine(chip, lineName, append(opts, AsOutputValue(value))...)
	if err != nil {
		return err
	}
	defer line.Close()

	return nil
}

// SetGPIOByNumber sets a GPIO line by number to the specified value.
// The line is automatically closed after the operation.
func SetGPIOByNumber(chip string, lineNumber, value int, opts ...Option) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("%w: value must be 0 or 1", ErrInvalidValue)
	}

	line, err := RequestLineByNumber(chip, lineNumber, append(opts, AsOutputValue(value))...)
	if err != nil {
		return err
	}
	defer line.Close()

	return nil
}

// GetGPIO reads the current value of a GPIO line.
// The line is automatically closed after the operation.
func GetGPIO(chip, lineName string, opts ...Option) (int, error) {
	line, err := RequestLine(chip, lineName, append(opts, AsInput())...)
	if err != nil {
		return 0, err
	}
	defer line.Close()

	value, err := line.Value()
	if err != nil {
		return 0, fmt.Errorf("%w: failed to read GPIO value: %w", ErrOperationFailed, err)
	}

	return value, nil
}

// GetGPIOByNumber reads the current value of a GPIO line by number.
// The line is automatically closed after the operation.
func GetGPIOByNumber(chip string, lineNumber int, opts ...Option) (int, error) {
	line, err := RequestLineByNumber(chip, lineNumber, append(opts, AsInput())...)
	if err != nil {
		return 0, err
	}
	defer line.Close()

	value, err := line.Value()
	if err != nil {
		return 0, fmt.Errorf("%w: failed to read GPIO value: %w", ErrOperationFailed, err)
	}

	return value, nil
}

// CreateToggleCallback creates a callback function that performs a GPIO toggle operation.
// This is useful for integration with state machines or other callback-based systems.
// The returned function can be called multiple times and will handle line management internally.
func CreateToggleCallback(chip, lineName string, duration time.Duration, opts ...Option) func(context.Context) error {
	return func(ctx context.Context) error {
		return ToggleGPIOCtx(ctx, chip, lineName, duration, opts...)
	}
}

// CreateToggleCallbackByNumber creates a callback function that performs a GPIO toggle operation by line number.
// This is useful for integration with state machines or other callback-based systems.
// The returned function can be called multiple times and will handle line management internally.
func CreateToggleCallbackByNumber(chip string, lineNumber int, duration time.Duration, opts ...Option) func(context.Context) error {
	return func(ctx context.Context) error {
		return ToggleGPIOByNumber(chip, lineNumber, duration, opts...)
	}
}

// CreateSetCallback creates a callback function that sets a GPIO to a specific value.
// The returned function can be called multiple times and will handle line management internally.
func CreateSetCallback(chip, lineName string, value int, opts ...Option) func(context.Context) error {
	return func(ctx context.Context) error {
		return SetGPIO(chip, lineName, value, opts...)
	}
}

// CreateSetCallbackByNumber creates a callback function that sets a GPIO by number to a specific value.
// The returned function can be called multiple times and will handle line management internally.
func CreateSetCallbackByNumber(chip string, lineNumber, value int, opts ...Option) func(context.Context) error {
	return func(ctx context.Context) error {
		return SetGPIOByNumber(chip, lineNumber, value, opts...)
	}
}

// PulseGPIO performs a single pulse on a GPIO line.
// This sets the line high for the specified duration, then low.
// Alias for ToggleGPIO for better semantic clarity.
func PulseGPIO(chip, lineName string, duration time.Duration, opts ...Option) error {
	return ToggleGPIO(chip, lineName, duration, opts...)
}

// PulseGPIOByNumber performs a single pulse on a GPIO line by number.
// This sets the line high for the specified duration, then low.
// Alias for ToggleGPIOByNumber for better semantic clarity.
func PulseGPIOByNumber(chip string, lineNumber int, duration time.Duration, opts ...Option) error {
	return ToggleGPIOByNumber(chip, lineNumber, duration, opts...)
}

// mapGpiocdevError maps gpiocdev errors to our package errors.
func mapGpiocdevError(err error, details string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, syscall.ENOENT):
		return fmt.Errorf("%w: %s", ErrChipNotFound, details)
	case errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM):
		return fmt.Errorf("%w: %s", ErrPermissionDenied, details)
	case errors.Is(err, gpiocdev.ErrNotFound):
		return fmt.Errorf("%w: %s", ErrLineNotFound, details)
	case errors.Is(err, gpiocdev.ErrClosed):
		return fmt.Errorf("%w: %s", ErrLineClosed, details)
	default:
		return fmt.Errorf("%w: %s: %w", ErrOperationFailed, details, err)
	}
}
