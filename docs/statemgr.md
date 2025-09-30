# State Manager Service

The state manager service (statemgr) provides centralized state machine management for u-bmc system components including hosts, chassis, and BMC subsystems. It orchestrates state transitions, coordinates actions across multiple services, and maintains consistent system state information throughout the BMC lifecycle. The state manager serves as the coordination point for complex operations that require synchronized actions from multiple hardware and software components.

## Overview

BMC systems require careful coordination of component states to ensure safe and reliable operation. The state manager implements this coordination through well-defined state machines that model the behavior of different system components and their interactions. Rather than allowing services to independently manage component states, the centralized approach ensures consistent behavior and prevents conflicting operations that could damage hardware or compromise system reliability.

The state manager integrates deeply with hardware management services including power management for host and chassis power sequencing, thermal management for temperature-aware state transitions, sensor monitoring for health-based state decisions, and LED management for visual state indication. This integration enables sophisticated coordination policies that consider multiple factors when making state transition decisions.

State machines in the statemgr service are designed to handle both normal operational transitions and error recovery scenarios. The service includes comprehensive error handling and recovery mechanisms that can safely navigate system failures while maintaining hardware protection and providing clear visibility into system status for operators and management tools.

## Architecture

The statemgr service implements a hierarchical state machine architecture where different system components are modeled as separate but interconnected state machines. This design enables independent management of different components while providing coordination mechanisms for operations that affect multiple components simultaneously.

### Component State Machines

The service manages several distinct types of state machines corresponding to different system components. Host state machines model individual compute nodes with states for power off, power on, booting, running, and various error conditions. Chassis state machines represent the overall system enclosure and shared infrastructure with states for power sequencing, thermal management, and overall system health. BMC state machines track the management controller itself including initialization, normal operation, maintenance modes, and shutdown procedures.

Each state machine type implements specific transition rules and validation logic appropriate for the component it represents. State machines include safety interlocks that prevent dangerous transitions and ensure that prerequisite conditions are met before allowing potentially harmful operations.

### State Coordination

Complex operations often require coordinated state changes across multiple components. The state manager implements coordination mechanisms that ensure atomic execution of multi-component operations while maintaining system safety and providing clear error handling when coordination cannot be completed successfully.

Coordination policies define the relationships between different component states and the actions required when dependencies change. For example, chassis power-down operations must ensure that all hosted compute nodes are safely shut down before proceeding with chassis-level power sequencing.

### Event Processing

State transitions are triggered by events from various sources including external management requests through the web API, hardware events from sensor monitoring and power management services, internal system events like thermal emergencies or component failures, and timer-based events for operations with temporal dependencies.

The event processing system ensures that state transitions are processed in the correct order and that concurrent events are handled safely without creating race conditions or inconsistent system states.

## State Machine Implementation

The statemgr service uses the pkg/state package to implement robust, testable state machines with comprehensive logging and monitoring integration. This implementation provides type-safe state definitions, validated transitions with comprehensive error handling, and extensive observability features.

### Host State Management

Host state machines model the lifecycle of individual compute nodes from initial power-up through normal operation to shutdown and error recovery. The state machine includes states for power-off with all systems inactive, power-on with basic power applied but systems not yet operational, boot sequence with BIOS/UEFI initialization and operating system loading, running state with full operational capability, and various error states for different types of system failures.

Transitions between host states consider multiple factors including power management policies and thermal conditions, hardware health status from sensor monitoring, external management requests and scheduling policies, and dependency relationships with chassis and other host systems.

The host state machine integrates with power management services to control physical power sequencing and with sensor monitoring to track system health throughout state transitions. LED management integration provides visual indication of host status for operators.

### Chassis State Management

Chassis state machines represent shared infrastructure and overall system-level operations. The chassis state machine coordinates operations that affect multiple hosts or the overall system environment including power supply sequencing and management, cooling system operation and thermal management coordination, and shared resource management like network switches or storage controllers.

Chassis states include power-off with all infrastructure inactive, power-sequencing with coordinated startup of power supplies and infrastructure, operational with full infrastructure support available, and maintenance modes for servicing and configuration activities.

Chassis state transitions coordinate with all hosted compute nodes to ensure safe operation and prevent conflicts between chassis-level operations and individual host requirements.

### BMC State Management

BMC state machines track the management controller's own operational status and lifecycle. This includes initialization states during BMC startup and service activation, operational states with full management capability available, maintenance modes for firmware updates or configuration changes, and shutdown states for safe BMC restart or system power-down.

BMC state management ensures that management capabilities remain available during system operations while providing safe mechanisms for BMC maintenance and updates that minimize disruption to managed systems.

## Integration with Hardware Services

The state manager coordinates closely with all hardware management services to ensure that state transitions are properly synchronized with physical hardware operations and system monitoring.

### Power Management Integration

State transitions that involve power operations coordinate with the power management service to ensure safe sequencing and proper timing. The state manager requests power operations through well-defined interfaces and monitors completion status to ensure that state transitions accurately reflect actual hardware status.

Power management integration includes coordination of host power operations with chassis power sequencing, management of power dependencies between different system components, and integration with emergency power-off procedures for safety-critical situations.

The state manager maintains power state consistency by tracking both requested and actual power states and resolving discrepancies through appropriate recovery procedures.

### Thermal Management Integration

Thermal conditions significantly impact safe state transition policies, and the state manager integrates closely with thermal management services to ensure that temperature considerations are properly incorporated into state transition decisions.

Thermal integration includes blocking power-on operations when thermal conditions are unsafe, coordinating emergency shutdown procedures when thermal limits are exceeded, and implementing thermal-aware scheduling policies that consider cooling capacity and temperature trends.

The state manager can implement sophisticated thermal policies that balance performance requirements with thermal constraints, enabling automatic load shedding or power reduction when thermal conditions require intervention.

### Sensor Integration

Comprehensive sensor monitoring provides the health information necessary for intelligent state transition decisions. The state manager integrates with sensor monitoring services to receive real-time health data and incorporate this information into transition policies.

Sensor integration enables health-based state decisions that prevent operations on failing hardware, predictive maintenance scheduling based on sensor trends and health indicators, and automated recovery procedures when sensor data indicates component failures or degraded performance.

The integration supports configurable health policies that can be customized based on operational requirements and risk tolerance for different types of deployments.

### LED Management Integration

Visual indication of system state is crucial for operator awareness and troubleshooting activities. The state manager coordinates with LED management services to ensure that visual indicators accurately reflect current system states and state transition activities.

LED integration includes coordinated LED patterns that indicate system state and transition status, error indication and alert signaling for problem conditions, and identification modes that help operators locate specific components during maintenance activities.

The LED integration can implement sophisticated indication policies that provide rich information about system status while remaining intuitive for operators with different levels of technical expertise.

## Configuration

The state manager service follows standard u-bmc configuration patterns with comprehensive options for state machine behavior, integration parameters, and operational policies.

### Basic Configuration

```go
stateMgr := statemgr.New(
    statemgr.WithServiceName("statemgr"),
    statemgr.WithServiceDescription("BMC State Management Service"),
    statemgr.WithTransitionTimeout(30 * time.Second),
    statemgr.WithEventBufferSize(100),
    statemgr.WithStateValidation(true),
    statemgr.WithMetrics(true),
    statemgr.WithTracing(true),
)
```

### State Machine Configuration

```go
stateMgr := statemgr.New(
    statemgr.WithHostStateMachines([]statemgr.HostConfig{
        {
            Name: "host.0",
            PowerDependencies: []string{"psu.0", "psu.1"},
            ThermalZone: "cpu_zone",
            HealthSensors: []string{"cpu0_temp", "dimm_temp"},
        },
    }),
    statemgr.WithChassisStateMachine(statemgr.ChassisConfig{
        Name: "chassis",
        PowerSequence: []string{"psu.main", "psu.aux"},
        CoolingPolicy: "balanced",
        HostDependencies: []string{"host.0", "host.1"},
    }),
    statemgr.WithBMCStateMachine(statemgr.BMCConfig{
        InitializationTimeout: 60 * time.Second,
        MaintenanceMode: true,
        UpdatePolicy: "graceful",
    }),
)
```

### Integration Configuration

```go
stateMgr := statemgr.New(
    statemgr.WithPowerManagerIntegration("powermgr"),
    statemgr.WithThermalManagerIntegration("thermalmgr"),
    statemgr.WithSensorMonitorIntegration("sensormon"),
    statemgr.WithLEDManagerIntegration("ledmgr"),
    statemgr.WithIntegrationTimeout(10 * time.Second),
    statemgr.WithRetryPolicy(3, 2*time.Second),
)
```

### Policy Configuration

```go
stateMgr := statemgr.New(
    statemgr.WithTransitionPolicies(map[string]statemgr.TransitionPolicy{
        "host_power_on": {
            ThermalCheck: true,
            HealthCheck: true,
            PowerCheck: true,
            MaxConcurrent: 2,
        },
        "chassis_power_off": {
            HostShutdownRequired: true,
            GracefulTimeout: 300 * time.Second,
            ForceTimeout: 30 * time.Second,
        },
    }),
    statemgr.WithEmergencyPolicies(statemgr.EmergencyConfig{
        ThermalShutdown: true,
        PowerFailureResponse: "graceful_shutdown",
        HealthFailureResponse: "isolate_component",
    }),
)
```

## NATS Integration and Endpoints

The state manager service provides comprehensive NATS endpoints for state management operations and integrates with other services through well-defined message patterns.

### State Query Endpoints

State information endpoints provide real-time access to current component states and historical state transition information. Clients can query individual component states, retrieve system-wide state summaries, and access detailed transition histories for troubleshooting and audit purposes.

```
statemgr.hosts.list           # List all host state machines
statemgr.host.get             # Get specific host state and status
statemgr.chassis.get          # Get chassis state and status  
statemgr.bmc.get              # Get BMC state and status
statemgr.system.status        # Get overall system status summary
```

### State Control Endpoints

State control endpoints enable external systems and operators to request state transitions through validated, safe interfaces. All state change requests are subject to policy validation and safety checks before execution.

```
statemgr.host.power.on        # Request host power-on transition
statemgr.host.power.off       # Request host power-off transition
statemgr.host.reset           # Request host reset operation
statemgr.chassis.power.cycle  # Request chassis power cycle
statemgr.bmc.maintenance      # Enter BMC maintenance mode
```

### Event Subscription Endpoints

Event subscription enables other services and external clients to receive real-time notifications of state changes and transition events. Event streams provide comprehensive information about state transitions including timing, success status, and any error conditions encountered.

```
statemgr.events.transitions   # All state transition events
statemgr.events.errors        # State transition error events
statemgr.events.host.*        # Host-specific state events
statemgr.events.chassis.*     # Chassis-specific state events
```

## State Transition Processing

State transitions in the statemgr service follow carefully designed patterns that ensure safety, consistency, and comprehensive error handling throughout the transition process.

### Transition Validation

Before executing any state transition, the state manager performs comprehensive validation to ensure that the requested transition is safe and appropriate given current system conditions. Validation includes checking current state compatibility with the requested transition, verifying that prerequisite conditions are met including thermal, power, and health requirements, and confirming that system policies allow the requested operation.

Validation failures result in clear error messages that explain why the transition cannot be executed and provide guidance on resolving blocking conditions.

### Coordination Sequence

Complex transitions that require coordination across multiple services follow well-defined sequences that ensure proper ordering and error handling. The coordination sequence begins with transition planning where all required actions and their dependencies are identified. Service coordination follows where required actions are requested from appropriate hardware management services with proper timeout and error handling.

Status monitoring ensures that all coordinated actions complete successfully before finalizing the state transition, while rollback procedures handle partial failures by safely reversing completed actions when coordination cannot be completed successfully.

### Error Recovery

Comprehensive error recovery mechanisms handle various failure scenarios that may occur during state transitions. Recovery procedures distinguish between different types of failures and apply appropriate recovery strategies ranging from simple retry operations for transient failures to comprehensive rollback procedures for more serious issues.

Error recovery includes automatic retry mechanisms with exponential backoff for transient failures, safe state rollback procedures that return systems to known good states when transitions fail, and alert generation for failures that require operator intervention or indicate serious system problems.

## Monitoring and Observability

The state manager service provides extensive monitoring and observability features that enable detailed analysis of system behavior and support troubleshooting of complex state transition issues.

### State History Tracking

Comprehensive state history tracking maintains detailed records of all state transitions including transition timing and duration, success and failure status with detailed error information, and the triggering events and conditions that initiated each transition.

State history supports both real-time monitoring and historical analysis, enabling identification of patterns and trends that may indicate system issues or optimization opportunities.

### Transition Metrics

Detailed metrics track state transition performance and reliability including transition success rates and failure modes, transition timing and performance characteristics, and system utilization and resource consumption during transitions.

Metrics integration enables alerting based on transition performance trends and supports capacity planning and system optimization activities.

### Event Correlation

Event correlation capabilities help identify relationships between different system events and state transitions, supporting complex troubleshooting scenarios where problems may involve interactions between multiple system components.

Correlation analysis can identify cascading failure patterns, dependency-related issues, and timing-sensitive interactions that may not be apparent from individual component monitoring.

## Safety and Reliability

The state manager service implements comprehensive safety and reliability mechanisms that protect hardware and ensure predictable system behavior even during failure conditions.

### Safety Interlocks

Safety interlocks prevent dangerous operations that could damage hardware or compromise system safety. These interlocks include thermal protection that prevents power operations when temperature conditions are unsafe, power sequencing protection that ensures proper ordering of power operations, and hardware health protection that blocks operations on components that have failed health checks.

Interlock implementations include both automatic protection mechanisms and override capabilities for authorized maintenance operations that may need to bypass normal safety restrictions.

### Consistency Guarantees

The state manager maintains strong consistency guarantees for system state information, ensuring that state information accurately reflects actual hardware status and that concurrent operations do not create inconsistent system states.

Consistency mechanisms include atomic transition operations that either complete fully or are safely rolled back, state synchronization with hardware management services to detect and resolve state inconsistencies, and conflict resolution procedures for handling concurrent state change requests.

### Recovery Procedures

Comprehensive recovery procedures handle various types of system failures and ensure that systems can be restored to operational status even after serious failures. Recovery includes automatic recovery for common failure scenarios, guided recovery procedures for complex failures that require operator intervention, and emergency procedures for safety-critical situations.

Recovery procedures integrate with hardware management services to ensure that recovery operations are coordinated safely and that hardware protection mechanisms remain active throughout recovery activities.

## Development and Testing

The state manager service follows standard u-bmc development practices with comprehensive testing infrastructure that validates both normal operation and error handling scenarios.

### Package Documentation

Detailed package documentation is available at pkg.go.dev covering all aspects of state manager integration including state machine definition and configuration patterns, integration with hardware management services, error handling and recovery procedures, and monitoring and observability features.

Documentation includes practical examples that demonstrate common integration patterns and provide templates for platform-specific state machine configurations.

### Testing Infrastructure

Comprehensive test suites validate all aspects of state manager functionality including state machine transition logic and safety validation, integration with hardware management services under various conditions, error handling and recovery mechanisms, and performance characteristics under different load conditions.

Testing infrastructure includes mock hardware services and failure injection capabilities that enable thorough testing of error handling and recovery mechanisms under realistic failure scenarios.

### State Machine Validation

State machine validation tools verify that state machine definitions are correct and complete, including validation of transition logic and safety constraints, verification of integration dependencies and coordination requirements, and analysis of reachability and deadlock conditions in complex state machine configurations.

Validation tools integrate with the development workflow to ensure that state machine changes maintain safety properties and do not introduce problematic behaviors.

## Platform Integration

The state manager service adapts to different hardware platforms through configurable state machine definitions and integration parameters while maintaining consistent behavior and safety properties across different deployments.

### Hardware-Specific Configuration

Different hardware platforms may require customized state machine configurations that reflect their specific power sequencing requirements, thermal management policies, and component dependencies. The state manager supports flexible configuration that enables platform-specific customization while maintaining compatibility with standard integration patterns.

Platform-specific configurations can define custom states and transitions that reflect unique hardware capabilities or requirements while building on the standard state machine framework provided by the service.

### Resource Dependencies

The state manager coordinates resource dependencies that may vary between different hardware platforms including power supply dependencies and sequencing requirements, thermal zone assignments and cooling coordination, and shared resource management for components that serve multiple hosts or chassis.

Dependency management ensures that resource conflicts are avoided and that resource availability is properly coordinated across different system components and operations.

## Future Enhancements

The state manager service provides a foundation for advanced state management capabilities that may be implemented in future u-bmc releases based on operational requirements and user feedback.

### Advanced Coordination

Future enhancements may include more sophisticated coordination mechanisms such as distributed state management for multi-BMC systems, advanced dependency analysis and optimization for complex system configurations, and predictive state management that anticipates required transitions based on system trends and usage patterns.

### Policy Framework

Advanced policy frameworks could enable more flexible and sophisticated state management policies including user-defined policies for custom operational requirements, machine learning integration for adaptive policy optimization, and integration with external policy engines for enterprise policy management.

### Performance Optimization

Continued performance optimization may focus on state transition parallelization for faster system operations, advanced caching strategies for frequently accessed state information, and optimization of coordination overhead for large-scale systems with many managed components.

The modular design of the state manager service and its standardized integration patterns provide a solid foundation for these enhancements while maintaining compatibility with existing hardware integrations and platform configurations.

## References

Additional information about the state manager service and its integration with the broader u-bmc system is available in related documentation including system architecture details in `docs/architecture.md`, power management integration in `docs/powermgr.md`, thermal management coordination in `docs/thermalmgr.md`, and platform-specific configuration guidance in `docs/porting.md`.

Package-level documentation and detailed API references are maintained on pkg.go.dev for comprehensive development and integration guidance. The pkg/state package provides the foundational state machine implementation used by the service with detailed documentation of state machine patterns and best practices.