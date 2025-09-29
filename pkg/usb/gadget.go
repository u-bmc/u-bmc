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
	"syscall"
)

const (
	configfsPath   = "/sys/kernel/config"
	gadgetPath     = "/sys/kernel/config/usb_gadget"
	udcPath        = "/sys/class/udc"
	dwc3DriverPath = "/sys/bus/platform/drivers/dwc3"
)

// CreateGadget creates a new USB gadget with the specified configuration.
func CreateGadget(ctx context.Context, config *GadgetConfig) error {
	if config == nil {
		return ErrInvalidConfig
	}

	if config.Name == "" {
		return ErrInvalidConfig
	}

	// Check if configfs is mounted
	if err := ensureConfigFSMounted(); err != nil {
		return err
	}

	gadgetDir := filepath.Join(gadgetPath, config.Name)

	// Check if gadget already exists
	if _, err := os.Stat(gadgetDir); err == nil {
		return ErrGadgetExists
	}

	// Create gadget directory
	if err := os.MkdirAll(gadgetDir, 0755); err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("failed to create gadget directory: %w", err)
	}

	// Set basic gadget attributes
	if err := writeGadgetAttributes(gadgetDir, config); err != nil {
		os.RemoveAll(gadgetDir)
		return fmt.Errorf("failed to set gadget attributes: %w", err)
	}

	// Create string descriptors
	if err := createStringDescriptors(gadgetDir, config); err != nil {
		os.RemoveAll(gadgetDir)
		return fmt.Errorf("failed to create string descriptors: %w", err)
	}

	// Create configuration
	configDir := filepath.Join(gadgetDir, "configs/c.1")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		os.RemoveAll(gadgetDir)
		return fmt.Errorf("failed to create configuration directory: %w", err)
	}

	// Set configuration attributes
	if err := writeConfigAttributes(configDir, config); err != nil {
		os.RemoveAll(gadgetDir)
		return fmt.Errorf("failed to set configuration attributes: %w", err)
	}

	// Create configuration string descriptors
	if err := createConfigStringDescriptors(configDir); err != nil {
		os.RemoveAll(gadgetDir)
		return fmt.Errorf("failed to create configuration string descriptors: %w", err)
	}

	// Create functions based on configuration
	if err := createGadgetFunctions(gadgetDir, configDir, config); err != nil {
		os.RemoveAll(gadgetDir)
		return fmt.Errorf("failed to create gadget functions: %w", err)
	}

	return nil
}

// DestroyGadget removes a USB gadget.
func DestroyGadget(ctx context.Context, name string) error {
	if name == "" {
		return ErrInvalidConfig
	}

	gadgetDir := filepath.Join(gadgetPath, name)

	// Check if gadget exists
	if _, err := os.Stat(gadgetDir); os.IsNotExist(err) {
		return ErrGadgetNotFound
	}

	// Unbind gadget first
	if err := UnbindGadget(ctx, name); err != nil && !IsNotBoundError(err) {
		return fmt.Errorf("failed to unbind gadget: %w", err)
	}

	// Remove gadget directory
	if err := os.RemoveAll(gadgetDir); err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("failed to remove gadget directory: %w", err)
	}

	return nil
}

// BindGadget binds a USB gadget to an available UDC.
func BindGadget(ctx context.Context, name string) error {
	if name == "" {
		return ErrInvalidConfig
	}

	gadgetDir := filepath.Join(gadgetPath, name)

	// Check if gadget exists
	if _, err := os.Stat(gadgetDir); os.IsNotExist(err) {
		return ErrGadgetNotFound
	}

	// Find available UDC
	udc, err := findAvailableUDC()
	if err != nil {
		return err
	}

	// Bind to UDC
	udcFile := filepath.Join(gadgetDir, "UDC")
	if err := writeFile(udcFile, udc); err != nil {
		return fmt.Errorf("failed to bind gadget to UDC: %w", err)
	}

	return nil
}

// UnbindGadget unbinds a USB gadget from its UDC.
func UnbindGadget(ctx context.Context, name string) error {
	if name == "" {
		return ErrInvalidConfig
	}

	gadgetDir := filepath.Join(gadgetPath, name)

	// Check if gadget exists
	if _, err := os.Stat(gadgetDir); os.IsNotExist(err) {
		return ErrGadgetNotFound
	}

	// Unbind from UDC
	udcFile := filepath.Join(gadgetDir, "UDC")
	if err := writeFile(udcFile, ""); err != nil {
		return fmt.Errorf("failed to unbind gadget from UDC: %w", err)
	}

	return nil
}

// GetGadgetStatus returns the current status of a USB gadget.
func GetGadgetStatus(ctx context.Context, name string) (*GadgetStatus, error) {
	if name == "" {
		return nil, ErrInvalidConfig
	}

	gadgetDir := filepath.Join(gadgetPath, name)

	// Check if gadget exists
	if _, err := os.Stat(gadgetDir); os.IsNotExist(err) {
		return nil, ErrGadgetNotFound
	}

	status := &GadgetStatus{
		Name: name,
	}

	// Check if bound to UDC
	udcFile := filepath.Join(gadgetDir, "UDC")
	udcContent, err := readFile(udcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read UDC file: %w", err)
	}

	udc := strings.TrimSpace(udcContent)
	if udc != "" {
		status.Bound = true
		status.UDC = udc

		// Get USB state
		statePath := filepath.Join(udcPath, udc, "state")
		if stateContent, err := readFile(statePath); err == nil {
			status.State = strings.TrimSpace(stateContent)
		}
	}

	return status, nil
}

// ListGadgets returns a list of all USB gadgets.
func ListGadgets(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(gadgetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigFSNotMounted
		}
		return nil, fmt.Errorf("failed to list gadgets: %w", err)
	}

	var gadgets []string
	for _, entry := range entries {
		if entry.IsDir() {
			gadgets = append(gadgets, entry.Name())
		}
	}

	return gadgets, nil
}

// Helper functions

func ensureConfigFSMounted() error {
	if _, err := os.Stat(configfsPath); os.IsNotExist(err) {
		return ErrConfigFSNotMounted
	}

	if _, err := os.Stat(gadgetPath); os.IsNotExist(err) {
		return ErrConfigFSNotMounted
	}

	return nil
}

func writeGadgetAttributes(gadgetDir string, config *GadgetConfig) error {
	attrs := map[string]string{
		"bcdUSB":    "0x0200", // USB 2.0
		"idVendor":  config.VendorID,
		"idProduct": config.ProductID,
		"bcdDevice": "0x0100", // Device release 1.00
	}

	for attr, value := range attrs {
		attrPath := filepath.Join(gadgetDir, attr)
		if err := writeFile(attrPath, value); err != nil {
			return fmt.Errorf("failed to write %s: %w", attr, err)
		}
	}

	return nil
}

func createStringDescriptors(gadgetDir string, config *GadgetConfig) error {
	stringsDir := filepath.Join(gadgetDir, "strings/0x409")
	if err := os.MkdirAll(stringsDir, 0755); err != nil {
		return fmt.Errorf("failed to create strings directory: %w", err)
	}

	strings := map[string]string{
		"serialnumber": config.SerialNumber,
		"manufacturer": config.Manufacturer,
		"product":      config.Product,
	}

	for str, value := range strings {
		strPath := filepath.Join(stringsDir, str)
		if err := writeFile(strPath, value); err != nil {
			return fmt.Errorf("failed to write %s: %w", str, err)
		}
	}

	return nil
}

func writeConfigAttributes(configDir string, config *GadgetConfig) error {
	maxPower := config.MaxPower
	if maxPower == 0 {
		maxPower = 250 // Default 500mA
	}

	attrs := map[string]string{
		"MaxPower": fmt.Sprintf("%d", maxPower),
	}

	for attr, value := range attrs {
		attrPath := filepath.Join(configDir, attr)
		if err := writeFile(attrPath, value); err != nil {
			return fmt.Errorf("failed to write %s: %w", attr, err)
		}
	}

	return nil
}

func createConfigStringDescriptors(configDir string) error {
	stringsDir := filepath.Join(configDir, "strings/0x409")
	if err := os.MkdirAll(stringsDir, 0755); err != nil {
		return fmt.Errorf("failed to create config strings directory: %w", err)
	}

	configPath := filepath.Join(stringsDir, "configuration")
	if err := writeFile(configPath, "Config 1: KVM"); err != nil {
		return fmt.Errorf("failed to write configuration string: %w", err)
	}

	return nil
}

func createGadgetFunctions(gadgetDir, configDir string, config *GadgetConfig) error {
	if config.EnableKeyboard {
		if err := createKeyboardFunction(gadgetDir, configDir); err != nil {
			return fmt.Errorf("failed to create keyboard function: %w", err)
		}
	}

	if config.EnableMouse {
		if err := createMouseFunction(gadgetDir, configDir); err != nil {
			return fmt.Errorf("failed to create mouse function: %w", err)
		}
	}

	if config.EnableMassStorage {
		if err := createMassStorageFunction(gadgetDir, configDir); err != nil {
			return fmt.Errorf("failed to create mass storage function: %w", err)
		}
	}

	return nil
}

func findAvailableUDC() (string, error) {
	entries, err := os.ReadDir(udcPath)
	if err != nil {
		return "", ErrUDCNotFound
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		udcName := entry.Name()
		statePath := filepath.Join(udcPath, udcName, "state")

		if stateContent, err := readFile(statePath); err == nil {
			state := strings.TrimSpace(stateContent)
			if state == "not attached" {
				return udcName, nil
			}
		}
	}

	return "", ErrUDCNotFound
}

func writeFile(path, content string) error {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err == syscall.ENOENT {
				return ErrFileNotFound
			}
		}
	}
	return err
}

func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrFileNotFound
		}
		if os.IsPermission(err) {
			return "", ErrPermissionDenied
		}
		return "", err
	}
	return string(content), nil
}

// IsNotBoundError returns true if the error indicates the gadget is not bound.
func IsNotBoundError(err error) bool {
	return err == ErrGadgetNotBound
}
