// SPDX-License-Identifier: BSD-3-Clause

// Package ledmgr provides LED management for BMC components including hosts, chassis, and management controllers.
//
// The LED manager service handles visual indicators for system status, power state, errors, and identification
// purposes. It operates as a backend service that receives LED control requests from statemgr via NATS IPC
// and executes the corresponding physical operations through configurable backends.
//
// # Architecture Overview
//
// The ledmgr service works in coordination with statemgr:
//  1. statemgr detects state changes (power on/off, errors, etc.)
//  2. statemgr sends LED control requests to ledmgr via NATS IPC
//  3. ledmgr executes the physical LED operations using backends
//  4. ledmgr responds with success/failure status
//  5. LEDs provide visual feedback to operators
//
// # LED Types
//
// The LED manager supports four main LED types for each component:
//   - Power LED: Indicates power state (on/off/transitioning)
//   - Status LED: Shows general operational status (ok/warning/critical)
//   - Error LED: Displays error conditions and faults
//   - Identify LED: Used for physical identification of components
//
// # LED States
//
// Standard LED states supported across all backends:
//   - Off: LED is turned off
//   - On: LED is solid on
//   - Blink: LED blinks at normal interval (500ms default)
//   - FastBlink: LED blinks rapidly for urgent conditions
//
// # Backend System
//
// The service supports configurable backends for LED control:
//
//	type LEDBackend interface {
//		SetLEDState(ctx context.Context, componentName string, ledType LEDType, state LEDState) error
//		GetLEDState(ctx context.Context, componentName string, ledType LEDType) (LEDState, error)
//		Initialize(ctx context.Context, config *config) error
//		Close() error
//	}
//
// Multiple backends can be implemented:
//   - GPIO Backend: Direct hardware control via GPIO lines (default)
//   - I2C Backend: Control via I2C LED controllers
//   - Custom Backend: Platform-specific control mechanisms
//
// # Basic Usage
//
// The LED manager is typically used as a service in BMC systems:
//
//	ledmgr := ledmgr.New(
//		ledmgr.WithServiceName("ledmgr"),
//		ledmgr.WithHostManagement(true),
//		ledmgr.WithChassisManagement(true),
//		ledmgr.WithBMCManagement(true),
//		ledmgr.WithDefaultBackend(ledmgr.BackendTypeGPIO),
//	)
//
//	// Run as part of BMC service framework
//	if err := ledmgr.Run(ctx, ipcConn); err != nil {
//		log.Fatal(err)
//	}
//
// # IPC Communication
//
// The service exposes NATS-based endpoints for LED control:
//
//	// LED control operations
//	ledmgr.{component}.{led_type}.control -> LEDControlRequest
//
//	// LED status queries
//	ledmgr.{component}.{led_type}.status -> LEDStatusRequest
//
// # GPIO Backend Implementation
//
// The default GPIO backend provides direct hardware control:
//
//	config := ledmgr.New(
//		ledmgr.WithGPIOChip("/dev/gpiochip0"),
//		ledmgr.WithComponents(map[string]ledmgr.ComponentConfig{
//			"host.0": {
//				LEDs: map[ledmgr.LEDType]ledmgr.LEDConfig{
//					ledmgr.LEDTypePower: {
//						Backend: ledmgr.BackendTypeGPIO,
//						GPIO: ledmgr.LEDGPIOConfig{
//							Line: "power-led-0",
//							ActiveState: ledmgr.ActiveHigh,
//						},
//					},
//				},
//			},
//		}),
//	)
//
// # I2C Backend Implementation
//
// The I2C backend supports LED controllers:
//
//	config := ledmgr.New(
//		ledmgr.WithI2CDevice("/dev/i2c-0"),
//		ledmgr.WithComponents(map[string]ledmgr.ComponentConfig{
//			"host.0": {
//				LEDs: map[ledmgr.LEDType]ledmgr.LEDConfig{
//					ledmgr.LEDTypePower: {
//						Backend: ledmgr.BackendTypeI2C,
//						I2C: ledmgr.LEDI2CConfig{
//							DevicePath: "/dev/i2c-0",
//							SlaveAddress: 0x20,
//							Register: 0x01,
//							OnValue: 0xFF,
//							OffValue: 0x00,
//							BlinkValue: 0x55,
//						},
//					},
//				},
//			},
//		}),
//	)
//
// # LED Control Patterns
//
// Power State Indication:
//
//	// Power on - solid green
//	SetLEDState(ctx, "host.0", LEDTypePower, LEDStateOn)
//
//	// Power off - LED off
//	SetLEDState(ctx, "host.0", LEDTypePower, LEDStateOff)
//
//	// Power transitioning - blinking
//	SetLEDState(ctx, "host.0", LEDTypePower, LEDStateBlink)
//
// Error Indication:
//
//	// Critical error - fast blinking red
//	SetLEDState(ctx, "host.0", LEDTypeError, LEDStateFastBlink)
//
//	// Warning - slow blinking amber
//	SetLEDState(ctx, "host.0", LEDTypeStatus, LEDStateBlink)
//
// Identification:
//
//	// Identify component - blinking blue
//	SetLEDState(ctx, "host.0", LEDTypeIdentify, LEDStateBlink)
//
// # Custom Backend Implementation
//
// Implement the LEDBackend interface for custom control mechanisms:
//
//	type CustomLEDBackend struct {
//		// custom fields
//	}
//
//	func (b *CustomLEDBackend) SetLEDState(ctx context.Context, componentName string, ledType LEDType, state LEDState) error {
//		// implement custom LED control logic
//		return nil
//	}
//
//	func (b *CustomLEDBackend) GetLEDState(ctx context.Context, componentName string, ledType LEDType) (LEDState, error) {
//		// implement custom LED status reading
//		return LEDStateOff, nil
//	}
//
//	// ... implement other methods
//
// # Error Handling
//
// The package provides specific error types for LED operations:
//
//	err := backend.SetLEDState(ctx, "host.0", LEDTypePower, LEDStateOn)
//	if err != nil {
//		switch {
//		case errors.Is(err, ledmgr.ErrComponentNotFound):
//			log.Error("Host not configured")
//		case errors.Is(err, ledmgr.ErrLEDOperationFailed):
//			log.Error("Hardware LED operation failed")
//		case errors.Is(err, ledmgr.ErrBackendNotSupported):
//			log.Error("Backend doesn't support this operation")
//		case errors.Is(err, ledmgr.ErrInvalidLEDState):
//			log.Error("Invalid LED state requested")
//		default:
//			log.Errorf("Unexpected error: %v", err)
//		}
//	}
//
// # Integration with State Manager
//
// The LED manager works seamlessly with the state manager:
//
//  1. Host powers on via statemgr API
//  2. statemgr transitions host to ON state
//  3. statemgr sends LED control request to ledmgr
//  4. ledmgr sets power LED to solid on
//  5. Operator sees visual confirmation of power state
//
// # Resource Management
//
// Always ensure proper cleanup of LED manager resources:
//
//	ledmgr := ledmgr.New(config...)
//	defer ledmgr.Close()
//
//	// GPIO/I2C resources are automatically managed
//	// IPC connections are cleaned up on context cancellation
//
// # Thread Safety
//
// The LEDMgr service is thread-safe and can handle concurrent operations.
// Individual LED operations are serialized per component to prevent conflicts.
//
// # Platform Considerations
//
// This package is designed for BMC systems with LED control capabilities:
//
// Supported Platforms:
//   - ASPEED AST2400/AST2500/AST2600 BMCs
//   - Nuvoton NPCM7xx BMCs
//   - Standard BMCs with GPIO or I2C LED controllers
//   - Custom BMC implementations
//
// Requirements:
//   - GPIO character device support (/dev/gpiochipN) for GPIO backend
//   - I2C device support (/dev/i2c-N) for I2C backend
//   - Appropriate hardware connections (LED circuits)
//   - Proper electrical design (current limiting, protection)
//
// # Performance Considerations
//
// LED operations are typically very fast (microseconds to milliseconds):
//
//   - GPIO operations: ~1-10µs per operation
//   - I2C operations: ~100µs-1ms per operation
//   - Blink timing: Controlled by software timers
//
// The service is optimized for:
//   - Low latency LED control
//   - Reliable operation under high load
//   - Graceful handling of hardware failures
//   - Minimal resource usage
//   - Consistent visual feedback
package ledmgr
