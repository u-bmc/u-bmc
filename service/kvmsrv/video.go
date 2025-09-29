// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/u-bmc/u-bmc/pkg/kvm"
	"github.com/u-bmc/u-bmc/pkg/log"
)

// videoCapture manages video capture from V4L2 devices.
type videoCapture struct {
	config  *kvm.VideoConfig
	frameCh chan *VideoFrame
	stopCh  chan struct{}
	doneCh  chan struct{}
	running atomic.Bool
	mu      sync.RWMutex

	// Statistics
	frameCount   atomic.Uint64
	droppedCount atomic.Uint64
	errorCount   atomic.Uint64

	// Current frame info
	currentWidth  atomic.Uint32
	currentHeight atomic.Uint32
	currentFormat atomic.Uint32
}

// VideoFrame represents a captured video frame.
type VideoFrame struct {
	Data      []byte
	Width     uint32
	Height    uint32
	Format    kvm.PixelFormat
	Timestamp time.Time
	Sequence  uint64
}

// newVideoCapture creates a new video capture manager.
func newVideoCapture(ctx context.Context, config *kvm.VideoConfig) (*videoCapture, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid video config: %w", err)
	}

	vc := &videoCapture{
		config:  config,
		frameCh: make(chan *VideoFrame, 10), // Buffer for frame processing
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}

	return vc, nil
}

// start begins video capture.
func (vc *videoCapture) start(ctx context.Context) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.running.Load() {
		return nil
	}

	// Initialize video capture
	if err := vc.initializeCapture(ctx); err != nil {
		return fmt.Errorf("failed to initialize video capture: %w", err)
	}

	vc.running.Store(true)

	// Start capture goroutine
	go vc.captureLoop(ctx)

	return nil
}

// stop stops video capture.
func (vc *videoCapture) stop(ctx context.Context) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if !vc.running.Load() {
		return nil
	}

	close(vc.stopCh)

	// Wait for capture loop to finish with timeout
	select {
	case <-vc.doneCh:
	case <-time.After(5 * time.Second):
		return ErrTimeout
	}

	vc.running.Store(false)
	close(vc.frameCh)

	return nil
}

// frames returns the channel for receiving video frames.
func (vc *videoCapture) frames() <-chan *VideoFrame {
	return vc.frameCh
}

// getStatus returns the current status of video capture.
func (vc *videoCapture) getStatus() *VideoCaptureStatus {
	return &VideoCaptureStatus{
		Device:  vc.config.Device,
		Width:   vc.currentWidth.Load(),
		Height:  vc.currentHeight.Load(),
		FPS:     vc.config.FPS,
		Frames:  vc.frameCount.Load(),
		Dropped: vc.droppedCount.Load(),
		Active:  vc.running.Load(),
	}
}

// initializeCapture initializes the video capture device.
func (vc *videoCapture) initializeCapture(ctx context.Context) error {
	l := log.GetGlobalLogger()

	// For now, simulate device initialization
	// In a real implementation, this would:
	// 1. Open the V4L2 device
	// 2. Query capabilities
	// 3. Set format and frame size
	// 4. Allocate buffers
	// 5. Start streaming

	l.InfoContext(ctx, "Initializing video capture device",
		"device", vc.config.Device,
		"width", vc.config.Width,
		"height", vc.config.Height,
		"format", vc.config.Format)

	// Set current format
	vc.currentWidth.Store(vc.config.Width)
	vc.currentHeight.Store(vc.config.Height)
	vc.currentFormat.Store(uint32(vc.config.Format))

	return nil
}

// captureLoop runs the main video capture loop.
func (vc *videoCapture) captureLoop(ctx context.Context) {
	defer close(vc.doneCh)

	l := log.GetGlobalLogger()
	l.InfoContext(ctx, "Starting video capture loop")

	ticker := time.NewTicker(time.Second / time.Duration(vc.config.FPS))
	defer ticker.Stop()

	sequence := uint64(0)

	for {
		select {
		case <-vc.stopCh:
			l.InfoContext(ctx, "Video capture loop stopping")
			return

		case <-ctx.Done():
			l.InfoContext(ctx, "Video capture loop cancelled")
			return

		case <-ticker.C:
			if err := vc.captureFrame(ctx, sequence); err != nil {
				vc.errorCount.Add(1)
				l.WarnContext(ctx, "Failed to capture frame", "error", err, "sequence", sequence)
				continue
			}
			sequence++
		}
	}
}

// captureFrame captures a single frame.
func (vc *videoCapture) captureFrame(ctx context.Context, sequence uint64) error {
	// For demonstration, create a simulated frame
	// In a real implementation, this would:
	// 1. Wait for buffer availability
	// 2. Dequeue buffer from driver
	// 3. Process frame data
	// 4. Queue buffer back to driver

	width := vc.currentWidth.Load()
	height := vc.currentHeight.Load()
	format := kvm.PixelFormat(vc.currentFormat.Load())

	// Create simulated frame data
	var frameSize int
	switch format {
	case kvm.PixelFormatYUYV:
		frameSize = int(width * height * 2) // 2 bytes per pixel for YUYV
	case kvm.PixelFormatMJPEG:
		frameSize = int(width * height / 2) // Estimated JPEG size
	default:
		frameSize = int(width * height * 3) // Default to RGB24
	}

	// Create simulated frame data with pattern
	frameData := make([]byte, frameSize)
	if format == kvm.PixelFormatYUYV {
		// Create a simple test pattern for YUYV
		vc.generateTestPattern(frameData, width, height, sequence)
	}

	frame := &VideoFrame{
		Data:      frameData,
		Width:     width,
		Height:    height,
		Format:    format,
		Timestamp: time.Now(),
		Sequence:  sequence,
	}

	// Try to send frame, drop if channel is full
	select {
	case vc.frameCh <- frame:
		vc.frameCount.Add(1)
	default:
		vc.droppedCount.Add(1)
	}

	return nil
}

// generateTestPattern generates a simple test pattern for YUYV format.
func (vc *videoCapture) generateTestPattern(data []byte, width, height uint32, sequence uint64) {
	// Create a simple moving pattern
	offset := int(sequence % uint64(width))

	for y := uint32(0); y < height; y++ {
		for x := uint32(0); x < width; x += 2 {
			idx := int((y*width + x) * 2)
			if idx+3 >= len(data) {
				break
			}

			// Create a vertical stripe pattern that moves
			stripe := (int(x) + offset) % 64
			var y1, y2, u, v byte

			if stripe < 32 {
				// White stripes
				y1, y2 = 235, 235 // White luminance
				u, v = 128, 128   // Neutral chrominance
			} else {
				// Black stripes
				y1, y2 = 16, 16 // Black luminance
				u, v = 128, 128 // Neutral chrominance
			}

			// Add some color variation based on position
			if y < height/3 {
				u = 240 // Blue tint
			} else if y < 2*height/3 {
				v = 240 // Red tint
			}

			// YUYV format: Y0 U Y1 V
			data[idx] = y1
			data[idx+1] = u
			data[idx+2] = y2
			data[idx+3] = v
		}
	}
}

// ConvertYUYVToJPEG converts YUYV frame data to JPEG.
func ConvertYUYVToJPEG(data []byte, width, height uint32, quality int) ([]byte, error) {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Convert YUYV to YCbCr image
	// 2. Encode as JPEG with specified quality
	// 3. Return JPEG data

	if len(data) < int(width*height*2) {
		return nil, ErrInvalidFrame
	}

	// For now, return simulated JPEG header + data
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG SOI + APP0 marker
	jpegData := make([]byte, len(jpegHeader)+len(data)/4)
	copy(jpegData, jpegHeader)

	// Simulate compression by taking every 4th byte
	for i := 0; i < len(data)/4 && i+len(jpegHeader) < len(jpegData); i++ {
		jpegData[len(jpegHeader)+i] = data[i*4]
	}

	return jpegData, nil
}

// ConvertYUYVToRGBA converts YUYV frame data to RGBA.
func ConvertYUYVToRGBA(data []byte, width, height uint32) ([]byte, error) {
	if len(data) < int(width*height*2) {
		return nil, ErrInvalidFrame
	}

	rgba := make([]byte, width*height*4)

	for y := uint32(0); y < height; y++ {
		for x := uint32(0); x < width; x += 2 {
			// YUYV format: Y0 U Y1 V
			srcIdx := int((y*width + x) * 2)
			if srcIdx+3 >= len(data) {
				break
			}

			y0 := float64(data[srcIdx])
			u := float64(data[srcIdx+1]) - 128
			y1 := float64(data[srcIdx+2])
			v := float64(data[srcIdx+3]) - 128

			// Convert YUV to RGB for first pixel
			r0 := y0 + 1.402*v
			g0 := y0 - 0.344*u - 0.714*v
			b0 := y0 + 1.772*u

			// Convert YUV to RGB for second pixel
			r1 := y1 + 1.402*v
			g1 := y1 - 0.344*u - 0.714*v
			b1 := y1 + 1.772*u

			// Clamp values and write to RGBA buffer
			dstIdx0 := int((y*width + x) * 4)
			dstIdx1 := int((y*width + x + 1) * 4)

			if dstIdx0+3 < len(rgba) {
				rgba[dstIdx0] = clampByte(r0)
				rgba[dstIdx0+1] = clampByte(g0)
				rgba[dstIdx0+2] = clampByte(b0)
				rgba[dstIdx0+3] = 255 // Alpha
			}

			if dstIdx1+3 < len(rgba) {
				rgba[dstIdx1] = clampByte(r1)
				rgba[dstIdx1+1] = clampByte(g1)
				rgba[dstIdx1+2] = clampByte(b1)
				rgba[dstIdx1+3] = 255 // Alpha
			}
		}
	}

	return rgba, nil
}

// clampByte clamps a float64 value to byte range.
func clampByte(val float64) byte {
	if val < 0 {
		return 0
	}
	if val > 255 {
		return 255
	}
	return byte(val)
}
