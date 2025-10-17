// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

// Package usb provides USB gadget functionality for BMC environments.
//
// This package provides a simplified interface for managing USB gadgets via the Linux
// configfs subsystem, focusing on common BMC use cases like HID devices (keyboard, mouse)
// and mass storage emulation.
//
// # Design Philosophy
//
// Rather than creating complex state management, this package provides:
//   - Stateless functions for USB gadget configuration
//   - Simple interfaces for HID device operations
//   - Direct configfs manipulation with proper error handling
//   - Callback-based integration with other systems
//
// # Basic Usage
//
// For simple USB gadget setup:
//
//	// Initialize a USB gadget
//	gadget := &usb.GadgetConfig{
//		Name:         "kvm-gadget",
//		VendorID:     "0x1d6b",
//		ProductID:    "0x0104",
//		Manufacturer: "U-BMC",
//		Product:      "Virtual KVM",
//	}
//
//	err := usb.CreateGadget(ctx, gadget)
//	if err != nil {
//		log.Printf("Failed to create gadget: %v", err)
//	}
//
// # HID Device Operations
//
// For keyboard and mouse input:
//
//	// Send keyboard report
//	err := usb.SendKeyboardReport(ctx, "/dev/hidg0", modifier, keys)
//	if err != nil {
//		log.Printf("Keyboard report failed: %v", err)
//	}
//
//	// Send mouse report
//	err = usb.SendMouseReport(ctx, "/dev/hidg1", x, y, buttons)
//	if err != nil {
//		log.Printf("Mouse report failed: %v", err)
//	}
//
// # Mass Storage
//
// For virtual media:
//
//	// Mount an ISO image
//	err := usb.SetMassStorageFile(ctx, "/path/to/image.iso", true) // true for CD-ROM mode
//	if err != nil {
//		log.Printf("Failed to mount image: %v", err)
//	}
//
//	// Unmount
//	err = usb.SetMassStorageFile(ctx, "", false)
//
// # Configuration Management
//
// The package handles configfs operations transparently:
//
//	// Enable/disable gadget
//	err := usb.BindGadget(ctx, "kvm-gadget")
//	err = usb.UnbindGadget(ctx, "kvm-gadget")
//
//	// Get gadget status
//	status, err := usb.GetGadgetStatus(ctx, "kvm-gadget")
//
// # Error Handling
//
// The package provides specific error types for different failure scenarios:
//
//	err := usb.CreateGadget(ctx, gadget)
//	if err != nil {
//		switch {
//		case errors.Is(err, usb.ErrGadgetExists):
//			log.Info("Gadget already exists, continuing")
//		case errors.Is(err, usb.ErrConfigFSNotMounted):
//			log.Fatal("ConfigFS not available")
//		case errors.Is(err, usb.ErrPermissionDenied):
//			log.Fatal("Insufficient permissions for USB operations")
//		default:
//			log.Fatalf("Unexpected error: %v", err)
//		}
//	}
//
// # Platform Requirements
//
// This package requires:
//   - Linux with configfs support (CONFIG_CONFIGFS_FS)
//   - USB gadget support (CONFIG_USB_GADGET)
//   - HID gadget support (CONFIG_USB_G_HID)
//   - Mass storage gadget support (CONFIG_USB_MASS_STORAGE)
//   - Appropriate permissions for /sys/kernel/config and /dev/hidgX
package usb
