# Power Manager Service

The power manager service (powermgr) provides comprehensive power control and sequencing capabilities for u-bmc systems including hosts, chassis components, and BMC subsystems. It manages GPIO-based power controls, monitors power states through various interfaces, and coordinates with other services to ensure safe power operations. The power manager serves as the primary interface between logical power management requests and physical hardware control mechanisms.

## Overview

Power management in BMC systems requires careful coordination of multiple power domains, proper sequencing to prevent damage to sensitive components, and integration with monitoring systems to ensure reliable operation. The powermgr service centralizes these responsibilities while providing a clean interface for other services and external management tools.

The service abstracts the complexity of different power control mechanisms including GPIO-based controls for power buttons, reset signals, and power-good monitoring, I2C and PMBus interfaces for advanced power supply management, and integration with platform-specific power sequencing requirements. This abstraction enables consistent power management behavior across different hardware platforms while supporting the specific requirements of each deployment.

Power operations are coordinated with thermal management to prevent unsafe power-on operations when cooling is inadequate, state management to ensure consistent system state tracking, and sensor monitoring to incorporate health information into power management decisions. The service also implements comprehensive safety mechanisms including power-good monitoring, overcurrent protection integration, and emergency shutdown capabilities for thermal and hardware fault conditions.

## Architecture

The powermgr service implements a layered architecture that separates physical hardware control from logical power management operations. This separation enables support for diverse hardware platforms while maintaining consistent behavior and safety properties across different deployments.

### Hardware Abstraction Layer

The hardware abstraction layer provides uniform interfaces for different types of power control hardware. GPIO controls handle basic power button, reset button, and power-good signal operations using the pkg/gpio package for reliable hardware access. I2C and PMBus interfaces enable advanced power supply monitoring and control through standardized protocols supported by modern power management hardware.

Platform-specific control mechanisms can be integrated through pluggable hardware adapters that implement standard interfaces while providing access to unique hardware capabilities. This design enables support for specialized power management hardware while maintaining compatibility with standard power management operations.

### Power Domain Management

Power domain management organizes power controls into logical groups that correspond to system components like individual hosts, chassis infrastructure, and BMC subsystems. Each power domain includes its own state tracking, sequencing rules, and safety policies that reflect the specific requirements of the managed component.

Power domains can have complex dependency relationships where some components must be powered before others, and power-down operations must follow specific sequences to prevent hardware damage or data loss. The power manager coordinates these dependencies while providing clear error handling when sequencing requirements cannot be satisfied.

### Safety and Monitoring

Comprehensive safety mechanisms protect hardware and ensure reliable operation during both normal operations and failure conditions. Power-good monitoring verifies that power operations complete successfully and that power remains stable during operation. Overcurrent and fault monitoring integrates with hardware protection systems to detect and respond to electrical faults.

Temperature monitoring integration prevents power operations when thermal conditions are unsafe, while health monitoring incorporates sensor data into power management decisions to avoid operations on failing hardware components.

## Power Control Mechanisms

The powermgr service supports multiple power control mechanisms that can be used individually or in combination based on platform requirements and available hardware capabilities.

### GPIO Power Control

GPIO-based power control provides the foundation for most BMC power management operations through direct control of power button, reset button, and power-good signals. The service uses the pkg/gpio package to provide reliable, low-latency access to GPIO hardware with appropriate error handling and safety mechanisms.

GPIO operations include momentary button presses for power-on and reset operations with configurable timing to match hardware requirements, power-good monitoring for reliable power state detection, and emergency power-off capabilities through force-off signals or relay controls.

The GPIO interface supports both active-high and active-low signal configurations, configurable timing parameters for different hardware requirements, and debouncing mechanisms to prevent spurious operations from electrical noise or mechanical switch bounce.

### Advanced Power Supply Control

Modern server and BMC hardware often includes advanced power supply management capabilities through I2C and PMBus interfaces. The powermgr service integrates with these interfaces to provide detailed power supply monitoring and control beyond basic GPIO operations.

Advanced power supply features include detailed power consumption monitoring with real-time current, voltage, and power measurements, temperature monitoring for power supply health assessment, fault detection and reporting for electrical and thermal failures, and efficiency optimization through dynamic power supply configuration.

PMBus integration enables sophisticated power management policies including power limiting and budgeting across multiple power supplies, automatic load balancing and redundancy management, and predictive failure detection based on power supply health monitoring.

### Platform-Specific Integration

Different hardware platforms may require specialized power control mechanisms that go beyond standard GPIO and PMBus interfaces. The powermgr service supports platform-specific integration through configurable hardware adapters that can implement custom power control logic while maintaining compatibility with standard power management operations.

Platform-specific features might include complex power sequencing requirements for multi-node systems, integration with platform management controllers or service processors, and support for specialized power control hardware like intelligent power distribution units or rack-level power management systems.

## Power Sequencing and Coordination

Safe power operations require careful sequencing and coordination between different system components and power domains. The powermgr service implements comprehensive sequencing mechanisms that ensure proper operation order while providing flexibility for different platform requirements.

### Startup Sequencing

Power-on operations follow carefully designed sequences that ensure components are powered in the correct order with appropriate timing between steps. Startup sequencing typically begins with chassis infrastructure including power supplies, cooling systems, and shared resources, followed by individual compute nodes with proper spacing to prevent power supply overload, and finally auxiliary systems and management interfaces.

Each step in the startup sequence includes validation checks to ensure that previous steps completed successfully before proceeding. Power-good monitoring verifies that power is stable before continuing, while thermal monitoring ensures that cooling systems are operational before powering components that generate significant heat.

Startup sequencing can be customized per platform to accommodate specific hardware requirements while maintaining safety properties and providing clear error handling when sequencing cannot be completed successfully.

### Shutdown Sequencing

Power-down operations reverse the startup sequence while ensuring that data integrity and hardware safety are maintained throughout the process. Shutdown sequencing begins with graceful shutdown of operating systems and applications to prevent data loss, followed by coordinated power-down of compute nodes with proper timing to prevent power supply instability, and finally chassis infrastructure shutdown with appropriate safety checks.

Emergency shutdown procedures provide faster power-down capabilities when immediate shutdown is required for safety reasons, such as thermal emergencies or electrical fault conditions. Emergency procedures balance speed requirements with hardware safety to provide rapid response while minimizing the risk of hardware damage.

### Dependency Management

Complex systems often have intricate power dependencies where some components must be powered before others can operate safely. The powermgr service manages these dependencies through configurable dependency graphs that define the relationships between different power domains and components.

Dependency management ensures that prerequisite components are powered and stable before dependent components are started, coordinates shutdown operations to respect dependency relationships, and provides clear error reporting when dependency requirements cannot be satisfied due to hardware failures or configuration issues.

## Integration with System Services

The powermgr service integrates closely with other u-bmc services to provide coordinated power management that considers thermal, health, and operational requirements from across the system.

### State Manager Integration

Power operations coordinate closely with the state manager service to ensure that power state changes are properly reflected in system state tracking and that power operations are consistent with overall system state management policies. The integration ensures that power operations triggered through the power manager are properly coordinated with state transitions managed by the state manager.

State coordination includes validation of power operations against current system state and policies, coordination of complex operations that involve both power changes and state transitions, and consistent error handling when power operations cannot be completed due to state management constraints.

### Thermal Management Integration

Thermal conditions significantly impact safe power operation, and the powermgr service integrates closely with thermal management to ensure that power operations consider current and predicted thermal conditions. This integration prevents power-on operations when cooling capacity is insufficient and coordinates emergency power-off operations when thermal limits are exceeded.

Thermal integration includes pre-operation validation that checks thermal conditions before allowing power-on operations, continuous monitoring that can trigger power reduction or emergency shutdown when thermal conditions deteriorate, and coordination with thermal management policies that may require power operations for thermal protection.

### Sensor Monitoring Integration

Comprehensive sensor monitoring provides the health and environmental information necessary for intelligent power management decisions. The powermgr service integrates with sensor monitoring to incorporate real-time hardware health data into power management operations.

Sensor integration enables health-based power decisions that prevent operations on failing hardware, environmental monitoring that considers factors like ambient temperature and humidity in power management decisions, and predictive maintenance capabilities that can proactively manage power operations based on trending sensor data.

### Emergency Response Coordination

The powermgr service plays a critical role in emergency response scenarios where immediate power operations may be required to protect hardware or ensure safety. Emergency coordination enables rapid response to thermal emergencies, electrical fault conditions, and other critical situations that require immediate power management action.

Emergency response includes integration with monitoring systems that can trigger immediate power operations, coordination with other services to ensure that emergency operations are properly communicated and logged, and override capabilities that enable emergency operations even when normal safety interlocks might prevent them.

## Configuration

The powermgr service follows standard u-bmc configuration patterns with comprehensive options for hardware integration, safety policies, and operational parameters.

### Basic Configuration

```go
powerMgr := powermgr.New(
    powermgr.WithServiceName("powermgr"),
    powermgr.WithServiceDescription("BMC Power Management Service"),
    powermgr.WithGPIOChip("/dev/gpiochip0"),
    powermgr.WithOperationTimeout(30 * time.Second),
    powermgr.WithPowerGoodTimeout(5 * time.Second),
    powermgr.WithMetrics(true),
    powermgr.WithTracing(true),
)
```

### Hardware Configuration

```go
powerMgr := powermgr.New(
    powermgr.WithGPIOLines(map[string]powermgr.GPIOConfig{
        "host.0.power": {
            Line: "power-button-0",
            ActiveLow: false,
            PulseWidth: 200 * time.Millisecond,
        },
        "host.0.reset": {
            Line: "reset-button-0", 
            ActiveLow: true,
            PulseWidth: 100 * time.Millisecond,
        },
        "host.0.power-good": {
            Line: "power-good-0",
            ActiveLow: false,
            Direction: "input",
        },
    }),
    powermgr.WithI2CDevices(map[string]powermgr.I2CConfig{
        "psu.0": {
            Bus: "/dev/i2c-1",
            Address: 0x58,
            Protocol: "pmbus",
        },
    }),
)
```

### Power Domain Configuration

```go
powerMgr := powermgr.New(
    powermgr.WithPowerDomains([]powermgr.PowerDomain{
        {
            Name: "chassis",
            Type: powermgr.DomainTypeChassis,
            Controls: []string{"chassis.power", "chassis.reset"},
            PowerGood: []string{"chassis.power-good"},
            Dependencies: []string{},
            SequenceDelay: 2 * time.Second,
        },
        {
            Name: "host.0",
            Type: powermgr.DomainTypeHost,
            Controls: []string{"host.0.power", "host.0.reset"},
            PowerGood: []string{"host.0.power-good"},
            Dependencies: []string{"chassis"},
            SequenceDelay: 1 * time.Second,
        },
    }),
)
```

### Safety and Policy Configuration

```go
powerMgr := powermgr.New(
    powermgr.WithSafetyPolicies(powermgr.SafetyConfig{
        ThermalCheck: true,
        HealthCheck: true,
        PowerGoodRequired: true,
        MaxConcurrentOperations: 2,
        EmergencyOverride: false,
    }),
    powermgr.WithSequencingPolicies(powermgr.SequencingConfig{
        StartupDelay: 5 * time.Second,
        ShutdownDelay: 10 * time.Second,
        ForceOffDelay: 30 * time.Second,
        PowerGoodTimeout: 10 * time.Second,
        RetryAttempts: 3,
    }),
    powermgr.WithEmergencyPolicies(powermgr.EmergencyConfig{
        ThermalShutdown: true,
        FaultResponse: "immediate_off",
        OverrideTimeout: 60 * time.Second,
    }),
)
```

### Integration Configuration

```go
powerMgr := powermgr.New(
    powermgr.WithStateManagerIntegration("statemgr"),
    powermgr.WithThermalManagerIntegration("thermalmgr"),
    powermgr.WithSensorMonitorIntegration("sensormon"),
    powermgr.WithIntegrationTimeout(10 * time.Second),
    powermgr.WithEventPublishing(true),
    powermgr.WithAuditLogging(true),
)
```

## NATS Integration and Endpoints

The powermgr service provides comprehensive NATS endpoints for power management operations and integrates with other services through well-defined message patterns.

### Power Control Endpoints

Power control endpoints provide the primary interface for requesting power operations on system components. All power operations are subject to safety validation and coordination with other system services before execution.

```
powermgr.hosts.power.on       # Power on specific host
powermgr.hosts.power.off      # Power off specific host  
powermgr.hosts.reset          # Reset specific host
powermgr.chassis.power.on     # Power on chassis
powermgr.chassis.power.off    # Power off chassis
powermgr.system.emergency.off # Emergency system shutdown
```

### Status and Monitoring Endpoints

Status endpoints provide real-time information about power states, ongoing operations, and system health related to power management.

```
powermgr.hosts.status         # Get host power status
powermgr.chassis.status       # Get chassis power status
powermgr.system.status        # Get overall power system status
powermgr.operations.list      # List active power operations
powermgr.health.check         # Power system health check
```

### Configuration and Management Endpoints

Configuration endpoints enable dynamic management of power policies and operational parameters without requiring service restart.

```
powermgr.config.get           # Get current power configuration
powermgr.config.update        # Update power configuration
powermgr.policies.get         # Get current power policies
powermgr.policies.update      # Update power policies
powermgr.domains.list         # List configured power domains
```

## Power Operation Processing

Power operations in the powermgr service follow carefully designed patterns that ensure safety, reliability, and comprehensive coordination with other system components.

### Operation Validation

Before executing any power operation, comprehensive validation ensures that the operation is safe and appropriate given current system conditions. Validation includes checking current power state compatibility with requested operations, verifying that thermal conditions allow the requested operation, confirming that health monitoring indicates components are suitable for the operation, and ensuring that system policies and dependencies permit the operation.

Validation failures provide clear error messages that explain why the operation cannot be executed and guidance on resolving blocking conditions.

### Hardware Control Execution

Hardware control execution translates validated power operations into specific hardware control sequences. Execution includes appropriate timing and sequencing for hardware requirements, comprehensive error detection and reporting for hardware control failures, power-good monitoring to verify successful operation completion, and rollback capabilities when operations cannot be completed successfully.

Hardware control integrates with platform-specific hardware adapters to provide access to specialized power control mechanisms while maintaining consistent behavior and error handling across different platforms.

### Coordination and Notification

Power operations coordinate with other system services and provide comprehensive notification of operation status and completion. Coordination includes real-time status updates to the state manager service, thermal management notification for operations that affect thermal conditions, event publication for monitoring and audit systems, and error reporting that enables appropriate response by other system components.

Coordination mechanisms ensure that power operations are properly integrated with broader system management activities and that all relevant services are informed of power state changes that may affect their operations.

## Safety Mechanisms and Protection

The powermgr service implements comprehensive safety mechanisms that protect hardware and ensure reliable operation even during failure conditions or unexpected scenarios.

### Hardware Protection

Hardware protection mechanisms prevent operations that could damage components or compromise system reliability. Protection includes power sequencing validation that ensures proper operation order, overcurrent and fault monitoring integration with hardware protection systems, thermal interlock integration that prevents operations when thermal conditions are unsafe, and power-good monitoring that verifies stable power delivery before proceeding with dependent operations.

Hardware protection mechanisms can be configured per platform to accommodate different hardware requirements while maintaining essential safety properties across all deployments.

### Operational Safety

Operational safety mechanisms ensure that power operations behave predictably and safely even when multiple operations are requested concurrently or when system conditions change during operation execution. Safety measures include operation serialization to prevent conflicting concurrent operations, timeout mechanisms that prevent operations from hanging indefinitely, abort capabilities that enable safe cancellation of operations when conditions change, and comprehensive logging that provides audit trails for all power operations.

Operational safety integrates with system monitoring to detect and respond to conditions that may affect operation safety during execution.

### Emergency Response

Emergency response capabilities provide immediate power control when safety conditions require rapid response. Emergency mechanisms include thermal emergency shutdown that triggers immediate power-off when temperature limits are exceeded, fault response procedures that provide appropriate power control when electrical or hardware faults are detected, and override capabilities that enable emergency operations even when normal safety interlocks might prevent them.

Emergency response procedures balance speed requirements with hardware safety to provide rapid response while minimizing risk of hardware damage during emergency conditions.

## Monitoring and Observability

The powermgr service provides extensive monitoring and observability features that enable detailed analysis of power system behavior and support troubleshooting of complex power management issues.

### Operation Tracking

Comprehensive operation tracking maintains detailed records of all power operations including operation timing and duration, success and failure status with detailed error information, hardware control sequences and their results, and coordination activities with other system services.

Operation tracking supports both real-time monitoring and historical analysis, enabling identification of patterns and trends that may indicate hardware issues or optimization opportunities.

### Power State Monitoring

Real-time power state monitoring provides continuous visibility into system power status including individual component power states and transitions, power consumption and efficiency metrics where available, fault and alarm conditions from power management hardware, and trend analysis for predictive maintenance and capacity planning.

Power state monitoring integrates with broader system monitoring to provide comprehensive visibility into power-related aspects of system operation and performance.

### Performance Metrics

Detailed performance metrics track power management system efficiency and reliability including operation success rates and failure modes, operation timing and performance characteristics, hardware utilization and reliability statistics, and system-wide power efficiency and consumption trends.

Performance metrics enable optimization of power management policies and identification of opportunities for improved efficiency and reliability.

## Development and Testing

The powermgr service follows standard u-bmc development practices with comprehensive testing infrastructure that validates both normal operation and error handling scenarios.

### Package Documentation

Detailed package documentation is available at pkg.go.dev covering all aspects of power manager integration including hardware configuration and platform-specific integration patterns, safety mechanism configuration and policy development, integration with other u-bmc services, and troubleshooting and diagnostic procedures.

Documentation includes practical examples that demonstrate common integration patterns and provide templates for platform-specific power management configurations.

### Testing Infrastructure

Comprehensive test suites validate all aspects of power manager functionality including hardware control logic and safety mechanism operation, integration with other services under various operational conditions, error handling and recovery procedures for different failure scenarios, and performance characteristics under different load and concurrency conditions.

Testing infrastructure includes mock hardware interfaces and failure injection capabilities that enable thorough testing of error handling and safety mechanisms under realistic failure scenarios without requiring specialized hardware or risking damage to development systems.

### Hardware Simulation

Hardware simulation capabilities enable development and testing of power management functionality without requiring access to specialized BMC hardware. Simulation includes GPIO interface simulation with configurable timing and failure modes, power supply simulation with realistic power delivery and fault characteristics, and thermal condition simulation that enables testing of thermal integration features.

Simulation capabilities integrate with the standard development workflow to enable comprehensive testing of power management features throughout the development process.

## Platform Integration

The powermgr service adapts to different hardware platforms through configurable hardware interfaces and power management policies while maintaining consistent behavior and safety properties across different deployments.

### Hardware Abstraction

Hardware abstraction mechanisms enable support for diverse power control hardware while maintaining consistent power management behavior. Abstraction includes configurable GPIO mappings for different pin assignments and signal polarities, pluggable hardware adapters for specialized power control mechanisms, and standardized interfaces that isolate power management logic from hardware-specific implementation details.

Hardware abstraction enables rapid platform bring-up while ensuring that power management behavior remains consistent and reliable across different hardware configurations.

### Platform-Specific Policies

Different platforms may require customized power management policies that reflect their specific operational requirements, hardware capabilities, and deployment environments. The powermgr service supports flexible policy configuration that enables platform-specific customization while maintaining compatibility with standard power management operations.

Platform-specific policies can define custom sequencing requirements for complex hardware configurations, specialized safety requirements for high-reliability or safety-critical deployments, and performance optimization settings that take advantage of specific hardware capabilities.

## Future Enhancements

The powermgr service provides a foundation for advanced power management capabilities that may be implemented in future u-bmc releases based on operational requirements and user feedback.

### Advanced Power Management

Future enhancements may include more sophisticated power management features such as dynamic power budgeting and allocation across multiple components, power efficiency optimization through intelligent load balancing and supply management, and predictive power management that anticipates power requirements based on workload patterns and system trends.

### Enhanced Integration

Advanced integration capabilities could include tighter integration with workload management systems for power-aware scheduling, integration with facilities management systems for comprehensive power and cooling coordination, and support for distributed power management in multi-BMC and rack-scale deployments.

### Intelligent Automation

Machine learning and intelligent automation features might enable adaptive power management policies that optimize for specific operational requirements, predictive failure detection and prevention based on power system behavior analysis, and automated optimization of power management parameters based on observed system behavior and performance characteristics.

The modular design of the powermgr service and its standardized integration patterns provide a solid foundation for these enhancements while maintaining compatibility with existing hardware integrations and platform configurations.

## References

Additional information about the powermgr service and its integration with the broader u-bmc system is available in related documentation including system architecture details in `docs/architecture.md`, state management coordination in `docs/statemgr.md`, thermal management integration in `docs/thermalmgr.md`, and platform-specific configuration guidance in `docs/porting.md`.

Package-level documentation and detailed API references are maintained on pkg.go.dev for comprehensive development and integration guidance. The pkg/gpio package provides the foundational GPIO control mechanisms used by the service with detailed documentation of GPIO usage patterns and best practices for BMC applications.