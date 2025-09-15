// SPDX-License-Identifier: BSD-3-Clause

// Package powermgr provides physical power management operations for BMC systems.
//
// This package handles the execution of power-related operations on BMC components
// including hosts, chassis, and management controllers. It operates independently
// of state management, focusing solely on the physical execution of power commands
// received via NATS IPC.
//
// # Key Concepts
//
// Component Types: The power manager supports three main component types:
//   - Host: Server/compute nodes that can be powered on, off, or reset
//   - Chassis: Physical enclosures with power control capabilities
//   - Management Controller: BMC itself with reset and power management features
//
// Power Operations: Standard power operations include:
//   - Power On: Bring a component to a powered state
//   - Power Off: Gracefully or forcefully power down a component
//   - Reset: Restart a component (warm, cold, hard reset variants)
//   - Emergency Shutdown: Immediate power termination for safety
//
// Power Control Mechanisms: The service supports multiple physical control methods:
//   - GPIO: Direct hardware control via GPIO lines (primary method)
//   - IPMI: Standards-based power control (future extension)
//   - Custom: Platform-specific control mechanisms
//
// Power Monitoring: Real-time power monitoring and control:
//   - Power consumption reading
//   - Power capping enforcement
//   - Thermal management integration
//   - Efficiency monitoring
//
// # Basic Usage
//
// The power manager is typically used as a service in BMC systems:
//
//	powermgr := powermgr.New(
//		powermgr.WithServiceName("powermgr"),
//		powermgr.WithHostSupport(true),
//		powermgr.WithChassisSupport(true),
//		powermgr.WithBMCSupport(true),
//	)
//
//	// Run as part of BMC service framework
//	if err := powermgr.Run(ctx, ipcConn); err != nil {
//		log.Fatal(err)
//	}
//
// # IPC Communication
//
// The service exposes NATS-based endpoints for power operations:
//
//	// Host power operations
//	powermgr.host.{id}.power.on
//	powermgr.host.{id}.power.off
//	powermgr.host.{id}.power.reset
//	powermgr.host.{id}.power.status
//	powermgr.host.{id}.power.consumption
//	powermgr.host.{id}.power.cap
//
//	// Chassis power operations
//	powermgr.chassis.{id}.power.on
//	powermgr.chassis.{id}.power.off
//	powermgr.chassis.{id}.power.status
//	powermgr.chassis.{id}.power.consumption
//
//	// BMC power operations
//	powermgr.bmc.{id}.power.reset
//	powermgr.bmc.{id}.power.status
//
// # GPIO Integration
//
// The primary control mechanism uses GPIO for physical power operations:
//
//	config := powermgr.NewConfig(
//		powermgr.WithGPIOChip("gpiochip0"),
//		powermgr.WithHostGPIOConfig(map[string]powermgr.GPIOConfig{
//			"host.0": {
//				PowerButton: powermgr.GPIOLineConfig{
//					Line: "power-button-0",
//					ActiveState: gpio.ActiveLow,
//				},
//				ResetButton: powermgr.GPIOLineConfig{
//					Line: "reset-button-0",
//					ActiveState: gpio.ActiveLow,
//				},
//				PowerStatus: powermgr.GPIOLineConfig{
//					Line: "power-good-0",
//					Direction: gpio.DirectionInput,
//				},
//			},
//		}),
//	)
//
// # Power Control Patterns
//
// Momentary Button Press (Soft Power):
//
//	// Triggered via IPC, implemented with GPIO
//	// 200ms pulse on power button line
//	powerButton.Toggle(200 * time.Millisecond)
//
// Force Power Off (Hard Power):
//
//	// Hold power button for 4+ seconds
//	powerButton.SetValue(1)
//	time.Sleep(4 * time.Second)
//	powerButton.SetValue(0)
//
// Reset Operation:
//
//	// Brief pulse on reset line
//	resetButton.Toggle(100 * time.Millisecond)
//
// Power Status Monitoring:
//
//	// Read power-good signal
//	powered, err := powerStatus.GetValue()
//
// # Power Capping
//
// The service supports dynamic power capping for energy management:
//
//	// Set power cap for a host
//	req := &schemav1alpha1.SetPowerCapRequest{
//		ComponentName: "host.0",
//		CapWatts: 300,
//		CapDuration: durationpb.New(time.Hour),
//	}
//
//	// Monitor power consumption
//	consumption, err := powermgr.GetPowerConsumption("host.0")
//
// # Extensible Architecture
//
// The power manager supports multiple control backends:
//
//	// GPIO-based control (primary)
//	gpioBackend := powermgr.NewGPIOBackend(gpioManager)
//
//	// IPMI-based control (future)
//	ipmiBackend := powermgr.NewIPMIBackend(ipmiClient)
//
//	// Custom platform control
//	customBackend := powermgr.NewCustomBackend(platformAPI)
//
//	powermgr := powermgr.New(
//		powermgr.WithBackends(gpioBackend, ipmiBackend, customBackend),
//	)
//
// # Error Handling
//
// The package provides specific error types for power operations:
//
//	err := powermgr.PowerOn("host.0")
//	if err != nil {
//		switch {
//		case errors.Is(err, powermgr.ErrComponentNotFound):
//			log.Error("Host not configured")
//		case errors.Is(err, powermgr.ErrPowerOperationFailed):
//			log.Error("Hardware power operation failed")
//		case errors.Is(err, powermgr.ErrPowerCapExceeded):
//			log.Error("Operation would exceed power limits")
//		case errors.Is(err, powermgr.ErrSafetyInterlock):
//			log.Error("Safety system prevented operation")
//		default:
//			log.Errorf("Unexpected error: %v", err)
//		}
//	}
//
// # Safety Features
//
// Built-in safety mechanisms protect hardware:
//
//	// Thermal protection
//	if temperature > thermalLimit {
//		return powermgr.ErrThermalProtection
//	}
//
//	// Power supply protection
//	if totalPower > supplyCapacity {
//		return powermgr.ErrPowerSupplyOverload
//	}
//
//	// Interlock checking
//	if !safetyInterlockOK {
//		return powermgr.ErrSafetyInterlock
//	}
//
// # Integration with State Manager
//
// The power manager works in coordination with the state manager:
//
//  1. State manager receives API requests
//  2. State manager validates state transitions
//  3. State manager calls power manager via IPC
//  4. Power manager executes physical operation
//  5. Power manager reports completion/failure
//  6. State manager updates component state
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
//   - Platform-specific power monitoring capabilities
//
// # Performance Considerations
//
// Power operations are typically fast (milliseconds) but some operations
// require longer timing:
//
//   - Soft power: ~200ms button press
//   - Hard power: ~4s button hold
//   - Reset: ~100ms pulse
//   - Power monitoring: Real-time readings
//   - Power capping: Dynamic adjustment
//
// The service is optimized for:
//   - Low latency power control
//   - Reliable operation under high load
//   - Graceful handling of hardware failures
//   - Minimal resource usage
package powermgr
