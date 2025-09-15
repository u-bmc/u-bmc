// SPDX-License-Identifier: BSD-3-Clause

// Package powermgr provides physical power management operations for BMC systems.
//
// This package handles the execution of power-related operations on BMC components
// including hosts, chassis, and management controllers. It operates as a backend
// service that receives power action requests from statemgr via NATS IPC and
// executes the corresponding physical operations through configurable backends.
//
// # Architecture Overview
//
// The powermgr service works in coordination with statemgr:
//  1. statemgr receives API requests and manages state transitions
//  2. statemgr sends power action requests to powermgr via NATS IPC
//  3. powermgr executes the physical power operations using backends
//  4. powermgr responds with success/failure status
//  5. statemgr updates component state based on the response
//
// # Component Types
//
// The power manager supports three main component types, each handled in separate files:
//   - Host: Server/compute nodes that can be powered on, off, or reset (host.go)
//   - Chassis: Physical enclosures with power control capabilities (chassis.go)
//   - Management Controller: BMC itself with reset and power management features (bmc.go)
//
// # Power Operations
//
// Standard power operations are mapped from existing protobuf actions:
//   - ChassisAction: ON, OFF, POWER_CYCLE, EMERGENCY_SHUTDOWN
//   - HostAction: ON, OFF, REBOOT, FORCE_OFF, FORCE_RESTART
//   - ManagementControllerAction: REBOOT, WARM_RESET, COLD_RESET, HARD_RESET, FACTORY_RESET
//
// # Backend System
//
// The service supports configurable backends for power control:
//
//	type PowerBackend interface {
//		PowerOn(ctx context.Context, componentName string) error
//		PowerOff(ctx context.Context, componentName string, force bool) error
//		Reset(ctx context.Context, componentName string) error
//		GetPowerStatus(ctx context.Context, componentName string) (bool, error)
//	}
//
// Multiple backends can be implemented:
//   - GPIO Backend: Direct hardware control via GPIO lines (default)
//   - IPMI Backend: Standards-based power control (future)
//   - Custom Backend: Platform-specific control mechanisms
//
// # Basic Usage
//
// The power manager is typically used as a service in BMC systems:
//
//	powermgr := powermgr.New(
//		powermgr.WithServiceName("powermgr"),
//		powermgr.WithHostManagement(true),
//		powermgr.WithChassisManagement(true),
//		powermgr.WithBMCManagement(true),
//	)
//
//	// Run as part of BMC service framework
//	if err := powermgr.Run(ctx, ipcConn); err != nil {
//		log.Fatal(err)
//	}
//
// # IPC Communication
//
// The service exposes NATS-based endpoints that receive existing protobuf messages:
//
//	// Host power operations
//	powermgr.host.{id}.action -> ChangeHostStateRequest
//
//	// Chassis power operations
//	powermgr.chassis.{id}.action -> ChangeChassisStateRequest
//
//	// BMC power operations
//	powermgr.bmc.{id}.action -> ChangeManagementControllerStateRequest
//
// # GPIO Backend Implementation
//
// The default GPIO backend provides direct hardware control:
//
//	config := powermgr.NewConfig(
//		powermgr.WithGPIOChip("/dev/gpiochip0"),
//		powermgr.WithComponents(map[string]powermgr.ComponentConfig{
//			"host.0": {
//				GPIO: powermgr.GPIOConfig{
//					PowerButton: powermgr.GPIOLineConfig{
//						Line: "power-button-0",
//						ActiveState: gpio.ActiveLow,
//					},
//					ResetButton: powermgr.GPIOLineConfig{
//						Line: "reset-button-0",
//						ActiveState: gpio.ActiveLow,
//					},
//					PowerStatus: powermgr.GPIOLineConfig{
//						Line: "power-good-0",
//						Direction: gpio.DirectionInput,
//					},
//				},
//			},
//		}),
//	)
//
// # Power Control Patterns
//
// Momentary Button Press (Soft Power):
//
//	// 200ms pulse on power button line
//	powerButton.Toggle(200 * time.Millisecond)
//
// Force Power Off (Hard Power):
//
//	// Hold power button for 4+ seconds
//	powerButton.Hold(ctx, 4 * time.Second)
//
// Reset Operation:
//
//	// Brief pulse on reset line
//	resetButton.Toggle(100 * time.Millisecond)
//
// Power Status Reading:
//
//	// Read power-good signal
//	powered, err := powerStatus.GetValue()
//
// # Custom Backend Implementation
//
// Implement the PowerBackend interface for custom control mechanisms:
//
//	type CustomBackend struct {
//		// custom fields
//	}
//
//	func (b *CustomBackend) PowerOn(ctx context.Context, componentName string) error {
//		// implement custom power on logic
//		return nil
//	}
//
//	func (b *CustomBackend) PowerOff(ctx context.Context, componentName string, force bool) error {
//		// implement custom power off logic
//		return nil
//	}
//
//	// ... implement other methods
//
// # Error Handling
//
// The package provides specific error types for power operations:
//
//	err := backend.PowerOn(ctx, "host.0")
//	if err != nil {
//		switch {
//		case errors.Is(err, powermgr.ErrComponentNotFound):
//			log.Error("Host not configured")
//		case errors.Is(err, powermgr.ErrPowerOperationFailed):
//			log.Error("Hardware power operation failed")
//		case errors.Is(err, powermgr.ErrBackendNotSupported):
//			log.Error("Backend doesn't support this operation")
//		case errors.Is(err, powermgr.ErrCallbackFailed):
//			log.Error("Callback function failed")
//		default:
//			log.Errorf("Unexpected error: %v", err)
//		}
//	}
//
// # Integration with State Manager
//
// The power manager works seamlessly with the state manager:
//
//  1. API client sends ChangeHostStateRequest to statemgr
//  2. statemgr validates state transition (OFF -> ON)
//  3. statemgr forwards ChangeHostStateRequest to powermgr
//  4. powermgr executes physical power on operation
//  5. powermgr responds with success/failure
//  6. statemgr updates host state to ON or ERROR
//
// # Resource Management
//
// Always ensure proper cleanup of power manager resources:
//
//	powermgr := powermgr.New(config...)
//	defer powermgr.Close()
//
//	// GPIO resources are automatically managed
//	// IPC connections are cleaned up on context cancellation
//
// # Thread Safety
//
// The PowerMgr service is thread-safe and can handle concurrent operations.
// Individual power operations are serialized per component to prevent conflicts.
//
// # Platform Considerations
//
// This package is designed for BMC systems with hardware power control capabilities:
//
// Supported Platforms:
//   - ASPEED AST2400/AST2500/AST2600 BMCs
//   - Nuvoton NPCM7xx BMCs
//   - Standard IPMI-compliant BMCs
//   - Custom BMC implementations with GPIO or equivalent control
//
// Requirements:
//   - GPIO character device support (/dev/gpiochipN)
//   - Appropriate hardware connections (power/reset buttons)
//   - Proper electrical design (isolation, protection)
//   - Platform-specific power control capabilities
//
// # Performance Considerations
//
// Power operations are typically fast (milliseconds) but some operations
// require longer timing:
//
//   - Soft power: ~200ms button press
//   - Hard power: ~4s button hold
//   - Reset: ~100ms pulse
//   - Status reading: Real-time
//
// The service is optimized for:
//   - Low latency power control
//   - Reliable operation under high load
//   - Graceful handling of hardware failures
//   - Minimal resource usage
package powermgr
