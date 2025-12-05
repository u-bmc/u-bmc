# LED Manager Service

The LED manager service (ledmgr) provides centralized control and coordination of status, identification, and power indication LEDs across u-bmc systems. It manages GPIO-based LED controls, coordinates LED states with system operations, and provides visual feedback for system status, component identification, and operational activities. The LED manager integrates closely with state management and hardware services to ensure that LED indications accurately reflect current system conditions and operations.

## Overview

Visual indication through LEDs serves critical roles in BMC systems including providing immediate visual feedback about system health and operational status, enabling easy identification of specific components during maintenance and troubleshooting activities, and indicating power states and transitions for both automated systems and human operators. The ledmgr service centralizes these responsibilities while providing consistent LED behavior across different hardware platforms.

The service abstracts the complexity of different LED control mechanisms including direct GPIO control for simple on/off and blinking patterns, PWM control for brightness adjustment and complex pattern generation, and I2C-based LED controllers for advanced multi-color and programmable LED systems. This abstraction enables consistent LED management behavior while supporting the specific capabilities of different hardware platforms.

LED operations coordinate with system state changes to provide accurate visual indication of component status and transitions, hardware operations to indicate activity and completion status, and maintenance operations to support component identification and troubleshooting procedures. The service implements intelligent LED management policies that balance informational requirements with power consumption and visual clarity considerations.

## Architecture

The ledmgr service implements a layered architecture that separates physical LED control from logical LED management operations. This design enables support for diverse LED hardware while maintaining consistent visual indication behavior and coordination with other system services.

### Hardware Abstraction Layer

The hardware abstraction layer provides uniform interfaces for different types of LED control hardware. GPIO-based LED control handles basic on/off operations and simple blinking patterns through direct pin control using the pkg/gpio package. PWM-based control enables brightness adjustment and complex timing patterns through pulse-width modulation interfaces. I2C LED controllers support advanced features including multi-color LEDs, programmable patterns, and coordinated control of multiple LED elements.

The abstraction layer includes automatic capability detection that identifies available LED control features and provides appropriate interfaces based on hardware capabilities. Platform-specific LED controllers can be integrated through pluggable hardware adapters that implement standard interfaces while providing access to specialized LED control features.

### LED Function Management

LED function management organizes LED controls into logical categories that correspond to their informational roles within the system. Status LEDs provide continuous indication of component health and operational state with standardized colors and patterns that communicate specific status information. Identification LEDs enable targeted indication of specific components during maintenance activities with clear visual patterns that help operators locate components in complex systems.

Power LEDs indicate power states and transitions with patterns that reflect current power status and ongoing power operations. Activity LEDs provide feedback about ongoing operations and system activity levels with patterns that communicate operation progress and completion status.

### Pattern and Coordination Engine

The pattern and coordination engine manages complex LED behaviors that involve timing, coordination between multiple LEDs, and integration with system events. Pattern management includes support for blinking, fading, and complex multi-phase patterns with configurable timing parameters that can be customized based on operational requirements and hardware capabilities.

Coordination mechanisms ensure that LED indications remain consistent and informative even when multiple system events occur simultaneously. Priority management resolves conflicts between different indication requirements while ensuring that critical status information remains visible and actionable.

## LED Control Mechanisms

The ledmgr service supports multiple LED control mechanisms that can be used individually or in combination based on platform requirements and available hardware capabilities.

### GPIO LED Control

GPIO-based LED control provides the foundation for most BMC LED operations through direct control of LED power and simple timing patterns. The service uses the pkg/gpio package to provide reliable, low-latency LED control with appropriate error handling and power management features.

GPIO LED operations include basic on/off control for status indication with support for both active-high and active-low LED configurations, simple blinking patterns with configurable on/off timing for attention-getting behaviors, and power management integration that can disable LEDs during low-power operations while maintaining essential status indication.

GPIO control supports multiple LEDs per function with coordinated control that can provide redundant indication or distributed visual feedback across different system areas. Configuration options include timing parameters, power management policies, and coordination rules that ensure consistent behavior across different LED elements.

### PWM LED Control

PWM-based LED control enables advanced LED behaviors including brightness adjustment, smooth fading transitions, and complex timing patterns that provide rich visual feedback. PWM control integrates with hardware PWM controllers to provide precise timing and brightness control while minimizing CPU overhead for LED management.

PWM capabilities include brightness adjustment that can provide subtle status indication without being distracting in operational environments, fading patterns that provide smooth transitions between different status states, and complex multi-phase patterns that can communicate detailed status information through visual sequences.

PWM control can coordinate multiple LEDs to provide synchronized patterns and behaviors that enhance visual communication while maintaining clear indication of individual component status and system-wide conditions.

### Advanced LED Controllers

Advanced LED controllers accessed through I2C or other interfaces provide sophisticated LED management capabilities including multi-color LEDs that can communicate different types of status information through color coding, programmable pattern engines that can execute complex LED behaviors without CPU intervention, and coordinated multi-LED systems that can provide rich visual feedback across large numbers of LED elements.

Advanced controller integration includes automatic capability detection and configuration, standardized interfaces that abstract controller-specific implementation details, and power management integration that optimizes LED operations for different operational modes and power constraints.

## LED Function Categories

The ledmgr service organizes LED control around standardized function categories that provide consistent visual communication across different hardware platforms and deployment scenarios.

### Status LEDs

Status LEDs provide continuous indication of component health and operational state using standardized color and pattern conventions. Status indication includes healthy operation with steady green indication for components operating within normal parameters, warning conditions with amber or yellow indication for components experiencing non-critical issues that may require attention, and error conditions with red indication for components experiencing failures or critical issues requiring immediate attention.

Status LED patterns can include blinking behaviors that indicate transitional states or ongoing diagnostic activities, and coordinated behaviors that show relationships between different system components and their collective health status.

Status LEDs integrate with health monitoring systems to provide real-time visual feedback based on sensor data, system diagnostics, and operational metrics collected from throughout the system.

### Identification LEDs

Identification LEDs enable targeted visual identification of specific components during maintenance and troubleshooting activities. Identification behaviors include steady bright indication for persistent component identification during maintenance procedures, distinctive blinking patterns that make components easy to locate in complex systems with many similar elements, and coordinated indication that can highlight component relationships and dependencies during troubleshooting activities.

Identification LED control can be triggered through management interfaces to support remote troubleshooting scenarios where operators need to identify components without physical access to management systems. Identification patterns are designed to be clearly distinguishable from status indications while remaining compatible with operational status LED behaviors.

Identification LEDs support timeout mechanisms that automatically clear identification states after configured periods to prevent confusion and ensure that identification indications remain current and actionable.

### Power LEDs

Power LEDs indicate current power states and power transition activities using patterns that clearly communicate power status to both automated systems and human operators. Power indication includes off states with no illumination when components are fully powered down, power-on states with steady indication when components are fully operational, and transition states with distinctive patterns that indicate ongoing power operations like startup or shutdown sequences.

Power LED patterns coordinate with actual power operations to provide accurate real-time feedback about power state changes and can indicate different types of power operations including normal startup and shutdown sequences, emergency power operations, and maintenance-related power cycling activities.

Power LEDs integrate with power management systems to ensure that visual indications accurately reflect actual hardware power states and operational status throughout all power management operations.

### Activity LEDs

Activity LEDs provide visual feedback about ongoing system operations and activity levels. Activity indication includes operation progress feedback for long-running operations that helps operators understand system activity levels and operation completion status, system load indication that provides visual feedback about overall system utilization and performance characteristics, and maintenance activity indication that shows when systems are undergoing configuration changes or diagnostic procedures.

Activity LED patterns can include rate-based blinking that reflects activity levels, coordinated patterns that show distributed operations across multiple system components, and completion indicators that signal successful operation completion or error conditions requiring attention.

Activity LEDs balance informational value with visual clarity to provide useful feedback without creating distracting or confusing visual environments during normal system operation.

## Integration with System Services

The ledmgr service integrates closely with other u-bmc services to provide accurate and timely visual indication that reflects current system conditions and operational activities.

### State Manager Integration

LED indications coordinate closely with system state changes managed by the state manager service to provide accurate visual feedback about component states and state transitions. Integration ensures that LED patterns accurately reflect current system states and that LED changes occur in coordination with actual state transitions.

State integration includes automatic LED updates when components change states, coordination of LED patterns during complex state transitions that involve multiple components, and error indication when state transitions fail or encounter problems that require operator attention.

LED behavior can be customized based on state management policies to provide appropriate visual feedback for different operational modes and system configurations while maintaining consistent indication behavior across different deployment scenarios.

### Power Manager Integration

Power state indications coordinate with power management operations to provide accurate visual feedback about power states and power transition activities. Integration ensures that power LEDs reflect actual hardware power states and that LED indications remain accurate throughout all power management operations.

Power integration includes real-time updates during power operations, coordination with power sequencing to provide appropriate indication timing, and emergency indication capabilities that can provide immediate visual feedback during power-related emergency situations.

LED power management integration can also optimize LED power consumption during different system power modes while maintaining essential status and identification capabilities throughout all operational scenarios.

### Thermal Manager Integration

LED indications can reflect thermal conditions and thermal management activities to provide visual feedback about system thermal status and cooling operations. Thermal integration enables LED patterns that indicate thermal warning conditions, cooling system status, and thermal emergency situations that may require immediate operator attention.

Thermal LED integration includes coordination with thermal management policies to provide appropriate warning timelines and indication patterns, integration with emergency thermal response procedures to provide immediate visual alerts, and status indication that reflects ongoing thermal management activities and cooling system operation.

### Hardware Monitoring Integration

LED status indications integrate with comprehensive hardware monitoring to provide visual feedback based on sensor data, component health information, and system performance metrics. Monitoring integration enables intelligent LED behaviors that reflect actual system conditions rather than just operational states.

Hardware monitoring integration includes health-based status indication that reflects sensor data and component diagnostics, predictive indication that can provide early warning of developing problems, and comprehensive status display that reflects overall system health and performance characteristics.

## Configuration

The ledmgr service follows standard u-bmc configuration patterns with comprehensive options for LED hardware, behavior policies, and integration parameters.

### Basic Configuration

```go
ledMgr := ledmgr.New(
    ledmgr.WithServiceName("ledmgr"),
    ledmgr.WithServiceDescription("BMC LED Management Service"),
    ledmgr.WithGPIOChip("/dev/gpiochip0"),
    ledmgr.WithUpdateInterval(100 * time.Millisecond),
    ledmgr.WithPowerManagement(true),
    ledmgr.WithMetrics(true),
    ledmgr.WithTracing(true),
)
```

### LED Hardware Configuration

```go
ledMgr := ledmgr.New(
    ledmgr.WithGPIOLEDs(map[string]ledmgr.GPIOLEDConfig{
        "status": {
            Line: "status-led",
            ActiveHigh: true,
            PowerDomain: "chassis",
        },
        "identify": {
            Line: "identify-led", 
            ActiveHigh: true,
            PowerDomain: "chassis",
        },
        "power": {
            Line: "power-led",
            ActiveHigh: false,
            PowerDomain: "always-on",
        },
    }),
    ledmgr.WithPWMLEDs(map[string]ledmgr.PWMLEDConfig{
        "activity": {
            Device: "/sys/class/pwm/pwmchip0/pwm0",
            MaxBrightness: 255,
            PowerDomain: "chassis",
        },
    }),
    ledmgr.WithI2CLEDs(map[string]ledmgr.I2CLEDConfig{
        "multi-status": {
            Bus: "/dev/i2c-1",
            Address: 0x60,
            Controller: "pca9685",
            Channels: []int{0, 1, 2}, // RGB channels
        },
    }),
)
```

### LED Function Configuration

```go
ledMgr := ledmgr.New(
    ledmgr.WithLEDFunctions(map[string]ledmgr.LEDFunction{
        "host.0.status": {
            Type: ledmgr.StatusLED,
            LEDs: []string{"status"},
            Patterns: map[string]ledmgr.LEDPattern{
                "healthy": {Color: "green", Mode: "steady"},
                "warning": {Color: "yellow", Mode: "blink", Period: 1000},
                "error": {Color: "red", Mode: "steady"},
                "critical": {Color: "red", Mode: "blink", Period: 250},
            },
            DefaultPattern: "healthy",
            PowerPolicy: "operational",
        },
        "host.0.identify": {
            Type: ledmgr.IdentifyLED,
            LEDs: []string{"identify"},
            Patterns: map[string]ledmgr.LEDPattern{
                "identify": {Color: "blue", Mode: "blink", Period: 500},
                "off": {Mode: "off"},
            },
            DefaultPattern: "off",
            Timeout: 300 * time.Second,
        },
        "chassis.power": {
            Type: ledmgr.PowerLED,
            LEDs: []string{"power"},
            Patterns: map[string]ledmgr.LEDPattern{
                "on": {Color: "green", Mode: "steady"},
                "off": {Mode: "off"},
                "transition": {Color: "yellow", Mode: "fade", Period: 2000},
            },
            DefaultPattern: "off",
        },
    }),
)
```

### Pattern and Timing Configuration

```go
ledMgr := ledmgr.New(
    ledmgr.WithDefaultPatterns(map[string]ledmgr.LEDPattern{
        "blink_slow": {Mode: "blink", Period: 2000, DutyCycle: 50},
        "blink_fast": {Mode: "blink", Period: 250, DutyCycle: 50},
        "fade_slow": {Mode: "fade", Period: 3000},
        "fade_fast": {Mode: "fade", Period: 1000},
        "heartbeat": {Mode: "pulse", Period: 2000, Phases: []int{10, 90}},
    }),
    ledmgr.WithTimingPrecision(10 * time.Millisecond),
    ledmgr.WithBrightnessLevels(map[string]int{
        "dim": 64,
        "normal": 128,
        "bright": 255,
    }),
)
```

### Integration Configuration

```go
ledMgr := ledmgr.New(
    ledmgr.WithStateManagerIntegration("statemgr"),
    ledmgr.WithPowerManagerIntegration("powermgr"),
    ledmgr.WithThermalManagerIntegration("thermalmgr"),
    ledmgr.WithSensorMonitorIntegration("sensormon"),
    ledmgr.WithIntegrationTimeout(5 * time.Second),
    ledmgr.WithEventSubscriptions([]string{
        "statemgr.events.transitions",
        "powermgr.events.operations", 
        "thermalmgr.events.thermal",
        "sensormon.events.threshold",
    }),
)
```

## NATS Integration and Endpoints

The ledmgr service provides comprehensive NATS endpoints for LED control and status monitoring while integrating with other services through well-defined event subscription patterns.

### LED Control Endpoints

LED control endpoints provide the primary interface for requesting LED state changes and pattern updates. All LED operations include validation to ensure that requested patterns are supported by the hardware and compatible with current system policies.

```
ledmgr.leds.list              # List all configured LEDs and functions
ledmgr.led.get                # Get current LED state and pattern
ledmgr.led.set                # Set LED pattern and state
ledmgr.function.get           # Get LED function status
ledmgr.function.set           # Set LED function pattern
ledmgr.identify.start         # Start component identification
ledmgr.identify.stop          # Stop component identification
```

### Status and Information Endpoints

Status endpoints provide real-time information about LED states, active patterns, and hardware status related to LED management operations.

```
ledmgr.status.get             # Get overall LED system status
ledmgr.patterns.list          # List available LED patterns
ledmgr.hardware.status        # Get LED hardware status and capabilities
ledmgr.functions.list         # List configured LED functions
ledmgr.power.status           # Get LED power management status
```

### Configuration Management Endpoints

Configuration endpoints enable dynamic management of LED patterns, function assignments, and operational parameters without requiring service restart.

```
ledmgr.config.get             # Get current LED configuration
ledmgr.config.update          # Update LED configuration
ledmgr.patterns.create        # Create custom LED patterns
ledmgr.patterns.update        # Update existing LED patterns
ledmgr.functions.configure    # Configure LED function assignments
```

## LED Pattern Processing

LED pattern processing in the ledmgr service implements sophisticated timing and coordination mechanisms that provide rich visual feedback while maintaining efficient resource utilization and reliable operation.

### Pattern Engine

The pattern engine processes LED pattern definitions and generates appropriate hardware control sequences based on configured timing parameters and hardware capabilities. Pattern processing includes support for basic on/off patterns with configurable timing, blinking patterns with adjustable duty cycles and frequencies, fading patterns that provide smooth brightness transitions, and complex multi-phase patterns that can communicate detailed status information.

Pattern execution includes hardware-appropriate timing optimization that matches pattern timing to hardware capabilities and system performance requirements, power management integration that can adjust pattern behavior based on system power states, and error handling that provides graceful degradation when hardware limitations prevent full pattern execution.

### Coordination and Priority Management

LED coordination mechanisms ensure that LED behaviors remain clear and informative when multiple status conditions or operations occur simultaneously. Coordination includes priority-based pattern selection that ensures critical status information takes precedence over less important indications, graceful pattern transitions that avoid jarring visual changes when LED functions change states, and conflict resolution that provides clear indication when multiple conditions require LED attention.

Priority management considers the relative importance of different LED functions and system conditions to ensure that the most important information is always visible while providing mechanisms for less critical information to be indicated when primary conditions are resolved.

### Performance Optimization

LED pattern processing includes performance optimizations that minimize system resource usage while maintaining precise timing and visual quality. Optimizations include efficient timing mechanisms that minimize CPU overhead for LED management, hardware acceleration utilization when available to offload pattern generation from the main CPU, and adaptive performance scaling that adjusts LED update rates based on system load and performance requirements.

Performance optimization ensures that LED management remains responsive and accurate even during periods of high system activity or resource constraint while maintaining the visual quality necessary for effective system status communication.

## Monitoring and Observability

The ledmgr service provides extensive monitoring and observability features that enable detailed analysis of LED system behavior and support troubleshooting of visual indication issues.

### LED Status Tracking

Comprehensive LED status tracking maintains detailed records of all LED state changes and pattern executions including timing information for pattern changes and execution, success and failure status for LED control operations, hardware status and capability information, and correlation with system events that trigger LED changes.

Status tracking supports both real-time monitoring of LED system operation and historical analysis that can identify patterns and trends in LED usage and system status indication requirements.

### Visual Indication Metrics

Detailed metrics track LED system performance and effectiveness including LED operation success rates and hardware failure detection, pattern execution timing and accuracy measurements, power consumption tracking for LED operations, and usage statistics that show LED function utilization and effectiveness.

Visual indication metrics help optimize LED configuration and patterns for maximum effectiveness while maintaining efficient resource utilization and reliable operation across different system conditions.

### Hardware Health Monitoring

LED hardware health monitoring provides continuous assessment of LED control hardware and can detect failures or degraded operation before they impact visual indication effectiveness. Health monitoring includes LED control hardware status and failure detection, pattern execution accuracy and timing validation, and power delivery monitoring for LED control systems.

Hardware health monitoring integrates with broader system health monitoring to provide comprehensive visibility into LED system operation and enable proactive maintenance of LED control systems.

## Development and Testing

The ledmgr service follows standard u-bmc development practices with comprehensive testing infrastructure that validates LED control functionality across different hardware configurations and operational scenarios.

### Package Documentation

Detailed package documentation is available at pkg.go.dev covering all aspects of LED manager integration including LED hardware configuration and pattern development, integration with system services for coordinated LED behavior, performance optimization and power management considerations, and troubleshooting procedures for LED control issues.

Documentation includes practical examples that demonstrate common LED configuration patterns and provide templates for platform-specific LED management implementations.

### Testing Infrastructure

Comprehensive test suites validate all aspects of LED manager functionality including LED control logic and pattern execution under various conditions, integration with system services and event coordination, hardware abstraction and platform-specific LED controller support, and performance characteristics under different system load conditions.

Testing infrastructure includes LED hardware simulation capabilities that enable development and testing without requiring specialized LED hardware, and failure injection mechanisms that validate error handling and recovery procedures for LED control failures.

### Hardware Simulation

LED hardware simulation enables development and testing of LED management functionality across different hardware configurations without requiring access to specialized BMC hardware. Simulation includes GPIO LED simulation with configurable timing and failure characteristics, PWM controller simulation with realistic brightness and timing behavior, and advanced LED controller simulation that models complex multi-LED systems and programmable pattern capabilities.

Simulation capabilities integrate with standard development workflows to enable comprehensive testing of LED management features throughout the development process while supporting validation of LED behaviors across different hardware platforms and configurations.

## Platform Integration

The ledmgr service adapts to different hardware platforms through configurable LED hardware interfaces and pattern definitions while maintaining consistent visual indication behavior across different deployments.

### Hardware Abstraction

LED hardware abstraction mechanisms enable support for diverse LED control hardware while maintaining consistent LED management behavior and visual indication standards. Abstraction includes configurable GPIO mappings for different LED assignments and control polarities, pluggable LED controller adapters for specialized LED control hardware, and standardized pattern interfaces that isolate LED management logic from hardware-specific implementation details.

Hardware abstraction enables rapid platform bring-up while ensuring that LED indication behavior remains consistent and effective across different hardware configurations and deployment scenarios.

### Visual Standards

The ledmgr service supports configurable visual standards that can be customized for different operational environments and user preferences while maintaining clear and effective visual communication. Visual standards include color coding conventions for different types of status information, timing and pattern standards that provide appropriate visual feedback for different operational scenarios, and brightness and power management policies that balance visual effectiveness with power consumption requirements.

Visual standards can be customized per platform and deployment to accommodate different operational requirements while maintaining compatibility with standard LED management operations and integration patterns.

## Future Enhancements

The ledmgr service provides a foundation for advanced LED management capabilities that may be implemented in future u-bmc releases based on operational requirements and user feedback.

### Advanced Pattern Capabilities

Future enhancements may include more sophisticated LED pattern capabilities such as multi-LED coordination for complex visual displays, adaptive patterns that adjust based on ambient lighting conditions or operational requirements, and integration with external display systems for comprehensive visual status indication.

### Intelligent LED Management

Advanced LED management features could include machine learning integration for adaptive LED behavior optimization, predictive LED indication that provides early warning of developing system conditions, and automatic pattern optimization based on observed effectiveness and user feedback.

### Integration Enhancements

Enhanced integration capabilities might include deeper integration with facilities management systems for coordinated visual indication across multiple systems, support for external LED display systems and status boards, and integration with mobile device applications for remote LED control and status monitoring.

The modular design of the ledmgr service and its standardized integration patterns provide a solid foundation for these enhancements while maintaining compatibility with existing hardware configurations and platform integrations.

## References

Additional information about the ledmgr service and its integration with the broader u-bmc system is available in related documentation including system architecture details in [docs/architecture.md](architecture.md), state management coordination in [docs/statemgr.md](statemgr.md), power management integration in [docs/powermgr.md](powermgr.md), and platform-specific configuration guidance in [docs/porting.md](porting.md).

Package-level documentation and detailed API references are maintained on pkg.go.dev for comprehensive development and integration guidance. The pkg/gpio package provides foundational GPIO control mechanisms used by the service with detailed documentation of GPIO usage patterns and best practices for LED control applications.