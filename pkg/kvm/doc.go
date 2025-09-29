// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

// Package kvm provides video capture and VNC server functionality for BMC environments.
//
// This package provides a simplified interface for capturing video from V4L2 devices
// and serving it via VNC protocol, focusing on common BMC KVM use cases.
//
// # Design Philosophy
//
// Rather than creating complex abstractions, this package provides:
//   - Simple video capture from V4L2 devices (/dev/videoN)
//   - Efficient frame encoding (YUYV to JPEG and RGBA)
//   - VNC server implementation with TCP and WebSocket support
//   - Stateless functions for integration with other systems
//   - Context-based cancellation and resource management
//
// # Basic Usage
//
// For simple video capture and VNC serving:
//
//	// Create video capture configuration
//	config := &kvm.VideoConfig{
//		Device: "/dev/video0",
//		Width:  640,
//		Height: 480,
//		Format: kvm.PixelFormatMJPEG,
//	}
//
//	// Start video capture
//	capture, err := kvm.NewVideoCapture(ctx, config)
//	if err != nil {
//		log.Printf("Failed to create video capture: %v", err)
//	}
//	defer capture.Close()
//
//	// Create VNC server
//	vncConfig := &kvm.VNCConfig{
//		Address: ":5900",
//		Width:   640,
//		Height:  480,
//	}
//
//	server, err := kvm.NewVNCServer(ctx, vncConfig)
//	if err != nil {
//		log.Printf("Failed to create VNC server: %v", err)
//	}
//
// # Video Capture
//
// For video frame capture:
//
//	// Get frames from capture
//	frameChan := capture.Frames()
//	for frame := range frameChan {
//		// Process frame
//		jpegData := frame.JPEG()
//		rgbaData := frame.RGBA()
//
//		// Send to VNC clients
//		server.SendFrame(rgbaData)
//	}
//
// # VNC Server
//
// For VNC server management:
//
//	// Start VNC server
//	err := server.Start(ctx)
//	if err != nil {
//		log.Printf("Failed to start VNC server: %v", err)
//	}
//
//	// Handle client connections
//	clients := server.Clients()
//	for client := range clients {
//		log.Printf("Client connected: %s", client.RemoteAddr())
//	}
//
// # Frame Processing
//
// For custom frame processing:
//
//	// Convert YUYV to JPEG
//	jpegData, err := kvm.EncodeJPEG(yuvData, width, height)
//	if err != nil {
//		log.Printf("JPEG encoding failed: %v", err)
//	}
//
//	// Convert YUYV to RGBA
//	rgbaData, err := kvm.ConvertYUYVToRGBA(yuvData, width, height)
//	if err != nil {
//		log.Printf("RGBA conversion failed: %v", err)
//	}
//
// # HTTP Streaming
//
// For HTTP MJPEG streaming:
//
//	// Create HTTP streamer
//	streamer := kvm.NewHTTPStreamer(capture)
//
//	// Serve MJPEG stream
//	http.HandleFunc("/stream", streamer.ServeHTTP)
//	http.ListenAndServe(":8080", nil)
//
// # Configuration Options
//
// The package provides various configuration options:
//
//	config := &kvm.VideoConfig{
//		Device:    "/dev/video0",
//		Width:     1920,
//		Height:    1080,
//		Format:    kvm.PixelFormatYUYV,
//		FPS:       30,
//		BufferCount: 4,
//	}
//
//	vncConfig := &kvm.VNCConfig{
//		Address:     ":5900",
//		Width:       1920,
//		Height:      1080,
//		PixelFormat: kvm.VNCPixelFormat32,
//		Encodings:   []kvm.VNCEncoding{kvm.EncodingRaw, kvm.EncodingRRE},
//	}
//
// # Error Handling
//
// The package provides specific error types for different failure scenarios:
//
//	err := kvm.NewVideoCapture(ctx, config)
//	if err != nil {
//		switch {
//		case errors.Is(err, kvm.ErrDeviceNotFound):
//			log.Fatal("Video device not available")
//		case errors.Is(err, kvm.ErrPermissionDenied):
//			log.Fatal("Insufficient permissions for video device")
//		case errors.Is(err, kvm.ErrUnsupportedFormat):
//			log.Fatal("Video format not supported")
//		default:
//			log.Fatalf("Unexpected error: %v", err)
//		}
//	}
//
// # Platform Requirements
//
// This package requires:
//   - Linux with V4L2 support
//   - Video capture device (/dev/videoN)
//   - Appropriate permissions for video device access
//   - Network access for VNC connections
package kvm
