# IPC Service

The IPC service provides the communication backbone for all u-bmc services through an embedded NATS server and connection management layer. It enables low-latency, reliable messaging between services using request-reply patterns, publish-subscribe messaging, and optional persistent streaming through JetStream. The IPC service is the first service started by the operator and serves as the foundation for all inter-service communication within u-bmc.

## Overview

Rather than relying on external message brokers or communication mechanisms, u-bmc includes its own embedded NATS server that provides high-performance, in-process messaging. This design eliminates external dependencies, reduces operational complexity, and provides better control over message routing and service coordination. The embedded approach also enables sophisticated features like message tracing, service discovery, and coordinated shutdown procedures.

The IPC service handles both the NATS server lifecycle and connection management for u-bmc services. Services connect to the embedded NATS instance through standardized connection patterns and use well-defined subject hierarchies for different types of communication. This structured approach ensures predictable message routing and enables comprehensive monitoring of service interactions.

## Architecture

The IPC service implements a dual-layer architecture with the embedded NATS server providing core messaging capabilities and a connection management layer that handles u-bmc-specific requirements like service registration, health monitoring, and graceful shutdown coordination.

### Embedded NATS Server

The embedded NATS server runs within the u-bmc process space and provides all standard NATS functionality including core request-reply and publish-subscribe messaging, JetStream for persistent messaging and event sourcing, clustering capabilities for distributed deployments, and comprehensive monitoring and metrics collection.

The server configuration is optimized for BMC workloads with appropriate buffer sizes, connection limits, and performance tuning for the constrained resources typical of BMC hardware. Security features including authentication and authorization are configured to ensure that only authorized services can access the message bus.

### Connection Management

The connection management layer provides u-bmc-specific abstractions over raw NATS connections. This includes automatic connection establishment and recovery, service registration and discovery mechanisms, structured subject hierarchies for different message types, and integration with the broader u-bmc telemetry and logging systems.

Services connect through standardized connection factories that handle authentication, subject prefix configuration, and connection monitoring. The connection layer also provides helper functions for common messaging patterns used throughout u-bmc.

### Message Patterns

The IPC service supports several distinct messaging patterns optimized for different use cases within u-bmc. Request-reply messaging handles synchronous service-to-service communication with timeout and error handling. Publish-subscribe messaging enables event distribution and asynchronous notifications between services. Stream-based messaging through JetStream provides persistence and replay capabilities for critical system events and audit trails.

Each messaging pattern uses specific subject hierarchies and message formats to ensure consistency and enable comprehensive monitoring of service interactions throughout the system.

## Subject Hierarchies

The IPC service enforces structured subject hierarchies that organize messages by service, operation type, and specific functionality. This organization enables efficient message routing, comprehensive monitoring, and clear service boundaries.

### Service-Specific Subjects

Each service uses a dedicated subject prefix that identifies its messages and provides namespace isolation. For example, the power management service uses subjects like `powermgr.hosts.power.set` for power control operations and `powermgr.chassis.status.get` for status queries.

The hierarchical structure enables wildcard subscriptions for monitoring and debugging while maintaining precise routing for operational messages. Services can subscribe to their own subject trees while monitoring tools can observe broader patterns across multiple services.

### System-Wide Subjects

System-wide subjects handle cross-cutting concerns like health monitoring, event distribution, and emergency coordination. These subjects follow standardized patterns that all services understand and use for system-wide coordination.

Health check subjects enable the operator service to monitor all services through a consistent interface, while emergency subjects provide immediate notification channels for critical system events that require coordinated responses from multiple services.

### Event Streaming

JetStream subjects organize persistent events into logical streams based on their source and content. Operational events from service activities, system events like power state changes and thermal alerts, and audit events for security and compliance tracking each use dedicated stream hierarchies.

The streaming organization enables efficient querying and replay of historical events while maintaining clear data ownership and access control boundaries between different types of system information.

## Service Integration

The IPC service provides the foundation for all service-to-service communication within u-bmc through standardized connection patterns and messaging interfaces. Services integrate with IPC through well-defined APIs that handle connection management, message routing, and error handling.

### Connection Establishment

Services establish IPC connections through factory functions that handle authentication, configuration, and connection monitoring. The connection process includes service registration with the IPC layer, subject prefix configuration based on service identity, and establishment of standard health check and lifecycle endpoints.

Connection factories provide appropriate configuration for different service types, with hardware management services receiving different buffer and timeout settings than interface services like the web server. This customization ensures optimal performance for each service's communication patterns.

### Message Handling

Services implement message handlers using standardized patterns that integrate with NATS subscription mechanisms while providing u-bmc-specific features like request validation, response formatting, and error handling. Handler registration includes subject pattern specification, message type validation, and integration with service-specific logging and metrics.

The message handling layer provides automatic serialization and deserialization for protobuf messages, timeout and retry handling for requests, and integration with distributed tracing for comprehensive observability across service interactions.

### Service Discovery

The IPC service provides service discovery mechanisms that enable services to find and communicate with each other without requiring static configuration. Services register their capabilities and endpoints during startup, and other services can query for available services and their supported operations.

Service discovery includes health status integration so that services can avoid routing requests to failed or unavailable services. The discovery system also supports service versioning and capability negotiation for future compatibility requirements.

## Configuration

The IPC service uses the standard u-bmc configuration pattern with comprehensive options for server behavior, connection management, and integration features.

### Basic Configuration

```go
ipcService := ipc.New(
    ipc.WithServiceName("ipc"),
    ipc.WithServiceDescription("Embedded NATS IPC Service"),
    ipc.WithServerPort(4222),
    ipc.WithClusterPort(6222),
    ipc.WithMonitorPort(8222),
    ipc.WithMaxConnections(100),
    ipc.WithMaxSubscriptions(1000),
    ipc.WithMetrics(true),
    ipc.WithTracing(true),
)
```

### Server Configuration

```go
ipcService := ipc.New(
    ipc.WithServerConfig(nats.Options{
        Host:           "127.0.0.1",
        Port:           4222,
        MaxConnections: 50,
        MaxSubs:        500,
        MaxPayload:     1024 * 1024, // 1MB
        WriteDeadline:  2 * time.Second,
    }),
    ipc.WithTLSConfig(&tls.Config{
        MinVersion: tls.VersionTLS13,
    }),
    ipc.WithAuthentication(true),
    ipc.WithAuthorization(true),
)
```

### JetStream Configuration

```go
ipcService := ipc.New(
    ipc.WithJetStream(true),
    ipc.WithJetStreamMaxMemory(64 * 1024 * 1024), // 64MB
    ipc.WithJetStreamMaxStorage(256 * 1024 * 1024), // 256MB
    ipc.WithStreamRetention(24 * time.Hour),
    ipc.WithEventStreams("system.events", "service.events", "audit.events"),
)
```

### Connection Management

```go
ipcService := ipc.New(
    ipc.WithConnectionTimeout(5 * time.Second),
    ipc.WithReconnectWait(2 * time.Second),
    ipc.WithMaxReconnectAttempts(10),
    ipc.WithHealthCheckInterval(30 * time.Second),
    ipc.WithServiceRegistry(true),
    ipc.WithServiceDiscovery(true),
)
```

## Performance and Tuning

The IPC service includes extensive configuration options for optimizing performance based on deployment requirements and hardware constraints. These optimizations ensure reliable operation on BMC-class hardware while providing sufficient performance for system management workloads.

### Memory Management

Buffer sizes and memory limits are configurable based on expected message volumes and available system memory. The service includes automatic buffer management that adjusts allocation based on actual usage patterns while maintaining performance guarantees for critical system communications.

Connection pooling and message batching reduce memory overhead while maintaining low latency for individual operations. The memory management system integrates with system monitoring to provide alerts when resource usage approaches configured limits.

### Network Optimization

Network configuration options optimize for the low-latency, high-reliability requirements of BMC communication. TCP keepalive settings ensure rapid detection of connection failures, while buffer sizes are tuned for the typical message sizes used by u-bmc services.

The embedded server approach eliminates network round-trips for most communications, but configuration options support distributed deployments where services run on separate systems with network-based NATS communication.

### Throughput Tuning

Message throughput optimization includes batching strategies for high-volume communications like sensor data collection, priority queuing for critical system messages, and flow control mechanisms that prevent resource exhaustion during traffic spikes.

Performance monitoring provides real-time visibility into message rates, processing latencies, and resource utilization to support performance tuning and capacity planning activities.

## Security

The IPC service implements comprehensive security features that protect inter-service communications while maintaining the performance and reliability required for BMC operations.

### Authentication and Authorization

Service authentication ensures that only legitimate u-bmc services can connect to the message bus. Authentication mechanisms include service identity verification based on cryptographic credentials, connection source validation for additional security layers, and integration with the broader u-bmc security framework.

Authorization controls determine which subjects each service can access, preventing services from interfering with operations outside their designated scope. Authorization policies are configured based on service roles and responsibilities within the system architecture.

### Message Security

Message-level security features protect sensitive information transmitted between services. This includes automatic encryption for sensitive message types, message integrity verification to detect tampering, and audit logging for security-relevant communications.

The security framework integrates with hardware security features where available, including TPM-based key management and secure boot verification of service identities.

### Network Security

Network security features protect against external attacks and unauthorized access attempts. The embedded server configuration includes network isolation that prevents external access to internal communications, TLS encryption for any network-based communications, and monitoring and alerting for suspicious connection patterns.

Security policies can be customized based on deployment requirements, with more restrictive policies available for high-security environments and more permissive policies for development and testing scenarios.

## Monitoring and Observability

The IPC service provides extensive monitoring capabilities that enable comprehensive observability of service communications and system interactions. These capabilities support both operational monitoring and detailed debugging of complex service interaction patterns.

### Message Tracing

Distributed tracing integration tracks messages across service boundaries and provides detailed timing information for request-reply interactions. Trace data includes message routing paths, processing times at each service, and error conditions or timeouts that affect message delivery.

Tracing integration with OpenTelemetry enables correlation with broader system traces and supports advanced analysis of service interaction patterns and performance characteristics.

### Metrics Collection

Comprehensive metrics track all aspects of IPC operation including message rates and volumes by service and subject, connection statistics and health information, resource usage including memory and network utilization, and error rates and failure modes for different types of communications.

Metrics are exposed through standard interfaces that integrate with monitoring systems and support alerting based on operational thresholds and performance trends.

### Service Health Monitoring

The IPC service provides centralized health monitoring for all connected services through standardized health check mechanisms. Health information includes service availability and response times, resource usage and performance metrics, and error rates and recent failure history.

Health monitoring enables the operator service to make informed decisions about service restarts and system recovery procedures while providing visibility into overall system health and performance.

## Error Handling and Recovery

Robust error handling and recovery mechanisms ensure that communication failures do not compromise system reliability or availability. The IPC service implements multiple layers of error detection and recovery to handle both transient and persistent failure conditions.

### Connection Recovery

Automatic connection recovery handles network interruptions and service restarts without requiring manual intervention. Recovery mechanisms include exponential backoff for reconnection attempts, message buffering during disconnection periods, and automatic resubscription to required subjects after reconnection.

Connection recovery integrates with service health monitoring to distinguish between temporary network issues and permanent service failures, enabling appropriate recovery strategies for different failure modes.

### Message Delivery Guarantees

Message delivery guarantees ensure that critical system communications are not lost during failure conditions. The service supports multiple delivery modes including at-most-once delivery for non-critical messages where performance is prioritized, at-least-once delivery for important messages that must be processed, and exactly-once delivery for critical operations that cannot tolerate duplication.

JetStream integration provides persistent messaging capabilities for the most critical system communications, ensuring that messages are not lost even during system-wide failures or restarts.

### Failure Detection and Response

Comprehensive failure detection mechanisms identify and respond to various types of communication failures. Detection includes network connectivity monitoring, service response time tracking, and message delivery confirmation systems.

Response mechanisms range from automatic retry and reconnection for transient failures to coordinated service restart procedures for persistent failures that cannot be resolved through normal recovery mechanisms.

## Development and Integration

The IPC service follows standard u-bmc development practices and provides comprehensive APIs and tools for service integration and development.

### Package Documentation

Detailed package documentation is available at pkg.go.dev covering all aspects of IPC service integration including connection management APIs, message handling patterns, configuration options and examples, and error handling and recovery procedures.

The documentation includes practical examples of common integration patterns used by other u-bmc services, providing templates for new service development and integration activities.

### Testing Infrastructure

Comprehensive test suites cover all aspects of IPC functionality including embedded server lifecycle and configuration, connection management and recovery scenarios, message routing and delivery guarantees, and integration with service monitoring and health checking systems.

Test infrastructure includes mock services and failure injection capabilities that enable thorough testing of error handling and recovery mechanisms under various failure conditions.

### Development Tools

Development and debugging tools support service development and system troubleshooting activities. These tools include message monitoring and inspection utilities, connection status and health monitoring interfaces, performance analysis and profiling capabilities, and configuration validation and testing frameworks.

The development tools integrate with standard Go development workflows and provide both command-line and programmatic interfaces for different types of development and operational activities.

## Future Enhancements

The IPC service provides a foundation for advanced messaging and communication capabilities that may be implemented in future u-bmc releases based on operational requirements and user feedback.

### Advanced Messaging Patterns

Future enhancements may include more sophisticated messaging patterns such as message routing and transformation capabilities, advanced queuing and priority management systems, and integration with external messaging systems for distributed deployments.

### Enhanced Security

Advanced security features could include end-to-end encryption for all service communications, integration with hardware security modules for key management, and advanced audit and compliance capabilities for regulated environments.

### Performance Optimization

Continued performance optimization may focus on specialized optimizations for BMC hardware constraints, advanced caching and batching strategies for high-volume communications, and integration with hardware acceleration features where available.

The modular design of the IPC service and its standardized interfaces provide a solid foundation for these enhancements while maintaining compatibility with existing service implementations and integration patterns.

## References

Additional information about the IPC service and its role in the broader u-bmc architecture is available in related documentation including system architecture details in [docs/architecture.md](architecture.md), service development guidance in individual service documentation, and platform-specific integration information in [docs/porting.md](porting.md).

Package-level documentation and detailed API references are maintained on pkg.go.dev for comprehensive development and integration guidance.