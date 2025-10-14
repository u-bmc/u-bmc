// SPDX-License-Identifier: BSD-3-Clause

// Package kvmsrv provides KVM (Keyboard, Video, Mouse) functionality for BMC environments.
//
// This service integrates video capture, USB gadget emulation, and VNC serving to provide
// remote console access to managed systems. It captures video frames from V4L2 devices,
// emulates keyboard and mouse input via USB gadgets, and serves the console via VNC protocol.
//
// # Service Overview
//
// The KVM service operates independently with minimal IPC dependencies:
//   - Captures video from /dev/video0 or configured V4L2 device
//   - Provides USB HID keyboard and mouse emulation
//   - Serves VNC on standard port 5900 (configurable)
//   - Optional HTTP MJPEG streaming support
//   - Optional mass storage emulation for virtual media
//
// # Configuration
//
// The service supports comprehensive configuration:
//
//	service := kvmsrv.New(
//		kvmsrv.WithServiceName("kvm-server"),
//		kvmsrv.WithVideoDevice("/dev/video0"),
//		kvmsrv.WithVideoResolution(1920, 1080),
//		kvmsrv.WithVNCPort(5900),
//		kvmsrv.WithHTTPPort(8080),
//		kvmsrv.WithUSB(true),
//		kvmsrv.WithHTTP(true),
//	)
//
// # USB Gadget Integration
//
// The service automatically manages USB gadget functionality:
//   - Creates and configures USB composite gadget
//   - Provides HID keyboard and mouse devices
//   - Supports mass storage for virtual media mounting
//   - Handles gadget lifecycle (creation, binding, cleanup)
//
// # Video Processing
//
// Efficient video frame processing pipeline:
//   - V4L2 video capture with configurable formats
//   - YUYV to JPEG conversion for HTTP streaming
//   - YUYV to RGBA conversion for VNC clients
//   - Frame buffering and rate control
//   - Automatic format negotiation
//
// # VNC Protocol Support
//
// Full VNC server implementation:
//   - TCP connections on standard VNC port
//   - WebSocket support for web-based clients
//   - Multiple encoding support (Raw, RRE, Hextile)
//   - Configurable pixel formats
//   - Client authentication (optional)
//   - Multi-client support with reasonable limits
//
// # HTTP Streaming
//
// Optional HTTP MJPEG streaming:
//   - Multipart HTTP response streaming
//   - Configurable JPEG quality
//   - Concurrent client support
//   - Bandwidth-efficient frame delivery
//
// # Resource Management
//
// The service properly manages system resources:
//   - Automatic cleanup on shutdown
//   - Context-based cancellation
//   - Graceful client disconnection handling
//   - USB gadget lifecycle management
//   - Video device exclusive access
//
// # Error Recovery
//
// Built-in error recovery mechanisms:
//   - Automatic reconnection to video devices
//   - USB gadget recreation on failure
//   - Client connection error handling
//   - Graceful degradation when components fail
//
// # Integration
//
// While primarily standalone, the service can integrate with other BMC services:
//   - Power management for coordinated system control
//   - State management for system status updates
//   - Security management for access control
//   - Inventory management for device enumeration
//
// # Performance Considerations
//
// The service is optimized for BMC environments:
//   - Low memory footprint
//   - Efficient frame processing
//   - Minimal CPU overhead
//   - Configurable quality vs. performance trade-offs
//   - Resource pooling and reuse
package kvmsrv
