// SPDX-License-Identifier: BSD-3-Clause

// Package i2c provides a comprehensive Go interface for communicating with I2C, I3C, SMBus, and PMBus devices
// on Linux systems through the standard /dev/i2c-* character device interface.
//
// # Overview
//
// This package offers a unified, production-grade interface for multiple inter-integrated circuit protocols:
//
//   - I2C (Inter-Integrated Circuit): The foundational two-wire serial communication protocol
//   - I3C (Improved Inter-Integrated Circuit): The newer protocol with enhanced features and performance
//   - SMBus (System Management Bus): A subset of I2C with additional specifications for system management
//   - PMBus (Power Management Bus): An extension of SMBus specifically designed for power management applications
//
// The package is designed to work exclusively with Linux's I2C subsystem and does not implement
// bare-metal bus communication. It leverages the kernel's I2C drivers and provides a high-level,
// type-safe interface while maintaining access to low-level operations when needed.
//
// # Supported Protocols
//
// ## I2C (Inter-Integrated Circuit)
//
// Basic two-wire serial communication supporting:
//   - 7-bit and 10-bit addressing
//   - Standard (100 kHz), Fast (400 kHz), and Fast+ (1 MHz) modes
//   - Raw read/write operations
//   - Combined transactions
//
// ## I3C (Improved Inter-Integrated Circuit)
//
// Enhanced protocol with backward I2C compatibility:
//   - Dynamic address assignment
//   - In-band interrupts
//   - Higher data rates
//   - Hot-join capability
//   - Common Command Codes (CCCs)
//
// ## SMBus (System Management Bus)
//
// Standardized subset of I2C with defined transaction types:
//   - Quick Command
//   - Send/Receive Byte
//   - Read/Write Byte Data
//   - Read/Write Word Data
//   - Read/Write Block Data
//   - Process Call operations
//   - Packet Error Checking (PEC)
//
// ## PMBus (Power Management Bus)
//
// SMBus-based protocol for power management:
//   - Standard PMBus commands
//   - Linear and direct data formats
//   - Coefficient-based scaling
//   - Manufacturer-specific extensions
//   - Telemetry and monitoring
//
// # Architecture
//
// The package follows a layered architecture:
//
//	┌─────────────────────────────────────┐
//	│          Application Layer          │
//	├─────────────────────────────────────┤
//	│  PMBus  │  SMBus  │  I3C  │   I2C   │
//	├─────────────────────────────────────┤
//	│          Connection Layer           │
//	├─────────────────────────────────────┤
//	│         Linux I2C Subsystem         │
//	└─────────────────────────────────────┘
//
// # Basic Usage
//
// ## Simple I2C Communication
//
//	cfg := i2c.NewConfig(
//		i2c.WithBus(1),
//		i2c.WithAddress(0x48),
//		i2c.WithProtocol(i2c.ProtocolI2C),
//	)
//
//	conn, err := i2c.Open(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer conn.Close()
//
//	// Read 2 bytes
//	data, err := conn.Read(2)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// ## SMBus Operations
//
//	cfg := i2c.NewConfig(
//		i2c.WithBus(1),
//		i2c.WithAddress(0x48),
//		i2c.WithProtocol(i2c.ProtocolSMBus),
//		i2c.WithPEC(true), // Enable Packet Error Checking
//	)
//
//	conn, err := i2c.Open(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer conn.Close()
//
//	// Read byte from register 0x10
//	value, err := conn.ReadByteData(0x10)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Write word to register 0x20
//	err = conn.WriteWordData(0x20, 0x1234)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// ## PMBus Operations
//
//	cfg := i2c.NewConfig(
//		i2c.WithBus(1),
//		i2c.WithAddress(0x40),
//		i2c.WithProtocol(i2c.ProtocolPMBus),
//		i2c.WithPMBusFormat(i2c.PMBusFormatLinear),
//	)
//
//	conn, err := i2c.Open(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer conn.Close()
//
//	// Read voltage (LINEAR11 format)
//	voltage, err := conn.ReadVoltage(i2c.PMBusVoutMode0)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Set voltage limit
//	err = conn.WriteVoltage(i2c.PMBusVoutOVFaultLimit, 12.5)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Configuration
//
// The package uses a functional options pattern for configuration:
//
//	cfg := i2c.NewConfig(
//		i2c.WithBus(1),                          // I2C bus number
//		i2c.WithAddress(0x48),                   // Device address
//		i2c.WithProtocol(i2c.ProtocolSMBus),     // Protocol type
//		i2c.WithRetries(3),                      // Retry count
//		i2c.WithTimeout(1*time.Second),          // Operation timeout
//		i2c.WithPEC(true),                       // Enable PEC for SMBus
//		i2c.WithForceAddress(false),             // Use I2C_SLAVE vs I2C_SLAVE_FORCE
//	)
//
// # Error Handling
//
// The package provides comprehensive error types for different failure scenarios:
//
//   - ErrBusNotFound: I2C bus device not available
//   - ErrDeviceNotResponding: Device fails to acknowledge
//   - ErrProtocolViolation: Protocol-specific error
//   - ErrTimeout: Operation timeout
//   - ErrChecksumMismatch: PEC validation failure
//   - ErrInvalidData: Data format or range error
//
// Errors are wrapped with context to provide clear diagnostic information:
//
//	if err != nil {
//		if errors.Is(err, i2c.ErrDeviceNotResponding) {
//			// Handle device communication failure
//			log.Printf("Device at address 0x%02x not responding", addr)
//		}
//	}
//
// # Thread Safety
//
// Connection instances are not thread-safe. Applications requiring concurrent access
// should implement appropriate synchronization:
//
//	type SafeConn struct {
//		conn *i2c.Conn
//		mu   sync.Mutex
//	}
//
//	func (sc *SafeConn) ReadByteData(reg uint8) (uint8, error) {
//		sc.mu.Lock()
//		defer sc.mu.Unlock()
//		return sc.conn.ReadByteData(reg)
//	}
//
// # Performance Considerations
//
// - Connection pooling: Reuse connections when possible to avoid repeated device file operations
// - Batch operations: Use block read/write operations for multi-byte transfers
// - Protocol selection: Choose the most appropriate protocol for your use case
// - Bus speed: Configure the I2C bus speed appropriately for your devices
//
// # Linux Integration
//
// The package integrates with Linux I2C subsystem features:
//
//   - Automatic bus detection via /sys/bus/i2c/devices/
//   - Support for I2C device tree overlays
//   - Integration with Linux I2C multiplexers
//   - Compatibility with I2C userspace drivers
//   - Support for I2C bus recovery mechanisms
//
// # Compatibility
//
// Minimum requirements:
//   - Linux kernel 2.6.5+ (for basic I2C support)
//   - Linux kernel 4.19+ (for full I3C support)
//
// # Examples
//
// See the examples directory for complete working examples including:
//   - Temperature sensor reading (I2C)
//   - EEPROM programming (I2C)
//   - Battery monitoring (SMBus)
//   - Power supply control (PMBus)
//   - Multi-device communication (I3C)
//
// # Contributing
//
// When contributing to this package:
//   - Follow the existing error handling patterns
//   - Add appropriate validation for new configuration options
//   - Include comprehensive documentation for new features
//   - Ensure compatibility with existing Linux I2C infrastructure
//   - Add protocol-specific tests when implementing new protocol features
package i2c
