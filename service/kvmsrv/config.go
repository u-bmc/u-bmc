// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"fmt"
	"time"

	"github.com/u-bmc/u-bmc/pkg/kvm"
	"github.com/u-bmc/u-bmc/pkg/usb"
)

// Default configuration constants.
const (
	DefaultServiceName        = "kvmsrv"
	DefaultServiceDescription = "KVM (Keyboard, Video, Mouse) service for BMC environments"
	DefaultServiceVersion     = "1.0.0"
	DefaultVideoDevice        = "/dev/video0"
	DefaultVideoWidth         = 640
	DefaultVideoHeight        = 480
	DefaultVideoFPS           = 30
	DefaultVNCPort            = 5900
	DefaultVNCWebSocketPort   = 5901
	DefaultHTTPPort           = 8080
	DefaultVNCMaxClients      = 5
	DefaultHTTPMaxClients     = 10
	DefaultJPEGQuality        = 75
	DefaultUSBGadgetName      = "kvm-gadget"
	DefaultUSBVendorID        = "0x1d6b" // Linux Foundation
	DefaultUSBProductID       = "0x0104" // Multifunction Composite Gadget
	DefaultUSBManufacturer    = "U-BMC"
	DefaultUSBProduct         = "Virtual KVM Device"
	DefaultClientTimeout      = 30 * time.Minute
	DefaultFrameTimeout       = 5 * time.Second
	DefaultBufferCount        = 4
)

// config represents the internal configuration for the KVM service.
type config struct {
	// Service configuration
	serviceName        string
	serviceDescription string
	serviceVersion     string

	// Video configuration
	videoDevice  string
	videoWidth   uint32
	videoHeight  uint32
	videoFPS     uint32
	bufferCount  uint32
	frameTimeout time.Duration

	// Network configuration
	vncPort          int
	vncWebSocketPort int
	httpPort         int

	// Feature flags
	enableVNC         bool
	enableHTTP        bool
	enableUSB         bool
	enableMassStorage bool

	// VNC configuration
	vncPassword   string
	vncMaxClients int

	// HTTP configuration
	httpMaxClients int
	jpegQuality    int

	// USB configuration
	usbGadgetName   string
	usbVendorID     string
	usbProductID    string
	usbManufacturer string
	usbProduct      string
	usbSerialNumber string

	// Client configuration
	clientTimeout time.Duration
}

// Option represents a configuration option for the KVM service.
type Option interface {
	apply(*config)
}

type serviceNameOption struct {
	name string
}

func (o *serviceNameOption) apply(c *config) {
	c.serviceName = o.name
}

// WithServiceName sets the service name.
func WithServiceName(name string) Option {
	return &serviceNameOption{name: name}
}

type serviceDescriptionOption struct {
	description string
}

func (o *serviceDescriptionOption) apply(c *config) {
	c.serviceDescription = o.description
}

// WithServiceDescription sets the service description.
func WithServiceDescription(description string) Option {
	return &serviceDescriptionOption{description: description}
}

type serviceVersionOption struct {
	version string
}

func (o *serviceVersionOption) apply(c *config) {
	c.serviceVersion = o.version
}

// WithServiceVersion sets the service version.
func WithServiceVersion(version string) Option {
	return &serviceVersionOption{version: version}
}

type videoDeviceOption struct {
	device string
}

func (o *videoDeviceOption) apply(c *config) {
	c.videoDevice = o.device
}

// WithVideoDevice sets the video capture device path.
func WithVideoDevice(device string) Option {
	return &videoDeviceOption{device: device}
}

type videoResolutionOption struct {
	width, height uint32
}

func (o *videoResolutionOption) apply(c *config) {
	c.videoWidth = o.width
	c.videoHeight = o.height
}

// WithVideoResolution sets the video capture resolution.
func WithVideoResolution(width, height uint32) Option {
	return &videoResolutionOption{width: width, height: height}
}

type videoFPSOption struct {
	fps uint32
}

func (o *videoFPSOption) apply(c *config) {
	c.videoFPS = o.fps
}

// WithVideoFPS sets the video capture frame rate.
func WithVideoFPS(fps uint32) Option {
	return &videoFPSOption{fps: fps}
}

type vncPortOption struct {
	port int
}

func (o *vncPortOption) apply(c *config) {
	c.vncPort = o.port
}

// WithVNCPort sets the VNC server port.
func WithVNCPort(port int) Option {
	return &vncPortOption{port: port}
}

type httpPortOption struct {
	port int
}

func (o *httpPortOption) apply(c *config) {
	c.httpPort = o.port
}

// WithHTTPPort sets the HTTP streaming port.
func WithHTTPPort(port int) Option {
	return &httpPortOption{port: port}
}

type enableVNCOption struct {
	enable bool
}

func (o *enableVNCOption) apply(c *config) {
	c.enableVNC = o.enable
}

// WithVNC enables or disables the VNC server.
func WithVNC(enable bool) Option {
	return &enableVNCOption{enable: enable}
}

type enableHTTPOption struct {
	enable bool
}

func (o *enableHTTPOption) apply(c *config) {
	c.enableHTTP = o.enable
}

// WithHTTP enables or disables HTTP streaming.
func WithHTTP(enable bool) Option {
	return &enableHTTPOption{enable: enable}
}

type enableUSBOption struct {
	enable bool
}

func (o *enableUSBOption) apply(c *config) {
	c.enableUSB = o.enable
}

// WithUSB enables or disables USB gadget functionality.
func WithUSB(enable bool) Option {
	return &enableUSBOption{enable: enable}
}

type vncPasswordOption struct {
	password string
}

func (o *vncPasswordOption) apply(c *config) {
	c.vncPassword = o.password
}

// WithVNCPassword sets the VNC authentication password.
func WithVNCPassword(password string) Option {
	return &vncPasswordOption{password: password}
}

type vncMaxClientsOption struct {
	maxClients int
}

func (o *vncMaxClientsOption) apply(c *config) {
	c.vncMaxClients = o.maxClients
}

// WithVNCMaxClients sets the maximum number of VNC clients.
func WithVNCMaxClients(maxClients int) Option {
	return &vncMaxClientsOption{maxClients: maxClients}
}

type httpMaxClientsOption struct {
	maxClients int
}

func (o *httpMaxClientsOption) apply(c *config) {
	c.httpMaxClients = o.maxClients
}

// WithHTTPMaxClients sets the maximum number of HTTP clients.
func WithHTTPMaxClients(maxClients int) Option {
	return &httpMaxClientsOption{maxClients: maxClients}
}

type jpegQualityOption struct {
	quality int
}

func (o *jpegQualityOption) apply(c *config) {
	c.jpegQuality = o.quality
}

// WithJPEGQuality sets the JPEG compression quality (1-100).
func WithJPEGQuality(quality int) Option {
	return &jpegQualityOption{quality: quality}
}

type usbGadgetNameOption struct {
	name string
}

func (o *usbGadgetNameOption) apply(c *config) {
	c.usbGadgetName = o.name
}

// WithUSBGadgetName sets the USB gadget name.
func WithUSBGadgetName(name string) Option {
	return &usbGadgetNameOption{name: name}
}

type clientTimeoutOption struct {
	timeout time.Duration
}

func (o *clientTimeoutOption) apply(c *config) {
	c.clientTimeout = o.timeout
}

// WithClientTimeout sets the client idle timeout.
func WithClientTimeout(timeout time.Duration) Option {
	return &clientTimeoutOption{timeout: timeout}
}

type frameTimeoutOption struct {
	timeout time.Duration
}

func (o *frameTimeoutOption) apply(c *config) {
	c.frameTimeout = o.timeout
}

// WithFrameTimeout sets the video frame capture timeout.
func WithFrameTimeout(timeout time.Duration) Option {
	return &frameTimeoutOption{timeout: timeout}
}

type bufferCountOption struct {
	count uint32
}

func (o *bufferCountOption) apply(c *config) {
	c.bufferCount = o.count
}

// WithBufferCount sets the number of video capture buffers.
func WithBufferCount(count uint32) Option {
	return &bufferCountOption{count: count}
}

// WithName is a backward compatibility alias for WithServiceName.
// Deprecated: Use WithServiceName instead.
func WithName(name string) Option {
	return WithServiceName(name)
}

// Validate validates the configuration and sets defaults where appropriate.
func (c *config) Validate() error {
	if c.serviceName == "" {
		c.serviceName = DefaultServiceName
	}

	if c.serviceDescription == "" {
		c.serviceDescription = DefaultServiceDescription
	}

	if c.serviceVersion == "" {
		c.serviceVersion = DefaultServiceVersion
	}

	if c.videoDevice == "" {
		return ErrInvalidConfiguration
	}

	if c.videoWidth == 0 {
		c.videoWidth = DefaultVideoWidth
	}

	if c.videoHeight == 0 {
		c.videoHeight = DefaultVideoHeight
	}

	if c.vncPort <= 0 {
		c.vncPort = DefaultVNCPort
	}
	if c.vncPort > 65535 {
		return ErrInvalidConfiguration
	}

	if c.httpPort <= 0 {
		c.httpPort = DefaultHTTPPort
	}
	if c.httpPort > 65535 {
		return ErrInvalidConfiguration
	}

	if c.vncWebSocketPort <= 0 {
		c.vncWebSocketPort = DefaultVNCWebSocketPort
	}
	if c.vncWebSocketPort > 65535 {
		return ErrInvalidConfiguration
	}

	if c.videoFPS == 0 {
		c.videoFPS = DefaultVideoFPS
	}

	if c.vncMaxClients <= 0 {
		c.vncMaxClients = DefaultVNCMaxClients
	}

	if c.httpMaxClients <= 0 {
		c.httpMaxClients = DefaultHTTPMaxClients
	}

	if c.jpegQuality <= 0 || c.jpegQuality > 100 {
		c.jpegQuality = DefaultJPEGQuality
	}

	if c.usbGadgetName == "" {
		c.usbGadgetName = DefaultUSBGadgetName
	}

	if c.usbVendorID == "" {
		c.usbVendorID = DefaultUSBVendorID
	}

	if c.usbProductID == "" {
		c.usbProductID = DefaultUSBProductID
	}

	if c.usbManufacturer == "" {
		c.usbManufacturer = DefaultUSBManufacturer
	}

	if c.usbProduct == "" {
		c.usbProduct = DefaultUSBProduct
	}

	if c.clientTimeout == 0 {
		c.clientTimeout = DefaultClientTimeout
	}

	if c.frameTimeout == 0 {
		c.frameTimeout = DefaultFrameTimeout
	}

	if c.bufferCount == 0 {
		c.bufferCount = DefaultBufferCount
	}

	return nil
}

// toVideoConfig converts the service config to a video capture config.
func (c *config) toVideoConfig() *kvm.VideoConfig {
	return &kvm.VideoConfig{
		Device:      c.videoDevice,
		Width:       c.videoWidth,
		Height:      c.videoHeight,
		Format:      kvm.PixelFormatYUYV,
		FPS:         c.videoFPS,
		BufferCount: c.bufferCount,
		Timeout:     c.frameTimeout,
	}
}

// toVNCConfig converts the service config to a VNC server config.
func (c *config) toVNCConfig() *kvm.VNCConfig {
	return &kvm.VNCConfig{
		Address:          formatAddress(c.vncPort),
		Width:            c.videoWidth,
		Height:           c.videoHeight,
		PixelFormat:      kvm.VNCPixelFormat32,
		Encodings:        []kvm.VNCEncoding{kvm.EncodingRaw},
		DesktopName:      "U-BMC KVM Console",
		Password:         c.vncPassword,
		MaxClients:       c.vncMaxClients,
		IdleTimeout:      c.clientTimeout,
		EnableWebSocket:  true,
		WebSocketAddress: formatAddress(c.vncWebSocketPort),
	}
}

// toHTTPConfig converts the service config to an HTTP streaming config.
func (c *config) toHTTPConfig() *kvm.HTTPConfig {
	return &kvm.HTTPConfig{
		Address:    formatAddress(c.httpPort),
		Path:       "/stream",
		Quality:    c.jpegQuality,
		MaxClients: c.httpMaxClients,
		Boundary:   "frame",
	}
}

// toUSBGadgetConfig converts the service config to a USB gadget config.
func (c *config) toUSBGadgetConfig() *usb.GadgetConfig {
	return &usb.GadgetConfig{
		Name:              c.usbGadgetName,
		VendorID:          c.usbVendorID,
		ProductID:         c.usbProductID,
		SerialNumber:      c.usbSerialNumber,
		Manufacturer:      c.usbManufacturer,
		Product:           c.usbProduct,
		MaxPower:          250, // 500mA
		EnableKeyboard:    true,
		EnableMouse:       true,
		EnableMassStorage: c.enableMassStorage,
	}
}

// formatAddress formats a port number as a network address.
func formatAddress(port int) string {
	return fmt.Sprintf(":%d", port)
}
