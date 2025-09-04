// SPDX-License-Identifier: BSD-3-Clause

//nolint:gosec
package i2c

import (
	"fmt"
	"time"
)

// I3C Common Command Codes (CCCs).
const (
	// Broadcast CCCs (sent to all devices).
	I3CCCC_ENEC_BC    = 0x00 // Enable Events Command (Broadcast)
	I3CCCC_DISEC_BC   = 0x01 // Disable Events Command (Broadcast)
	I3CCCC_ENTAS0_BC  = 0x02 // Enter Activity State 0 (Broadcast)
	I3CCCC_ENTAS1_BC  = 0x03 // Enter Activity State 1 (Broadcast)
	I3CCCC_ENTAS2_BC  = 0x04 // Enter Activity State 2 (Broadcast)
	I3CCCC_ENTAS3_BC  = 0x05 // Enter Activity State 3 (Broadcast)
	I3CCCC_RSTDAA_BC  = 0x06 // Reset Dynamic Address Assignment (Broadcast)
	I3CCCC_ENTDAA_BC  = 0x07 // Enter Dynamic Address Assignment (Broadcast)
	I3CCCC_DEFSLVS_BC = 0x08 // Define List of Slaves (Broadcast)
	I3CCCC_SETMWL_BC  = 0x09 // Set Max Write Length (Broadcast)
	I3CCCC_SETMRL_BC  = 0x0A // Set Max Read Length (Broadcast)
	I3CCCC_ENTTM_BC   = 0x0B // Enter Test Mode (Broadcast)
	I3CCCC_SETBUSCON  = 0x0C // Set Bus Context
	I3CCCC_ENDXFER    = 0x12 // Data Transfer Ending Procedure
	I3CCCC_ENTHDR0    = 0x20 // Enter HDR Mode 0
	I3CCCC_ENTHDR1    = 0x21 // Enter HDR Mode 1
	I3CCCC_ENTHDR2    = 0x22 // Enter HDR Mode 2
	I3CCCC_ENTHDR3    = 0x23 // Enter HDR Mode 3
	I3CCCC_ENTHDR4    = 0x24 // Enter HDR Mode 4
	I3CCCC_ENTHDR5    = 0x25 // Enter HDR Mode 5
	I3CCCC_ENTHDR6    = 0x26 // Enter HDR Mode 6
	I3CCCC_ENTHDR7    = 0x27 // Enter HDR Mode 7
	I3CCCC_SETXTIME   = 0x28 // Exchange Timing Information
	I3CCCC_SETAASA    = 0x29 // Set All Addresses to Static Addresses
	I3CCCC_RSTACT     = 0x2A // Target Reset Action
	I3CCCC_DEFGRPA    = 0x2B // Define List of Group Address
	I3CCCC_RSTGRPA    = 0x2C // Reset Group Address
	I3CCCC_MLANE      = 0x2D // Multi-Lane Data Transfer
	I3CCCC_SETDASA_BC = 0x87 // Set Dynamic Address from Static Address (Broadcast)

	// Direct CCCs (sent to specific device).
	I3CCCC_ENEC_DC     = 0x80 // Enable Events Command (Direct)
	I3CCCC_DISEC_DC    = 0x81 // Disable Events Command (Direct)
	I3CCCC_ENTAS0_DC   = 0x82 // Enter Activity State 0 (Direct)
	I3CCCC_ENTAS1_DC   = 0x83 // Enter Activity State 1 (Direct)
	I3CCCC_ENTAS2_DC   = 0x84 // Enter Activity State 2 (Direct)
	I3CCCC_ENTAS3_DC   = 0x85 // Enter Activity State 3 (Direct)
	I3CCCC_RSTDAA_DC   = 0x86 // Reset Dynamic Address Assignment (Direct)
	I3CCCC_SETDASA     = 0x87 // Set Dynamic Address from Static Address
	I3CCCC_SETNEWDA    = 0x88 // Set New Dynamic Address
	I3CCCC_SETMWL_DC   = 0x89 // Set Max Write Length (Direct)
	I3CCCC_SETMRL_DC   = 0x8A // Set Max Read Length (Direct)
	I3CCCC_GETMWL      = 0x8B // Get Max Write Length
	I3CCCC_GETMRL      = 0x8C // Get Max Read Length
	I3CCCC_GETPID      = 0x8D // Get Provisioned ID
	I3CCCC_GETBCR      = 0x8E // Get Bus Characteristics Register
	I3CCCC_GETDCR      = 0x8F // Get Device Characteristics Register
	I3CCCC_GETSTATUS   = 0x90 // Get Device Status
	I3CCCC_GETACCMST   = 0x91 // Get Accept Mastership
	I3CCCC_SETBRGTGT   = 0x93 // Set Bridge Targets
	I3CCCC_GETMXDS     = 0x94 // Get Max Data Speed
	I3CCCC_GETCAPS     = 0x95 // Get Optional Feature Capabilities
	I3CCCC_SETROUTE    = 0x96 // Set Route
	I3CCCC_D2DXFER     = 0x97 // Device to Device(s) Tunneling Control
	I3CCCC_SETXTIME_DC = 0x98 // Exchange Timing Information (Direct)
	I3CCCC_GETXTIME    = 0x99 // Get Exchange Timing Information
)

// I3C Event Enable/Disable bits.
const (
	I3C_EVENT_INTR     = 0x01 // In-Band Interrupt
	I3C_EVENT_MASTEREQ = 0x02 // Master Request
	I3C_EVENT_HOTJOIN  = 0x08 // Hot-Join
)

// I3C Bus Characteristics Register (BCR) bits.
const (
	I3C_BCR_DEVICE_ROLE     = 0x40 // 0=I3C Slave, 1=I3C Master
	I3C_BCR_ADVANCED_CAPS   = 0x20 // Advanced Capabilities
	I3C_BCR_VIRTUAL_TARGET  = 0x10 // Virtual Target
	I3C_BCR_OFFLINE_CAPABLE = 0x08 // Offline Capable
	I3C_BCR_IBI_PAYLOAD     = 0x04 // IBI Payload
	I3C_BCR_IBI_REQUEST     = 0x02 // IBI Request Capable
	I3C_BCR_MAX_DATA_SPEED  = 0x01 // Max Data Speed Limitation
)

// I3C Device Characteristics Register (DCR) values.
const (
	I3C_DCR_GENERIC       = 0x00 // Generic Device
	I3C_DCR_SENSOR        = 0x01 // Sensor
	I3C_DCR_DISPLAY       = 0x02 // Display
	I3C_DCR_INTERFACE     = 0x03 // Interface Device
	I3C_DCR_GPIO          = 0x04 // GPIO Expander
	I3C_DCR_ACTUATOR      = 0x05 // Actuator
	I3C_DCR_AUDIO         = 0x06 // Audio Device
	I3C_DCR_TIMING        = 0x07 // Timing Control
	I3C_DCR_COMMUNICATION = 0x08 // Communication
	I3C_DCR_MEMORY        = 0x09 // Memory
	I3C_DCR_POWER         = 0x0A // Power Management
	I3C_DCR_PROCESSING    = 0x0B // Processing Unit
)

// I3C speeds and timing.
const (
	I3C_SPEED_SDR_12_5MHZ = 12500000 // 12.5 MHz
	I3C_SPEED_SDR_25MHZ   = 25000000 // 25 MHz
	I3C_SPEED_HDR_DDR     = 25000000 // HDR-DDR mode
	I3C_SPEED_HDR_TSP     = 41700000 // HDR-TSP mode
	I3C_SPEED_HDR_TSL     = 41700000 // HDR-TSL mode
)

// I3CDeviceInfo represents information about an I3C device.
type I3CDeviceInfo struct {
	StaticAddress  uint8  // Static I2C address (if any)
	DynamicAddress uint8  // Assigned dynamic address
	PID            uint64 // 48-bit Provisioned ID
	BCR            uint8  // Bus Characteristics Register
	DCR            uint8  // Device Characteristics Register
	MaxWriteLength uint16 // Maximum write length
	MaxReadLength  uint16 // Maximum read length
	MaxDataSpeed   uint32 // Maximum data speed in Hz
	HasStaticAddr  bool   // Whether device has static address
	IBICapable     bool   // In-Band Interrupt capable
	MasterCapable  bool   // Master role capable
	HotJoinCapable bool   // Hot-Join capable
}

// I3CTransferMode represents different I3C transfer modes.
type I3CTransferMode int

const (
	// I3C_MODE_SDR uses Single Data Rate mode (standard I3C).
	I3C_MODE_SDR I3CTransferMode = iota
	// I3C_MODE_HDR_DDR uses High Data Rate Double Data Rate mode.
	I3C_MODE_HDR_DDR
	// I3C_MODE_HDR_TSP uses High Data Rate Ternary Symbol Pure mode.
	I3C_MODE_HDR_TSP
	// I3C_MODE_HDR_TSL uses High Data Rate Ternary Symbol Legacy mode.
	I3C_MODE_HDR_TSL
)

// EnterDynamicAddressAssignment initiates the Dynamic Address Assignment procedure.
// This allows devices without static addresses to receive dynamic addresses.
func (c *Conn) EnterDynamicAddressAssignment() ([]I3CDeviceInfo, error) {
	if c.config.Protocol != ProtocolI3C {
		return nil, ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return nil, ErrI3CNotSupported
	}

	// Send ENTDAA broadcast command
	if err := c.SendCommonCommandCode(I3CCCC_ENTDAA_BC, nil); err != nil {
		return nil, fmt.Errorf("%w: failed to send ENTDAA: %w", ErrI3CDynamicAddressFailed, err)
	}

	var devices []I3CDeviceInfo
	nextAddress := uint8(0x08) // Start from first available dynamic address

	// Continue DAA process until no more devices respond
	for nextAddress <= 0x7F {
		// Read device response (PID + BCR + DCR)
		response := make([]byte, 8)
		n, err := c.Read(response)
		if err != nil || n < 8 {
			break // No more devices
		}

		// Parse device information
		device := I3CDeviceInfo{
			DynamicAddress: nextAddress,
			PID:            uint64(response[0])<<40 | uint64(response[1])<<32 | uint64(response[2])<<24 | uint64(response[3])<<16 | uint64(response[4])<<8 | uint64(response[5]),
			BCR:            response[6],
			DCR:            response[7],
		}

		// Parse capabilities from BCR
		device.MasterCapable = (device.BCR & I3C_BCR_DEVICE_ROLE) != 0
		device.IBICapable = (device.BCR & I3C_BCR_IBI_REQUEST) != 0

		// Assign dynamic address to device using SETDASA broadcast CCC
		if err := c.SendCommonCommandCode(I3CCCC_SETDASA_BC, []byte{nextAddress}); err != nil {
			return nil, fmt.Errorf("%w: failed to assign address 0x%02x: %w", ErrI3CDynamicAddressFailed, nextAddress, err)
		}

		devices = append(devices, device)
		nextAddress++

		// Small delay between assignments
		time.Sleep(1 * time.Millisecond)
	}

	return devices, nil
}

// SendCommonCommandCode sends an I3C Common Command Code (CCC).
func (c *Conn) SendCommonCommandCode(ccc uint8, data []byte) error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}

	// TODO: Implement via Linux i3cdev (I3C_PRIV_XFER). Writing CCCs over I2C is invalid.
	return ErrI3CNotSupported
}

// GetDeviceCharacteristics reads the Bus and Device Characteristics Registers.
func (c *Conn) GetDeviceCharacteristics() (uint8, uint8, error) {
	if c.config.Protocol != ProtocolI3C {
		return 0, 0, ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return 0, 0, ErrI3CNotSupported
	}
	if !c.addrSet {
		return 0, 0, ErrInvalidAddress
	}

	// Send GETBCR direct CCC
	if err := c.SendCommonCommandCode(I3CCCC_GETBCR, nil); err != nil {
		return 0, 0, fmt.Errorf("%w: GETBCR failed: %w", ErrI3CCCCFailed, err)
	}

	// Read BCR response
	bcrData := make([]byte, 1)
	if _, err := c.Read(bcrData); err != nil {
		return 0, 0, fmt.Errorf("%w: failed to read BCR: %w", ErrI3CCCCFailed, err)
	}
	bcr := bcrData[0]

	// Send GETDCR direct CCC
	if err := c.SendCommonCommandCode(I3CCCC_GETDCR, nil); err != nil {
		return 0, 0, fmt.Errorf("%w: GETDCR failed: %w", ErrI3CCCCFailed, err)
	}

	// Read DCR response
	dcrData := make([]byte, 1)
	if _, err := c.Read(dcrData); err != nil {
		return 0, 0, fmt.Errorf("%w: failed to read DCR: %w", ErrI3CCCCFailed, err)
	}
	dcr := dcrData[0]

	return bcr, dcr, nil
}

// GetProvisionedID reads the 48-bit Provisioned ID of the device.
func (c *Conn) GetProvisionedID() (uint64, error) {
	if c.config.Protocol != ProtocolI3C {
		return 0, ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return 0, ErrI3CNotSupported
	}
	if !c.addrSet {
		return 0, ErrInvalidAddress
	}

	// Send GETPID direct CCC
	if err := c.SendCommonCommandCode(I3CCCC_GETPID, nil); err != nil {
		return 0, fmt.Errorf("%w: GETPID failed: %w", ErrI3CCCCFailed, err)
	}

	// Read 6-byte PID response
	pidData := make([]byte, 6)
	if _, err := c.Read(pidData); err != nil {
		return 0, fmt.Errorf("%w: failed to read PID: %w", ErrI3CCCCFailed, err)
	}

	// Assemble 48-bit PID
	pid := uint64(pidData[0])<<40 | uint64(pidData[1])<<32 | uint64(pidData[2])<<24 |
		uint64(pidData[3])<<16 | uint64(pidData[4])<<8 | uint64(pidData[5])

	return pid, nil
}

// EnableEvents enables specific I3C events for the device.
func (c *Conn) EnableEvents(events uint8) error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}
	if !c.addrSet {
		return ErrInvalidAddress
	}

	// Send ENEC direct CCC with event mask
	if err := c.SendCommonCommandCode(I3CCCC_ENEC_DC, []byte{events}); err != nil {
		return fmt.Errorf("%w: ENEC failed: %w", ErrI3CCCCFailed, err)
	}

	return nil
}

// DisableEvents disables specific I3C events for the device.
func (c *Conn) DisableEvents(events uint8) error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}
	if !c.addrSet {
		return ErrInvalidAddress
	}

	// Send DISEC direct CCC with event mask
	if err := c.SendCommonCommandCode(I3CCCC_DISEC_DC, []byte{events}); err != nil {
		return fmt.Errorf("%w: DISEC failed: %w", ErrI3CCCCFailed, err)
	}

	return nil
}

// SetMaxWriteLength sets the maximum write length for the device.
func (c *Conn) SetMaxWriteLength(length uint16) error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}
	if !c.addrSet {
		return ErrInvalidAddress
	}

	// Send SETMWL direct CCC
	data := []byte{uint8(length), uint8(length >> 8)}
	if err := c.SendCommonCommandCode(I3CCCC_SETMWL_DC, data); err != nil {
		return fmt.Errorf("%w: SETMWL failed: %w", ErrI3CCCCFailed, err)
	}

	return nil
}

// SetMaxReadLength sets the maximum read length for the device.
func (c *Conn) SetMaxReadLength(length uint16) error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}

	// Send SETMRL direct CCC
	data := []byte{uint8(c.currentAddr << 1), uint8(length), uint8(length >> 8)}
	if err := c.SendCommonCommandCode(I3CCCC_SETMRL_DC, data); err != nil {
		return fmt.Errorf("%w: SETMRL failed: %w", ErrI3CCCCFailed, err)
	}

	return nil
}

// GetMaxWriteLength gets the maximum write length for the device.
func (c *Conn) GetMaxWriteLength() (uint16, error) {
	if c.config.Protocol != ProtocolI3C {
		return 0, ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return 0, ErrI3CNotSupported
	}

	// Send GETMWL direct CCC
	if err := c.SendCommonCommandCode(I3CCCC_GETMWL, []byte{uint8(c.currentAddr << 1)}); err != nil {
		return 0, fmt.Errorf("%w: GETMWL failed: %w", ErrI3CCCCFailed, err)
	}

	// Read 2-byte response
	lengthData := make([]byte, 2)
	if _, err := c.Read(lengthData); err != nil {
		return 0, fmt.Errorf("%w: failed to read MWL: %w", ErrI3CCCCFailed, err)
	}

	return uint16(lengthData[0]) | uint16(lengthData[1])<<8, nil
}

// GetMaxReadLength gets the maximum read length for the device.
func (c *Conn) GetMaxReadLength() (uint16, error) {
	if c.config.Protocol != ProtocolI3C {
		return 0, ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return 0, ErrI3CNotSupported
	}

	// Send GETMRL direct CCC
	if err := c.SendCommonCommandCode(I3CCCC_GETMRL, []byte{uint8(c.currentAddr << 1)}); err != nil {
		return 0, fmt.Errorf("%w: GETMRL failed: %w", ErrI3CCCCFailed, err)
	}

	// Read 2-byte response
	lengthData := make([]byte, 2)
	if _, err := c.Read(lengthData); err != nil {
		return 0, fmt.Errorf("%w: failed to read MRL: %w", ErrI3CCCCFailed, err)
	}

	return uint16(lengthData[0]) | uint16(lengthData[1])<<8, nil
}

// ResetDynamicAddress resets the dynamic address assignment for the device.
func (c *Conn) ResetDynamicAddress() error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}
	if !c.addrSet {
		return ErrInvalidAddress
	}

	// Send RSTDAA direct CCC
	if err := c.SendCommonCommandCode(I3CCCC_RSTDAA_DC, nil); err != nil {
		return fmt.Errorf("%w: RSTDAA failed: %w", ErrI3CCCCFailed, err)
	}

	// Address is no longer valid
	c.addrSet = false
	return nil
}

// SetDynamicAddress sets a new dynamic address for the device.
func (c *Conn) SetDynamicAddress(newAddress uint8) error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}

	if !c.config.IsValidAddress(uint16(newAddress)) {
		return fmt.Errorf("%w: 0x%02x", ErrInvalidAddress, newAddress)
	}

	// Send SETNEWDA direct CCC
	data := []byte{uint8(c.currentAddr << 1), newAddress << 1}
	if err := c.SendCommonCommandCode(I3CCCC_SETNEWDA, data); err != nil {
		return fmt.Errorf("%w: SETNEWDA failed: %w", ErrI3CCCCFailed, err)
	}

	// Update current address
	c.currentAddr = uint16(newAddress)
	return nil
}

// EnableHotJoin enables Hot-Join capability for new devices.
func (c *Conn) EnableHotJoin() error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}

	// Enable Hot-Join events globally
	if err := c.SendCommonCommandCode(I3CCCC_ENEC_BC, []byte{I3C_EVENT_HOTJOIN}); err != nil {
		return fmt.Errorf("%w: failed to enable Hot-Join: %w", ErrI3CHotJoinFailed, err)
	}

	return nil
}

// EnterHDRMode switches the bus to High Data Rate mode.
func (c *Conn) EnterHDRMode(mode I3CTransferMode) error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}

	var ccc uint8
	switch mode {
	case I3C_MODE_HDR_DDR:
		ccc = I3CCCC_ENTHDR0
	case I3C_MODE_HDR_TSP:
		ccc = I3CCCC_ENTHDR1
	case I3C_MODE_HDR_TSL:
		ccc = I3CCCC_ENTHDR2
	default:
		return fmt.Errorf("%w: unsupported HDR mode", ErrInvalidData)
	}

	if err := c.SendCommonCommandCode(ccc, nil); err != nil {
		return fmt.Errorf("%w: failed to enter HDR mode: %w", ErrI3CCCCFailed, err)
	}

	return nil
}

// ExitHDRMode returns the bus to standard SDR mode.
func (c *Conn) ExitHDRMode() error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}

	// TODO: Implement correct HDR exit sequence per mode using i3cdev.
	return ErrI3CNotSupported
}

// ScanI3CBus scans for I3C devices and returns their information.
func (c *Conn) ScanI3CBus() ([]I3CDeviceInfo, error) {
	if c.config.Protocol != ProtocolI3C {
		return nil, ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return nil, ErrI3CNotSupported
	}

	var devices []I3CDeviceInfo

	// First, perform Dynamic Address Assignment to discover new devices
	newDevices, err := c.EnterDynamicAddressAssignment()
	if err != nil {
		return nil, fmt.Errorf("failed to perform DAA: %w", err)
	}

	devices = append(devices, newDevices...)

	// Scan for legacy I2C devices
	for addr := uint8(0x08); addr <= 0x77; addr++ {
		// Skip addresses already assigned to I3C devices
		skip := false
		for _, dev := range devices {
			if dev.StaticAddress == addr || dev.DynamicAddress == addr {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Try to communicate with potential I2C device
		oldAddr := c.currentAddr
		oldAddrSet := c.addrSet

		if err := c.SetAddress(uint16(addr)); err != nil {
			continue
		}

		// Try a simple read to test presence
		if _, err := c.ReadByte(); err == nil {
			// Found I2C device
			device := I3CDeviceInfo{
				StaticAddress:  addr,
				HasStaticAddr:  true,
				IBICapable:     false,
				MasterCapable:  false,
				HotJoinCapable: false,
			}
			devices = append(devices, device)
		}

		// Restore previous address
		if oldAddrSet {
			if err := c.SetAddress(oldAddr); err != nil {
				c.addrSet = false
			}
		} else {
			c.addrSet = false
		}
	}

	return devices, nil
}

// ValidateI3CDevice performs I3C-specific device validation.
func (c *Conn) ValidateI3CDevice() error {
	if c.config.Protocol != ProtocolI3C {
		return ErrProtocolViolation
	}

	if !c.supportsI3C() {
		return ErrI3CNotSupported
	}
	if !c.addrSet {
		return ErrInvalidAddress
	}

	// Try to read device characteristics as validation
	if _, _, err := c.GetDeviceCharacteristics(); err != nil {
		return fmt.Errorf("%w: failed to get device characteristics: %w", ErrDeviceNotResponding, err)
	}

	return nil
}

// GetI3CCapabilities returns I3C-specific capability information.
func (c *Conn) GetI3CCapabilities() ([]string, error) {
	if c.config.Protocol != ProtocolI3C {
		return nil, ErrProtocolViolation
	}

	var capabilities []string

	if c.supportsI3C() {
		capabilities = append(capabilities, "I3C Basic Operations")

		// Check if device responds to basic I3C commands
		if _, _, err := c.GetDeviceCharacteristics(); err == nil {
			capabilities = append(capabilities, "Device Characteristics")
		}

		if _, err := c.GetProvisionedID(); err == nil {
			capabilities = append(capabilities, "Provisioned ID")
		}

		if _, err := c.GetMaxWriteLength(); err == nil {
			capabilities = append(capabilities, "Max Length Configuration")
		}

		// Check for advanced features
		bcr, _, err := c.GetDeviceCharacteristics()
		if err == nil {
			if bcr&I3C_BCR_IBI_REQUEST != 0 {
				capabilities = append(capabilities, "In-Band Interrupts")
			}
			if bcr&I3C_BCR_DEVICE_ROLE != 0 {
				capabilities = append(capabilities, "Master Capability")
			}
			if bcr&I3C_BCR_ADVANCED_CAPS != 0 {
				capabilities = append(capabilities, "Advanced Capabilities")
			}
		}
	}

	return capabilities, nil
}

// supportsI3C checks if the adapter supports I3C operations.
// Note: This is a placeholder - actual I3C support detection would require
// kernel I3C subsystem integration.
func (c *Conn) supportsI3C() bool {
	// For now, assume I3C support if the protocol is set to I3C
	// In a real implementation, this would check kernel capabilities
	return c.config.Protocol == ProtocolI3C
}

// GetDeviceType returns a human-readable device type based on DCR.
func GetDeviceType(dcr uint8) string {
	switch dcr {
	case I3C_DCR_GENERIC:
		return "Generic Device"
	case I3C_DCR_SENSOR:
		return "Sensor"
	case I3C_DCR_DISPLAY:
		return "Display"
	case I3C_DCR_INTERFACE:
		return "Interface Device"
	case I3C_DCR_GPIO:
		return "GPIO Expander"
	case I3C_DCR_ACTUATOR:
		return "Actuator"
	case I3C_DCR_AUDIO:
		return "Audio Device"
	case I3C_DCR_TIMING:
		return "Timing Control"
	case I3C_DCR_COMMUNICATION:
		return "Communication"
	case I3C_DCR_MEMORY:
		return "Memory"
	case I3C_DCR_POWER:
		return "Power Management"
	case I3C_DCR_PROCESSING:
		return "Processing Unit"
	default:
		return fmt.Sprintf("Unknown (0x%02X)", dcr)
	}
}
