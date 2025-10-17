// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/usb"
)

// usbManager manages USB gadget functionality.
type usbManager struct {
	config  *usb.GadgetConfig
	gadget  string
	running atomic.Bool
	mu      sync.RWMutex
	stopCh  chan struct{}
	doneCh  chan struct{}

	// HID device paths
	keyboardDevice string
	mouseDevice    string

	// Status tracking
	gadgetBound      atomic.Bool
	keyboardReady    atomic.Bool
	mouseReady       atomic.Bool
	massStorageReady atomic.Bool
}

// newUSBManager creates a new USB manager.
func newUSBManager(ctx context.Context, config *usb.GadgetConfig) (*usbManager, error) {
	if config == nil {
		return nil, ErrInvalidConfiguration
	}

	um := &usbManager{
		config:         config,
		gadget:         config.Name,
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
		keyboardDevice: "/dev/hidg0",
		mouseDevice:    "/dev/hidg1",
	}

	return um, nil
}

// start initializes and starts the USB gadget.
func (um *usbManager) start(ctx context.Context) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if um.running.Load() {
		return nil
	}

	l := log.GetGlobalLogger()
	l.InfoContext(ctx, "Starting USB gadget", "name", um.gadget)

	// Create USB gadget
	if err := um.createGadget(ctx); err != nil {
		return fmt.Errorf("failed to create USB gadget: %w", err)
	}

	// Bind gadget to UDC
	if err := um.bindGadget(ctx); err != nil {
		return fmt.Errorf("failed to bind USB gadget: %w", err)
	}

	um.running.Store(true)

	// Start monitoring goroutine
	go um.monitorLoop(ctx)

	l.InfoContext(ctx, "USB gadget started successfully")
	return nil
}

// stop stops and cleans up the USB gadget.
func (um *usbManager) stop(ctx context.Context) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if !um.running.Load() {
		return nil
	}

	l := log.GetGlobalLogger()
	l.InfoContext(ctx, "Stopping USB gadget")

	close(um.stopCh)

	// Wait for monitor loop to finish
	select {
	case <-um.doneCh:
	case <-time.After(5 * time.Second):
		l.WarnContext(ctx, "USB gadget monitor loop timeout")
	}

	// Unbind and destroy gadget
	if err := um.cleanupGadget(ctx); err != nil {
		l.WarnContext(ctx, "Failed to cleanup USB gadget", "error", err)
	}

	um.running.Store(false)
	return nil
}

// getStatus returns the current USB gadget status.
func (um *usbManager) getStatus() *USBGadgetStatus {
	status, err := usb.GetGadgetStatus(context.Background(), um.gadget)
	if err != nil {
		return &USBGadgetStatus{
			Name:        um.gadget,
			Bound:       false,
			Keyboard:    false,
			Mouse:       false,
			MassStorage: false,
		}
	}

	return &USBGadgetStatus{
		Name:        status.Name,
		Bound:       status.Bound,
		UDC:         status.UDC,
		State:       status.State,
		Keyboard:    um.keyboardReady.Load(),
		Mouse:       um.mouseReady.Load(),
		MassStorage: um.massStorageReady.Load(),
	}
}

// sendKeyboardInput sends keyboard input via HID.
func (um *usbManager) sendKeyboardInput(ctx context.Context, modifier byte, keys []byte) error {
	if !um.keyboardReady.Load() {
		return ErrResourceUnavailable
	}

	return usb.SendKeyboardReport(ctx, um.keyboardDevice, modifier, keys)
}

// sendMouseInput sends mouse input via HID.
func (um *usbManager) sendMouseInput(ctx context.Context, x, y uint16, buttons byte) error {
	if !um.mouseReady.Load() {
		return ErrResourceUnavailable
	}

	return usb.SendMouseReport(ctx, um.mouseDevice, x, y, buttons)
}

// sendWheelInput sends mouse wheel input via HID.
func (um *usbManager) sendWheelInput(ctx context.Context, wheel int8) error {
	if !um.mouseReady.Load() {
		return ErrResourceUnavailable
	}

	return usb.SendWheelReport(ctx, um.mouseDevice, wheel)
}

// setMassStorageFile sets the mass storage backing file.
func (um *usbManager) setMassStorageFile(ctx context.Context, filePath string, cdromMode bool) error {
	if !um.massStorageReady.Load() {
		return ErrResourceUnavailable
	}

	return usb.SetMassStorageFile(ctx, um.gadget, filePath, cdromMode)
}

// createGadget creates the USB gadget.
func (um *usbManager) createGadget(ctx context.Context) error {
	l := log.GetGlobalLogger()

	// Check if gadget already exists
	gadgets, err := usb.ListGadgets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list existing gadgets: %w", err)
	}

	for _, existing := range gadgets {
		if existing == um.gadget {
			l.InfoContext(ctx, "USB gadget already exists, destroying first")
			if err := usb.DestroyGadget(ctx, um.gadget); err != nil {
				l.WarnContext(ctx, "Failed to destroy existing gadget", "error", err)
			}
			break
		}
	}

	// Create new gadget
	if err := usb.CreateGadget(ctx, um.config); err != nil {
		return fmt.Errorf("failed to create gadget: %w", err)
	}

	l.InfoContext(ctx, "USB gadget created successfully")
	return nil
}

// bindGadget binds the USB gadget to a UDC.
func (um *usbManager) bindGadget(ctx context.Context) error {
	l := log.GetGlobalLogger()

	if err := usb.BindGadget(ctx, um.gadget); err != nil {
		return fmt.Errorf("failed to bind gadget: %w", err)
	}

	um.gadgetBound.Store(true)
	l.InfoContext(ctx, "USB gadget bound to UDC")

	// Wait for HID devices to become available
	if um.config.EnableKeyboard {
		if err := um.waitForHIDDevice(ctx, um.keyboardDevice); err != nil {
			l.WarnContext(ctx, "Keyboard HID device not ready", "error", err)
		} else {
			um.keyboardReady.Store(true)
			l.InfoContext(ctx, "Keyboard HID device ready")
		}
	}

	if um.config.EnableMouse {
		if err := um.waitForHIDDevice(ctx, um.mouseDevice); err != nil {
			l.WarnContext(ctx, "Mouse HID device not ready", "error", err)
		} else {
			um.mouseReady.Store(true)
			l.InfoContext(ctx, "Mouse HID device ready")
		}
	}

	if um.config.EnableMassStorage {
		um.massStorageReady.Store(true)
		l.InfoContext(ctx, "Mass storage device ready")
	}

	return nil
}

// waitForHIDDevice waits for a HID device to become available.
func (um *usbManager) waitForHIDDevice(ctx context.Context, devicePath string) error {
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return ErrTimeout
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Try to test the device by sending a null report
			if err := usb.SendKeyboardReport(ctx, devicePath, 0, []byte{0, 0, 0, 0, 0, 0}); err != nil {
				continue // Device not ready yet
			}
			return nil // Device is ready
		}
	}
}

// cleanupGadget unbinds and destroys the USB gadget.
func (um *usbManager) cleanupGadget(ctx context.Context) error {
	l := log.GetGlobalLogger()

	// Reset status flags
	um.gadgetBound.Store(false)
	um.keyboardReady.Store(false)
	um.mouseReady.Store(false)
	um.massStorageReady.Store(false)

	// Unbind gadget
	if err := usb.UnbindGadget(ctx, um.gadget); err != nil && !usb.IsNotBoundError(err) {
		l.WarnContext(ctx, "Failed to unbind USB gadget", "error", err)
	}

	// Destroy gadget
	if err := usb.DestroyGadget(ctx, um.gadget); err != nil {
		return fmt.Errorf("failed to destroy USB gadget: %w", err)
	}

	l.InfoContext(ctx, "USB gadget cleanup completed")
	return nil
}

// monitorLoop monitors the USB gadget status.
func (um *usbManager) monitorLoop(ctx context.Context) {
	defer close(um.doneCh)

	l := log.GetGlobalLogger()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-um.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			um.checkGadgetStatus(ctx, l)
		}
	}
}

// checkGadgetStatus checks and logs the current gadget status.
func (um *usbManager) checkGadgetStatus(ctx context.Context, l *slog.Logger) {
	status, err := usb.GetGadgetStatus(ctx, um.gadget)
	if err != nil {
		l.WarnContext(ctx, "Failed to get USB gadget status", "error", err)
		return
	}

	// Update bound status
	wasBound := um.gadgetBound.Load()
	um.gadgetBound.Store(status.Bound)

	// Log status changes
	if status.Bound != wasBound {
		if status.Bound {
			l.InfoContext(ctx, "USB gadget bound", "udc", status.UDC, "state", status.State)
		} else {
			l.WarnContext(ctx, "USB gadget unbound")
		}
	}
}
