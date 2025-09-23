// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package mount

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
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
		{"mqueue", "/dev/mqueue", "mqueue", unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID, "mode=1777"},
		{"tmpfs", "/dev/shm", "tmpfs", unix.MS_NODEV | unix.MS_NOSUID, "mode=1777"},
		{"devpts", "/dev/pts", "devpts", unix.MS_NOEXEC | unix.MS_NOSUID, "gid=5,mode=620,ptmxmode=0666"}, // allow non-root to open /dev/pts/ptmx; OK for single-tenant BMC
		{"tmpfs", "/run", "tmpfs", unix.MS_NODEV | unix.MS_NOSUID, "mode=755"},
		{"tmpfs", "/tmp", "tmpfs", unix.MS_NODEV | unix.MS_NOSUID, "mode=1777"},
	}

	var errs []error
	for _, m := range mounts {
		_, err := ensureMount(m)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to mount %s: %w", m.target, err))
			continue
		}
	}

	// Always perform final verification of all mounts
	if err := verifyMounts(mounts); err != nil {
		errs = append(errs, fmt.Errorf("mount verification failed: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// ensureMount creates the mount point if necessary and performs the mount operation.
// It returns true if the mount was already present (EBUSY), false if a new mount was created.
func ensureMount(m mountSpec) (bool, error) {
	if info, err := os.Stat(m.target); err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("%w %s: %w", ErrMountPointCreation, m.target, err)
		}
		if err := os.MkdirAll(m.target, 0o755); err != nil {
			return false, fmt.Errorf("%w %s: %w", ErrMountPointCreation, m.target, err)
		}
	} else if !info.IsDir() {
		return false, fmt.Errorf("%w %s: not a directory", ErrMountPointCreation, m.target)
	}

	if err := unix.Mount(m.source, m.target, m.fstype, m.flags, m.data); err != nil {
		if errors.Is(err, unix.EBUSY) {
			return true, nil // Mount already exists
		}
		return false, fmt.Errorf("%w %s (type=%s): %w", ErrMountFailed, m.target, m.fstype, err)
	}

	return false, nil
}

// verifyMounts reads /proc/mounts and verifies that the specified mounts
// match the expected filesystem types.
func verifyMounts(specs []mountSpec) error {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProcMountsRead, err)
	}
	defer file.Close()

	// Build a map of expected mounts for quick lookup
	expected := make(map[string]string)
	for _, spec := range specs {
		expected[spec.target] = spec.fstype
	}

	scanner := bufio.NewScanner(file)
	var errs []error
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		mountPoint := fields[1]
		fsType := fields[2]

		if expectedFsType, exists := expected[mountPoint]; exists {
			if fsType != expectedFsType {
				errs = append(errs, fmt.Errorf("%w: %s expected %s, got %s",
					ErrMountVerification, mountPoint, expectedFsType, fsType))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrProcMountsRead, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
