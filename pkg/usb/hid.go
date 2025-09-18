// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package usb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HID report descriptors
var (
	// keyboardReportDescriptor is the standard USB HID keyboard report descriptor
	keyboardReportDescriptor = []byte{
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x06, // USAGE (Keyboard)
		0xa1, 0x01, // COLLECTION (Application)
		0x05, 0x07, //   USAGE_PAGE (Keyboard)
		0x19, 0xe0, //   USAGE_MINIMUM (Keyboard LeftControl)
		0x29, 0xe7, //   USAGE_MAXIMUM (Keyboard Right GUI)
		0x15, 0x00, //   LOGICAL_MINIMUM (0)
		0x25, 0x01, //   LOGICAL_MAXIMUM (1)
		0x75, 0x01, //   REPORT_SIZE (1)
		0x95, 0x08, //   REPORT_COUNT (8)
		0x81, 0x02, //   INPUT (Data,Var,Abs)
		0x95, 0x01, //   REPORT_COUNT (1)
		0x75, 0x08, //   REPORT_SIZE (8)
		0x81, 0x03, //   INPUT (Cnst,Var,Abs)
		0x95, 0x05, //   REPORT_COUNT (5)
		0x75, 0x01, //   REPORT_SIZE (1)
		0x05, 0x08, //   USAGE_PAGE (LEDs)
		0x19, 0x01, //   USAGE_MINIMUM (Num Lock)
		0x29, 0x05, //   USAGE_MAXIMUM (Kana)
		0x91, 0x02, //   OUTPUT (Data,Var,Abs)
		0x95, 0x01, //   REPORT_COUNT (1)
		0x75, 0x03, //   REPORT_SIZE (3)
		0x91, 0x03, //   OUTPUT (Cnst,Var,Abs)
		0x95, 0x06, //   REPORT_COUNT (6)
		0x75, 0x08, //   REPORT_SIZE (8)
		0x15, 0x00, //   LOGICAL_MINIMUM (0)
		0x25, 0x65, //   LOGICAL_MAXIMUM (101)
		0x05, 0x07, //   USAGE_PAGE (Keyboard)
		0x19, 0x00, //   USAGE_MINIMUM (Reserved)
		0x29, 0x65, //   USAGE_MAXIMUM (Keyboard Application)
		0x81, 0x00, //   INPUT (Data,Ary,Abs)
		0xc0, // END_COLLECTION
	}

	// mouseReportDescriptor is the USB HID absolute mouse report descriptor
	mouseReportDescriptor = []byte{
		0x05, 0x01, // Usage Page (Generic Desktop Ctrls)
		0x09, 0x02, // Usage (Mouse)
		0xA1, 0x01, // Collection (Application)
		0x85, 0x01, //     Report ID (1)
		0x09, 0x01, //     Usage (Pointer)
		0xA1, 0x00, //     Collection (Physical)
		0x05, 0x09, //         Usage Page (Button)
		0x19, 0x01, //         Usage Minimum (0x01)
		0x29, 0x03, //         Usage Maximum (0x03)
		0x15, 0x00, //         Logical Minimum (0)
		0x25, 0x01, //         Logical Maximum (1)
		0x75, 0x01, //         Report Size (1)
		0x95, 0x03, //         Report Count (3)
		0x81, 0x02, //         Input (Data, Var, Abs)
		0x95, 0x01, //         Report Count (1)
		0x75, 0x05, //         Report Size (5)
		0x81, 0x03, //         Input (Cnst, Var, Abs)
		0x05, 0x01, //         Usage Page (Generic Desktop Ctrls)
		0x09, 0x30, //         Usage (X)
		0x09, 0x31, //         Usage (Y)
		0x16, 0x00, 0x00, //   Logical Minimum (0)
		0x26, 0xFF, 0x7F, //   Logical Maximum (32767)
		0x36, 0x00, 0x00, //   Physical Minimum (0)
		0x46, 0xFF, 0x7F, //   Physical Maximum (32767)
		0x75, 0x10, //         Report Size (16)
		0x95, 0x02, //         Report Count (2)
		0x81, 0x02, //         Input (Data, Var, Abs)
		0xC0,       //     End Collection
		0x85, 0x02, //     Report ID (2)
		0x09, 0x38, //     Usage (Wheel)
		0x15, 0x81, //     Logical Minimum (-127)
		0x25, 0x7F, //     Logical Maximum (127)
		0x35, 0x00, //     Physical Minimum (0)
		0x45, 0x00, //     Physical Maximum (0)
		0x75, 0x08, //     Report Size (8)
		0x95, 0x01, //     Report Count (1)
		0x81, 0x06, //     Input (Data, Var, Rel)
		0xC0, // End Collection
	}
)

// KeyboardReport represents a USB HID keyboard report.
type KeyboardReport struct {
	Modifier byte    // Modifier keys (Ctrl, Shift, Alt, etc.)
	Reserved byte    // Reserved byte
	Keys     [6]byte // Key codes
}

// MouseReport represents a USB HID absolute mouse report.
type MouseReport struct {
	ReportID byte   // Report ID (1 for mouse movement)
	Buttons  byte   // Button states
	X        uint16 // X coordinate (0-32767)
	Y        uint16 // Y coordinate (0-32767)
}

// WheelReport represents a USB HID mouse wheel report.
type WheelReport struct {
	ReportID byte // Report ID (2 for wheel)
	Wheel    int8 // Wheel movement (-127 to 127)
}

// SendKeyboardReport sends a keyboard HID report to the specified device.
func SendKeyboardReport(ctx context.Context, devicePath string, modifier byte, keys []byte) error {
	if devicePath == "" {
		return ErrInvalidConfig
	}

	// Prepare the report
	report := KeyboardReport{
		Modifier: modifier,
		Reserved: 0,
	}

	// Copy keys, limiting to 6 keys maximum
	keyCount := len(keys)
	if keyCount > 6 {
		keyCount = 6
	}
	copy(report.Keys[:keyCount], keys[:keyCount])

	// Convert to byte slice
	reportBytes := []byte{
		report.Modifier,
		report.Reserved,
		report.Keys[0],
		report.Keys[1],
		report.Keys[2],
		report.Keys[3],
		report.Keys[4],
		report.Keys[5],
	}

	return writeHIDReport(devicePath, reportBytes)
}

// SendMouseReport sends an absolute mouse HID report to the specified device.
func SendMouseReport(ctx context.Context, devicePath string, x, y uint16, buttons byte) error {
	if devicePath == "" {
		return ErrInvalidConfig
	}

	report := MouseReport{
		ReportID: 1,
		Buttons:  buttons,
		X:        x,
		Y:        y,
	}

	// Convert to byte slice
	reportBytes := []byte{
		report.ReportID,
		report.Buttons,
		byte(report.X),      // X low byte
		byte(report.X >> 8), // X high byte
		byte(report.Y),      // Y low byte
		byte(report.Y >> 8), // Y high byte
	}

	return writeHIDReport(devicePath, reportBytes)
}

// SendWheelReport sends a mouse wheel HID report to the specified device.
func SendWheelReport(ctx context.Context, devicePath string, wheel int8) error {
	if devicePath == "" {
		return ErrInvalidConfig
	}

	// Skip sending if wheel is zero
	if wheel == 0 {
		return nil
	}

	report := WheelReport{
		ReportID: 2,
		Wheel:    wheel,
	}

	// Convert to byte slice
	reportBytes := []byte{
		report.ReportID,
		byte(report.Wheel),
	}

	return writeHIDReport(devicePath, reportBytes)
}

// createKeyboardFunction creates a HID keyboard function for the gadget.
func createKeyboardFunction(gadgetDir, configDir string) error {
	functionDir := filepath.Join(gadgetDir, "functions/hid.usb0")
	if err := os.MkdirAll(functionDir, 0755); err != nil {
		return fmt.Errorf("failed to create keyboard function directory: %w", err)
	}

	// Set keyboard attributes
	attrs := map[string]string{
		"protocol":        "1", // Keyboard
		"subclass":        "1", // Boot interface
		"report_length":   "8", // 8 bytes
		"no_out_endpoint": "0", // Enable output endpoint for LEDs
	}

	for attr, value := range attrs {
		attrPath := filepath.Join(functionDir, attr)
		if err := writeFile(attrPath, value); err != nil {
			return fmt.Errorf("failed to write keyboard %s: %w", attr, err)
		}
	}

	// Write report descriptor
	reportDescPath := filepath.Join(functionDir, "report_desc")
	if err := os.WriteFile(reportDescPath, keyboardReportDescriptor, 0644); err != nil {
		return fmt.Errorf("failed to write keyboard report descriptor: %w", err)
	}

	// Link function to configuration
	linkPath := filepath.Join(configDir, "hid.usb0")
	if err := os.Symlink(functionDir, linkPath); err != nil {
		return fmt.Errorf("failed to link keyboard function to configuration: %w", err)
	}

	return nil
}

// createMouseFunction creates a HID mouse function for the gadget.
func createMouseFunction(gadgetDir, configDir string) error {
	functionDir := filepath.Join(gadgetDir, "functions/hid.usb1")
	if err := os.MkdirAll(functionDir, 0755); err != nil {
		return fmt.Errorf("failed to create mouse function directory: %w", err)
	}

	// Set mouse attributes
	attrs := map[string]string{
		"protocol":        "2", // Mouse
		"subclass":        "0", // No subclass
		"report_length":   "6", // 6 bytes for absolute mouse
		"no_out_endpoint": "1", // No output endpoint needed
	}

	for attr, value := range attrs {
		attrPath := filepath.Join(functionDir, attr)
		if err := writeFile(attrPath, value); err != nil {
			return fmt.Errorf("failed to write mouse %s: %w", attr, err)
		}
	}

	// Write report descriptor
	reportDescPath := filepath.Join(functionDir, "report_desc")
	if err := os.WriteFile(reportDescPath, mouseReportDescriptor, 0644); err != nil {
		return fmt.Errorf("failed to write mouse report descriptor: %w", err)
	}

	// Link function to configuration
	linkPath := filepath.Join(configDir, "hid.usb1")
	if err := os.Symlink(functionDir, linkPath); err != nil {
		return fmt.Errorf("failed to link mouse function to configuration: %w", err)
	}

	return nil
}

// writeHIDReport writes a HID report to the specified device with timeout.
func writeHIDReport(devicePath string, report []byte) error {
	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		return ErrHIDDeviceNotFound
	}

	file, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("failed to open HID device: %w", err)
	}
	defer file.Close()

	// Set write deadline
	if err := file.SetWriteDeadline(time.Now().Add(10 * time.Millisecond)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if _, err := file.Write(report); err != nil {
		if os.IsTimeout(err) {
			return ErrOperationTimeout
		}
		return ErrHIDOperationFailed
	}

	return nil
}

// GetHIDKeyboardState reads the LED state from a keyboard HID device.
func GetHIDKeyboardState(ctx context.Context, devicePath string) (byte, error) {
	if devicePath == "" {
		return 0, ErrInvalidConfig
	}

	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		return 0, ErrHIDDeviceNotFound
	}

	file, err := os.OpenFile(devicePath, os.O_RDONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return 0, ErrPermissionDenied
		}
		return 0, fmt.Errorf("failed to open HID device: %w", err)
	}
	defer file.Close()

	// Set read deadline
	if err := file.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		return 0, fmt.Errorf("failed to set read deadline: %w", err)
	}

	buf := make([]byte, 1)
	if _, err := file.Read(buf); err != nil {
		if os.IsTimeout(err) {
			return 0, ErrOperationTimeout
		}
		return 0, ErrHIDOperationFailed
	}

	return buf[0], nil
}
