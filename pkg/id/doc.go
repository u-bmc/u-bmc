// SPDX-License-Identifier: BSD-3-Clause

// Package id provides UUID-based identifier generation and management for
// persistent and ephemeral identification needs. The package wraps Google's
// UUID library with convenient functions for generating new identifiers and
// managing persistent identifiers that survive application restarts.
//
// This package is particularly useful for system services that need stable,
// unique identifiers for configuration management, device identification,
// session tracking, and other scenarios where consistent identification
// across restarts is required.
//
// # Core Functionality
//
// The package provides three main functions:
//
//   - NewID: Generates a new random UUID for one-time use
//   - GetOrCreatePersistentID: Retrieves an existing UUID from disk or creates
//     a new one if none exists, ensuring the same ID is returned on subsequent calls
//   - UpdatePersistentID: Generates a new UUID and updates the persistent storage,
//     useful for identifier rotation or reset scenarios
//
// # Basic Usage
//
// Generating a new ephemeral UUID:
//
//	sessionID := id.NewID()
//	log.Printf("New session: %s", sessionID)
//
// Creating or retrieving a persistent device identifier:
//
//	deviceID, err := id.GetOrCreatePersistentID("device.uuid", "/var/lib/myapp")
//	if err != nil {
//		log.Fatalf("Failed to get device ID: %v", err)
//	}
//	log.Printf("Device ID: %s", deviceID)
//
// # Persistent ID Management
//
// The persistent ID functions are designed for scenarios where you need
// stable identifiers across application restarts:
//
//	func initializeDevice() error {
//		// This will create /var/lib/bmc/device.uuid on first run
//		// and read the existing ID on subsequent runs
//		deviceID, err := id.GetOrCreatePersistentID("device.uuid", "/var/lib/bmc")
//		if err != nil {
//			return fmt.Errorf("failed to initialize device ID: %w", err)
//		}
//
//		log.Printf("BMC Device ID: %s", deviceID)
//
//		// Store the ID for use throughout the application
//		config.SetDeviceID(deviceID)
//		return nil
//	}
//
// # Multiple Persistent Identifiers
//
// Applications often need multiple types of persistent identifiers:
//
//	type SystemIdentifiers struct {
//		DeviceID    string
//		SessionID   string
//		InstanceID  string
//	}
//
//	func loadSystemIdentifiers(configDir string) (*SystemIdentifiers, error) {
//		deviceID, err := id.GetOrCreatePersistentID("device.uuid", configDir)
//		if err != nil {
//			return nil, fmt.Errorf("failed to get device ID: %w", err)
//		}
//
//		// Session ID changes on each restart
//		sessionID := id.NewID()
//
//		// Instance ID persists but can be rotated
//		instanceID, err := id.GetOrCreatePersistentID("instance.uuid", configDir)
//		if err != nil {
//			return nil, fmt.Errorf("failed to get instance ID: %w", err)
//		}
//
//		return &SystemIdentifiers{
//			DeviceID:   deviceID,
//			SessionID:  sessionID,
//			InstanceID: instanceID,
//		}, nil
//	}
//
// # Identifier Rotation
//
// For security or compliance reasons, you might need to rotate identifiers:
//
//	func rotateInstanceID(configDir string) error {
//		oldID, err := id.GetOrCreatePersistentID("instance.uuid", configDir)
//		if err != nil {
//			return fmt.Errorf("failed to get current instance ID: %w", err)
//		}
//
//		newID, err := id.UpdatePersistentID("instance.uuid", configDir)
//		if err != nil {
//			return fmt.Errorf("failed to rotate instance ID: %w", err)
//		}
//
//		log.Printf("Rotated instance ID from %s to %s", oldID, newID)
//		return nil
//	}
//
// # Configuration Management
//
// Integration with configuration systems:
//
//	type BMCConfig struct {
//		DeviceID     string `json:"device_id"`
//		Name         string `json:"name"`
//		Location     string `json:"location"`
//	}
//
//	func loadConfig(configPath, dataDir string) (*BMCConfig, error) {
//		config := &BMCConfig{}
//
//		// Load existing config if it exists
//		if data, err := os.ReadFile(configPath); err == nil {
//			if err := json.Unmarshal(data, config); err != nil {
//				return nil, fmt.Errorf("failed to parse config: %w", err)
//			}
//		}
//
//		// Ensure device ID is set
//		if config.DeviceID == "" {
//			deviceID, err := id.GetOrCreatePersistentID("device.uuid", dataDir)
//			if err != nil {
//				return nil, fmt.Errorf("failed to get device ID: %w", err)
//			}
//			config.DeviceID = deviceID
//
//			// Save updated config
//			if err := saveConfig(config, configPath); err != nil {
//				return nil, fmt.Errorf("failed to save config: %w", err)
//			}
//		}
//
//		return config, nil
//	}
//
// # Error Handling
//
// The package functions can fail for various filesystem-related reasons:
//
//	deviceID, err := id.GetOrCreatePersistentID("device.uuid", "/var/lib/bmc")
//	if err != nil {
//		switch {
//		case errors.Is(err, id.ErrDirectoryCreation):
//			log.Printf("Failed to create directory: %v", err)
//		case errors.Is(err, id.ErrFileCreation):
//			log.Printf("Failed to create ID file: %v", err)
//		case errors.Is(err, id.ErrFileRead):
//			log.Printf("Failed to read ID file: %v", err)
//		case errors.Is(err, id.ErrInvalidUUID):
//			log.Printf("Existing ID file contains invalid UUID: %v", err)
//		default:
//			log.Printf("Unexpected error: %v", err)
//		}
//		return err
//	}
//
// # Concurrent Access
//
// The persistent ID functions are safe for concurrent access from multiple
// goroutines within the same process. However, be aware that:
//
//   - Multiple processes accessing the same ID file may race during creation
//   - File locking is not implemented, so external coordination may be needed
//   - The underlying file.AtomicCreateFile ensures atomic creation operations
//
// Example of safe concurrent usage:
//
//	func startWorkers(numWorkers int) error {
//		var wg sync.WaitGroup
//		errors := make(chan error, numWorkers)
//
//		for i := 0; i < numWorkers; i++ {
//			wg.Add(1)
//			go func(workerID int) {
//				defer wg.Done()
//
//				// Each worker gets the same persistent device ID
//				deviceID, err := id.GetOrCreatePersistentID("device.uuid", "/var/lib/bmc")
//				if err != nil {
//					errors <- fmt.Errorf("worker %d failed to get device ID: %w", workerID, err)
//					return
//				}
//
//				// But each worker gets its own session ID
//				sessionID := id.NewID()
//
//				log.Printf("Worker %d: device=%s session=%s", workerID, deviceID, sessionID)
//			}(i)
//		}
//
//		wg.Wait()
//		close(errors)
//
//		for err := range errors {
//			return err
//		}
//
//		return nil
//	}
//
// # File Format and Storage
//
// Persistent IDs are stored as plain text files containing the UUID string.
// The files are created with standard permissions and can be read by any
// process with appropriate filesystem access:
//
//	$ cat /var/lib/bmc/device.uuid
//	a1b2c3d4-e5f6-7890-abcd-ef1234567890
//
// This simple format makes the IDs easily accessible from shell scripts,
// configuration management tools, and other applications.
//
// # Best Practices
//
// When using this package:
//
//   - Use descriptive filenames for different types of IDs (e.g., "device.uuid",
//     "instance.uuid", "session.uuid")
//   - Store persistent IDs in appropriate system directories (/var/lib for
//     system services, user config directories for user applications)
//   - Set proper directory permissions to control access to ID files
//   - Consider ID rotation policies for security-sensitive applications
//   - Document the purpose and lifecycle of each persistent identifier
//   - Use ephemeral IDs (NewID) for temporary identifiers that shouldn't persist
package id
