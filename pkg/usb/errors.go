// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package usb

import "errors"

var (
	// ErrConfigFSNotMounted indicates that configfs is not mounted at /sys/kernel/config.
	ErrConfigFSNotMounted = errors.New("configfs not mounted")

	// ErrGadgetExists indicates that a USB gadget with the specified name already exists.
	ErrGadgetExists = errors.New("USB gadget already exists")

	// ErrGadgetNotFound indicates that the specified USB gadget could not be found.
	ErrGadgetNotFound = errors.New("USB gadget not found")

	// ErrPermissionDenied indicates insufficient permissions for USB operations.
	ErrPermissionDenied = errors.New("permission denied for USB operation")

	// ErrInvalidConfig indicates that the provided gadget configuration is invalid.
	ErrInvalidConfig = errors.New("invalid USB gadget configuration")

	// ErrHIDDeviceNotFound indicates that the specified HID device could not be found.
	ErrHIDDeviceNotFound = errors.New("HID device not found")

	// ErrHIDOperationFailed indicates that a HID operation failed.
	ErrHIDOperationFailed = errors.New("HID operation failed")

	// ErrMassStorageOperationFailed indicates that a mass storage operation failed.
	ErrMassStorageOperationFailed = errors.New("mass storage operation failed")

	// ErrUDCNotFound indicates that no USB Device Controller was found.
	ErrUDCNotFound = errors.New("USB Device Controller not found")

	// ErrGadgetBound indicates that the gadget is already bound to a UDC.
	ErrGadgetBound = errors.New("USB gadget already bound")

	// ErrGadgetNotBound indicates that the gadget is not bound to a UDC.
	ErrGadgetNotBound = errors.New("USB gadget not bound")

	// ErrInvalidHIDReport indicates that an invalid HID report was provided.
	ErrInvalidHIDReport = errors.New("invalid HID report")

	// ErrFileNotFound indicates that a required file was not found.
	ErrFileNotFound = errors.New("file not found")

	// ErrOperationTimeout indicates that an operation timed out.
	ErrOperationTimeout = errors.New("operation timed out")
)
