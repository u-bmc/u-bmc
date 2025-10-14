// SPDX-License-Identifier: BSD-3-Clause

// Package thermalmgr provides a thermal management service for BMC systems.
// This service implements PID-based temperature control and cooling device management
// through NATS-based IPC endpoints, integrating with hwmon interfaces for thermal regulation.
//
// # Overview
//
// The thermalmgr service is responsible for maintaining system temperatures within safe
// operating limits by controlling cooling devices such as fans, pumps, and other thermal
// management hardware. It operates autonomously using PID control loops and responds to
// thermal management commands via NATS messaging.
//
// Key responsibilities:
//   - Monitor temperature sensors and maintain thermal zones
//   - Control cooling devices through hwmon interfaces
//   - Implement PID-based thermal control loops
//   - Respond to thermal management commands via NATS IPC
//   - Communicate with sensormon for temperature readings
//   - Notify powermgr of critical thermal conditions
//   - Provide thermal zone and cooling device status information
//
// # Service Architecture
//
// The service follows the standard BMC service pattern:
//   - NATS-based IPC for inter-service communication
//   - Microservice endpoints for thermal management operations
//   - Context-aware operations with timeout support
//   - Structured logging with slog
//   - OpenTelemetry integration for observability
//
// # Thermal Management Model
//
// Thermal Zones: Logical groupings of temperature sensors and cooling devices that work
// together to maintain temperature targets. Each zone operates independently with its own
// PID controller and thermal thresholds.
//
// Cooling Devices: Hardware components that can reduce system temperature, including fans,
// water pumps, heat exchangers, and liquid coolers. Each device has controllable power
// levels and operational status.
//
// PID Control: Software-based control loops that continuously adjust cooling device output
// based on temperature error, using proportional, integral, and derivative terms for
// optimal thermal regulation.
//
// # NATS IPC Endpoints
//
// The service provides the following endpoints:
//
// Thermal Zone Management:
//   - thermalmgr.zones.list - List all thermal zones
//   - thermalmgr.zone.get - Get thermal zone information
//   - thermalmgr.zone.set - Update thermal zone configuration
//   - thermalmgr.zone.control - Control thermal zone operation
//
// Cooling Device Management:
//   - thermalmgr.devices.list - List all cooling devices
//   - thermalmgr.device.get - Get cooling device information
//   - thermalmgr.device.set - Update cooling device configuration
//   - thermalmgr.device.control - Control cooling device operation
//
// Thermal Control:
//   - thermalmgr.control.start - Start thermal management
//   - thermalmgr.control.stop - Stop thermal management
//   - thermalmgr.control.status - Get thermal management status
//   - thermalmgr.control.emergency - Handle emergency thermal conditions
//
// # Integration with Other Services
//
// Sensormon Integration:
// The thermalmgr service coordinates with sensormon to receive temperature readings
// and thermal threshold alerts. It subscribes to sensor data streams and responds
// to thermal events.
//
// Powermgr Integration:
// In critical thermal conditions, thermalmgr communicates with powermgr to request
// emergency shutdowns or power reduction when cooling alone is insufficient.
//
// Statemgr Integration:
// Thermal management operations are coordinated with system state to ensure proper
// thermal control during power transitions and operational state changes.
//
// # Configuration
//
// The service supports configuration of:
//   - Thermal zones and their associated sensors/devices
//   - PID controller parameters for each zone
//   - Temperature thresholds and emergency limits
//   - Cooling device mappings and power ranges
//   - Control loop timing and sample rates
//   - Emergency response procedures
//
// # Error Handling and Resilience
//
// The service implements robust error handling:
//   - Graceful degradation when sensors or devices fail
//   - Fallback to safe cooling levels during errors
//   - Automatic recovery and re-initialization
//   - Emergency thermal protection procedures
//   - Comprehensive logging of thermal events
//
// # Performance Characteristics
//
// The service is designed for real-time thermal management:
//   - Low-latency response to thermal events
//   - Efficient PID control loop execution
//   - Minimal resource usage for continuous operation
//   - Concurrent thermal zone management
//   - Non-blocking IPC communication
//
// # Safety and Reliability
//
// Critical safety features:
//   - Fail-safe cooling defaults
//   - Emergency thermal shutdown coordination
//   - Temperature monitoring redundancy
//   - Thermal runaway protection
//   - Hardware abstraction layer fault tolerance
package thermalmgr
