# API Documentation

The u-bmc API provides comprehensive access to BMC functionality through ConnectRPC with automatic REST transcoding. The API is designed for both programmatic access and human-readable REST operations, with consistent authentication, error handling, and observability across all endpoints. All API operations are defined in protobuf schemas under `schema/v1alpha1/` which serve as the authoritative API contract.

## Overview

The u-bmc API follows modern RPC design principles with type-safe interfaces, comprehensive error handling, and extensive observability features. The primary protocol is ConnectRPC which provides efficient binary serialization, streaming capabilities, and robust error semantics. REST endpoints are automatically generated from protobuf HTTP annotations to provide compatibility with standard HTTP tooling.

All API access flows through the websrv service which handles authentication, authorization, request routing, and protocol translation. The web server coordinates with backend services through NATS IPC to fulfill API requests while maintaining security boundaries and providing comprehensive audit logging for all management operations.

The API design emphasizes consistency across different resource types and operations while providing the flexibility necessary for diverse BMC management scenarios. Resource hierarchies follow logical system organization, and operation semantics remain consistent whether accessed through ConnectRPC or REST endpoints.

## Protocol Support

The u-bmc API supports multiple protocols to accommodate different client requirements and integration scenarios while maintaining consistent behavior and security properties across all access methods.

### ConnectRPC Protocol

ConnectRPC serves as the primary API protocol, providing efficient RPC semantics with modern features including type-safe request and response handling, comprehensive error information with structured error codes and details, streaming support for real-time data feeds and long-running operations, and automatic retry and timeout handling in client libraries.

ConnectRPC clients are available for major programming languages including Go, JavaScript/TypeScript, Python, and Java. These clients provide idiomatic interfaces that handle connection management, authentication, and error handling automatically while exposing the full functionality of the u-bmc API.

The protocol supports both unary RPCs for simple request-response operations and streaming RPCs for real-time monitoring, event feeds, and operations that produce incremental results over time.

### REST Compatibility

REST endpoints provide broad compatibility with existing HTTP tooling and enable integration scenarios where ConnectRPC clients are not available or appropriate. All REST endpoints are automatically generated from the same protobuf definitions used for ConnectRPC, ensuring consistent behavior and eliminating maintenance overhead for dual API implementations.

REST endpoint URLs follow conventional patterns with resource hierarchies that match the logical organization of BMC components. HTTP methods map naturally to CRUD operations while maintaining the semantic richness of the underlying protobuf operations. Query parameters and request bodies follow standard REST conventions while preserving type safety and validation from the protobuf schema.

Content negotiation supports both JSON and binary protobuf payloads, enabling clients to choose the most appropriate serialization format for their requirements. Error responses follow standard HTTP status code conventions while providing detailed error information compatible with both automated tooling and human debugging.

### Authentication and Authorization

All API access requires authentication through mechanisms integrated with the u-bmc user management system. Authentication methods include session-based authentication for web interfaces with secure cookie management and CSRF protection, token-based authentication for API clients with configurable token lifetimes and refresh capabilities, and certificate-based authentication for high-security scenarios with mutual TLS verification.

Authorization decisions are made by the security manager service based on user roles, requested operations, and current system context. The authorization system supports both simple role-based access control and more sophisticated attribute-based policies that consider factors like request source, time constraints, and current system state.

Authentication state is maintained consistently across requests and protocols, enabling seamless transitions between different types of client interactions while maintaining appropriate security boundaries and audit capabilities.

## API Structure and Organization

The u-bmc API is organized around logical resource hierarchies that reflect the structure of BMC-managed systems and provide intuitive navigation for both human operators and programmatic clients.

### System Resources

System-level resources provide high-level information and control capabilities for overall system management. The system resource hierarchy includes general system information like hardware configuration, firmware versions, and operational status, power management for system-wide power operations and policy configuration, thermal management for temperature monitoring and cooling control across the entire system, and health monitoring that aggregates status information from all system components.

System resources enable operations that affect multiple components or provide system-wide views of operational status and configuration information.

### Host Resources

Host resources represent individual compute nodes managed by the BMC, including their power states, hardware configuration, and operational status. Host resource operations include power control for individual hosts with support for graceful and forced power operations, status monitoring that provides real-time information about host operational status and health, hardware inventory that details host hardware configuration and component information, and console access for serial console and remote management interface access.

Host resources support both individual host operations and bulk operations that can affect multiple hosts simultaneously while maintaining appropriate safety checks and coordination between related operations.

### Chassis Resources

Chassis resources represent the physical enclosure and shared infrastructure that supports multiple hosts or system components. Chassis operations include power supply management for shared power infrastructure, cooling system control for fans, thermal zones, and environmental management, shared resource management for network infrastructure, storage controllers, and other shared components, and infrastructure monitoring that provides visibility into shared system health and performance.

Chassis resources enable coordination of infrastructure operations while providing appropriate isolation and safety mechanisms for operations that affect multiple hosted systems.

### Component Resources

Component resources provide detailed access to individual hardware components including sensors, power supplies, storage devices, and network interfaces. Component operations include detailed monitoring and status reporting for individual hardware elements, configuration management for component-specific settings and operational parameters, health monitoring that provides component-level diagnostic information and failure prediction, and maintenance operations that support component replacement, testing, and lifecycle management.

Component resources provide the detailed access necessary for comprehensive system management while maintaining appropriate abstraction to support diverse hardware configurations and component types.

## Endpoint Reference

The following sections provide detailed information about major API endpoint categories and their operations. Complete endpoint documentation with request/response schemas is available through the API documentation generated from protobuf definitions.

### System Management Endpoints

System management endpoints provide high-level system control and monitoring capabilities that span multiple subsystems and components.

System status endpoints return comprehensive system health and operational information including overall system state, component health summaries, active alerts and warnings, and operational metrics. These endpoints provide the information necessary for system-wide monitoring and health assessment.

System configuration endpoints enable management of system-wide settings including network configuration, time synchronization, logging policies, and security settings. Configuration changes are validated and applied consistently across all relevant system components.

System operation endpoints support system-wide operations including graceful shutdown and restart procedures, maintenance mode transitions, and system-wide diagnostic operations. These endpoints coordinate complex operations that require synchronization across multiple services and components.

### Power Management Endpoints

Power management endpoints provide comprehensive control over system power operations including individual hosts, chassis infrastructure, and overall system power management.

Host power endpoints support individual host power operations including power on, power off, reset, and force off operations. Each operation includes appropriate safety checks and coordination with thermal and health monitoring systems. Power operations support both immediate execution and scheduled operations for maintenance and operational planning.

Chassis power endpoints manage shared power infrastructure including power supply operations, power distribution management, and emergency power procedures. Chassis power operations coordinate with hosted systems to ensure safe operation and prevent conflicts between infrastructure and host power requirements.

System power endpoints provide overall power management including system-wide power budgeting, emergency shutdown procedures, and power efficiency optimization. These endpoints enable comprehensive power management policies that balance operational requirements with efficiency and safety considerations.

### Hardware Monitoring Endpoints

Hardware monitoring endpoints provide detailed access to sensor data, component health information, and system performance metrics collected throughout the BMC-managed system.

Sensor endpoints provide access to real-time and historical sensor data including temperature, voltage, current, fan speed, and other environmental measurements. Sensor data includes threshold information, trend analysis, and alert generation for conditions that require attention.

Component health endpoints aggregate health information from individual system components including diagnostic results, failure prediction information, and maintenance recommendations. Health data supports both reactive maintenance based on current conditions and proactive maintenance based on trend analysis and predictive algorithms.

System performance endpoints provide system-wide performance metrics including resource utilization, throughput measurements, and efficiency calculations. Performance data supports capacity planning, optimization activities, and operational decision making.

### User and Security Management Endpoints

User and security management endpoints handle authentication, authorization, and security policy management for the BMC system.

User management endpoints support user account creation, modification, and removal along with password management, role assignment, and access control configuration. User operations include comprehensive audit logging for security and compliance requirements.

Authentication endpoints handle login, logout, and session management operations including multi-factor authentication support and integration with external authentication systems. Authentication operations provide the foundation for all other API access and maintain appropriate security boundaries.

Security policy endpoints enable configuration of system security policies including access control rules, audit policies, and security monitoring configuration. Security operations support both operational security management and compliance with regulatory and organizational security requirements.

## Request and Response Patterns

The u-bmc API follows consistent patterns for request structure, response formatting, and error handling across all endpoints and protocols.

### Request Structure

API requests follow standardized patterns that provide consistency and predictability across different operations and resource types. Request identification includes unique request IDs that support distributed tracing and audit logging, authentication tokens or session information for security verification, and operation context information that may affect processing behavior.

Request parameters are validated against protobuf schemas to ensure type safety and completeness before processing. Validation includes structural validation of request format and content, semantic validation of parameter values and relationships, and authorization validation that confirms the requester has permission for the requested operation.

Request processing includes comprehensive logging and monitoring that provides visibility into API usage patterns, performance characteristics, and error conditions for operational monitoring and troubleshooting.

### Response Structure

API responses provide comprehensive information about operation results including detailed status information that indicates operation success or failure, resource data that provides the information requested by the operation, and metadata that supports client processing and caching decisions.

Response formatting is consistent across protocols with automatic serialization to appropriate formats based on client preferences and protocol requirements. Response data includes comprehensive error information when operations fail, pagination support for operations that return large result sets, and caching information that enables efficient client behavior.

Response validation ensures that all returned data conforms to API schemas and that sensitive information is appropriately filtered based on client authorization levels and security policies.

### Error Handling

Comprehensive error handling provides clear information about operation failures and guidance for resolving problems. Error responses include structured error codes that enable programmatic error handling, detailed error messages that provide human-readable descriptions of problems, and contextual information that helps identify the cause and potential solutions for errors.

Error categories include validation errors for malformed or invalid requests, authorization errors for operations that exceed client permissions, resource errors for operations that cannot be completed due to system state or resource constraints, and system errors for internal problems that prevent operation completion.

Error handling integrates with distributed tracing and logging systems to provide comprehensive debugging information while maintaining appropriate security boundaries for error details and system information.

## Streaming and Real-Time APIs

The u-bmc API provides extensive streaming capabilities that enable real-time monitoring, event-driven integration, and efficient handling of long-running operations.

### Event Streams

Event streaming endpoints provide real-time notification of system changes and operational activities. Event streams include system state changes for power transitions, health status updates, and configuration modifications, hardware events for sensor threshold violations, component failures, and performance alerts, and operational events for user activities, maintenance operations, and system administration tasks.

Event streams support filtering and subscription management that enables clients to receive only relevant events while minimizing bandwidth and processing overhead. Stream management includes automatic reconnection and backfill capabilities that ensure clients maintain consistent event coverage even during network interruptions.

Event stream integration with JetStream provides persistence and replay capabilities for critical events, enabling audit trails and historical analysis of system activities and changes.

### Sensor Data Streams

Sensor data streaming provides efficient access to real-time sensor information without requiring continuous polling. Sensor streams include configurable update rates that balance data freshness with resource consumption, threshold-based updates that provide immediate notification of significant changes, and batch updates that provide efficient delivery of data from multiple sensors.

Sensor streaming supports both individual sensor subscriptions for focused monitoring and bulk subscriptions that provide comprehensive sensor coverage with efficient resource utilization. Stream processing includes data aggregation and filtering capabilities that reduce bandwidth requirements while maintaining data quality and timeliness.

### Operation Progress Streams

Long-running operations provide progress streaming that enables clients to monitor operation status and completion without requiring periodic status polling. Progress streams include detailed progress information for complex operations like firmware updates, system diagnostics, and bulk configuration changes, error reporting that provides immediate notification of operation problems, and completion notification that signals successful operation completion or final error status.

Progress streaming enables responsive user interfaces and automation systems that can provide appropriate feedback and take timely action based on operation status and results.

## Authentication and Security

API security implementation provides comprehensive protection for BMC management operations while maintaining usability and integration capabilities required for operational environments.

### Authentication Methods

Multiple authentication methods support diverse client requirements and security policies. Session authentication provides secure browser-based access with HTTP-only cookies, CSRF protection, and configurable session timeouts. Token authentication supports API clients with bearer tokens that include configurable lifetimes, refresh capabilities, and scope-based permissions.

Certificate authentication enables high-security scenarios with mutual TLS verification, certificate-based identity verification, and integration with PKI infrastructure. Authentication methods can be combined to support complex scenarios that require different authentication approaches for different types of access.

Authentication integration with external systems supports LDAP, Active Directory, and other enterprise authentication systems while maintaining local authentication capabilities for scenarios where external systems are unavailable.

### Authorization Framework

Comprehensive authorization controls ensure that API access is properly restricted based on user roles, system policies, and operational context. Role-based access control provides foundational authorization with configurable roles that map to different levels of system access and operational capability.

Attribute-based authorization enables sophisticated policies that consider additional factors including request source, time constraints, system operational state, and resource-specific permissions. Authorization policies support both permissive and restrictive approaches that can be configured based on operational requirements and risk tolerance.

Authorization decisions integrate with comprehensive audit logging that provides detailed records of access attempts, permission grants and denials, and policy enforcement activities for security monitoring and compliance requirements.

### Security Monitoring

Security monitoring capabilities provide continuous assessment of API access patterns and security-relevant activities. Monitoring includes access pattern analysis that can identify suspicious behavior and potential security threats, authentication failure tracking that detects brute force attacks and credential compromise attempts, and authorization violation monitoring that identifies privilege escalation attempts and policy violations.

Security monitoring integrates with alerting systems to provide immediate notification of security events that require attention, and with audit systems to provide comprehensive records for forensic analysis and compliance reporting.

## Performance and Optimization

API performance optimization ensures responsive operation within the resource constraints typical of BMC hardware while supporting the concurrency requirements of operational management scenarios.

### Request Processing Optimization

Efficient request processing minimizes latency and resource consumption for API operations. Processing optimization includes intelligent request routing that directs requests to appropriate backend services, connection pooling and reuse that reduces connection setup overhead, and request batching that improves efficiency for bulk operations.

Processing includes comprehensive caching strategies that reduce backend service load for frequently accessed information while maintaining data freshness and consistency requirements for management operations.

### Response Optimization

Response optimization reduces bandwidth requirements and improves client performance through intelligent data formatting and delivery mechanisms. Optimization includes response compression that reduces network bandwidth requirements, selective field inclusion that enables clients to request only necessary data, and efficient serialization that balances processing requirements with wire format efficiency.

Response caching provides appropriate cache control headers and etag support that enable efficient client caching while maintaining consistency requirements for dynamic management information.

### Concurrency and Rate Limiting

Concurrency management ensures stable API performance under varying load conditions while protecting backend services from overload. Concurrency controls include request rate limiting that prevents individual clients from overwhelming the system, connection limits that manage resource consumption, and operation queuing that provides fair access during periods of high demand.

Rate limiting policies can be configured based on client authentication, operation types, and system capacity to provide appropriate access levels while maintaining system stability and responsiveness for all users.

## Integration Examples

The following examples demonstrate common API integration patterns using both ConnectRPC and REST interfaces.

### ConnectRPC Client Example

```go
package main

import (
    "context"
    "crypto/tls"
    "net/http"
    
    "connectrpc.com/connect"
    "github.com/u-bmc/u-bmc/schema/v1alpha1/systemv1alpha1connect"
    "github.com/u-bmc/u-bmc/schema/v1alpha1"
)

func main() {
    client := systemv1alpha1connect.NewSystemServiceClient(
        http.DefaultClient,
        "https://bmc.example.com:8443",
        connect.WithTLSConfig(&tls.Config{
            ServerName: "bmc.example.com",
        }),
    )
    
    // Get system status
    resp, err := client.GetSystemStatus(context.Background(), 
        connect.NewRequest(&v1alpha1.GetSystemStatusRequest{}))
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("System Status: %s\n", resp.Msg.Status)
    fmt.Printf("Health: %s\n", resp.Msg.Health)
}
```

### REST Client Example

```bash
# Authentication
curl -X POST https://bmc.example.com:8443/api/v1alpha1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "secret"}' \
  -c cookies.txt

# Get system status
curl -X GET https://bmc.example.com:8443/api/v1alpha1/system/status \
  -b cookies.txt \
  -H "Accept: application/json"

# Power on host
curl -X POST https://bmc.example.com:8443/api/v1alpha1/hosts/host.0/power \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"action": "HOST_POWER_ACTION_ON"}'

# Get sensor data
curl -X GET "https://bmc.example.com:8443/api/v1alpha1/sensors?filter=type:temperature" \
  -b cookies.txt \
  -H "Accept: application/json"
```

### Event Streaming Example

```javascript
// ConnectRPC streaming example
const client = createPromiseClient(SystemService, transport);

const stream = client.streamSystemEvents({});
for await (const event of stream) {
  console.log(`Event: ${event.type} - ${event.description}`);
  
  if (event.type === "SYSTEM_EVENT_POWER_CHANGE") {
    handlePowerEvent(event);
  }
}

// Server-Sent Events (REST)
const eventSource = new EventSource(
  'https://bmc.example.com:8443/api/v1alpha1/events/stream',
  { withCredentials: true }
);

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Received event:', data);
};
```

## Error Codes and Troubleshooting

Comprehensive error handling provides structured error information that enables both programmatic error handling and effective human troubleshooting.

### Common Error Patterns

Authentication errors indicate problems with client authentication including invalid credentials, expired tokens, and session timeout conditions. These errors provide guidance on authentication renewal and credential management.

Authorization errors indicate insufficient permissions for requested operations including role-based restrictions, resource-specific permissions, and policy violations. Error details include information about required permissions and potential resolution approaches.

Validation errors indicate problems with request format or content including missing required fields, invalid parameter values, and constraint violations. Validation errors provide specific information about validation failures and requirements for correct request formatting.

Resource errors indicate problems with system resources or state including unavailable components, conflicting operations, and resource constraint violations. Resource errors provide information about current system state and guidance for resolving resource conflicts.

### Debugging and Diagnostics

API debugging capabilities provide detailed information for troubleshooting client integration and operational issues. Debugging includes request/response logging that provides detailed information about API interactions, distributed tracing that follows requests across service boundaries, and performance monitoring that identifies bottlenecks and optimization opportunities.

Diagnostic endpoints provide system information that supports troubleshooting including service health status, resource utilization information, and configuration validation results. Diagnostic information balances detail requirements with security considerations to provide useful troubleshooting information without exposing sensitive system details.

## Future API Evolution

The u-bmc API provides a foundation for enhanced functionality that may be implemented in future releases based on operational requirements and user feedback.

### Enhanced Protocol Support

Future enhancements may include GraphQL endpoints for flexible client-driven data queries, WebSocket support for enhanced real-time communication, and gRPC-Web support for browser-based clients requiring high-performance streaming capabilities.

### Advanced Integration Features

Advanced integration capabilities could include webhook support for external system integration, bulk operation APIs for efficient management of large systems, and federation capabilities for managing multiple BMC systems through unified interfaces.

### API Versioning and Compatibility

The API versioning strategy ensures backward compatibility while enabling evolution of API capabilities. Version management includes semantic versioning for API schemas and client libraries, deprecation policies that provide appropriate migration timelines, and compatibility testing that ensures new versions maintain compatibility with existing clients.

API documentation and client libraries are maintained consistently across versions to support smooth migration and upgrade activities while preserving investment in existing integration work.

## References

Additional information about API usage and integration is available in related documentation including system architecture details in [docs/architecture.md](architecture.md), service-specific documentation for backend service integration, security framework information in [docs/websrv.md](websrv.md) and security service documentation, and platform-specific deployment guidance in [docs/porting.md](porting.md).

The authoritative API schema is maintained in protobuf files under `schema/v1alpha1/` with comprehensive annotations for endpoint generation and client library support. Generated documentation and client libraries are available through standard distribution channels with version-specific compatibility information and migration guidance.