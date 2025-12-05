# Operator Service

The operator service is the foundational supervision and lifecycle management component of u-bmc. It orchestrates the startup sequence of all other services, monitors their health, and coordinates graceful shutdown procedures. The operator ensures that services start in the correct dependency order and provides centralized process management for the entire BMC system.

## Overview

The operator service acts as the process supervisor for all u-bmc services. Rather than using external process managers like systemd or supervisord, u-bmc includes its own lightweight operator that understands the specific dependencies and requirements of BMC services. This design provides better control over service interactions and enables more sophisticated coordination during critical operations like power management and thermal emergencies.

The operator implements a dependency-aware startup sequence, beginning with foundational services like IPC and progressing through core management services before finally starting user-facing services like the web server. During shutdown, this process is reversed while maintaining service dependencies to ensure clean termination.

## Architecture

The operator follows a hierarchical service management model where services are organized into startup groups based on their dependencies. The startup sequence is carefully orchestrated to ensure that each service has access to its required dependencies before initialization.

### Startup Sequence

The operator manages services in several distinct phases during startup. The IPC service starts first as it provides the communication backbone for all other services. Core hardware management services like power management, sensor monitoring, and thermal management start next, as they provide essential system monitoring capabilities. Management services including user management, security, and inventory follow, building on the foundation provided by the core services. Finally, protocol and interface services like the web server, IPMI compatibility, and planned KVM services start to provide external access to the system.

This phased approach ensures that critical system functions are available before less essential services and that each service can reliably communicate with its dependencies through the NATS IPC layer.

### Service Dependencies

The operator maintains a comprehensive understanding of service dependencies through its configuration. Hardware management services depend on IPC for communication and may have interdependencies for coordinated responses to system events. Management services typically depend on both IPC and relevant hardware services, while interface services depend on the full stack of underlying services.

For example, the thermal management service requires access to sensor data from the sensor monitoring service and may need to coordinate emergency responses with the power management service. The operator ensures these dependencies are satisfied before allowing thermalmgr to start.

### Health Monitoring

Beyond initial startup coordination, the operator continuously monitors service health through multiple mechanisms. Services report their status through NATS heartbeat messages, and the operator can detect service failures through connection monitoring. When service failures are detected, the operator can initiate restart procedures or coordinate system-wide responses depending on the criticality of the failed service.

The health monitoring system distinguishes between different types of service failures and responds appropriately. Transient failures in non-critical services may trigger simple restarts, while failures in critical services like power management may require more comprehensive recovery procedures.

## Configuration

The operator service uses the standard u-bmc configuration pattern with a `New()` function accepting option functions. This allows for flexible configuration of startup behavior, timeout values, and service-specific parameters.

### Basic Configuration

```go
op := operator.New(
    operator.WithServiceName("operator"),
    operator.WithServiceDescription("BMC Service Supervisor"),
    operator.WithStartupTimeout(30 * time.Second),
    operator.WithShutdownTimeout(15 * time.Second),
    operator.WithHealthCheckInterval(5 * time.Second),
    operator.WithMetrics(true),
    operator.WithTracing(true),
)
```

### Service Registration

Services are registered with the operator during system initialization, typically in the platform-specific main function. Each service registration includes dependency information and startup parameters.

```go
// Register core services
op.RegisterService("ipc", ipcService,
    operator.WithServiceGroup("foundation"),
    operator.WithCritical(true),
    operator.WithStartupDelay(0),
)

op.RegisterService("powermgr", powerService,
    operator.WithServiceGroup("core"),
    operator.WithDependencies("ipc"),
    operator.WithCritical(true),
    operator.WithStartupDelay(1 * time.Second),
)

op.RegisterService("websrv", webService,
    operator.WithServiceGroup("interface"),
    operator.WithDependencies("ipc", "powermgr", "statemgr"),
    operator.WithCritical(false),
    operator.WithStartupDelay(2 * time.Second),
)
```

### Timeout Configuration

The operator provides configurable timeout values for different phases of service lifecycle management. Startup timeouts control how long the operator waits for services to complete initialization, while shutdown timeouts limit the time allowed for graceful service termination.

```go
op := operator.New(
    operator.WithStartupTimeout(45 * time.Second),
    operator.WithShutdownTimeout(20 * time.Second),
    operator.WithServiceStartTimeout(10 * time.Second),
    operator.WithServiceStopTimeout(5 * time.Second),
    operator.WithHealthCheckTimeout(3 * time.Second),
)
```

## Service Lifecycle Management

The operator manages the complete lifecycle of u-bmc services from initial startup through runtime monitoring to graceful shutdown. This comprehensive lifecycle management ensures system reliability and proper coordination between services.

### Service Startup

During startup, the operator processes services in dependency order within each startup group. Services within the same group that have no interdependencies may start concurrently to reduce overall startup time. The operator monitors each service's initialization and only proceeds to the next group once all services in the current group have successfully started.

Service startup includes several phases: pre-initialization setup where the operator prepares the service's runtime environment, actual service initialization where the service performs its startup procedures, and post-initialization verification where the operator confirms the service is ready to handle requests.

If a service fails to start within its configured timeout, the operator can take several actions depending on the service's criticality. Critical service failures typically result in system startup failure, while non-critical service failures may be logged and skipped to allow system operation with reduced functionality.

### Runtime Monitoring

Once all services are running, the operator transitions to runtime monitoring mode where it continuously tracks service health and performance. This monitoring includes regular health check requests sent through NATS, monitoring of service response times, and detection of service crashes or hangs.

The operator maintains service statistics including uptime, restart counts, and health check success rates. These statistics help identify problematic services and can inform decisions about service restart policies or system maintenance requirements.

When service issues are detected, the operator can initiate various recovery actions. Simple service restarts are attempted first for transient failures, while persistent failures may trigger more comprehensive recovery procedures including dependency service restarts or system-wide failsafe modes.

### Graceful Shutdown

System shutdown reverses the startup process while maintaining service dependencies. Interface services stop first to prevent new external requests, followed by management services, core services, and finally the IPC foundation. This sequence ensures that services can complete their shutdown procedures while still having access to required dependencies.

The operator coordinates shutdown timing to allow services to complete critical operations before termination. For example, the power management service may need time to safely sequence power supplies, while the thermal management service ensures fans remain operational until shutdown is complete.

Emergency shutdown procedures bypass normal graceful shutdown timeouts when immediate system halt is required, such as during thermal emergencies or critical hardware failures detected by hardware management services.

## Integration with Services

The operator integrates deeply with all u-bmc services through standardized interfaces and communication patterns. This integration enables sophisticated coordination and monitoring capabilities while maintaining clean service boundaries.

### NATS Integration

All communication between the operator and managed services occurs through the NATS IPC layer. This communication includes service lifecycle commands, health check requests and responses, and event notifications for service state changes.

Services implement standard NATS endpoints for operator communication including health check endpoints that report service status and readiness, lifecycle endpoints for startup and shutdown coordination, and event publishing for status changes and significant operational events.

The operator subscribes to service event streams to maintain real-time awareness of service status and can publish system-wide events to coordinate responses to significant system changes or emergency conditions.

### Service Interface Standardization

All managed services implement a common interface that the operator uses for lifecycle management. This interface includes methods for service initialization and startup, health checking and status reporting, graceful shutdown and cleanup, and configuration validation and updates.

The standardized interface ensures consistent behavior across all services while allowing each service to implement its specific functionality. Services can extend the base interface with additional capabilities while maintaining compatibility with operator management.

Service implementation follows patterns established by existing services like `statemgr` and `powermgr`, ensuring consistency in configuration, error handling, and operational behavior across the entire system.

## Error Handling and Recovery

The operator implements comprehensive error handling and recovery mechanisms to maintain system reliability in the face of service failures or unexpected conditions. These mechanisms range from simple service restarts to complex system-wide recovery procedures.

### Service Restart Policies

Different services have different restart policies based on their criticality and typical failure modes. Critical services like power management have aggressive restart policies with multiple retry attempts and escalation to system-wide recovery if restarts fail. Non-critical services may have more relaxed restart policies that prioritize system stability over service availability.

The operator tracks restart attempts and implements exponential backoff to prevent restart loops that could impact system stability. Services that repeatedly fail to start may be marked as failed and excluded from further restart attempts until manual intervention occurs.

Restart policies can be customized per service based on operational requirements and observed failure patterns. Some services may benefit from immediate restart attempts, while others may require delays to allow transient conditions to clear.

### System-Wide Recovery

When critical service failures cannot be resolved through normal restart procedures, the operator can initiate system-wide recovery actions. These actions may include restarting entire service groups to clear complex dependency issues, activating failsafe modes that provide minimal functionality while troubleshooting occurs, or coordinating emergency system responses for hardware-related failures.

System-wide recovery procedures are coordinated with hardware management services to ensure that critical system functions like thermal management and power sequencing remain operational during recovery operations.

The operator maintains detailed logs of all recovery actions to support post-incident analysis and system improvement efforts.

### Monitoring and Alerting

The operator provides extensive monitoring capabilities that integrate with the broader u-bmc telemetry system. Service health metrics, startup and shutdown timing, restart frequencies, and error rates are all tracked and made available for monitoring and alerting systems.

Integration with OpenTelemetry provides distributed tracing capabilities that help diagnose complex issues involving multiple services and their interactions.

## Security and Permissions

The operator service runs with elevated privileges necessary to manage other system services and coordinate system-wide operations. These privileges are carefully managed to minimize security exposure while enabling necessary functionality.

### Process Management

The operator requires permissions to start, stop, and monitor system processes corresponding to u-bmc services. These permissions are typically provided through process capabilities or appropriate user group membership rather than full root privileges where possible.

Service isolation ensures that individual services run with minimal required permissions, with the operator coordinating activities that require elevated privileges on behalf of managed services.

### Resource Access

Hardware management services may require access to system resources like GPIO devices, I2C buses, or hardware monitoring interfaces. The operator coordinates access to these resources and can implement resource locking or sharing policies to prevent conflicts between services.

Configuration and state persistence may require access to specific filesystem locations or storage devices. The operator manages these access requirements while maintaining appropriate security boundaries between services.

## Platform Integration

The operator service integrates with platform-specific configuration and hardware management requirements through its service registration and configuration system. This integration allows the operator to adapt to different hardware platforms while maintaining consistent service management behavior.

### Hardware-Specific Services

Different hardware platforms may require different combinations of services or platform-specific service configurations. The operator's flexible service registration system accommodates these requirements while maintaining dependency management and startup coordination.

Platform-specific main functions register the appropriate services for their hardware configuration, including any platform-specific GPIO, I2C, or thermal management requirements.

Service configuration can be customized per platform while maintaining compatibility with the operator's management interfaces and coordination mechanisms.

### Resource Management

The operator coordinates access to shared hardware resources among multiple services. This coordination prevents conflicts and ensures that hardware resources are used efficiently and safely.

Resource management includes GPIO line allocation for power and LED control, I2C bus access coordination for sensor monitoring and power management, and thermal zone assignment for temperature and fan control.

## Observability and Debugging

The operator provides comprehensive observability features that support both routine monitoring and detailed debugging of service interactions and system behavior.

### Logging and Metrics

All operator activities are logged using structured logging that integrates with the broader u-bmc logging system. Log entries include service lifecycle events, health check results, error conditions and recovery actions, and performance metrics for startup and shutdown operations.

Metrics integration provides quantitative data on service performance and reliability including service startup times, health check success rates, restart frequencies, and resource usage patterns.

### Diagnostic Tools

The operator exposes diagnostic interfaces through NATS that allow real-time inspection of service status and system state. These interfaces support troubleshooting and system analysis without requiring direct access to individual services.

Diagnostic capabilities include service dependency visualization, real-time service status monitoring, service restart and recovery history, and system resource usage analysis.

## Development and Testing

The operator service follows standard u-bmc development practices including comprehensive testing, clear documentation, and integration with the broader development toolchain.

### Package Documentation

Detailed package documentation is available at pkg.go.dev for the operator service and its supporting packages. This documentation includes API references, configuration examples, and integration guidance for developers working with the operator service.

The operator service documentation builds on the foundational patterns established by other u-bmc services while highlighting the unique aspects of service supervision and lifecycle management.

### Testing Infrastructure

The operator service includes comprehensive test suites that cover service lifecycle management scenarios, error handling and recovery procedures, dependency management and coordination, and integration with other u-bmc services.

Test scenarios include simulated service failures, resource constraint conditions, and platform-specific configuration requirements to ensure robust operation across different deployment environments.

## Future Enhancements

The operator service provides a foundation for advanced system management capabilities that may be implemented in future u-bmc releases.

### Advanced Monitoring

Future enhancements may include more sophisticated service health monitoring with predictive failure detection, resource usage optimization and automatic scaling, and integration with external monitoring and alerting systems.

### Dynamic Configuration

Support for dynamic service configuration updates and runtime service deployment and management could enable more flexible system operation and maintenance procedures.

### High Availability

Advanced deployment scenarios may benefit from operator clustering and high availability capabilities that provide redundant service management and automatic failover in distributed BMC environments.

The operator service's modular design and standardized interfaces provide a solid foundation for these advanced capabilities while maintaining compatibility with existing service implementations and platform integrations.

## References

For additional information about the operator service and its integration with other u-bmc components, see the related documentation including the system architecture overview in [docs/architecture.md](architecture.md), platform porting guidance in [docs/porting.md](porting.md), and individual service documentation for integration details.

Package-level documentation and API references are available on pkg.go.dev for detailed implementation information and development guidance.