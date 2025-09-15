// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

// Package gpio provides a simplified abstraction for GPIO operations in BMC environments.
//
// This package provides a thin wrapper around the gpio-cdev library, focusing on common
// BMC use cases while avoiding reinventing functionality that the underlying library
// already provides well.
//
// # Design Philosophy
//
// Rather than creating complex abstractions, this package provides:
//   - Simple convenience functions for common operations
//   - Configuration helpers that map to gpio-cdev options
//   - Stateless callback functions for integration with other systems
//   - Direct access to underlying gpio-cdev functionality when needed
//
// # Basic Usage
//
// For simple GPIO operations, use the convenience functions:
//
//	// Toggle a GPIO for 200ms (like gpioset --mode=time)
//	err := gpio.ToggleGPIO("/dev/gpiochip0", "power-button", 200*time.Millisecond)
//	if err != nil {
//		log.Printf("Failed to toggle GPIO: %v", err)
//	}
//
//	// Set a GPIO to a specific value
//	err = gpio.SetGPIO("/dev/gpiochip0", "reset-line", 1)
//	if err != nil {
//		log.Printf("Failed to set GPIO: %v", err)
//	}
//
//	// Read a GPIO value
//	value, err := gpio.GetGPIO("/dev/gpiochip0", "power-status")
//	if err != nil {
//		log.Printf("Failed to read GPIO: %v", err)
//	}
//
// # Callback Functions
//
// For integration with state machines or other callback-based systems:
//
//	// Create a toggle callback
//	togglePower := gpio.CreateToggleCallback("/dev/gpiochip0", "power-button",
//		200*time.Millisecond, gpio.AsOutput())
//
//	// Use in a state machine or other system
//	err := togglePower(ctx)
//	if err != nil {
//		log.Printf("Power toggle failed: %v", err)
//	}
//
// # Advanced Usage
//
// For more complex scenarios, use the wrapper functions with options:
//
//	// Request a line with specific configuration
//	line, err := gpio.RequestLine("/dev/gpiochip0", "status-led",
//		gpio.AsOutput(),
//		gpio.WithInitialValue(0),
//		gpio.WithConsumer("status-controller"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer line.Close()
//
//	// Use the line directly (it's a *gpiocdev.Line)
//	err = line.SetValue(1)
//
// # Configuration Options
//
// The package provides configuration options that map directly to gpio-cdev options:
//
//	gpio.AsInput()           // Configure as input
//	gpio.AsOutput()          // Configure as output with initial value 0
//	gpio.AsOutputValue(1)    // Configure as output with specific initial value
//	gpio.WithPullUp()        // Enable internal pull-up resistor
//	gpio.WithPullDown()      // Enable internal pull-down resistor
//	gpio.WithActiveLow()     // Configure as active-low
//	gpio.WithConsumer("app") // Set consumer name
//	gpio.WithEdgeDetection(gpio.EdgeBoth) // Enable edge detection
//
// # BMC Power Control Example
//
//	// Power button toggle
//	powerToggle := gpio.CreateToggleCallback("/dev/gpiochip0", "power-btn",
//		200*time.Millisecond, gpio.AsOutput())
//
//	// Reset button toggle
//	resetToggle := gpio.CreateToggleCallback("/dev/gpiochip0", "reset-btn",
//		100*time.Millisecond, gpio.AsOutput())
//
//	// Power status reader
//	getPowerStatus := func() (bool, error) {
//		value, err := gpio.GetGPIO("/dev/gpiochip0", "power-good")
//		return value == 1, err
//	}
//
// # Error Handling
//
// The package provides specific error types for different failure scenarios:
//
//	err := gpio.ToggleGPIO("/dev/gpiochip0", "missing-line", time.Second)
//	if err != nil {
//		switch {
//		case errors.Is(err, gpio.ErrChipNotFound):
//			log.Fatal("GPIO chip not available")
//		case errors.Is(err, gpio.ErrLineNotFound):
//			log.Fatal("GPIO line not found")
//		case errors.Is(err, gpio.ErrPermissionDenied):
//			log.Fatal("Insufficient permissions for GPIO access")
//		default:
//			log.Fatalf("Unexpected error: %v", err)
//		}
//	}
//
// # Platform Support
//
// This package requires Linux with GPIO character device support (/dev/gpiochipN).
// Ensure your kernel has CONFIG_GPIO_CDEV enabled and appropriate permissions.
package gpio
