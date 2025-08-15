// SPDX-License-Identifier: BSD-3-Clause

package mount

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

type mountSpec struct {
	source string
	target string
	fstype string
	flags  uintptr
	data   string
}

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

	for _, m := range mounts {
		if err := ensureMount(m); err != nil {
			return fmt.Errorf("failed to mount %s: %w", m.target, err)
		}
	}

	return nil
}

func ensureMount(m mountSpec) error {
	if _, err := os.Stat(m.target); os.IsNotExist(err) {
		if err := os.MkdirAll(m.target, 0755); err != nil {
			return fmt.Errorf("failed to create mount point %s: %w", m.target, err)
		}
	}

	if err := unix.Mount(m.source, m.target, m.fstype, m.flags, m.data); err != nil && err != unix.EBUSY {
		return fmt.Errorf("mount failed: %w", err)
	}

	return nil
}
