// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Compile-time assertion that KVMSrv implements service.Service.
var _ service.Service = (*KVMSrv)(nil)

// KVMSrv provides KVM (Keyboard, Video, Mouse) functionality for BMC environments.
type KVMSrv struct {
	config *config

	// Components
	videoCapture *videoCapture
	usbGadget    *usbManager
	vncServer    *vncManager
	httpServer   *httpManager

	// State
	running bool
	mu      sync.RWMutex

	// Observability
	logger *slog.Logger
	tracer trace.Tracer

	// Channels
	shutdown chan struct{}
	done     chan struct{}
}

// New creates a new KVM service instance.
func New(opts ...Option) *KVMSrv {
	cfg := &config{
		serviceName:        DefaultServiceName,
		serviceDescription: DefaultServiceDescription,
		serviceVersion:     DefaultServiceVersion,
		videoDevice:        DefaultVideoDevice,
		videoWidth:         DefaultVideoWidth,
		videoHeight:        DefaultVideoHeight,
		videoFPS:           DefaultVideoFPS,
		vncPort:            DefaultVNCPort,
		vncWebSocketPort:   DefaultVNCWebSocketPort,
		httpPort:           DefaultHTTPPort,
		enableVNC:          true,
		enableHTTP:         true,
		enableUSB:          true,
		enableMassStorage:  true,
		vncPassword:        "",
		vncMaxClients:      DefaultVNCMaxClients,
		httpMaxClients:     DefaultHTTPMaxClients,
		jpegQuality:        DefaultJPEGQuality,
		usbGadgetName:      DefaultUSBGadgetName,
		usbVendorID:        DefaultUSBVendorID,
		usbProductID:       DefaultUSBProductID,
		usbManufacturer:    DefaultUSBManufacturer,
		usbProduct:         DefaultUSBProduct,
		usbSerialNumber:    "",
		clientTimeout:      DefaultClientTimeout,
		frameTimeout:       DefaultFrameTimeout,
		bufferCount:        DefaultBufferCount,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &KVMSrv{
		config:   cfg,
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Name returns the service name.
func (s *KVMSrv) Name() string {
	return s.config.serviceName
}

// Run starts the KVM service.
func (s *KVMSrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	s.tracer = otel.Tracer(s.config.serviceName)

	ctx, span := s.tracer.Start(ctx, "kvmsrv.Run")
	defer span.End()

	s.logger = log.GetGlobalLogger().With("service", s.config.serviceName)
	s.logger.InfoContext(ctx, "Starting KVM service",
		"version", s.config.serviceVersion,
		"video_device", s.config.videoDevice,
		"vnc_enabled", s.config.enableVNC,
		"http_enabled", s.config.enableHTTP,
		"usb_enabled", s.config.enableUSB)

	// Validate configuration
	if err := s.config.Validate(); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Invalid configuration", "error", err)
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Set running state
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		close(s.done)
	}()

	// Initialize components
	if err := s.initializeComponents(ctx); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to initialize components", "error", err)
		return err
	}

	// Start components
	if err := s.startComponents(ctx); err != nil {
		span.RecordError(err)
		s.logger.ErrorContext(ctx, "Failed to start components", "error", err)
		s.cleanupComponents(ctx)
		return err
	}

	s.logger.InfoContext(ctx, "KVM service started successfully")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		s.logger.InfoContext(ctx, "KVM service shutting down", "reason", ctx.Err())
	case <-s.shutdown:
		s.logger.InfoContext(ctx, "KVM service shutdown requested")
	}

	// Cleanup
	s.cleanupComponents(ctx)
	s.logger.InfoContext(ctx, "KVM service stopped")

	return ctx.Err()
}

// initializeComponents initializes all service components.
func (s *KVMSrv) initializeComponents(ctx context.Context) error {
	var err error

	// Initialize USB gadget if enabled
	if s.config.enableUSB {
		s.logger.InfoContext(ctx, "Initializing USB gadget")
		s.usbGadget, err = newUSBManager(ctx, s.config.toUSBGadgetConfig())
		if err != nil {
			return fmt.Errorf("failed to initialize USB gadget: %w", err)
		}
	}

	// Initialize video capture
	s.logger.InfoContext(ctx, "Initializing video capture", "device", s.config.videoDevice)
	s.videoCapture, err = newVideoCapture(ctx, s.config.toVideoConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize video capture: %w", err)
	}

	// Initialize VNC server if enabled
	if s.config.enableVNC {
		s.logger.InfoContext(ctx, "Initializing VNC server", "port", s.config.vncPort)
		s.vncServer, err = newVNCManager(ctx, s.config.toVNCConfig())
		if err != nil {
			return fmt.Errorf("failed to initialize VNC server: %w", err)
		}
	}

	// Initialize HTTP server if enabled
	if s.config.enableHTTP {
		s.logger.InfoContext(ctx, "Initializing HTTP server", "port", s.config.httpPort)
		s.httpServer, err = newHTTPManager(ctx, s.config.toHTTPConfig())
		if err != nil {
			return fmt.Errorf("failed to initialize HTTP server: %w", err)
		}
	}

	return nil
}

// startComponents starts all initialized components.
func (s *KVMSrv) startComponents(ctx context.Context) error {
	// Start video capture
	if s.videoCapture != nil {
		if err := s.videoCapture.start(ctx); err != nil {
			return fmt.Errorf("failed to start video capture: %w", err)
		}
		s.logger.InfoContext(ctx, "Video capture started")
	}

	// Start USB gadget
	if s.usbGadget != nil {
		if err := s.usbGadget.start(ctx); err != nil {
			return fmt.Errorf("failed to start USB gadget: %w", err)
		}
		s.logger.InfoContext(ctx, "USB gadget started")
	}

	// Start VNC server
	if s.vncServer != nil {
		if err := s.vncServer.start(ctx, s.videoCapture.frames()); err != nil {
			return fmt.Errorf("failed to start VNC server: %w", err)
		}
		s.logger.InfoContext(ctx, "VNC server started")
	}

	// Start HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.start(ctx, s.videoCapture.frames()); err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}
		s.logger.InfoContext(ctx, "HTTP server started")
	}

	return nil
}

// cleanupComponents cleans up all components.
func (s *KVMSrv) cleanupComponents(ctx context.Context) {
	// Stop HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.stop(ctx); err != nil {
			s.logger.WarnContext(ctx, "Failed to stop HTTP server", "error", err)
		}
		s.httpServer = nil
	}

	// Stop VNC server
	if s.vncServer != nil {
		if err := s.vncServer.stop(ctx); err != nil {
			s.logger.WarnContext(ctx, "Failed to stop VNC server", "error", err)
		}
		s.vncServer = nil
	}

	// Stop USB gadget
	if s.usbGadget != nil {
		if err := s.usbGadget.stop(ctx); err != nil {
			s.logger.WarnContext(ctx, "Failed to stop USB gadget", "error", err)
		}
		s.usbGadget = nil
	}

	// Stop video capture
	if s.videoCapture != nil {
		if err := s.videoCapture.stop(ctx); err != nil {
			s.logger.WarnContext(ctx, "Failed to stop video capture", "error", err)
		}
		s.videoCapture = nil
	}
}

// IsRunning returns whether the service is currently running.
func (s *KVMSrv) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetStatus returns the current status of the KVM service.
func (s *KVMSrv) GetStatus() *ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &ServiceStatus{
		Running: s.running,
	}

	if s.videoCapture != nil {
		status.VideoCapture = s.videoCapture.getStatus()
	}

	if s.usbGadget != nil {
		status.USBGadget = s.usbGadget.getStatus()
	}

	if s.vncServer != nil {
		status.VNCServer = s.vncServer.getStatus()
	}

	if s.httpServer != nil {
		status.HTTPServer = s.httpServer.getStatus()
	}

	return status
}

// SetMassStorageFile sets the mass storage backing file.
func (s *KVMSrv) SetMassStorageFile(ctx context.Context, filePath string, cdromMode bool) error {
	if s.usbGadget == nil {
		return ErrUSBGadgetInitFailed
	}

	return s.usbGadget.setMassStorageFile(ctx, filePath, cdromMode)
}

// SendKeyboardInput sends keyboard input via USB HID.
func (s *KVMSrv) SendKeyboardInput(ctx context.Context, modifier byte, keys []byte) error {
	if s.usbGadget == nil {
		return ErrUSBGadgetInitFailed
	}

	return s.usbGadget.sendKeyboardInput(ctx, modifier, keys)
}

// SendMouseInput sends mouse input via USB HID.
func (s *KVMSrv) SendMouseInput(ctx context.Context, x, y uint16, buttons byte) error {
	if s.usbGadget == nil {
		return ErrUSBGadgetInitFailed
	}

	return s.usbGadget.sendMouseInput(ctx, x, y, buttons)
}

// SendWheelInput sends mouse wheel input via USB HID.
func (s *KVMSrv) SendWheelInput(ctx context.Context, wheel int8) error {
	if s.usbGadget == nil {
		return ErrUSBGadgetInitFailed
	}

	return s.usbGadget.sendWheelInput(ctx, wheel)
}

// Shutdown gracefully shuts down the service.
func (s *KVMSrv) Shutdown(ctx context.Context) error {
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		return nil
	}

	close(s.shutdown)

	// Wait for shutdown to complete or timeout
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return ErrOperationFailed
	}
}

// ServiceStatus represents the current status of the KVM service.
// ServiceStatus represents the current status of the KVM service and its components.
type ServiceStatus struct {
	Running      bool                // whether the service is currently running
	VideoCapture *VideoCaptureStatus // video capture component status
	USBGadget    *USBGadgetStatus    // USB gadget component status
	VNCServer    *VNCServerStatus    // VNC server component status
	HTTPServer   *HTTPServerStatus   // HTTP server component status
}

// VideoCaptureStatus represents the status of video capture.
// VideoCaptureStatus represents the status of video capture functionality.
type VideoCaptureStatus struct {
	Device  string // video device path
	Width   uint32 // video frame width
	Height  uint32 // video frame height
	FPS     uint32 // frames per second
	Frames  uint64 // total frames captured
	Dropped uint64 // total frames dropped
	Active  bool   // whether video capture is active
}

// USBGadgetStatus represents the status of the USB gadget.
// USBGadgetStatus represents the status of USB gadget functionality.
type USBGadgetStatus struct {
	Name        string // gadget name
	Bound       bool   // whether gadget is bound to UDC
	UDC         string // USB device controller name
	State       string // current gadget state
	Keyboard    bool   // whether keyboard function is enabled
	Mouse       bool   // whether mouse function is enabled
	MassStorage bool   // whether mass storage function is enabled
}

// VNCServerStatus represents the status of the VNC server.
// VNCServerStatus represents the status of the VNC server.
type VNCServerStatus struct {
	Address    string // server listen address
	Clients    int    // current number of connected clients
	MaxClients int    // maximum allowed clients
	Active     bool   // whether server is active
}

// HTTPServerStatus represents the status of the HTTP server.
// HTTPServerStatus represents the status of the HTTP server.
type HTTPServerStatus struct {
	Address    string // server listen address
	Clients    int    // current number of connected clients
	MaxClients int    // maximum allowed clients
	Active     bool   // whether server is active
}
