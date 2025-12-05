# Web Server Service

The web server service (websrv) provides the primary external interface for u-bmc through HTTPS endpoints that serve both ConnectRPC and REST APIs. It acts as the gateway between external clients and the internal u-bmc service ecosystem, handling authentication, request routing, and protocol translation. The web server integrates deeply with the IPC layer to coordinate with backend services while providing a secure, standards-compliant interface for management tools, web interfaces, and automation systems.

## Overview

The websrv service implements a modern, high-performance web server optimized for BMC workloads and constraints. Rather than relying on external web servers or reverse proxies, u-bmc includes its own purpose-built HTTP server that understands BMC-specific requirements like hardware resource constraints, security considerations, and integration with internal service architectures.

The server provides dual protocol support with ConnectRPC serving as the primary API protocol for modern clients and REST endpoints automatically transcoded from protobuf annotations for compatibility with existing tools and scripts. This dual approach enables both high-performance native clients and broad compatibility with standard HTTP tooling.

Security is built into every aspect of the web server design, with mandatory TLS encryption, comprehensive authentication and authorization integration, and careful attention to common web security vulnerabilities. The server operates within the broader u-bmc security framework while providing the necessary external access points for system management.

## Architecture

The websrv service implements a layered architecture that separates protocol handling, request processing, and backend integration. This separation enables clean implementation of different protocol requirements while maintaining consistent security and monitoring across all access paths.

### HTTP Server Foundation

The foundation layer implements a high-performance HTTP server optimized for BMC hardware constraints and workloads. The server supports HTTP/1.1, HTTP/2, and optionally HTTP/3 with QUIC for modern clients requiring enhanced performance. TLS 1.3 is mandatory for all connections, with configurable cipher suites and certificate management supporting both development and production scenarios.

Connection management includes appropriate timeouts, rate limiting, and resource management to prevent resource exhaustion on constrained BMC hardware. The server integrates with system monitoring to provide visibility into connection patterns and resource usage.

### ConnectRPC Integration

ConnectRPC provides the primary API interface using efficient protobuf serialization and modern RPC semantics. The integration supports both binary protobuf and JSON serialization, enabling clients to choose the most appropriate format for their requirements. Streaming RPCs enable efficient handling of long-running operations and real-time data feeds.

The ConnectRPC layer integrates directly with internal service NATS endpoints, translating HTTP requests into appropriate IPC calls and handling response aggregation and formatting. This integration maintains the benefits of the internal service architecture while providing a clean external interface.

### REST Transcoding

REST endpoints are automatically generated from protobuf service definitions and their HTTP annotations, providing compatibility with standard HTTP tooling without requiring separate API implementations. The transcoding layer handles URL path parameter extraction, query parameter processing, and request/response body transformation between JSON and protobuf formats.

REST endpoint documentation is generated automatically from protobuf comments and annotations, ensuring consistency between the primary ConnectRPC interface and the compatibility REST interface.

### Request Processing Pipeline

All requests flow through a comprehensive processing pipeline that handles authentication, authorization, request validation, rate limiting, and logging. The pipeline is configurable to support different security requirements and operational needs while maintaining consistent behavior across all API endpoints.

Request processing integrates with distributed tracing to provide comprehensive visibility into request handling across the web server and backend services, enabling detailed performance analysis and troubleshooting capabilities.

## Protocol Support

The websrv service provides comprehensive protocol support designed to meet diverse client requirements while maintaining security and performance standards appropriate for BMC environments.

### ConnectRPC Protocol

ConnectRPC serves as the primary protocol for u-bmc API access, providing efficient RPC semantics with modern features like streaming, cancellation, and comprehensive error handling. The protocol supports both unary and streaming RPC patterns, enabling efficient implementation of both simple request-response operations and complex real-time data feeds.

Client libraries are available for major programming languages, providing idiomatic interfaces that handle connection management, authentication, and error handling automatically. The protocol design emphasizes type safety and clear error semantics to reduce client implementation complexity.

Streaming capabilities enable efficient handling of operations like sensor monitoring, event feeds, and long-running management operations without requiring polling or complex state management in clients.

### REST Compatibility

REST endpoints provide broad compatibility with existing HTTP tooling and scripts that expect traditional REST semantics. All REST endpoints are generated automatically from the same protobuf definitions used for ConnectRPC, ensuring consistent behavior and eliminating the maintenance burden of dual API implementations.

URL patterns follow REST conventions with resource hierarchies and HTTP method semantics that map naturally to the underlying service operations. Query parameters and request bodies are handled according to standard REST practices while maintaining compatibility with the protobuf type system.

Error responses follow standard HTTP status code conventions while providing detailed error information compatible with both human operators and automated tooling.

### Content Type Support

The web server supports multiple content types to accommodate different client requirements and use cases. Binary protobuf provides maximum efficiency for clients that support it, while JSON offers broad compatibility and human readability for debugging and manual operations. Content negotiation automatically selects appropriate formats based on client preferences and endpoint capabilities.

Form data and multipart uploads are supported for operations that require file transfers or complex data structures that are more naturally expressed in form formats than JSON or protobuf.

## Security Implementation

Security permeates every aspect of the websrv design and implementation, reflecting the critical importance of BMC security and the exposure inherent in providing external network access to system management capabilities.

### Transport Security

All communications require TLS encryption using modern cipher suites and protocol versions. The server supports TLS 1.3 exclusively by default, with TLS 1.2 available as a compatibility option for environments that require it. Certificate management supports both self-signed certificates for development and testing scenarios and integration with certificate authorities for production deployments.

HSTS (HTTP Strict Transport Security) is enabled by default to prevent downgrade attacks, and the server includes comprehensive security headers to protect against common web vulnerabilities like XSS, clickjacking, and content type confusion.

### Authentication Integration

Authentication integrates with the broader u-bmc user management system to provide consistent identity verification across all access methods. The web server supports multiple authentication mechanisms including session-based authentication for web interfaces, token-based authentication for API clients, and certificate-based authentication for high-security scenarios.

Authentication state is managed consistently across requests and integrates with session management capabilities that provide appropriate security properties for different types of clients and use cases.

### Authorization Framework

Authorization decisions integrate with the u-bmc security manager service to provide fine-grained access control based on user roles, resource types, and operation contexts. The authorization system supports both simple role-based access control and more complex attribute-based policies that can consider factors like request source, time of day, and system state.

Authorization checks are performed consistently across all API endpoints and integrated into the request processing pipeline to ensure that security policies are enforced uniformly regardless of the protocol or endpoint used for access.

### Request Validation and Sanitization

Comprehensive input validation and sanitization protect against injection attacks and ensure that only valid requests are processed by backend services. Validation is based on protobuf schema definitions and includes both structural validation and semantic validation appropriate for each operation type.

Rate limiting and request size limits prevent resource exhaustion attacks while allowing legitimate usage patterns. The validation system integrates with monitoring and alerting to detect and respond to potential security threats.

## Configuration

The websrv service follows standard u-bmc configuration patterns with comprehensive options for server behavior, security settings, and integration parameters.

### Basic Configuration

```go
webServer := websrv.New(
    websrv.WithServiceName("websrv"),
    websrv.WithServiceDescription("U-BMC Web Server"),
    websrv.WithListenAddress("0.0.0.0:8443"),
    websrv.WithTLSConfig(&tls.Config{
        MinVersion: tls.VersionTLS13,
    }),
    websrv.WithCertificateFile("/etc/ubmc/tls/cert.pem"),
    websrv.WithPrivateKeyFile("/etc/ubmc/tls/key.pem"),
    websrv.WithMetrics(true),
    websrv.WithTracing(true),
)
```

### Security Configuration

```go
webServer := websrv.New(
    websrv.WithAuthenticationRequired(true),
    websrv.WithSessionTimeout(30 * time.Minute),
    websrv.WithMaxRequestSize(1024 * 1024), // 1MB
    websrv.WithRateLimit(100, time.Minute), // 100 requests per minute
    websrv.WithCSRFProtection(true),
    websrv.WithSecurityHeaders(map[string]string{
        "X-Frame-Options": "DENY",
        "X-Content-Type-Options": "nosniff",
        "Referrer-Policy": "strict-origin-when-cross-origin",
    }),
    websrv.WithCORSPolicy(&websrv.CORSConfig{
        AllowedOrigins: []string{"https://management.example.com"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowCredentials: true,
    }),
)
```

### Performance Configuration

```go
webServer := websrv.New(
    websrv.WithReadTimeout(30 * time.Second),
    websrv.WithWriteTimeout(30 * time.Second),
    websrv.WithIdleTimeout(120 * time.Second),
    websrv.WithMaxConnections(50),
    websrv.WithKeepAlive(true),
    websrv.WithCompression(true),
    websrv.WithHTTP2(true),
    websrv.WithHTTP3(false), // Optional QUIC support
)
```

### Backend Integration

```go
webServer := websrv.New(
    websrv.WithNATSConnection(natsConn),
    websrv.WithServiceTimeout(10 * time.Second),
    websrv.WithRetryPolicy(3, 1*time.Second),
    websrv.WithCircuitBreaker(true),
    websrv.WithLoadBalancing(websrv.RoundRobin),
)
```

## API Endpoints

The websrv service exposes a comprehensive set of API endpoints that provide access to all u-bmc functionality through both ConnectRPC and REST interfaces. All endpoints are generated from protobuf service definitions to ensure consistency and maintainability.

### System Management

System management endpoints provide high-level system information and control capabilities including system status and health information, power state management for hosts and chassis components, thermal monitoring and control interfaces, and configuration management for system-wide settings.

These endpoints integrate with multiple backend services to provide consolidated views of system state and enable coordinated operations that span multiple management domains.

### Hardware Management

Hardware management endpoints expose detailed control and monitoring capabilities for BMC-managed hardware including sensor data retrieval with real-time updates and historical information, power supply monitoring and control through PMBus interfaces, fan control with manual and automatic modes supporting PID-based thermal management, and LED control for status indication and identification purposes.

Hardware endpoints provide both immediate control capabilities and status information with appropriate error handling for hardware faults and unavailable components.

### User Management

User management endpoints handle authentication, authorization, and account management functions including user account creation, modification, and removal, authentication token management and session control, role assignment and permission management, and audit logging for security-relevant operations.

These endpoints integrate closely with the security framework to ensure consistent policy enforcement and comprehensive audit capabilities.

### Inventory and Asset Management

Inventory endpoints provide detailed information about system components and assets including component identification and manufacturing information, firmware version tracking and update status, asset tracking information for compliance and management purposes, and component health status and failure prediction information.

The inventory system integrates with hardware discovery mechanisms to provide accurate, real-time information about system configuration and component status.

## Integration Patterns

The websrv service implements several integration patterns that enable efficient communication with backend services while maintaining clean separation of concerns and appropriate error handling.

### Service Proxy Pattern

The service proxy pattern handles translation between HTTP requests and NATS IPC calls, managing connection pooling, timeout handling, and error translation. The proxy layer abstracts backend service complexity while providing consistent request handling and monitoring across all endpoints.

Proxy implementations include intelligent retry logic that handles transient failures without exposing clients to backend service instability, and circuit breaker patterns that provide graceful degradation when backend services are unavailable.

### Request Aggregation

Many API operations require coordination between multiple backend services to provide complete responses. The websrv service implements efficient request aggregation patterns that minimize IPC overhead while maintaining consistent error handling and timeout behavior.

Aggregation includes support for partial failures where some information may be unavailable without preventing successful completion of the overall operation, enabling robust operation even when some backend services are experiencing issues.

### Event Streaming

Real-time event streaming enables clients to receive immediate notifications of system changes without requiring polling. The streaming implementation supports both short-lived connections for specific operations and long-lived connections for continuous monitoring applications.

Event filtering and aggregation capabilities enable clients to receive only relevant events while minimizing bandwidth usage and processing overhead on both client and server sides.

## Performance and Scalability

The websrv service is designed to provide optimal performance within the constraints of BMC hardware while supporting the concurrency requirements of typical management workloads.

### Connection Management

Efficient connection management minimizes resource usage while supporting reasonable numbers of concurrent clients. Connection pooling and keep-alive mechanisms reduce connection setup overhead, while appropriate timeouts prevent resource leaks from abandoned connections.

The server includes monitoring and alerting capabilities that track connection usage patterns and provide visibility into resource utilization trends that can inform capacity planning decisions.

### Request Processing Optimization

Request processing optimization includes efficient routing and handler dispatch mechanisms, intelligent caching of frequently accessed data where appropriate, and streaming processing for large responses to minimize memory usage.

Processing pipelines are designed to minimize latency for common operations while maintaining security and monitoring requirements that are essential for BMC operations.

### Resource Management

Comprehensive resource management prevents individual requests or clients from consuming excessive server resources through request size limits, processing time limits, memory usage monitoring, and connection limits with appropriate queueing and rejection policies.

Resource management integrates with system monitoring to provide alerts when usage approaches configured limits, enabling proactive response to capacity issues.

## Monitoring and Observability

Extensive monitoring and observability features provide comprehensive visibility into web server operation and enable detailed analysis of client interactions and system performance.

### Request Tracing

Distributed tracing integration tracks requests from initial HTTP reception through backend service processing and response generation. Trace information includes detailed timing data for each processing stage, service interaction patterns and dependencies, and error conditions and retry attempts throughout request processing.

Tracing data integrates with broader system tracing to enable end-to-end analysis of client operations across the entire u-bmc system architecture.

### Metrics Collection

Comprehensive metrics track all aspects of web server operation including request rates, response times, and error rates by endpoint and client, connection statistics and resource utilization information, security events including authentication failures and suspicious request patterns, and backend service interaction statistics and performance data.

Metrics are exposed through standard interfaces that integrate with monitoring systems and support alerting based on operational thresholds and performance trends.

### Access Logging

Detailed access logging records all client interactions with configurable log levels and filtering capabilities. Log entries include comprehensive request information, authentication and authorization details, response status and timing information, and error details for failed operations.

Access logs integrate with audit and compliance systems to provide comprehensive records of system access and management operations for security and regulatory requirements.

## Development and Testing

The websrv service follows standard u-bmc development practices with comprehensive testing infrastructure and clear integration patterns for ongoing development activities.

### Package Documentation

Detailed package documentation is available at pkg.go.dev covering all aspects of web server configuration and integration including HTTP handler development patterns, authentication and authorization integration procedures, performance optimization guidance, and security best practices for BMC web services.

Documentation includes practical examples and templates that demonstrate common integration patterns and provide starting points for new feature development.

### Testing Infrastructure

Comprehensive test suites cover all aspects of web server functionality including HTTP protocol handling and security features, authentication and authorization mechanisms, backend service integration and error handling, and performance characteristics under various load conditions.

Test infrastructure includes mock backend services and failure injection capabilities that enable thorough testing of error handling and recovery mechanisms under realistic failure scenarios.

### API Documentation Generation

Automatic API documentation generation from protobuf definitions ensures that endpoint documentation remains current and accurate. Generated documentation includes endpoint descriptions, parameter specifications, example requests and responses, and error condition documentation.

Documentation generation integrates with the development workflow to ensure that API changes are properly documented and that breaking changes are clearly identified and communicated.

## Security Considerations

The websrv service implements comprehensive security measures that address the unique requirements and threat models associated with BMC systems and network-accessible management interfaces.

### Threat Model

The security implementation addresses multiple threat categories including network-based attacks from external sources, credential-based attacks attempting unauthorized access, injection attacks targeting web application vulnerabilities, and resource exhaustion attacks designed to disrupt system availability.

Security measures are designed to provide defense in depth with multiple layers of protection that maintain security even when individual measures are bypassed or compromised.

### Compliance and Auditing

Security features support compliance with relevant industry standards and regulatory requirements including comprehensive audit logging of security-relevant events, access control enforcement with detailed logging and monitoring, and security configuration options that support various compliance frameworks.

Audit capabilities provide the detailed logging and monitoring information required for security assessments and compliance verification activities.

### Incident Response

Security monitoring and alerting capabilities support incident response activities including real-time detection of suspicious activity patterns, automated response capabilities for common attack types, and comprehensive logging that supports forensic analysis and incident investigation.

Integration with broader system monitoring enables coordinated response to security incidents that may affect multiple system components or require coordinated defensive measures.

## Future Enhancements

The websrv service provides a foundation for advanced web interface capabilities that may be implemented in future u-bmc releases based on operational requirements and user feedback.

### Advanced Protocol Support

Future enhancements may include WebSocket support for real-time bidirectional communication, GraphQL endpoints for flexible client-driven data queries, and enhanced streaming capabilities for high-volume data feeds like detailed sensor monitoring.

### User Interface Integration

Integration with web-based management interfaces could include embedded web UI serving capabilities, single sign-on integration with enterprise identity systems, and advanced session management features for complex multi-user scenarios.

### Performance Optimization

Continued performance optimization may focus on advanced caching strategies for frequently accessed data, CDN integration for static content delivery in distributed scenarios, and enhanced HTTP/3 and QUIC support for improved performance over high-latency or unreliable network connections.

The modular design of the websrv service and its standardized integration patterns provide a solid foundation for these enhancements while maintaining compatibility with existing clients and backend service integrations.

## References

Additional information about the websrv service and its integration with the broader u-bmc system is available in related documentation including system architecture details in [docs/architecture.md](architecture.md), API usage examples and patterns in [docs/api.md](api.md), and security framework information in individual service documentation for `usermgr` and `securitymgr`.

Package-level documentation and detailed API references are maintained on pkg.go.dev for comprehensive development and integration guidance. The authoritative API schema is defined in the protobuf files under `schema/v1alpha1/` with comprehensive annotations for both ConnectRPC and REST endpoint generation.