// SPDX-License-Identifier: BSD-3-Clause

// Package websrv provides a high-performance web server implementation for the u-bmc
// system that supports modern HTTP protocols including HTTP/2 and HTTP/3 with TLS
// encryption. It serves as the primary web interface for BMC operations, providing
// both REST APIs via Connect RPC and optionally serving a web UI for browser-based
// management.
//
// The websrv service acts as the main entry point for external clients to interact
// with the BMC system, translating HTTP requests into internal NATS messages and
// routing them to appropriate backend services.
//
// # Core Features
//
//   - Multi-protocol support: HTTP/3 (QUIC), HTTP/2, and HTTP/1.1
//   - Automatic TLS certificate management (self-signed or Let's Encrypt)
//   - Connect RPC API serving with protocol transcoding
//   - Optional static web UI file serving
//   - Automatic HTTP to HTTPS redirection
//   - OpenTelemetry integration for observability
//   - Request validation and CORS support
//   - Health checks and gRPC reflection
//
// # Architecture
//
// The websrv service implements a multi-server architecture:
//
//   - HTTP/3 Server: Primary server using QUIC protocol over UDP
//   - HTTP/2 Server: Fallback server using HTTP/2 over TCP with TLS
//   - HTTP Redirect Server: Handles HTTP-to-HTTPS redirection on port 80
//   - Optional ACME HTTP-01 challenge handler for Let's Encrypt
//
// All servers run concurrently and provide the same API endpoints, with the
// client automatically selecting the best available protocol.
//
// # Protocol Support
//
// ## HTTP/3 (QUIC)
//
// The primary protocol offering the best performance with features like:
//   - Reduced connection establishment time
//   - Built-in multiplexing without head-of-line blocking
//   - Connection migration support
//   - Improved congestion control
//
// ## HTTP/2
//
// Fallback protocol for clients that don't support HTTP/3:
//   - Multiplexed streams over a single TCP connection
//   - Header compression (HPACK)
//   - Server push capabilities (if needed)
//   - Binary protocol efficiency
//
// ## HTTP/1.1
//
// Legacy support for older clients through HTTP/2 negotiation.
//
// # TLS Configuration
//
// The service supports flexible TLS certificate management:
//
// ## Self-Signed Certificates
//
// Automatically generated certificates suitable for development and internal use:
//
//	websrv := websrv.New(
//		websrv.WithCertificateType(cert.CertificateTypeSelfSigned),
//		websrv.WithHostname("bmc.local"),
//		websrv.WithAlternativeNames("192.168.1.100", "bmc.example.com"),
//	)
//
// ## Let's Encrypt Certificates
//
// Automatically obtained and renewed certificates for production use:
//
//	websrv := websrv.New(
//		websrv.WithCertificateType(cert.CertificateTypeLetsEncrypt),
//		websrv.WithHostname("bmc.example.com"),
//		websrv.WithCertEmail("admin@example.com"),
//	)
//
// # API Architecture
//
// The service implements the complete BMC API through Connect RPC handlers:
//
//   - System Information and Health
//   - Asset Management
//   - Chassis Control and Monitoring
//   - Host Power Management
//   - Management Controller Operations
//   - Sensor Data Access
//   - Thermal Zone Management
//   - User Account Management
//   - Authentication and Authorization
//
// Each API endpoint follows a consistent pattern:
//  1. Request validation and authentication
//  2. Protocol buffer marshaling
//  3. NATS message routing to backend services
//  4. Response unmarshaling and validation
//  5. OpenTelemetry tracing and logging
//
// # Service Integration
//
// The websrv service integrates with other BMC services via NATS messaging:
//
//	┌─────────────┐    HTTP/Connect RPC    ┌─────────────┐
//	│   Client    │ ────────────────────► │   websrv    │
//	└─────────────┘                       └─────────────┘
//	                                              │
//	                                              │ NATS
//	                                              ▼
//	┌─────────────┐    ┌─────────────┐    ┌─────────────┐
//	│  statemgr   │    │  powermgr   │    │ sensormon   │
//	└─────────────┘    └─────────────┘    └─────────────┘
//
// # Configuration Options
//
// The service provides extensive configuration options:
//
//	ws := websrv.New(
//		// Basic configuration
//		websrv.WithName("production-websrv"),
//		websrv.WithAddr(":8443"),
//
//		// TLS configuration
//		websrv.WithCertificateType(cert.CertificateTypeLetsEncrypt),
//		websrv.WithHostname("bmc.example.com"),
//		websrv.WithCertEmail("admin@example.com"),
//
//		// Web UI configuration
//		websrv.WithWebUI(true),
//		websrv.WithWebUIPath("/usr/share/bmc-webui"),
//
//		// Performance tuning
//		websrv.WithReadTimeout(10*time.Second),
//		websrv.WithWriteTimeout(10*time.Second),
//		websrv.WithIdleTimeout(120*time.Second),
//
//		// QUIC optimization
//		websrv.WithRmemMax("16777216"), // 16MB
//		websrv.WithWmemMax("16777216"), // 16MB
//	)
//
// # Web UI Integration
//
// When enabled, the service can serve static web UI files alongside the API:
//
//	// Enable web UI serving
//	ws := websrv.New(
//		websrv.WithWebUI(true),
//		websrv.WithWebUIPath("/var/www/bmc-ui"),
//	)
//
// The service intelligently routes requests:
//   - API requests (Content-Type: application/*) → Connect RPC handlers
//   - Browser requests (HTML, CSS, JS) → Static file server
//
// # Security Features
//
// ## Transport Security
//   - TLS 1.3 minimum version enforcement
//   - Modern cipher suite selection
//   - Perfect Forward Secrecy (PFS)
//   - HSTS headers for HTTPS enforcement
//
// ## Request Security
//   - CORS policy enforcement
//   - Request size limits
//   - Rate limiting (via middleware)
//   - Input validation and sanitization
//
// ## Certificate Security
//   - Automatic certificate rotation
//   - OCSP stapling support
//   - Certificate transparency logging
//
// # Observability
//
// The service provides comprehensive observability:
//
// ## Logging
//   - Structured logging with context
//   - Request/response correlation IDs
//   - Performance metrics logging
//   - Error tracking and categorization
//
// ## Metrics
//   - Request duration histograms
//   - Error rate counters
//   - Connection pool metrics
//   - Protocol-specific metrics
//
// ## Tracing
//   - Distributed tracing across service boundaries
//   - Request flow visualization
//   - Performance bottleneck identification
//   - Error propagation tracking
//
// # Performance Optimization
//
// ## QUIC Optimization
//
// The service automatically configures kernel parameters for optimal QUIC performance:
//
//	// These are set automatically
//	net.core.rmem_max = 7500000  // Receive buffer size
//	net.core.wmem_max = 7500000  // Send buffer size
//
// ## Connection Management
//   - Connection pooling and reuse
//   - Keep-alive optimization
//   - Graceful connection draining
//   - Resource cleanup on shutdown
//
// ## Memory Management
//   - Efficient buffer management
//   - Protocol buffer pooling
//   - Memory-mapped file serving
//   - Garbage collection optimization
//
// # Error Handling
//
// The service implements comprehensive error handling:
//
//   - Graceful degradation when services are unavailable
//   - Automatic retry logic with backoff
//   - Circuit breaker patterns for upstream services
//   - Detailed error responses with appropriate HTTP status codes
//
// # Usage Examples
//
// ## Basic Usage
//
//	package main
//
//	import (
//		"context"
//		"github.com/u-bmc/u-bmc/service/websrv"
//	)
//
//	func main() {
//		ws := websrv.New()
//		ctx := context.Background()
//
//		// Assumes IPC connection is provided by operator
//		err := ws.Run(ctx, ipcConn)
//		if err != nil {
//			log.Fatal("Web server failed:", err)
//		}
//	}
//
// ## Production Configuration
//
//	ws := websrv.New(
//		websrv.WithName("bmc-websrv"),
//		websrv.WithAddr(":443"),
//		websrv.WithCertificateType(cert.CertificateTypeLetsEncrypt),
//		websrv.WithHostname("bmc.example.com"),
//		websrv.WithCertEmail("admin@example.com"),
//		websrv.WithWebUI(true),
//		websrv.WithWebUIPath("/var/www/bmc-ui"),
//		websrv.WithReadTimeout(15*time.Second),
//		websrv.WithWriteTimeout(15*time.Second),
//	)
//
// ## Development Configuration
//
//	ws := websrv.New(
//		websrv.WithAddr(":8443"),
//		websrv.WithCertificateType(cert.CertificateTypeSelfSigned),
//		websrv.WithHostname("localhost"),
//		websrv.WithWebUI(true),
//		websrv.WithWebUIPath("./webui/dist"),
//	)
//
// # Health Checks
//
// The service provides built-in health check endpoints:
//
//   - `/grpc.health.v1.Health/Check` - gRPC health check
//   - Standard HTTP health endpoints via Connect protocol
//
// # Best Practices
//
// ## TLS Configuration
//   - Always use TLS 1.3 in production
//   - Regularly rotate certificates
//   - Monitor certificate expiration
//   - Use Let's Encrypt for public-facing deployments
//
// ## Performance
//   - Enable HTTP/3 for modern clients
//   - Configure appropriate buffer sizes for your network
//   - Monitor connection metrics
//   - Use CDN for static assets in large deployments
//
// ## Security
//   - Implement rate limiting
//   - Use strong authentication mechanisms
//   - Regularly update dependencies
//   - Monitor for security vulnerabilities
//
// ## Monitoring
//   - Set up alerts for high error rates
//   - Monitor certificate expiration
//   - Track performance metrics
//   - Implement log aggregation
//
// The websrv service is designed to be the robust, high-performance gateway
// for all BMC web-based interactions, providing both the flexibility needed
// for development and the reliability required for production deployments.
package websrv
