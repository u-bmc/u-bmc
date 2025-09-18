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
)

// Compile-time assertion that KVMSrv implements service.Service.
var _ service.Service = (*KVMSrv)(nil)

// KVMSrv provides KVM (Keyboard, Video, Mouse) functionality for BMC environments.
type KVMSrv struct {
	config

	// Components
	videoCapture *videoCapture
	usbGadget    *usbManager
	vncServer    *vncManager
	httpServer   *httpManager

	// State
	running bool
	mu      sync.RWMutex

	// Channels
	shutdown chan struct{}
	done     chan struct{}
}

// New creates a new KVM service instance.
func New(opts ...Option) *KVMSrv {
	cfg := &config{
		name: "kvmsrv",
		cfg:  DefaultConfig(),
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &KVMSrv{
		config:   *cfg,
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Name returns the service name.
func (s *KVMSrv) Name() string {
	return s.name
}

// Run starts the KVM service.
func (s *KVMSrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()
	l.InfoContext(ctx, "Starting KVM service", "service", s.name)

	// Validate configuration
	if err := s.cfg.Validate(); err != nil {
		l.ErrorContext(ctx, "Invalid configuration", "error", err)
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
	if err := s.initializeComponents(ctx, l); err != nil {
		l.ErrorContext(ctx, "Failed to initialize components", "error", err)
		return err
	}

	// Start components
	if err := s.startComponents(ctx, l); err != nil {
		l.ErrorContext(ctx, "Failed to start components", "error", err)
		s.cleanupComponents(ctx, l)
		return err
	}

	l.InfoContext(ctx, "KVM service started successfully")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		l.InfoContext(ctx, "KVM service shutting down", "reason", ctx.Err())
	case <-s.shutdown:
		l.InfoContext(ctx, "KVM service shutdown requested")
	}

	// Cleanup
	s.cleanupComponents(ctx, l)
	l.InfoContext(ctx, "KVM service stopped")

	return ctx.Err()
}

// initializeComponents initializes all service components.
func (s *KVMSrv) initializeComponents(ctx context.Context, l *slog.Logger) error {
	var err error

	// Initialize USB gadget if enabled
	if s.cfg.EnableUSB {
		l.InfoContext(ctx, "Initializing USB gadget")
		s.usbGadget, err = newUSBManager(ctx, s.cfg.ToUSBGadgetConfig())
		if err != nil {
			return fmt.Errorf("failed to initialize USB gadget: %w", err)
		}
	}

	// Initialize video capture
	l.InfoContext(ctx, "Initializing video capture", "device", s.cfg.VideoDevice)
	s.videoCapture, err = newVideoCapture(ctx, s.cfg.ToVideoConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize video capture: %w", err)
	}

	// Initialize VNC server if enabled
	if s.cfg.EnableVNC {
		l.InfoContext(ctx, "Initializing VNC server", "port", s.cfg.VNCPort)
		s.vncServer, err = newVNCManager(ctx, s.cfg.ToVNCConfig())
		if err != nil {
			return fmt.Errorf("failed to initialize VNC server: %w", err)
		}
	}

	// Initialize HTTP server if enabled
	if s.cfg.EnableHTTP {
		l.InfoContext(ctx, "Initializing HTTP server", "port", s.cfg.HTTPPort)
		s.httpServer, err = newHTTPManager(ctx, s.cfg.ToHTTPConfig())
		if err != nil {
			return fmt.Errorf("failed to initialize HTTP server: %w", err)
		}
	}

	return nil
}

// startComponents starts all initialized components.
func (s *KVMSrv) startComponents(ctx context.Context, l *slog.Logger) error {
	// Start video capture
	if s.videoCapture != nil {
		if err := s.videoCapture.start(ctx); err != nil {
			return fmt.Errorf("failed to start video capture: %w", err)
		}
		l.InfoContext(ctx, "Video capture started")
	}

	// Start USB gadget
	if s.usbGadget != nil {
		if err := s.usbGadget.start(ctx); err != nil {
			return fmt.Errorf("failed to start USB gadget: %w", err)
		}
		l.InfoContext(ctx, "USB gadget started")
	}

	// Start VNC server
	if s.vncServer != nil {
		if err := s.vncServer.start(ctx, s.videoCapture.frames()); err != nil {
			return fmt.Errorf("failed to start VNC server: %w", err)
		}
		l.InfoContext(ctx, "VNC server started")
	}

	// Start HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.start(ctx, s.videoCapture.frames()); err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}
		l.InfoContext(ctx, "HTTP server started")
	}

	return nil
}

// cleanupComponents cleans up all components.
func (s *KVMSrv) cleanupComponents(ctx context.Context, l *slog.Logger) {
	// Stop HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.stop(ctx); err != nil {
			l.WarnContext(ctx, "Failed to stop HTTP server", "error", err)
		}
		s.httpServer = nil
	}

	// Stop VNC server
	if s.vncServer != nil {
		if err := s.vncServer.stop(ctx); err != nil {
			l.WarnContext(ctx, "Failed to stop VNC server", "error", err)
		}
		s.vncServer = nil
	}

	// Stop USB gadget
	if s.usbGadget != nil {
		if err := s.usbGadget.stop(ctx); err != nil {
			l.WarnContext(ctx, "Failed to stop USB gadget", "error", err)
		}
		s.usbGadget = nil
	}

	// Stop video capture
	if s.videoCapture != nil {
		if err := s.videoCapture.stop(ctx); err != nil {
			l.WarnContext(ctx, "Failed to stop video capture", "error", err)
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
type ServiceStatus struct {
	Running      bool                `json:"running"`
	VideoCapture *VideoCaptureStatus `json:"video_capture,omitempty"`
	USBGadget    *USBGadgetStatus    `json:"usb_gadget,omitempty"`
	VNCServer    *VNCServerStatus    `json:"vnc_server,omitempty"`
	HTTPServer   *HTTPServerStatus   `json:"http_server,omitempty"`
}

// VideoCaptureStatus represents the status of video capture.
type VideoCaptureStatus struct {
	Device  string `json:"device"`
	Width   uint32 `json:"width"`
	Height  uint32 `json:"height"`
	FPS     uint32 `json:"fps"`
	Frames  uint64 `json:"frames"`
	Dropped uint64 `json:"dropped"`
	Active  bool   `json:"active"`
}

// USBGadgetStatus represents the status of the USB gadget.
type USBGadgetStatus struct {
	Name        string `json:"name"`
	Bound       bool   `json:"bound"`
	UDC         string `json:"udc,omitempty"`
	State       string `json:"state,omitempty"`
	Keyboard    bool   `json:"keyboard"`
	Mouse       bool   `json:"mouse"`
	MassStorage bool   `json:"mass_storage"`
}

// VNCServerStatus represents the status of the VNC server.
type VNCServerStatus struct {
	Address    string `json:"address"`
	Clients    int    `json:"clients"`
	MaxClients int    `json:"max_clients"`
	Active     bool   `json:"active"`
}

// HTTPServerStatus represents the status of the HTTP server.
type HTTPServerStatus struct {
	Address    string `json:"address"`
	Clients    int    `json:"clients"`
	MaxClients int    `json:"max_clients"`
	Active     bool   `json:"active"`
}
