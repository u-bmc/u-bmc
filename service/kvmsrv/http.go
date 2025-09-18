// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"context"
	"fmt"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/u-bmc/u-bmc/pkg/kvm"
	"github.com/u-bmc/u-bmc/pkg/log"
)

// httpManager manages HTTP MJPEG streaming functionality.
type httpManager struct {
	config   *kvm.HTTPConfig
	server   *http.Server
	listener net.Listener
	clients  sync.Map
	running  atomic.Bool
	mu       sync.RWMutex
	stopCh   chan struct{}
	doneCh   chan struct{}

	// Statistics
	clientCount      atomic.Int32
	maxClients       int32
	framesSent       atomic.Uint64
	bytesTransmitted atomic.Uint64
}

// httpClient represents a connected HTTP client.
type httpClient struct {
	id         string
	remoteAddr string
	connected  time.Time
	lastSeen   time.Time
	writer     *multipart.Writer
	response   http.ResponseWriter
	request    *http.Request
	mu         sync.Mutex
}

// newHTTPManager creates a new HTTP manager.
func newHTTPManager(ctx context.Context, config *kvm.HTTPConfig) (*httpManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid HTTP config: %w", err)
	}

	hm := &httpManager{
		config:     config,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
		maxClients: int32(config.MaxClients),
	}

	return hm, nil
}

// start starts the HTTP streaming server.
func (hm *httpManager) start(ctx context.Context, frameCh <-chan *VideoFrame) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.running.Load() {
		return nil
	}

	l := log.GetGlobalLogger()
	l.InfoContext(ctx, "Starting HTTP streaming server", "address", hm.config.Address)

	// Create TCP listener
	var err error
	hm.listener, err = net.Listen("tcp", hm.config.Address)
	if err != nil {
		return fmt.Errorf("failed to create HTTP listener: %w", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc(hm.config.Path, hm.handleStreamRequest)
	mux.HandleFunc("/", hm.handleRootRequest)

	hm.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	hm.running.Store(true)

	// Start frame distribution goroutine
	go hm.frameDistributor(ctx, frameCh)

	// Start server
	go func() {
		if err := hm.server.Serve(hm.listener); err != nil && err != http.ErrServerClosed {
			l.ErrorContext(ctx, "HTTP server failed", "error", err)
		}
	}()

	// Start client cleanup goroutine
	go hm.clientCleanup(ctx)

	l.InfoContext(ctx, "HTTP streaming server started successfully")
	return nil
}

// stop stops the HTTP streaming server.
func (hm *httpManager) stop(ctx context.Context) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.running.Load() {
		return nil
	}

	l := log.GetGlobalLogger()
	l.InfoContext(ctx, "Stopping HTTP streaming server")

	close(hm.stopCh)

	// Shutdown server
	if hm.server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		hm.server.Shutdown(shutdownCtx)
	}

	// Close all client connections
	hm.clients.Range(func(key, value interface{}) bool {
		hm.clients.Delete(key)
		return true
	})

	// Wait for goroutines to finish
	select {
	case <-hm.doneCh:
	case <-time.After(5 * time.Second):
		l.WarnContext(ctx, "HTTP server stop timeout")
	}

	hm.running.Store(false)
	l.InfoContext(ctx, "HTTP streaming server stopped")
	return nil
}

// getStatus returns the current HTTP server status.
func (hm *httpManager) getStatus() *HTTPServerStatus {
	return &HTTPServerStatus{
		Address:    hm.config.Address,
		Clients:    int(hm.clientCount.Load()),
		MaxClients: int(hm.maxClients),
		Active:     hm.running.Load(),
	}
}

// handleRootRequest handles requests to the root path.
func (hm *httpManager) handleRootRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Serve a simple HTML page with an embedded image
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>U-BMC KVM Console</title>
    <style>
        body { margin: 0; padding: 20px; background: #000; color: #fff; font-family: Arial, sans-serif; }
        .container { text-align: center; }
        img { border: 2px solid #333; max-width: 100%%; height: auto; }
        .info { margin-top: 20px; font-size: 14px; color: #ccc; }
    </style>
</head>
<body>
    <div class="container">
        <h1>U-BMC KVM Console</h1>
        <img src="%s" alt="Console Stream" />
        <div class="info">
            MJPEG Stream - Quality: %d%% - Refresh automatically
        </div>
    </div>
</body>
</html>`, hm.config.Path, hm.config.Quality)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleStreamRequest handles MJPEG streaming requests.
func (hm *httpManager) handleStreamRequest(w http.ResponseWriter, r *http.Request) {
	l := log.GetGlobalLogger()
	ctx := r.Context()
	remoteAddr := r.RemoteAddr

	l.InfoContext(ctx, "New HTTP streaming client", "remote", remoteAddr)

	// Check client limit
	if hm.clientCount.Load() >= hm.maxClients && hm.maxClients > 0 {
		l.WarnContext(ctx, "HTTP client limit reached, rejecting connection", "remote", remoteAddr)
		http.Error(w, "Server busy", http.StatusServiceUnavailable)
		return
	}

	// Set response headers for MJPEG streaming
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", hm.config.Boundary))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Connection", "close")

	// Create multipart writer
	writer := multipart.NewWriter(w)
	if err := writer.SetBoundary(hm.config.Boundary); err != nil {
		l.WarnContext(ctx, "Failed to set multipart boundary", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create client
	client := &httpClient{
		id:         fmt.Sprintf("http-%d", time.Now().UnixNano()),
		remoteAddr: remoteAddr,
		connected:  time.Now(),
		lastSeen:   time.Now(),
		writer:     writer,
		response:   w,
		request:    r,
	}

	hm.clients.Store(client.id, client)
	hm.clientCount.Add(1)

	defer func() {
		writer.Close()
		hm.clients.Delete(client.id)
		hm.clientCount.Add(-1)
		l.InfoContext(ctx, "HTTP streaming client disconnected", "remote", remoteAddr)
	}()

	// Wait for client to disconnect or server to stop
	select {
	case <-hm.stopCh:
		return
	case <-r.Context().Done():
		return
	}
}

// frameDistributor distributes video frames to connected HTTP clients.
func (hm *httpManager) frameDistributor(ctx context.Context, frameCh <-chan *VideoFrame) {
	defer close(hm.doneCh)

	l := log.GetGlobalLogger()

	for {
		select {
		case <-hm.stopCh:
			return
		case <-ctx.Done():
			return
		case frame := <-frameCh:
			if frame == nil {
				continue
			}

			// Convert frame to JPEG if needed
			var jpegData []byte
			var err error

			if frame.Format == kvm.PixelFormatYUYV {
				jpegData, err = ConvertYUYVToJPEG(frame.Data, frame.Width, frame.Height, hm.config.Quality)
				if err != nil {
					l.WarnContext(ctx, "Failed to convert frame to JPEG", "error", err)
					continue
				}
			} else if frame.Format == kvm.PixelFormatMJPEG {
				jpegData = frame.Data
			} else {
				// For other formats, try to convert to JPEG
				jpegData, err = ConvertYUYVToJPEG(frame.Data, frame.Width, frame.Height, hm.config.Quality)
				if err != nil {
					l.WarnContext(ctx, "Failed to encode frame as JPEG", "error", err)
					continue
				}
			}

			// Send to all clients
			hm.clients.Range(func(key, value interface{}) bool {
				if client, ok := value.(*httpClient); ok {
					if err := hm.sendFrameToClient(client, jpegData); err != nil {
						l.WarnContext(ctx, "Failed to send frame to HTTP client",
							"client", client.remoteAddr, "error", err)
						// Remove failed client
						hm.clients.Delete(key)
						hm.clientCount.Add(-1)
					}
				}
				return true
			})

			hm.framesSent.Add(1)
		}
	}
}

// sendFrameToClient sends a JPEG frame to a specific HTTP client.
func (hm *httpManager) sendFrameToClient(client *httpClient, jpegData []byte) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	// Create multipart header
	header := textproto.MIMEHeader{
		"Content-Type":   []string{"image/jpeg"},
		"Content-Length": []string{strconv.Itoa(len(jpegData))},
	}

	// Create part
	part, err := client.writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("failed to create multipart: %w", err)
	}

	// Write JPEG data
	if _, err := part.Write(jpegData); err != nil {
		return fmt.Errorf("failed to write JPEG data: %w", err)
	}

	// Flush the response
	if flusher, ok := client.response.(http.Flusher); ok {
		flusher.Flush()
	}

	hm.bytesTransmitted.Add(uint64(len(jpegData)))
	client.lastSeen = time.Now()

	return nil
}

// clientCleanup periodically cleans up stale client connections.
func (hm *httpManager) clientCleanup(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			hm.cleanupStaleClients(ctx)
		}
	}
}

// cleanupStaleClients removes clients that haven't been seen recently.
func (hm *httpManager) cleanupStaleClients(ctx context.Context) {
	l := log.GetGlobalLogger()
	now := time.Now()
	timeout := 5 * time.Minute // HTTP clients timeout after 5 minutes

	hm.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*httpClient); ok {
			if now.Sub(client.lastSeen) > timeout {
				l.InfoContext(ctx, "Cleaning up stale HTTP client",
					"client", client.remoteAddr, "idle", now.Sub(client.lastSeen))
				hm.clients.Delete(key)
				hm.clientCount.Add(-1)
			}
		}
		return true
	})
}
