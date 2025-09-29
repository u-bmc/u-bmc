// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"time"
)

// ExampleConfigurations provides example configurations for different use cases.

// ExampleBasicConfig returns a basic KVM configuration for simple setups.
func ExampleBasicConfig() *Config {
	return &Config{
		VideoDevice:       "/dev/video0",
		VideoWidth:        640,
		VideoHeight:       480,
		VideoFPS:          30,
		VNCPort:           5900,
		VNCWebSocketPort:  5901,
		HTTPPort:          8080,
		EnableVNC:         true,
		EnableHTTP:        true,
		EnableUSB:         true,
		EnableMassStorage: true,
		VNCPassword:       "",
		VNCMaxClients:     3,
		HTTPMaxClients:    5,
		JPEGQuality:       75,
		USBGadgetName:     "basic-kvm",
		USBVendorID:       "0x1d6b",
		USBProductID:      "0x0104",
		USBManufacturer:   "U-BMC",
		USBProduct:        "Basic KVM Device",
		USBSerialNumber:   "",
		ClientTimeout:     30 * time.Minute,
		FrameTimeout:      5 * time.Second,
		BufferCount:       4,
	}
}

// ExampleHighResConfig returns a high-resolution KVM configuration.
func ExampleHighResConfig() *Config {
	return &Config{
		VideoDevice:       "/dev/video0",
		VideoWidth:        1920,
		VideoHeight:       1080,
		VideoFPS:          30,
		VNCPort:           5900,
		VNCWebSocketPort:  5901,
		HTTPPort:          8080,
		EnableVNC:         true,
		EnableHTTP:        true,
		EnableUSB:         true,
		EnableMassStorage: true,
		VNCPassword:       "",
		VNCMaxClients:     2, // Fewer clients for high-res
		HTTPMaxClients:    3,
		JPEGQuality:       85, // Higher quality for better image
		USBGadgetName:     "hires-kvm",
		USBVendorID:       "0x1d6b",
		USBProductID:      "0x0104",
		USBManufacturer:   "U-BMC",
		USBProduct:        "HD KVM Device",
		USBSerialNumber:   "",
		ClientTimeout:     15 * time.Minute, // Shorter timeout for high-res
		FrameTimeout:      3 * time.Second,
		BufferCount:       6, // More buffers for smoother streaming
	}
}

// ExampleSecureConfig returns a KVM configuration with security features.
func ExampleSecureConfig() *Config {
	return &Config{
		VideoDevice:       "/dev/video0",
		VideoWidth:        1024,
		VideoHeight:       768,
		VideoFPS:          25,
		VNCPort:           5900,
		VNCWebSocketPort:  5901,
		HTTPPort:          8080,
		EnableVNC:         true,
		EnableHTTP:        false, // Disable HTTP for security
		EnableUSB:         true,
		EnableMassStorage: false,       // Disable mass storage for security
		VNCPassword:       "secure123", // Password protection
		VNCMaxClients:     1,           // Single client only
		HTTPMaxClients:    0,           // HTTP disabled
		JPEGQuality:       70,
		USBGadgetName:     "secure-kvm",
		USBVendorID:       "0x1d6b",
		USBProductID:      "0x0104",
		USBManufacturer:   "U-BMC",
		USBProduct:        "Secure KVM Device",
		USBSerialNumber:   "SEC001",
		ClientTimeout:     10 * time.Minute, // Shorter timeout
		FrameTimeout:      5 * time.Second,
		BufferCount:       4,
	}
}

// ExampleLowBandwidthConfig returns a configuration optimized for low bandwidth.
func ExampleLowBandwidthConfig() *Config {
	return &Config{
		VideoDevice:       "/dev/video0",
		VideoWidth:        800,
		VideoHeight:       600,
		VideoFPS:          15, // Lower frame rate
		VNCPort:           5900,
		VNCWebSocketPort:  5901,
		HTTPPort:          8080,
		EnableVNC:         true,
		EnableHTTP:        true,
		EnableUSB:         true,
		EnableMassStorage: true,
		VNCPassword:       "",
		VNCMaxClients:     5,
		HTTPMaxClients:    8,
		JPEGQuality:       60, // Lower quality for smaller frames
		USBGadgetName:     "lowbw-kvm",
		USBVendorID:       "0x1d6b",
		USBProductID:      "0x0104",
		USBManufacturer:   "U-BMC",
		USBProduct:        "Efficient KVM Device",
		USBSerialNumber:   "",
		ClientTimeout:     45 * time.Minute, // Longer timeout for slow connections
		FrameTimeout:      10 * time.Second, // More tolerance for slow capture
		BufferCount:       2,                // Fewer buffers to save memory
	}
}

// ExampleMultimediaConfig returns a configuration for multimedia workloads.
func ExampleMultimediaConfig() *Config {
	return &Config{
		VideoDevice:       "/dev/video0",
		VideoWidth:        1920,
		VideoHeight:       1080,
		VideoFPS:          60, // High frame rate
		VNCPort:           5900,
		VNCWebSocketPort:  5901,
		HTTPPort:          8080,
		EnableVNC:         true,
		EnableHTTP:        true,
		EnableUSB:         true,
		EnableMassStorage: true,
		VNCPassword:       "",
		VNCMaxClients:     2,
		HTTPMaxClients:    3,
		JPEGQuality:       95, // Very high quality
		USBGadgetName:     "media-kvm",
		USBVendorID:       "0x1d6b",
		USBProductID:      "0x0104",
		USBManufacturer:   "U-BMC",
		USBProduct:        "Media KVM Device",
		USBSerialNumber:   "",
		ClientTimeout:     20 * time.Minute,
		FrameTimeout:      2 * time.Second, // Fast capture
		BufferCount:       8,               // Many buffers for smooth playback
	}
}

// ExampleTestingConfig returns a configuration suitable for testing and development.
func ExampleTestingConfig() *Config {
	return &Config{
		VideoDevice:       "/dev/video0",
		VideoWidth:        320,
		VideoHeight:       240,
		VideoFPS:          10, // Very low frame rate for testing
		VNCPort:           5900,
		VNCWebSocketPort:  5901,
		HTTPPort:          8080,
		EnableVNC:         true,
		EnableHTTP:        true,
		EnableUSB:         true,
		EnableMassStorage: true,
		VNCPassword:       "test",
		VNCMaxClients:     10, // Allow many test clients
		HTTPMaxClients:    15,
		JPEGQuality:       50, // Low quality for fast processing
		USBGadgetName:     "test-kvm",
		USBVendorID:       "0x1d6b",
		USBProductID:      "0x0104",
		USBManufacturer:   "U-BMC",
		USBProduct:        "Test KVM Device",
		USBSerialNumber:   "TEST001",
		ClientTimeout:     5 * time.Minute,  // Short timeout for testing
		FrameTimeout:      15 * time.Second, // Very tolerant for test environments
		BufferCount:       2,
	}
}

// ExampleProductionConfig returns a robust configuration for production environments.
func ExampleProductionConfig() *Config {
	return &Config{
		VideoDevice:       "/dev/video0",
		VideoWidth:        1600,
		VideoHeight:       1200,
		VideoFPS:          30,
		VNCPort:           5900,
		VNCWebSocketPort:  5901,
		HTTPPort:          8080,
		EnableVNC:         true,
		EnableHTTP:        true,
		EnableUSB:         true,
		EnableMassStorage: true,
		VNCPassword:       "", // Set in production deployment
		VNCMaxClients:     3,
		HTTPMaxClients:    5,
		JPEGQuality:       80, // Good balance of quality and bandwidth
		USBGadgetName:     "prod-kvm",
		USBVendorID:       "0x1d6b",
		USBProductID:      "0x0104",
		USBManufacturer:   "U-BMC",
		USBProduct:        "Production KVM Device",
		USBSerialNumber:   "", // Set per device in production
		ClientTimeout:     30 * time.Minute,
		FrameTimeout:      5 * time.Second,
		BufferCount:       4,
	}
}

// ExampleConfigByUseCase returns appropriate configuration based on use case.
func ExampleConfigByUseCase(useCase string) *Config {
	switch useCase {
	case "basic", "default":
		return ExampleBasicConfig()
	case "hires", "high-resolution", "hd":
		return ExampleHighResConfig()
	case "secure", "security":
		return ExampleSecureConfig()
	case "lowbw", "low-bandwidth", "efficient":
		return ExampleLowBandwidthConfig()
	case "media", "multimedia", "video":
		return ExampleMultimediaConfig()
	case "test", "testing", "development", "dev":
		return ExampleTestingConfig()
	case "prod", "production":
		return ExampleProductionConfig()
	default:
		return DefaultConfig()
	}
}
