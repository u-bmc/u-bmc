// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package usb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SetMassStorageFile sets the backing file for the mass storage function.
func SetMassStorageFile(ctx context.Context, gadgetName, filePath string, cdromMode bool) error {
	if gadgetName == "" {
		return ErrInvalidConfig
	}

	gadgetDir := filepath.Join(gadgetPath, gadgetName)

	// Check if gadget exists
	if _, err := os.Stat(gadgetDir); os.IsNotExist(err) {
		return ErrGadgetNotFound
	}

	// Get mass storage function path
	functionDir := filepath.Join(gadgetDir, "functions/mass_storage.usb0/lun.0")
	if _, err := os.Stat(functionDir); os.IsNotExist(err) {
		return fmt.Errorf("mass storage function not found: %w", err)
	}

	// Set CD-ROM mode
	cdromPath := filepath.Join(functionDir, "cdrom")
	cdromValue := "0"
	if cdromMode {
		cdromValue = "1"
	}
	if err := writeFile(cdromPath, cdromValue); err != nil {
		return fmt.Errorf("failed to set cdrom mode: %w", err)
	}

	// Set the file path
	fileSysPath := filepath.Join(functionDir, "file")
	if err := writeFile(fileSysPath, filePath); err != nil {
		return fmt.Errorf("failed to set mass storage file: %w", err)
	}

	return nil
}

// GetMassStorageFile returns the current backing file for the mass storage function.
func GetMassStorageFile(ctx context.Context, gadgetName string) (string, bool, error) {
	if gadgetName == "" {
		return "", false, ErrInvalidConfig
	}

	gadgetDir := filepath.Join(gadgetPath, gadgetName)

	// Check if gadget exists
	if _, err := os.Stat(gadgetDir); os.IsNotExist(err) {
		return "", false, ErrGadgetNotFound
	}

	// Get mass storage function path
	functionDir := filepath.Join(gadgetDir, "functions/mass_storage.usb0/lun.0")
	if _, err := os.Stat(functionDir); os.IsNotExist(err) {
		return "", false, fmt.Errorf("mass storage function not found: %w", err)
	}

	// Get the file path
	fileSysPath := filepath.Join(functionDir, "file")
	filePath, err := readFile(fileSysPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read mass storage file: %w", err)
	}

	// Get CD-ROM mode
	cdromPath := filepath.Join(functionDir, "cdrom")
	cdromContent, err := readFile(cdromPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read cdrom mode: %w", err)
	}

	cdromMode := strings.TrimSpace(cdromContent) == "1"
	filePath = strings.TrimSpace(filePath)

	return filePath, cdromMode, nil
}

// SetMassStorageReadOnly sets the read-only flag for the mass storage function.
func SetMassStorageReadOnly(ctx context.Context, gadgetName string, readOnly bool) error {
	if gadgetName == "" {
		return ErrInvalidConfig
	}

	gadgetDir := filepath.Join(gadgetPath, gadgetName)

	// Check if gadget exists
	if _, err := os.Stat(gadgetDir); os.IsNotExist(err) {
		return ErrGadgetNotFound
	}

	// Get mass storage function path
	functionDir := filepath.Join(gadgetDir, "functions/mass_storage.usb0/lun.0")
	if _, err := os.Stat(functionDir); os.IsNotExist(err) {
		return fmt.Errorf("mass storage function not found: %w", err)
	}

	// Set read-only flag
	roPath := filepath.Join(functionDir, "ro")
	roValue := "0"
	if readOnly {
		roValue = "1"
	}
	if err := writeFile(roPath, roValue); err != nil {
		return fmt.Errorf("failed to set read-only flag: %w", err)
	}

	return nil
}

// GetMassStorageStatus returns the current status of the mass storage function.
func GetMassStorageStatus(ctx context.Context, gadgetName string) (*MassStorageStatus, error) {
	if gadgetName == "" {
		return nil, ErrInvalidConfig
	}

	gadgetDir := filepath.Join(gadgetPath, gadgetName)

	// Check if gadget exists
	if _, err := os.Stat(gadgetDir); os.IsNotExist(err) {
		return nil, ErrGadgetNotFound
	}

	// Get mass storage function path
	functionDir := filepath.Join(gadgetDir, "functions/mass_storage.usb0/lun.0")
	if _, err := os.Stat(functionDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("mass storage function not found: %w", err)
	}

	status := &MassStorageStatus{}

	// Get file path
	fileSysPath := filepath.Join(functionDir, "file")
	if filePath, err := readFile(fileSysPath); err == nil {
		status.FilePath = strings.TrimSpace(filePath)
	}

	// Get CD-ROM mode
	cdromPath := filepath.Join(functionDir, "cdrom")
	if cdromContent, err := readFile(cdromPath); err == nil {
		status.CDROMMode = strings.TrimSpace(cdromContent) == "1"
	}

	// Get read-only flag
	roPath := filepath.Join(functionDir, "ro")
	if roContent, err := readFile(roPath); err == nil {
		status.ReadOnly = strings.TrimSpace(roContent) == "1"
	}

	// Get removable flag
	removablePath := filepath.Join(functionDir, "removable")
	if removableContent, err := readFile(removablePath); err == nil {
		status.Removable = strings.TrimSpace(removableContent) == "1"
	}

	// Get inquiry string
	inquiryPath := filepath.Join(functionDir, "inquiry_string")
	if inquiryContent, err := readFile(inquiryPath); err == nil {
		status.InquiryString = strings.TrimSpace(inquiryContent)
	}

	return status, nil
}

// MassStorageStatus represents the current status of the mass storage function.
type MassStorageStatus struct {
	// FilePath is the current backing file path.
	FilePath string

	// CDROMMode indicates if the device appears as a CD-ROM.
	CDROMMode bool

	// ReadOnly indicates if the device is read-only.
	ReadOnly bool

	// Removable indicates if the device appears as removable.
	Removable bool

	// InquiryString is the SCSI inquiry string.
	InquiryString string
}

// createMassStorageFunction creates a mass storage function for the gadget.
func createMassStorageFunction(gadgetDir, configDir string) error {
	// Create base mass storage function
	functionDir := filepath.Join(gadgetDir, "functions/mass_storage.usb0")
	if err := os.MkdirAll(functionDir, 0755); err != nil {
		return fmt.Errorf("failed to create mass storage function directory: %w", err)
	}

	// Set mass storage attributes
	attrs := map[string]string{
		"stall": "1", // Enable stall responses
	}

	for attr, value := range attrs {
		attrPath := filepath.Join(functionDir, attr)
		if err := writeFile(attrPath, value); err != nil {
			return fmt.Errorf("failed to write mass storage %s: %w", attr, err)
		}
	}

	// Create LUN 0
	lunDir := filepath.Join(functionDir, "lun.0")
	if err := os.MkdirAll(lunDir, 0755); err != nil {
		return fmt.Errorf("failed to create mass storage LUN directory: %w", err)
	}

	// Set LUN attributes
	lunAttrs := map[string]string{
		"cdrom":          "1",                        // CD-ROM mode by default
		"ro":             "1",                        // Read-only by default
		"removable":      "1",                        // Removable media
		"file":           "",                         // No file initially
		"inquiry_string": "U-BMC   Virtual Media   ", // SCSI inquiry string
	}

	for attr, value := range lunAttrs {
		attrPath := filepath.Join(lunDir, attr)
		if err := writeFile(attrPath, value); err != nil {
			return fmt.Errorf("failed to write mass storage LUN %s: %w", attr, err)
		}
	}

	// Link function to configuration
	linkPath := filepath.Join(configDir, "mass_storage.usb0")
	if err := os.Symlink(functionDir, linkPath); err != nil {
		return fmt.Errorf("failed to link mass storage function to configuration: %w", err)
	}

	return nil
}
