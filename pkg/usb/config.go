// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package usb

// GadgetConfig represents the configuration for a USB gadget.
type GadgetConfig struct {
	// Name is the unique name for the gadget (used in configfs path).
	Name string

	// VendorID is the USB vendor ID (e.g., "0x1d6b").
	VendorID string

	// ProductID is the USB product ID (e.g., "0x0104").
	ProductID string

	// SerialNumber is the device serial number.
	SerialNumber string

	// Manufacturer is the manufacturer string.
	Manufacturer string

	// Product is the product description string.
	Product string

	// MaxPower is the maximum power consumption in 2mA units (default: 250).
	MaxPower int

	// EnableKeyboard enables HID keyboard functionality.
	EnableKeyboard bool

	// EnableMouse enables HID mouse functionality.
	EnableMouse bool

	// EnableMassStorage enables mass storage functionality.
	EnableMassStorage bool
}

// HIDKeyboardConfig represents HID keyboard configuration.
type HIDKeyboardConfig struct {
	// DevicePath is the path to the HID device (e.g., "/dev/hidg0").
	DevicePath string

	// Protocol is the HID protocol (1 for keyboard).
	Protocol int

	// Subclass is the HID subclass (1 for boot interface).
	Subclass int

	// ReportLength is the length of HID reports in bytes.
	ReportLength int
}

// HIDMouseConfig represents HID mouse configuration.
type HIDMouseConfig struct {
	// DevicePath is the path to the HID device (e.g., "/dev/hidg1").
	DevicePath string

	// Protocol is the HID protocol (2 for mouse).
	Protocol int

	// Subclass is the HID subclass (1 for boot interface).
	Subclass int

	// ReportLength is the length of HID reports in bytes.
	ReportLength int

	// AbsoluteMode indicates if the mouse operates in absolute coordinate mode.
	AbsoluteMode bool
}

// MassStorageConfig represents mass storage configuration.
type MassStorageConfig struct {
	// LUN is the logical unit number (usually 0).
	LUN int

	// ReadOnly indicates if the storage should be read-only.
	ReadOnly bool

	// Removable indicates if the storage should appear as removable.
	Removable bool

	// CDROMMode indicates if the storage should appear as a CD-ROM.
	CDROMMode bool

	// FilePath is the path to the backing file or block device.
	FilePath string

	// InquiryString is the SCSI inquiry string shown to the host.
	InquiryString string
}

// GadgetStatus represents the current status of a USB gadget.
type GadgetStatus struct {
	// Name is the gadget name.
	Name string

	// Bound indicates if the gadget is bound to a UDC.
	Bound bool

	// UDC is the name of the USB Device Controller if bound.
	UDC string

	// State is the current USB state (e.g., "configured", "suspended").
	State string
}

// DefaultGadgetConfig returns a default USB gadget configuration for BMC KVM use.
func DefaultGadgetConfig() *GadgetConfig {
	return &GadgetConfig{
		Name:              "kvm-gadget",
		VendorID:          "0x1d6b", // Linux Foundation
		ProductID:         "0x0104", // Multifunction Composite Gadget
		SerialNumber:      "",
		Manufacturer:      "U-BMC",
		Product:           "Virtual KVM Device",
		MaxPower:          250, // 500mA
		EnableKeyboard:    true,
		EnableMouse:       true,
		EnableMassStorage: true,
	}
}

// DefaultKeyboardConfig returns a default HID keyboard configuration.
func DefaultKeyboardConfig() *HIDKeyboardConfig {
	return &HIDKeyboardConfig{
		DevicePath:   "/dev/hidg0",
		Protocol:     1, // Keyboard
		Subclass:     1, // Boot interface
		ReportLength: 8, // Standard keyboard report
	}
}

// DefaultMouseConfig returns a default HID mouse configuration.
func DefaultMouseConfig() *HIDMouseConfig {
	return &HIDMouseConfig{
		DevicePath:   "/dev/hidg1",
		Protocol:     2, // Mouse
		Subclass:     1, // Boot interface
		ReportLength: 6, // Absolute mouse report
		AbsoluteMode: true,
	}
}

// DefaultMassStorageConfig returns a default mass storage configuration.
func DefaultMassStorageConfig() *MassStorageConfig {
	return &MassStorageConfig{
		LUN:           0,
		ReadOnly:      true,
		Removable:     true,
		CDROMMode:     true,
		FilePath:      "",
		InquiryString: "U-BMC   Virtual Media   ",
	}
}
