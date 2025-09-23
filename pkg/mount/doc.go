// SPDX-License-Identifier: BSD-3-Clause

// Package mount provides functionality for setting up essential filesystem mounts
// in a Linux system environment. This package is specifically designed for system
// initialization scenarios where virtual filesystems and essential mount points
// need to be established for proper system operation.
//
// The package handles the mounting of critical virtual filesystems including proc,
// sysfs, devtmpfs, and various security and debugging filesystems with appropriate
// security flags and mount options. It ensures that these filesystems are mounted
// with proper security restrictions such as nodev, noexec, and nosuid flags where
// appropriate.
//
// # Core Functionality
//
// The package provides a single primary function `SetupMounts()` that configures
// a comprehensive set of essential filesystem mounts. The function is designed to
// be idempotent and safe to call multiple times, as it handles already-mounted
// filesystems gracefully.
//
// # Supported Filesystems
//
// The package sets up the following virtual filesystems:
//
//   - /proc: Process information pseudo-filesystem
//   - /sys: Sysfs device and kernel information
//   - /sys/kernel/security: Security framework interface
//   - /sys/kernel/debug: Kernel debugging interface
//   - /sys/kernel/tracing: Kernel tracing interface
//   - /sys/fs/cgroup: Control groups v2 filesystem
//   - /sys/fs/pstore: Persistent storage for crash dumps
//   - /sys/fs/bpf: BPF filesystem for BPF maps and programs
//   - /dev: Device files filesystem
//   - /dev/mqueue: POSIX message queues
//   - /dev/shm: Shared memory tmpfs
//   - /dev/pts: Pseudo-terminal slave filesystem
//   - /run: Runtime data tmpfs
//   - /tmp: Temporary files tmpfs
//
// # Basic Usage
//
// Setting up all essential mounts for system initialization:
//
//	func initializeSystem() error {
//		log.Println("Setting up essential filesystem mounts...")
//
//		if err := mount.SetupMounts(); err != nil {
//			return fmt.Errorf("failed to setup mounts: %w", err)
//		}
//
//		log.Println("All essential mounts configured successfully")
//		return nil
//	}
//
// # BMC System Initialization
//
// In a BMC (Baseboard Management Controller) context:
//
//	func initializeBMC() error {
//		// Set up essential mounts first
//		if err := mount.SetupMounts(); err != nil {
//			log.Fatalf("Failed to setup essential mounts: %v", err)
//			return err
//		}
//
//		// Verify critical mounts are available
//		if _, err := os.Stat("/proc/version"); err != nil {
//			return fmt.Errorf("proc filesystem not properly mounted: %w", err)
//		}
//
//		if _, err := os.Stat("/sys/class"); err != nil {
//			return fmt.Errorf("sysfs not properly mounted: %w", err)
//		}
//
//		log.Println("BMC filesystem environment ready")
//		return nil
//	}
//
// # Container and Embedded Systems
//
// For containerized or embedded BMC systems:
//
//	func setupContainerEnvironment() error {
//		// Check if running in a container
//		if _, err := os.Stat("/.dockerenv"); err == nil {
//			log.Println("Container environment detected")
//		}
//
//		// Setup mounts (will handle already-mounted filesystems)
//		if err := mount.SetupMounts(); err != nil {
//			// In containers, some mounts might fail due to restrictions
//			log.Printf("Mount setup completed with some restrictions: %v", err)
//
//			// Verify essential mounts are still available
//			essential := []string{"/proc", "/sys", "/dev"}
//			for _, path := range essential {
//				if _, err := os.Stat(path); err != nil {
//					return fmt.Errorf("essential mount missing: %s", path)
//				}
//			}
//		}
//
//		return nil
//	}
//
// # Error Handling and Recovery
//
// Handling mount failures with appropriate recovery:
//
//	func setupWithRetry() error {
//		maxRetries := 3
//		for attempt := 1; attempt <= maxRetries; attempt++ {
//			err := mount.SetupMounts()
//			if err == nil {
//				log.Println("All mounts setup successfully")
//				return nil
//			}
//
//			log.Printf("Mount setup attempt %d failed: %v", attempt, err)
//
//			// Check specific error types
//			switch {
//			case errors.Is(err, mount.ErrMountPointCreation):
//				log.Println("Failed to create mount points - checking permissions")
//			case errors.Is(err, mount.ErrMountFailed):
//				log.Println("Mount operations failed - checking system state")
//			case errors.Is(err, mount.ErrMountVerification):
//				log.Println("Mount verification failed - may be partially mounted")
//			}
//
//			if attempt < maxRetries {
//				time.Sleep(time.Duration(attempt) * time.Second)
//			}
//		}
//
//		return fmt.Errorf("failed to setup mounts after %d attempts", maxRetries)
//	}
//
// # Verification and Health Checks
//
// Verifying mount status for health monitoring:
//
//	func verifyMountHealth() error {
//		// Check if critical paths are accessible
//		criticalPaths := map[string]string{
//			"/proc/cpuinfo":           "proc",
//			"/sys/class/net":          "sysfs",
//			"/dev/null":               "devtmpfs",
//			"/run/systemd":            "tmpfs",
//		}
//
//		var errors []error
//		for path, fstype := range criticalPaths {
//			if _, err := os.Stat(path); err != nil {
//				errors = append(errors, fmt.Errorf("%s filesystem check failed for %s: %w", fstype, path, err))
//			}
//		}
//
//		if len(errors) > 0 {
//			return fmt.Errorf("mount health check failed: %v", errors)
//		}
//
//		return nil
//	}
//
// # Integration with System Services
//
// Using mounts setup in service initialization:
//
//	type BMCService struct {
//		name     string
//		mounted  bool
//	}
//
//	func (s *BMCService) Initialize() error {
//		// Ensure filesystem environment is ready
//		if !s.mounted {
//			if err := mount.SetupMounts(); err != nil {
//				return fmt.Errorf("service %s: mount setup failed: %w", s.name, err)
//			}
//			s.mounted = true
//		}
//
//		// Verify service-specific requirements
//		if err := s.verifyRequirements(); err != nil {
//			return fmt.Errorf("service %s: requirements check failed: %w", s.name, err)
//		}
//
//		log.Printf("Service %s: filesystem environment ready", s.name)
//		return nil
//	}
//
//	func (s *BMCService) verifyRequirements() error {
//		// Check for hardware monitoring capabilities
//		if _, err := os.Stat("/sys/class/hwmon"); err != nil {
//			return fmt.Errorf("hardware monitoring interface not available: %w", err)
//		}
//
//		// Check for GPIO access
//		if _, err := os.Stat("/sys/class/gpio"); err != nil {
//			log.Println("GPIO interface not available - some features may be limited")
//		}
//
//		return nil
//	}
//
// # Security Considerations
//
// The package implements several security best practices:
//
//   - Uses nodev flag to prevent device file creation on inappropriate filesystems
//   - Uses noexec flag to prevent execution of binaries from mount points
//   - Uses nosuid flag to ignore setuid/setgid bits on files
//   - Sets appropriate permissions for device and temporary filesystems
//   - Enables namespace delegation for cgroup2 where supported
//
// Example of additional security hardening:
//
//	func hardenMountSecurity() error {
//		// Setup base mounts first
//		if err := mount.SetupMounts(); err != nil {
//			return err
//		}
//
//		// Additional security: remount /tmp with stricter options
//		if err := syscall.Mount("", "/tmp", "", syscall.MS_REMOUNT|syscall.MS_NODEV|syscall.MS_NOSUID|syscall.MS_NOEXEC, ""); err != nil {
//			log.Printf("Warning: failed to harden /tmp mount: %v", err)
//		}
//
//		// Restrict /proc access if needed
//		if err := syscall.Mount("", "/proc", "", syscall.MS_REMOUNT|syscall.MS_HIDEPID, "hidepid=2"); err != nil {
//			log.Printf("Warning: failed to restrict /proc visibility: %v", err)
//		}
//
//		return nil
//	}
//
// # Platform Compatibility
//
// This package is Linux-specific and uses Linux-specific mount options and
// filesystem types. It includes build constraints to ensure it only compiles
// on Linux systems:
//
//	//go:build linux
//
// # Performance and Resource Usage
//
// The mount operations are generally fast as they involve kernel filesystem
// registration rather than large data transfers. However, consider:
//
//   - Mount operations require appropriate privileges (typically root)
//   - Some filesystems may consume kernel memory for metadata
//   - Verification involves reading /proc/mounts which may be slow with many mounts
//   - Failed mount attempts may leave partial state requiring cleanup
//
// # Troubleshooting Common Issues
//
// Common issues and their solutions:
//
//	func troubleshootMounts() {
//		// Check if running with sufficient privileges
//		if os.Geteuid() != 0 {
//			log.Println("Warning: not running as root - some mounts may fail")
//		}
//
//		// Check if mount points exist and are directories
//		mountPoints := []string{"/proc", "/sys", "/dev", "/run", "/tmp"}
//		for _, point := range mountPoints {
//			if info, err := os.Stat(point); err != nil {
//				log.Printf("Mount point %s does not exist: %v", point, err)
//			} else if !info.IsDir() {
//				log.Printf("Mount point %s is not a directory", point)
//			}
//		}
//
//		// Check for conflicting mounts
//		if data, err := os.ReadFile("/proc/mounts"); err == nil {
//			log.Printf("Current mounts:\n%s", string(data))
//		}
//	}
package mount
