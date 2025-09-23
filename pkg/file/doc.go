// SPDX-License-Identifier: BSD-3-Clause

// Package file provides atomic file operations for safe and reliable file
// system interactions. The package focuses on atomic operations that ensure
// data integrity by using temporary files and atomic rename operations to
// prevent partial writes and data corruption.
//
// Atomic file operations are critical in system programming where data
// integrity is paramount. This package implements the common pattern of
// writing to a temporary file and then atomically renaming it to the target
// location, ensuring that other processes never see partially written files.
//
// # Core Operations
//
// The package provides two primary atomic operations:
//
//   - AtomicCreateFile: Creates a new file atomically, failing if the file
//     already exists. This is useful for creating lock files, configuration
//     files, or any scenario where you need to ensure exclusive creation.
//
//   - AtomicUpdateFile: Updates an existing file atomically by copying the
//     original content and appending new data. If the original file doesn't
//     exist, it creates a new file. This is useful for log files, configuration
//     updates, or any append-only scenarios.
//
// # Basic Usage
//
// Creating a new file atomically:
//
//	data := []byte("initial configuration data")
//	err := file.AtomicCreateFile("/etc/myapp/config.json", data, 0644)
//	if err != nil {
//		if errors.Is(err, os.ErrExist) {
//			log.Println("Configuration file already exists")
//		} else {
//			log.Fatalf("Failed to create config: %v", err)
//		}
//	}
//
// Updating a file atomically:
//
//	logEntry := []byte("2024-01-15 10:30:00 - Application started\n")
//	err := file.AtomicUpdateFile("/var/log/myapp.log", logEntry, 0644)
//	if err != nil {
//		log.Fatalf("Failed to update log: %v", err)
//	}
//
// # Lock File Pattern
//
// A common use case for atomic file creation is implementing lock files:
//
//	func acquireLock(lockPath string) error {
//		pidData := []byte(fmt.Sprintf("%d", os.Getpid()))
//		err := file.AtomicCreateFile(lockPath, pidData, 0644)
//		if err != nil {
//			if errors.Is(err, os.ErrExist) {
//				return fmt.Errorf("application is already running")
//			}
//			return fmt.Errorf("failed to acquire lock: %w", err)
//		}
//		return nil
//	}
//
//	func releaseLock(lockPath string) error {
//		return os.Remove(lockPath)
//	}
//
// # Configuration File Management
//
// For managing configuration files that need atomic updates:
//
//	type Config struct {
//		Database struct {
//			Host string `json:"host"`
//			Port int    `json:"port"`
//		} `json:"database"`
//		Features []string `json:"features"`
//	}
//
//	func saveConfig(cfg Config, path string) error {
//		data, err := json.MarshalIndent(cfg, "", "  ")
//		if err != nil {
//			return fmt.Errorf("failed to marshal config: %w", err)
//		}
//
//		// Add newline for better file formatting
//		data = append(data, '\n')
//
//		err = file.AtomicUpdateFile(path, data, 0644)
//		if err != nil {
//			return fmt.Errorf("failed to save config: %w", err)
//		}
//
//		return nil
//	}
//
// # Log File Rotation Helper
//
// Atomic updates are useful for log file operations:
//
//	func appendLog(message, logPath string) error {
//		timestamp := time.Now().Format("2006-01-02 15:04:05")
//		logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)
//
//		err := file.AtomicUpdateFile(logPath, []byte(logEntry), 0644)
//		if err != nil {
//			return fmt.Errorf("failed to append to log: %w", err)
//		}
//
//		return nil
//	}
//
//	func rotateLogs(currentPath, archivePath string) error {
//		// Read current log content
//		data, err := os.ReadFile(currentPath)
//		if err != nil && !os.IsNotExist(err) {
//			return fmt.Errorf("failed to read current log: %w", err)
//		}
//
//		if len(data) > 0 {
//			// Archive current log atomically
//			err = file.AtomicUpdateFile(archivePath, data, 0644)
//			if err != nil {
//				return fmt.Errorf("failed to archive log: %w", err)
//			}
//		}
//
//		// Clear current log by creating empty file
//		err = file.AtomicCreateFile(currentPath+".tmp", []byte{}, 0644)
//		if err != nil {
//			return fmt.Errorf("failed to create new log: %w", err)
//		}
//
//		return os.Rename(currentPath+".tmp", currentPath)
//	}
//
// # Error Handling
//
// The package provides specific error handling for common scenarios:
//
//	err := file.AtomicCreateFile(path, data, 0644)
//	if err != nil {
//		switch {
//		case errors.Is(err, os.ErrExist):
//			log.Printf("File already exists: %s", path)
//		case errors.Is(err, os.ErrPermission):
//			log.Printf("Permission denied: %s", path)
//		case errors.Is(err, file.ErrTemporaryFileCreation):
//			log.Printf("Failed to create temporary file: %v", err)
//		case errors.Is(err, file.ErrAtomicRename):
//			log.Printf("Failed to rename temporary file: %v", err)
//		default:
//			log.Printf("Unexpected error: %v", err)
//		}
//	}
//
// # Concurrent Safety
//
// The atomic operations provided by this package ensure atomic file replacement
// to prevent partial reads and writes. However, they do not serialize concurrent
// writers. The underlying filesystem operations ensure that:
//
//   - Temporary files are created with unique names to avoid conflicts
//   - Rename operations are atomic at the filesystem level
//   - Readers will never see partially written files
//
// However, concurrent writers may race against each other, resulting in
// last-write-wins behavior where earlier updates can be lost. For true
// concurrent safety, use external coordination (file locks, mutexes) or
// append-only operations with O_APPEND for scenarios requiring concurrent writes.
//
// Example of concurrent log writing (may lose updates):
//
//	func concurrentLogging(logPath string, workerID int, messages []string) {
//		for i, msg := range messages {
//			logEntry := fmt.Sprintf("Worker %d: Message %d: %s", workerID, i, msg)
//			err := appendLog(logEntry, logPath)
//			if err != nil {
//				log.Printf("Worker %d failed to log: %v", workerID, err)
//			}
//			time.Sleep(10 * time.Millisecond) // Simulate work
//		}
//	}
//
//	// Start multiple workers
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(id int) {
//			defer wg.Done()
//			messages := []string{"started", "processing", "completed"}
//			concurrentLogging("/var/log/workers.log", id, messages)
//		}(i)
//	}
//	wg.Wait()
//
// # Best Practices
//
// When using this package:
//
//   - Always check for os.ErrExist when using AtomicCreateFile if the file
//     might already exist
//   - Use appropriate file permissions (e.g., 0644 for regular files, 0600
//     for sensitive data)
//   - Consider disk space requirements as temporary files are created during
//     the operation
//   - Handle errors appropriately, especially in concurrent environments
//   - Clean up temporary files manually if operations are interrupted
//
// # Performance Considerations
//
// Atomic file operations have some performance implications:
//
//   - Extra disk I/O due to temporary file creation and rename operations
//   - Brief periods where both original and temporary files exist
//   - Potential for temporary file accumulation if operations are interrupted
//
// For high-frequency operations, consider batching updates or using alternative
// approaches like append-only files with periodic compaction.
package file
