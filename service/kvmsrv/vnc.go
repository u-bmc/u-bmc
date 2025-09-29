// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/u-bmc/u-bmc/pkg/kvm"
	"github.com/u-bmc/u-bmc/pkg/log"
)

// vncManager manages the VNC server functionality.
type vncManager struct {
	config     *kvm.VNCConfig
	listener   net.Listener
	httpServer *http.Server
	clients    sync.Map
	running    atomic.Bool
	mu         sync.RWMutex
	stopCh     chan struct{}
	doneCh     chan struct{}

	// Statistics
	clientCount      atomic.Int32
	maxClients       int32
	framesSent       atomic.Uint64
	bytesTransmitted atomic.Uint64
}

// vncClient represents a connected VNC client.
type vncClient struct {
	conn       net.Conn
	id         string
	remoteAddr string
	connected  time.Time
	lastSeen   time.Time
	protocol   string
	encoding   []kvm.VNCEncoding
	mu         sync.Mutex
}

// newVNCManager creates a new VNC manager.
func newVNCManager(ctx context.Context, config *kvm.VNCConfig) (*vncManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid VNC config: %w", err)
	}

	vm := &vncManager{
		config:     config,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
		maxClients: int32(config.MaxClients),
	}

	return vm, nil
}

// start starts the VNC server.
func (vm *vncManager) start(ctx context.Context, frameCh <-chan *VideoFrame) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.running.Load() {
		return nil
	}

	l := log.GetGlobalLogger()
	l.InfoContext(ctx, "Starting VNC server", "address", vm.config.Address)

	// Create TCP listener
	var err error
	vm.listener, err = net.Listen("tcp", vm.config.Address)
	if err != nil {
		return fmt.Errorf("failed to create VNC listener: %w", err)
	}

	vm.running.Store(true)

	// Start frame distribution goroutine
	go vm.frameDistributor(ctx, frameCh)

	// Start connection handler
	go vm.connectionHandler(ctx)

	// Start WebSocket server if enabled
	if vm.config.EnableWebSocket {
		go vm.startWebSocketServer(ctx)
	}

	// Start client cleanup goroutine
	go vm.clientCleanup(ctx)

	l.InfoContext(ctx, "VNC server started successfully")
	return nil
}

// stop stops the VNC server.
func (vm *vncManager) stop(ctx context.Context) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if !vm.running.Load() {
		return nil
	}

	l := log.GetGlobalLogger()
	l.InfoContext(ctx, "Stopping VNC server")

	close(vm.stopCh)

	// Close listener
	if vm.listener != nil {
		vm.listener.Close()
	}

	// Stop HTTP server
	if vm.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		vm.httpServer.Shutdown(shutdownCtx)
	}

	// Close all client connections
	vm.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*vncClient); ok {
			client.conn.Close()
		}
		vm.clients.Delete(key)
		return true
	})

	// Wait for goroutines to finish
	select {
	case <-vm.doneCh:
	case <-time.After(5 * time.Second):
		l.WarnContext(ctx, "VNC server stop timeout")
	}

	vm.running.Store(false)
	l.InfoContext(ctx, "VNC server stopped")
	return nil
}

// getStatus returns the current VNC server status.
func (vm *vncManager) getStatus() *VNCServerStatus {
	return &VNCServerStatus{
		Address:    vm.config.Address,
		Clients:    int(vm.clientCount.Load()),
		MaxClients: int(vm.maxClients),
		Active:     vm.running.Load(),
	}
}

// connectionHandler handles incoming VNC connections.
func (vm *vncManager) connectionHandler(ctx context.Context) {
	defer close(vm.doneCh)

	l := log.GetGlobalLogger()

	for {
		select {
		case <-vm.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		// Accept connection with timeout
		if err := vm.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
			continue
		}

		conn, err := vm.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue // Timeout, check for shutdown
			}
			if vm.running.Load() {
				l.WarnContext(ctx, "Failed to accept VNC connection", "error", err)
			}
			continue
		}

		// Check client limit
		if vm.clientCount.Load() >= vm.maxClients && vm.maxClients > 0 {
			l.WarnContext(ctx, "VNC client limit reached, rejecting connection",
				"remote", conn.RemoteAddr())
			conn.Close()
			continue
		}

		// Handle client in separate goroutine
		go vm.handleClient(ctx, conn)
	}
}

// handleClient handles an individual VNC client connection.
func (vm *vncManager) handleClient(ctx context.Context, conn net.Conn) {
	l := log.GetGlobalLogger()
	remoteAddr := conn.RemoteAddr().String()

	l.InfoContext(ctx, "New VNC client connected", "remote", remoteAddr)

	client := &vncClient{
		conn:       conn,
		id:         fmt.Sprintf("vnc-%d", time.Now().UnixNano()),
		remoteAddr: remoteAddr,
		connected:  time.Now(),
		lastSeen:   time.Now(),
		protocol:   "RFB 003.008",
		encoding:   vm.config.Encodings,
	}

	vm.clients.Store(client.id, client)
	vm.clientCount.Add(1)

	defer func() {
		conn.Close()
		vm.clients.Delete(client.id)
		vm.clientCount.Add(-1)
		l.InfoContext(ctx, "VNC client disconnected", "remote", remoteAddr)
	}()

	// Perform VNC handshake
	if err := vm.performHandshake(ctx, client); err != nil {
		l.WarnContext(ctx, "VNC handshake failed", "remote", remoteAddr, "error", err)
		return
	}

	// Send initial framebuffer
	if err := vm.sendInitialFramebuffer(ctx, client); err != nil {
		l.WarnContext(ctx, "Failed to send initial framebuffer", "remote", remoteAddr, "error", err)
		return
	}

	// Handle client messages
	vm.handleClientMessages(ctx, client)
}

// performHandshake performs the VNC protocol handshake.
func (vm *vncManager) performHandshake(ctx context.Context, client *vncClient) error {
	// This is a simplified handshake implementation
	// In a real implementation, this would:
	// 1. Send protocol version
	// 2. Handle security negotiation
	// 3. Perform authentication if required
	// 4. Exchange client/server init messages

	conn := client.conn

	// Send protocol version
	protocolVersion := "RFB 003.008\n"
	if _, err := conn.Write([]byte(protocolVersion)); err != nil {
		return fmt.Errorf("failed to send protocol version: %w", err)
	}

	// Read client protocol version
	buf := make([]byte, 12)
	if _, err := conn.Read(buf); err != nil {
		return fmt.Errorf("failed to read client protocol version: %w", err)
	}

	// Send security types (no authentication for simplicity)
	securityTypes := []byte{1, 1} // 1 security type: None
	if _, err := conn.Write(securityTypes); err != nil {
		return fmt.Errorf("failed to send security types: %w", err)
	}

	// Read client security choice
	securityChoice := make([]byte, 1)
	if _, err := conn.Read(securityChoice); err != nil {
		return fmt.Errorf("failed to read security choice: %w", err)
	}

	// Send security result (success)
	securityResult := []byte{0, 0, 0, 0} // Success
	if _, err := conn.Write(securityResult); err != nil {
		return fmt.Errorf("failed to send security result: %w", err)
	}

	// Read client init message
	clientInit := make([]byte, 1)
	if _, err := conn.Read(clientInit); err != nil {
		return fmt.Errorf("failed to read client init: %w", err)
	}

	return nil
}

// sendInitialFramebuffer sends the initial framebuffer to a client.
func (vm *vncManager) sendInitialFramebuffer(ctx context.Context, client *vncClient) error {
	// Send server init message
	serverInit := vm.createServerInitMessage()
	if _, err := client.conn.Write(serverInit); err != nil {
		return fmt.Errorf("failed to send server init: %w", err)
	}

	// Create and send initial framebuffer
	framebuffer := vm.createInitialFramebuffer()
	if err := vm.sendFrameToClient(client, framebuffer); err != nil {
		return fmt.Errorf("failed to send initial framebuffer: %w", err)
	}

	return nil
}

// createServerInitMessage creates a VNC server init message.
func (vm *vncManager) createServerInitMessage() []byte {
	// VNC Server Init message format:
	// width (2 bytes) + height (2 bytes) + pixel format (16 bytes) + name length (4 bytes) + name

	width := uint16(vm.config.Width)
	height := uint16(vm.config.Height)
	desktopName := vm.config.DesktopName
	nameLength := uint32(len(desktopName))

	// Simplified pixel format (32-bit RGBA)
	pixelFormat := []byte{
		32,     // bits per pixel
		24,     // depth
		0,      // big endian flag
		1,      // true color flag
		0, 255, // red max (255)
		0, 255, // green max (255)
		0, 255, // blue max (255)
		16,      // red shift
		8,       // green shift
		0,       // blue shift
		0, 0, 0, // padding
	}

	msg := make([]byte, 24+len(desktopName))

	// Width and height
	msg[0] = byte(width >> 8)
	msg[1] = byte(width)
	msg[2] = byte(height >> 8)
	msg[3] = byte(height)

	// Pixel format
	copy(msg[4:20], pixelFormat)

	// Name length
	msg[20] = byte(nameLength >> 24)
	msg[21] = byte(nameLength >> 16)
	msg[22] = byte(nameLength >> 8)
	msg[23] = byte(nameLength)

	// Name
	copy(msg[24:], []byte(desktopName))

	return msg
}

// createInitialFramebuffer creates an initial framebuffer.
func (vm *vncManager) createInitialFramebuffer() []byte {
	width := vm.config.Width
	height := vm.config.Height

	// Create a simple pattern (black screen with white border)
	framebuffer := make([]byte, width*height*4) // 4 bytes per pixel (RGBA)

	for y := uint32(0); y < height; y++ {
		for x := uint32(0); x < width; x++ {
			idx := (y*width + x) * 4

			// White border, black interior
			if x < 5 || x >= width-5 || y < 5 || y >= height-5 {
				framebuffer[idx] = 255   // R
				framebuffer[idx+1] = 255 // G
				framebuffer[idx+2] = 255 // B
				framebuffer[idx+3] = 255 // A
			} else {
				framebuffer[idx] = 0     // R
				framebuffer[idx+1] = 0   // G
				framebuffer[idx+2] = 0   // B
				framebuffer[idx+3] = 255 // A
			}
		}
	}

	return framebuffer
}

// frameDistributor distributes video frames to connected clients.
func (vm *vncManager) frameDistributor(ctx context.Context, frameCh <-chan *VideoFrame) {
	l := log.GetGlobalLogger()

	for {
		select {
		case <-vm.stopCh:
			return
		case <-ctx.Done():
			return
		case frame := <-frameCh:
			if frame == nil {
				continue
			}

			// Convert frame to RGBA if needed
			var rgbaData []byte
			var err error

			if frame.Format == kvm.PixelFormatYUYV {
				rgbaData, err = ConvertYUYVToRGBA(frame.Data, frame.Width, frame.Height)
				if err != nil {
					l.WarnContext(ctx, "Failed to convert frame to RGBA", "error", err)
					continue
				}
			} else {
				rgbaData = frame.Data
			}

			// Send to all clients
			vm.clients.Range(func(key, value interface{}) bool {
				if client, ok := value.(*vncClient); ok {
					if err := vm.sendFrameToClient(client, rgbaData); err != nil {
						l.WarnContext(ctx, "Failed to send frame to client",
							"client", client.remoteAddr, "error", err)
					}
				}
				return true
			})

			vm.framesSent.Add(1)
		}
	}
}

// sendFrameToClient sends a framebuffer update to a specific client.
func (vm *vncManager) sendFrameToClient(client *vncClient, framebuffer []byte) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	// Create framebuffer update message
	// Message type (0) + padding (1) + number of rectangles (2) + rectangle data
	width := uint16(vm.config.Width)
	height := uint16(vm.config.Height)

	header := []byte{
		0,    // FramebufferUpdate message type
		0,    // padding
		0, 1, // number of rectangles (1)
		0, 0, // x position
		0, 0, // y position
		byte(width >> 8), byte(width), // width
		byte(height >> 8), byte(height), // height
		0, 0, 0, 0, // encoding type (Raw)
	}

	// Send header
	if _, err := client.conn.Write(header); err != nil {
		return err
	}

	// Send framebuffer data
	if _, err := client.conn.Write(framebuffer); err != nil {
		return err
	}

	vm.bytesTransmitted.Add(uint64(len(header) + len(framebuffer)))
	client.lastSeen = time.Now()

	return nil
}

// handleClientMessages handles incoming messages from a VNC client.
func (vm *vncManager) handleClientMessages(ctx context.Context, client *vncClient) {
	l := log.GetGlobalLogger()
	buffer := make([]byte, 1024)

	for {
		select {
		case <-vm.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		// Set read timeout
		client.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, err := client.conn.Read(buffer)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue // Timeout, check for shutdown
			}
			if vm.running.Load() {
				l.WarnContext(ctx, "Failed to read from VNC client",
					"client", client.remoteAddr, "error", err)
			}
			return
		}

		if n > 0 {
			// Process client message
			vm.processClientMessage(ctx, client, buffer[:n])
			client.lastSeen = time.Now()
		}
	}
}

// processClientMessage processes a message from a VNC client.
func (vm *vncManager) processClientMessage(ctx context.Context, client *vncClient, message []byte) {
	if len(message) == 0 {
		return
	}

	messageType := message[0]

	switch messageType {
	case 0: // SetPixelFormat
		// Client wants to change pixel format
		// For simplicity, we ignore this and keep using RGBA32

	case 2: // SetEncodings
		// Client is telling us which encodings it supports
		// For simplicity, we'll stick with Raw encoding

	case 3: // FramebufferUpdateRequest
		// Client is requesting a framebuffer update
		// We continuously send updates, so we can ignore this

	case 4: // KeyEvent
		if len(message) >= 8 {
			// Extract key event data
			downFlag := message[1] != 0
			keysym := uint32(message[4])<<24 | uint32(message[5])<<16 |
				uint32(message[6])<<8 | uint32(message[7])

			vm.handleKeyEvent(ctx, downFlag, keysym)
		}

	case 5: // PointerEvent
		if len(message) >= 6 {
			// Extract pointer event data
			buttonMask := message[1]
			x := uint16(message[2])<<8 | uint16(message[3])
			y := uint16(message[4])<<8 | uint16(message[5])

			vm.handlePointerEvent(ctx, buttonMask, x, y)
		}

	case 6: // ClientCutText
		// Client is sending clipboard data
		// For a BMC implementation, we might ignore this

	default:
		// Unknown message type, ignore
	}
}

// handleKeyEvent handles a key event from a VNC client.
func (vm *vncManager) handleKeyEvent(ctx context.Context, downFlag bool, keysym uint32) {
	// This would integrate with the USB HID keyboard functionality
	// For now, we just log the event
	l := log.GetGlobalLogger()
	l.DebugContext(ctx, "VNC key event", "down", downFlag, "keysym", keysym)
}

// handlePointerEvent handles a pointer event from a VNC client.
func (vm *vncManager) handlePointerEvent(ctx context.Context, buttonMask byte, x, y uint16) {
	// This would integrate with the USB HID mouse functionality
	// For now, we just log the event
	l := log.GetGlobalLogger()
	l.DebugContext(ctx, "VNC pointer event", "buttons", buttonMask, "x", x, "y", y)
}

// startWebSocketServer starts the WebSocket VNC server.
func (vm *vncManager) startWebSocketServer(ctx context.Context) {
	l := log.GetGlobalLogger()

	mux := http.NewServeMux()
	mux.HandleFunc("/", vm.handleWebSocketConnection)

	vm.httpServer = &http.Server{
		Addr:    vm.config.WebSocketAddress,
		Handler: mux,
	}

	l.InfoContext(ctx, "Starting VNC WebSocket server", "address", vm.config.WebSocketAddress)

	if err := vm.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.ErrorContext(ctx, "VNC WebSocket server failed", "error", err)
	}
}

// handleWebSocketConnection handles WebSocket VNC connections.
func (vm *vncManager) handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	// This would implement WebSocket upgrade and VNC-over-WebSocket protocol
	// For simplicity, we'll just return a basic response
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("VNC WebSocket support not yet implemented"))
}

// clientCleanup periodically cleans up stale client connections.
func (vm *vncManager) clientCleanup(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-vm.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			vm.cleanupStaleClients(ctx)
		}
	}
}

// cleanupStaleClients removes clients that haven't been seen recently.
func (vm *vncManager) cleanupStaleClients(ctx context.Context) {
	l := log.GetGlobalLogger()
	now := time.Now()
	timeout := vm.config.IdleTimeout

	vm.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*vncClient); ok {
			if now.Sub(client.lastSeen) > timeout {
				l.InfoContext(ctx, "Cleaning up stale VNC client",
					"client", client.remoteAddr, "idle", now.Sub(client.lastSeen))
				client.conn.Close()
				vm.clients.Delete(key)
				vm.clientCount.Add(-1)
			}
		}
		return true
	})
}
