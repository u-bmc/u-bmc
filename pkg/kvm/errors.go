// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package kvm

import "errors"

var (
	// ErrDeviceNotFound indicates that the specified video device could not be found.
	ErrDeviceNotFound = errors.New("video device not found")

	// ErrPermissionDenied indicates insufficient permissions for video device access.
	ErrPermissionDenied = errors.New("permission denied for video device")

	// ErrUnsupportedFormat indicates that the video format is not supported.
	ErrUnsupportedFormat = errors.New("unsupported video format")

	// ErrInvalidConfig indicates that the provided configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrDeviceBusy indicates that the video device is already in use.
	ErrDeviceBusy = errors.New("video device busy")

	// ErrStreamingFailed indicates that video streaming failed.
	ErrStreamingFailed = errors.New("video streaming failed")

	// ErrEncodingFailed indicates that frame encoding failed.
	ErrEncodingFailed = errors.New("frame encoding failed")

	// ErrVNCServerFailed indicates that the VNC server failed to start or operate.
	ErrVNCServerFailed = errors.New("VNC server failed")

	// ErrClientDisconnected indicates that a VNC client disconnected.
	ErrClientDisconnected = errors.New("VNC client disconnected")

	// ErrInvalidFrame indicates that an invalid frame was provided.
	ErrInvalidFrame = errors.New("invalid video frame")

	// ErrBufferOverflow indicates that a buffer overflow occurred.
	ErrBufferOverflow = errors.New("buffer overflow")

	// ErrTimeout indicates that an operation timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrClosed indicates that a resource has been closed.
	ErrClosed = errors.New("resource closed")

	// ErrNetworkFailed indicates that a network operation failed.
	ErrNetworkFailed = errors.New("network operation failed")

	// ErrInvalidPixelFormat indicates that an invalid pixel format was specified.
	ErrInvalidPixelFormat = errors.New("invalid pixel format")

	// ErrFrameDropped indicates that a video frame was dropped.
	ErrFrameDropped = errors.New("video frame dropped")
)
