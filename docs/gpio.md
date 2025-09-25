# GPIO Package

A simplified Go package for GPIO operations in BMC environments, providing a thin wrapper around the `gpio-cdev` library.

## Design Philosophy

This package avoids reinventing functionality that the underlying `gpio-cdev` library already provides well. Instead, it focuses on:

- Simple convenience functions for common BMC operations
- Configuration helpers that map directly to `gpio-cdev` options
- Stateless callback functions for integration with other systems
- Direct access to underlying `gpio-cdev` functionality when needed

## Key Features

- **Stateless Operations**: Most functions handle line management internally and clean up automatically
- **Callback Functions**: Create reusable callbacks for state machines and event-driven systems
- **Direct Library Access**: When you need more control, get direct access to `*gpiocdev.Line` objects
- **BMC-Focused**: Designed specifically for common BMC use cases like power control and status monitoring

## Quick Examples

### Simple GPIO Toggle (like gpioset --mode=time)

```go
// Toggle a power button for 200ms
err := gpio.ToggleGPIO("/dev/gpiochip0", "power-button", 200*time.Millisecond)
```

### Set GPIO Value

```go
// Turn on a status LED
err := gpio.SetGPIO("/dev/gpiochip0", "status-led", 1)
```

### Read GPIO Value

```go
// Check power status
powered, err := gpio.GetGPIO("/dev/gpiochip0", "power-good")
```

### Callback Functions for State Machines

```go
// Create reusable callbacks
powerOn := gpio.CreateToggleCallback("/dev/gpiochip0", "power-button",
    200*time.Millisecond, gpio.AsOutput())

forceOff := gpio.CreateToggleCallback("/dev/gpiochip0", "power-button",
    4*time.Second, gpio.AsOutput())

// Use in your state machine
err := powerOn(ctx)
```

### Direct Library Access

```go
// Get direct access to gpiocdev.Line for advanced operations
line, err := gpio.RequestLine("/dev/gpiochip0", "complex-gpio",
    gpio.AsInput(),
    gpio.WithPullUp(),
    gpio.WithEdgeDetection(gpio.EdgeBoth))
defer line.Close()

// Use gpiocdev.Line methods directly
value, err := line.Value()
```

## Configuration Options

All configuration options map directly to `gpio-cdev` options:

- `gpio.AsInput()` - Configure as input
- `gpio.AsOutput()` - Configure as output (initial value 0)
- `gpio.AsOutputValue(1)` - Configure as output with specific initial value
- `gpio.WithPullUp()` - Enable internal pull-up resistor
- `gpio.WithPullDown()` - Enable internal pull-down resistor
- `gpio.WithActiveLow()` - Configure as active-low
- `gpio.WithOpenDrain()` - Configure as open-drain output
- `gpio.WithOpenSource()` - Configure as open-source output
- `gpio.WithEdgeDetection(gpio.EdgeBoth)` - Enable edge detection
- `gpio.WithDebounce(50*time.Millisecond)` - Set debounce period
- `gpio.WithConsumer("my-app")` - Set consumer name

## Migration from Complex GPIO Managers

If you were previously using a complex GPIO manager with line caching and state management, this package encourages a simpler approach:

**Before:**
```go
manager := gpio.NewManager()
defer manager.Close()

line, err := manager.RequestLine("chip", "line", options...)
defer line.Close()

err = line.SetValue(1)
```

**After:**
```go
// For simple operations, use convenience functions
err := gpio.SetGPIO("/dev/gpiochip0", "line", 1)

// For reusable operations, use callbacks
setter := gpio.CreateSetCallback("/dev/gpiochip0", "line", 1)
err := setter(ctx)

// For complex operations, use direct library access
line, err := gpio.RequestLine("/dev/gpiochip0", "line", gpio.AsOutput())
defer line.Close()
// Use line directly with gpiocdev methods
```

## Error Handling

The package provides specific error types for common GPIO failures:

- `ErrChipNotFound` - GPIO chip device not found
- `ErrLineNotFound` - GPIO line not found on chip
- `ErrPermissionDenied` - Insufficient permissions
- `ErrInvalidValue` - Invalid GPIO value (not 0 or 1)
- `ErrInvalidDuration` - Invalid duration for timed operations
- `ErrOperationFailed` - Generic operation failure
- `ErrLineClosed` - Operation on closed line

## Platform Support

Requires Linux with GPIO character device support (`/dev/gpiochipN`). Tested on:

- ASPEED AST2400/AST2500/AST2600
- Nuvoton NPCM7xx
- Raspberry Pi (development/testing)
- Generic Linux systems with CONFIG_GPIO_CDEV

## When to Use What

1. **Simple one-off operations**: Use convenience functions like `ToggleGPIO()`, `SetGPIO()`, `GetGPIO()`
2. **Repeated operations**: Use callback functions like `CreateToggleCallback()`
3. **Complex operations**: Use `RequestLine()` for direct `gpiocdev.Line` access
4. **Event monitoring**: Use `RequestLine()` with edge detection options
5. **Bulk operations**: Use `RequestLines()` for multiple lines at once

This approach provides the simplicity of convenience functions while maintaining access to the full power of the underlying `gpio-cdev` library when needed.
