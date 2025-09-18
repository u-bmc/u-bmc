// SPDX-License-Identifier: BSD-3-Clause

package kvmsrv

import "errors"

var (
	// ErrServiceNotConfigured indicates that the KVM service is not properly configured.
	ErrServiceNotConfigured = errors.New("KVM service not configured")

	// ErrVideoDeviceUnavailable indicates that the video capture device is not available.
	ErrVideoDeviceUnavailable = errors.New("video capture device unavailable")

	// ErrUSBGadgetInitFailed indicates that USB gadget initialization failed.
	ErrUSBGadgetInitFailed = errors.New("USB gadget initialization failed")

	// ErrVNCServerFailed indicates that the VNC server failed to start or operate.
	ErrVNCServerFailed = errors.New("VNC server failed")

	// ErrHTTPServerFailed indicates that the HTTP streaming server failed.
	ErrHTTPServerFailed = errors.New("HTTP streaming server failed")

	// ErrInvalidConfiguration indicates that the service configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid service configuration")

	// ErrResourceUnavailable indicates that a required resource is unavailable.
	ErrResourceUnavailable = errors.New("required resource unavailable")

	// ErrOperationFailed indicates that a KVM operation failed.
	ErrOperationFailed = errors.New("KVM operation failed")

	// ErrServiceShutdown indicates that the service is shutting down.
	ErrServiceShutdown = errors.New("KVM service shutting down")

	// ErrTimeout indicates that an operation timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrInvalidFrame indicates that an invalid video frame was provided.
	ErrInvalidFrame = errors.New("invalid video frame")
)
