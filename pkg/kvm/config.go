// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package kvm

import "time"

// PixelFormat represents a video pixel format.
type PixelFormat uint32

const (
	// PixelFormatYUYV represents YUYV 4:2:2 format.
	PixelFormatYUYV PixelFormat = 0x56595559

	// PixelFormatMJPEG represents Motion-JPEG format.
	PixelFormatMJPEG PixelFormat = 0x4745504A

	// PixelFormatRGB24 represents RGB24 format.
	PixelFormatRGB24 PixelFormat = 0x33424752

	// PixelFormatBGR24 represents BGR24 format.
	PixelFormatBGR24 PixelFormat = 0x33524742
)

// VideoConfig represents the configuration for video capture.
type VideoConfig struct {
	// Device is the path to the video device (e.g., "/dev/video0").
	Device string

	// Width is the desired frame width in pixels.
	Width uint32

	// Height is the desired frame height in pixels.
	Height uint32

	// Format is the desired pixel format.
	Format PixelFormat

	// FPS is the desired frames per second (0 for default).
	FPS uint32

	// BufferCount is the number of capture buffers to allocate.
	BufferCount uint32

	// Timeout is the maximum time to wait for a frame.
	Timeout time.Duration
}

// VNCPixelFormat represents a VNC pixel format.
type VNCPixelFormat int

const (
	// VNCPixelFormat32 represents 32-bit RGBA pixel format.
	VNCPixelFormat32 VNCPixelFormat = 32

	// VNCPixelFormat16 represents 16-bit RGB pixel format.
	VNCPixelFormat16 VNCPixelFormat = 16

	// VNCPixelFormat8 represents 8-bit indexed pixel format.
	VNCPixelFormat8 VNCPixelFormat = 8
)

// VNCEncoding represents a VNC encoding type.
type VNCEncoding int

const (
	// EncodingRaw represents raw pixel data encoding.
	EncodingRaw VNCEncoding = 0

	// EncodingCopyRect represents copy rectangle encoding.
	EncodingCopyRect VNCEncoding = 1

	// EncodingRRE represents rise-and-run-length encoding.
	EncodingRRE VNCEncoding = 2

	// EncodingHextile represents hextile encoding.
	EncodingHextile VNCEncoding = 5

	// EncodingZRLE represents zlib run-length encoding.
	EncodingZRLE VNCEncoding = 16

	// EncodingTight represents tight encoding.
	EncodingTight VNCEncoding = 7
)

// VNCConfig represents the configuration for a VNC server.
type VNCConfig struct {
	// Address is the network address to bind to (e.g., ":5900").
	Address string

	// Width is the framebuffer width in pixels.
	Width uint32

	// Height is the framebuffer height in pixels.
	Height uint32

	// PixelFormat is the VNC pixel format.
	PixelFormat VNCPixelFormat

	// Encodings is the list of supported encodings.
	Encodings []VNCEncoding

	// DesktopName is the name shown to VNC clients.
	DesktopName string

	// Password is the VNC password (empty for no authentication).
	Password string

	// MaxClients is the maximum number of concurrent clients (0 for unlimited).
	MaxClients int

	// IdleTimeout is the timeout for idle clients.
	IdleTimeout time.Duration

	// EnableWebSocket enables WebSocket support.
	EnableWebSocket bool

	// WebSocketAddress is the WebSocket server address.
	WebSocketAddress string
}

// HTTPConfig represents the configuration for HTTP streaming.
type HTTPConfig struct {
	// Address is the network address to bind to (e.g., ":8080").
	Address string

	// Path is the HTTP path for the stream (e.g., "/stream").
	Path string

	// Quality is the JPEG quality (1-100).
	Quality int

	// MaxClients is the maximum number of concurrent clients (0 for unlimited).
	MaxClients int

	// Boundary is the multipart boundary string.
	Boundary string
}

// FrameInfo represents information about a video frame.
type FrameInfo struct {
	// Width is the frame width in pixels.
	Width uint32

	// Height is the frame height in pixels.
	Height uint32

	// Format is the pixel format.
	Format PixelFormat

	// Timestamp is the frame capture timestamp.
	Timestamp time.Time

	// SequenceNumber is the frame sequence number.
	SequenceNumber uint64

	// BytesUsed is the number of bytes used in the frame buffer.
	BytesUsed uint32
}

// DefaultVideoConfig returns a default video capture configuration.
func DefaultVideoConfig() *VideoConfig {
	return &VideoConfig{
		Device:      "/dev/video0",
		Width:       640,
		Height:      480,
		Format:      PixelFormatYUYV,
		FPS:         30,
		BufferCount: 4,
		Timeout:     5 * time.Second,
	}
}

// DefaultVNCConfig returns a default VNC server configuration.
func DefaultVNCConfig() *VNCConfig {
	return &VNCConfig{
		Address:          ":5900",
		Width:            640,
		Height:           480,
		PixelFormat:      VNCPixelFormat32,
		Encodings:        []VNCEncoding{EncodingRaw},
		DesktopName:      "U-BMC KVM",
		Password:         "",
		MaxClients:       5,
		IdleTimeout:      30 * time.Minute,
		EnableWebSocket:  true,
		WebSocketAddress: ":5901",
	}
}

// DefaultHTTPConfig returns a default HTTP streaming configuration.
func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Address:    ":8080",
		Path:       "/stream",
		Quality:    75,
		MaxClients: 10,
		Boundary:   "frame",
	}
}

// Validate validates the video configuration.
func (c *VideoConfig) Validate() error {
	if c.Device == "" {
		return ErrInvalidConfig
	}
	if c.Width == 0 || c.Height == 0 {
		return ErrInvalidConfig
	}
	if c.BufferCount == 0 {
		c.BufferCount = 4
	}
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	return nil
}

// Validate validates the VNC configuration.
func (c *VNCConfig) Validate() error {
	if c.Address == "" {
		return ErrInvalidConfig
	}
	if c.Width == 0 || c.Height == 0 {
		return ErrInvalidConfig
	}
	if c.DesktopName == "" {
		c.DesktopName = "U-BMC KVM"
	}
	if len(c.Encodings) == 0 {
		c.Encodings = []VNCEncoding{EncodingRaw}
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 30 * time.Minute
	}
	return nil
}

// Validate validates the HTTP configuration.
func (c *HTTPConfig) Validate() error {
	if c.Address == "" {
		return ErrInvalidConfig
	}
	if c.Path == "" {
		c.Path = "/stream"
	}
	if c.Quality <= 0 || c.Quality > 100 {
		c.Quality = 75
	}
	if c.Boundary == "" {
		c.Boundary = "frame"
	}
	return nil
}
