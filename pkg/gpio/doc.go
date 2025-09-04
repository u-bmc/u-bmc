// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

// Package gpio provides a high-level abstraction for GPIO operations in BMC environments.
//
// This package wraps the low-level gpio-cdev functionality and provides a more convenient
// and BMC-specific interface for common GPIO operations such as power control, reset
// operations, status indicators, and general I/O.
//
// # Key Concepts
//
// GPIO Chip: A GPIO controller that manages a collection of GPIO lines. In BMC systems,
// you typically have multiple GPIO chips (e.g., /dev/gpiochip0, /dev/gpiochip1).
//
// GPIO Line: An individual GPIO pin within a chip. Lines can be configured as inputs
// or outputs and may have additional properties like pull-up/pull-down resistors.
//
// Line Configuration: Each GPIO line can be configured with specific properties such as
// direction (input/output), initial value, bias (pull-up/pull-down), and edge detection.
//
// # Basic Usage
//
// The simplest way to use this package is through the Manager type:
//
//	manager := gpio.NewManager()
//	defer manager.Close()
//
//	// Configure a power button (momentary press)
//	powerBtn, err := manager.RequestLine("gpiochip0", "power-button",
//		gpio.WithDirection(gpio.DirectionOutput),
//		gpio.WithInitialValue(0),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Perform a 200ms button press
//	if err := powerBtn.Toggle(200 * time.Millisecond); err != nil {
//		log.Fatal(err)
//	}
//
// # Advanced Usage
//
// For more complex scenarios, you can configure multiple lines at once:
//
//	config := gpio.NewConfig(
//		gpio.WithChip("gpiochip0"),
//		gpio.WithLines(map[string]gpio.LineConfig{
//			"power-led": {
//				Direction: gpio.DirectionOutput,
//				InitialValue: 0,
//			},
//			"reset-button": {
//				Direction: gpio.DirectionOutput,
//				InitialValue: 0,
//			},
//			"power-status": {
//				Direction: gpio.DirectionInput,
//				Bias: gpio.BiasPullUp,
//			},
//		}),
//	)
//
//	lines, err := manager.RequestLines(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use the configured lines
//	powerLed := lines["power-led"]
//	resetBtn := lines["reset-button"]
//	powerStatus := lines["power-status"]
//
// # BMC Common Patterns
//
// Power Control:
//
//	// Momentary power button press
//	powerBtn.Toggle(200 * time.Millisecond)
//
//	// Hard power off (hold for 4 seconds)
//	powerBtn.SetValue(1)
//	time.Sleep(4 * time.Second)
//	powerBtn.SetValue(0)
//
// Status Monitoring:
//
//	// Read power status
//	powered, err := powerStatus.GetValue()
//	if err != nil {
//		log.Printf("Failed to read power status: %v", err)
//	}
//
// LED Control:
//
//	// Turn on status LED
//	statusLed.SetValue(1)
//
//	// Blink pattern
//	for i := 0; i < 5; i++ {
//		statusLed.SetValue(1)
//		time.Sleep(100 * time.Millisecond)
//		statusLed.SetValue(0)
//		time.Sleep(100 * time.Millisecond)
//	}
//
// # Event Monitoring
//
// The package supports edge detection for monitoring GPIO state changes:
//
//	powerBtn, err := manager.RequestLine("gpiochip0", "power-button-input",
//		gpio.WithDirection(gpio.DirectionInput),
//		gpio.WithEdge(gpio.EdgeFalling),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Monitor for button presses
//	events := powerBtn.Events()
//	for event := range events {
//		fmt.Printf("Power button pressed at %v\n", event.Timestamp)
//	}
//
// # Error Handling
//
// The package provides specific error types for different failure scenarios:
//
//	line, err := manager.RequestLine("gpiochip0", "non-existent-line")
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
// # Resource Management
//
// Always ensure proper cleanup of GPIO resources:
//
//	manager := gpio.NewManager()
//	defer manager.Close() // Closes all managed lines
//
//	// Or for individual lines
//	line, err := manager.RequestLine(...)
//	if err != nil {
//		return err
//	}
//	defer line.Close()
//
// # Thread Safety
//
// The Manager type is thread-safe and can be used concurrently from multiple goroutines.
// Individual Line instances are also thread-safe for concurrent read/write operations.
//
// # Platform Considerations
//
// This package is designed for Linux systems with GPIO character device support
// (/dev/gpiochipN). Ensure your kernel has CONFIG_GPIO_CDEV enabled and that
// your user has appropriate permissions to access GPIO devices.
//
// Common BMC platforms supported:
//   - ASPEED AST2400/AST2500/AST2600
//   - Nuvoton NPCM7xx
//   - Raspberry Pi (for development/testing)
//   - Generic Linux systems with GPIO character device support
package gpio
