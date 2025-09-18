// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"fmt"
	"time"

	"github.com/u-bmc/u-bmc/pkg/kvm"
	"github.com/u-bmc/u-bmc/pkg/usb"
)

// Config represents the configuration for the KVM service.
type Config struct {
	// VideoDevice is the path to the video capture device.
	VideoDevice string

	// VideoWidth is the desired video capture width.
	VideoWidth uint32

	// VideoHeight is the desired video capture height.
	VideoHeight uint32

	// VideoFPS is the desired video capture frame rate.
	VideoFPS uint32

	// VNCPort is the port for the VNC server.
	VNCPort int

	// VNCWebSocketPort is the port for VNC WebSocket connections.
	VNCWebSocketPort int

	// HTTPPort is the port for HTTP MJPEG streaming.
	HTTPPort int

	// EnableVNC enables the VNC server.
	EnableVNC bool

	// EnableHTTP enables HTTP MJPEG streaming.
	EnableHTTP bool

	// EnableUSB enables USB gadget functionality.
	EnableUSB bool

	// EnableMassStorage enables USB mass storage emulation.
	EnableMassStorage bool

	// VNCPassword is the VNC authentication password (empty for no auth).
	VNCPassword string

	// VNCMaxClients is the maximum number of concurrent VNC clients.
	VNCMaxClients int

	// HTTPMaxClients is the maximum number of concurrent HTTP clients.
	HTTPMaxClients int

	// JPEGQuality is the JPEG compression quality (1-100).
	JPEGQuality int

	// USBGadgetName is the name for the USB gadget.
	USBGadgetName string

	// USBVendorID is the USB vendor ID.
	USBVendorID string

	// USBProductID is the USB product ID.
	USBProductID string

	// USBManufacturer is the USB manufacturer string.
	USBManufacturer string

	// USBProduct is the USB product string.
	USBProduct string

	// USBSerialNumber is the USB serial number.
	USBSerialNumber string

	// ClientTimeout is the timeout for idle clients.
	ClientTimeout time.Duration

	// FrameTimeout is the timeout for video frame capture.
	FrameTimeout time.Duration

	// BufferCount is the number of video capture buffers.
	BufferCount uint32
}

type config struct {
	name string
	cfg  *Config
}

type Option interface {
	apply(*config)
}

type nameOption struct {
	name string
}

func (o *nameOption) apply(c *config) {
	c.name = o.name
}

func WithName(name string) Option {
	return &nameOption{
		name: name,
	}
}

type configOption struct {
	cfg *Config
}

func (o *configOption) apply(c *config) {
	c.cfg = o.cfg
}

func WithConfig(cfg *Config) Option {
	return &configOption{
		cfg: cfg,
	}
}

// DefaultConfig returns a default configuration for the KVM service.
func DefaultConfig() *Config {
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
		VNCMaxClients:     5,
		HTTPMaxClients:    10,
		JPEGQuality:       75,
		USBGadgetName:     "kvm-gadget",
		USBVendorID:       "0x1d6b", // Linux Foundation
		USBProductID:      "0x0104", // Multifunction Composite Gadget
		USBManufacturer:   "U-BMC",
		USBProduct:        "Virtual KVM Device",
		USBSerialNumber:   "",
		ClientTimeout:     30 * time.Minute,
		FrameTimeout:      5 * time.Second,
		BufferCount:       4,
	}
}

// Validate validates the configuration and sets defaults where appropriate.
func (c *Config) Validate() error {
	if c.VideoDevice == "" {
		return ErrInvalidConfiguration
	}

	if c.VideoWidth == 0 || c.VideoHeight == 0 {
		return ErrInvalidConfiguration
	}

	if c.VNCPort <= 0 || c.VNCPort > 65535 {
		return ErrInvalidConfiguration
	}

	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return ErrInvalidConfiguration
	}

	if c.VNCWebSocketPort <= 0 || c.VNCWebSocketPort > 65535 {
		return ErrInvalidConfiguration
	}

	if c.VideoFPS == 0 {
		c.VideoFPS = 30
	}

	if c.VNCMaxClients <= 0 {
		c.VNCMaxClients = 5
	}

	if c.HTTPMaxClients <= 0 {
		c.HTTPMaxClients = 10
	}

	if c.JPEGQuality <= 0 || c.JPEGQuality > 100 {
		c.JPEGQuality = 75
	}

	if c.USBGadgetName == "" {
		c.USBGadgetName = "kvm-gadget"
	}

	if c.USBVendorID == "" {
		c.USBVendorID = "0x1d6b"
	}

	if c.USBProductID == "" {
		c.USBProductID = "0x0104"
	}

	if c.USBManufacturer == "" {
		c.USBManufacturer = "U-BMC"
	}

	if c.USBProduct == "" {
		c.USBProduct = "Virtual KVM Device"
	}

	if c.ClientTimeout == 0 {
		c.ClientTimeout = 30 * time.Minute
	}

	if c.FrameTimeout == 0 {
		c.FrameTimeout = 5 * time.Second
	}

	if c.BufferCount == 0 {
		c.BufferCount = 4
	}

	return nil
}

// ToVideoConfig converts the service config to a video capture config.
func (c *Config) ToVideoConfig() *kvm.VideoConfig {
	return &kvm.VideoConfig{
		Device:      c.VideoDevice,
		Width:       c.VideoWidth,
		Height:      c.VideoHeight,
		Format:      kvm.PixelFormatYUYV,
		FPS:         c.VideoFPS,
		BufferCount: c.BufferCount,
		Timeout:     c.FrameTimeout,
	}
}

// ToVNCConfig converts the service config to a VNC server config.
func (c *Config) ToVNCConfig() *kvm.VNCConfig {
	return &kvm.VNCConfig{
		Address:          formatAddress(c.VNCPort),
		Width:            c.VideoWidth,
		Height:           c.VideoHeight,
		PixelFormat:      kvm.VNCPixelFormat32,
		Encodings:        []kvm.VNCEncoding{kvm.EncodingRaw},
		DesktopName:      "U-BMC KVM Console",
		Password:         c.VNCPassword,
		MaxClients:       c.VNCMaxClients,
		IdleTimeout:      c.ClientTimeout,
		EnableWebSocket:  true,
		WebSocketAddress: formatAddress(c.VNCWebSocketPort),
	}
}

// ToHTTPConfig converts the service config to an HTTP streaming config.
func (c *Config) ToHTTPConfig() *kvm.HTTPConfig {
	return &kvm.HTTPConfig{
		Address:    formatAddress(c.HTTPPort),
		Path:       "/stream",
		Quality:    c.JPEGQuality,
		MaxClients: c.HTTPMaxClients,
		Boundary:   "frame",
	}
}

// ToUSBGadgetConfig converts the service config to a USB gadget config.
func (c *Config) ToUSBGadgetConfig() *usb.GadgetConfig {
	return &usb.GadgetConfig{
		Name:              c.USBGadgetName,
		VendorID:          c.USBVendorID,
		ProductID:         c.USBProductID,
		SerialNumber:      c.USBSerialNumber,
		Manufacturer:      c.USBManufacturer,
		Product:           c.USBProduct,
		MaxPower:          250, // 500mA
		EnableKeyboard:    true,
		EnableMouse:       true,
		EnableMassStorage: c.EnableMassStorage,
	}
}

// formatAddress formats a port number as a network address.
func formatAddress(port int) string {
	return fmt.Sprintf(":%d", port)
}
