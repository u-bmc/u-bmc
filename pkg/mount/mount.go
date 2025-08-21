// SPDX-License-Identifier: BSD-3-Clause
//go:build linux
// +build linux

// Package mount provides functionality for setting up essential filesystem mounts
// in a Linux system, including proc, sysfs, and various virtual filesystems.
package mount

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

var (
	// ErrMountPointCreation indicates failure to create a mount point directory
	ErrMountPointCreation = errors.New("failed to create mount point")

	// ErrMountFailed indicates the mount operation failed
	ErrMountFailed = errors.New("mount operation failed")

	// ErrMountVerification indicates the mounted filesystem doesn't match expected specification
	ErrMountVerification = errors.New("mount verification failed")

	// ErrProcMountsRead indicates failure to read /proc/mounts
	ErrProcMountsRead = errors.New("failed to read /proc/mounts")
)

type mountSpec struct {
	source string
	target string
	fstype string
	flags  uintptr
	data   string
}

// SetupMounts configures essential filesystem mounts for a Linux system.
// It mounts proc, sysfs, and various virtual filesystems with appropriate
// security flags. After mounting /proc, it verifies that existing mounts
// match the expected specifications.
func SetupMounts() error {
	mounts := []mountSpec{
		{"proc", "/proc", "proc", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, ""},
		{"sysfs", "/sys", "sysfs", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, ""},
		{"securityfs", "/sys/kernel/security", "securityfs", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, ""},
		{"debugfs", "/sys/kernel/debug", "debugfs", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, ""},
		{"tracefs", "/sys/kernel/tracing", "tracefs", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, ""},
		{"cgroup2", "/sys/fs/cgroup", "cgroup2", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, "nsdelegate,memory_recursiveprot"},
		{"pstore", "/sys/fs/pstore", "pstore", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, ""},
		{"bpf", "/sys/fs/bpf", "bpf", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, "mode=700"},
		{"devtmpfs", "/dev", "devtmpfs", unix.MS_NOSUID, "mode=755"},
		{"mqueue", "/dev/mqueue", "mqueue", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, ""},
		{"tmpfs", "/dev/shm", "tmpfs", unix.MS_NODEV | unix.MS_NOSUID, ""},
		{"devpts", "/dev/pts", "devpts", unix.MS_NOEXEC | unix.MS_NOSUID, "gid=5,mode=620,ptmxmode=000"},
		{"tmpfs", "/run", "tmpfs", unix.MS_NODEV | unix.MS_NOSUID, "mode=755"},
		{"tmpfs", "/tmp", "tmpfs", unix.MS_NODEV | unix.MS_NOSUID, "mode=755"},
	}

	procMounted := false
	for i, m := range mounts {
		wasMounted, err := ensureMount(m)
		if err != nil {
			return fmt.Errorf("failed to mount %s: %w", m.target, err)
		}

		// After mounting /proc (if it wasn't already mounted), verify existing mounts
		if m.target == "/proc" && !wasMounted {
			procMounted = true
			if err := verifyMounts(mounts[:i+1]); err != nil {
				return fmt.Errorf("mount verification failed: %w", err)
			}
		}
	}

	// If /proc was already mounted, verify all mounts at the end
	if !procMounted {
		if err := verifyMounts(mounts); err != nil {
			return fmt.Errorf("mount verification failed: %w", err)
		}
	}

	return nil
}

// ensureMount creates the mount point if necessary and performs the mount operation.
// It returns true if the mount was already present (EBUSY), false if a new mount was created.
func ensureMount(m mountSpec) (bool, error) {
	if _, err := os.Stat(m.target); os.IsNotExist(err) {
		if err := os.MkdirAll(m.target, 0755); err != nil {
			return false, fmt.Errorf("%w %s: %v", ErrMountPointCreation, m.target, err)
		}
	}

	if err := unix.Mount(m.source, m.target, m.fstype, m.flags, m.data); err != nil {
		if err == unix.EBUSY {
			return true, nil // Mount already exists
		}
		return false, fmt.Errorf("%w: %v", ErrMountFailed, err)
	}

	return false, nil
}

// verifyMounts reads /proc/mounts and verifies that the specified mounts
// match the expected filesystem types.
func verifyMounts(specs []mountSpec) error {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProcMountsRead, err)
	}
	defer file.Close()

	// Build a map of expected mounts for quick lookup
	expected := make(map[string]string)
	for _, spec := range specs {
		expected[spec.target] = spec.fstype
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		mountPoint := fields[1]
		fsType := fields[2]

		if expectedFsType, exists := expected[mountPoint]; exists {
			if fsType != expectedFsType {
				return fmt.Errorf("%w: %s expected %s, got %s",
					ErrMountVerification, mountPoint, expectedFsType, fsType)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrProcMountsRead, err)
	}

	return nil
}
