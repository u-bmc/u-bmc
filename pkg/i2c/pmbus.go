// SPDX-License-Identifier: BSD-3-Clause

//nolint:gosec
package i2c

import (
	"fmt"
	"math"
)

// PMBus command constants.
const (
	// Standard PMBus commands.
	PMBusPageCommand         = 0x00 // PAGE
	PMBusOperation           = 0x01 // OPERATION
	PMBusOnOffConfig         = 0x02 // ON_OFF_CONFIG
	PMBusClearFaults         = 0x03 // CLEAR_FAULTS
	PMBusPhase               = 0x04 // PHASE
	PMBusPagePlusWrite       = 0x05 // PAGE_PLUS_WRITE
	PMBusPagePlusRead        = 0x06 // PAGE_PLUS_READ
	PMBusZoneConfig          = 0x07 // ZONE_CONFIG
	PMBusZoneActive          = 0x08 // ZONE_ACTIVE
	PMBusWriteProtect        = 0x10 // WRITE_PROTECT
	PMBusStoreDefaultAll     = 0x11 // STORE_DEFAULT_ALL
	PMBusRestoreDefaultAll   = 0x12 // RESTORE_DEFAULT_ALL
	PMBusStoreDefaultCode    = 0x13 // STORE_DEFAULT_CODE
	PMBusRestoreDefaultCode  = 0x14 // RESTORE_DEFAULT_CODE
	PMBusStoreUserAll        = 0x15 // STORE_USER_ALL
	PMBusRestoreUserAll      = 0x16 // RESTORE_USER_ALL
	PMBusStoreUserCode       = 0x17 // STORE_USER_CODE
	PMBusRestoreUserCode     = 0x18 // RESTORE_USER_CODE
	PMBusCapability          = 0x19 // CAPABILITY
	PMBusQueryCommand        = 0x1A // QUERY
	PMBusSMBalertMask        = 0x1B // SMBALERT_MASK
	PMBusVoutMode            = 0x20 // VOUT_MODE
	PMBusVoutCommand         = 0x21 // VOUT_COMMAND
	PMBusVoutTrimCommand     = 0x22 // VOUT_TRIM
	PMBusVoutCalOffset       = 0x23 // VOUT_CAL_OFFSET
	PMBusVoutMax             = 0x24 // VOUT_MAX
	PMBusVoutMarginHigh      = 0x25 // VOUT_MARGIN_HIGH
	PMBusVoutMarginLow       = 0x26 // VOUT_MARGIN_LOW
	PMBusVoutTransitionRate  = 0x27 // VOUT_TRANSITION_RATE
	PMBusVoutDroopCommand    = 0x28 // VOUT_DROOP
	PMBusVoutScaleLoop       = 0x29 // VOUT_SCALE_LOOP
	PMBusVoutScaleMonitor    = 0x2A // VOUT_SCALE_MONITOR
	PMBusCoefficientsCmd     = 0x30 // COEFFICIENTS
	PMBusPoutMax             = 0x31 // POUT_MAX
	PMBusMaxDuty             = 0x32 // MAX_DUTY
	PMBusFrequencySwitch     = 0x33 // FREQUENCY_SWITCH
	PMBusPowerMode           = 0x34 // POWER_MODE
	PMBusVinOn               = 0x35 // VIN_ON
	PMBusVinOff              = 0x36 // VIN_OFF
	PMBusIoutCalGain         = 0x38 // IOUT_CAL_GAIN
	PMBusIoutCalOffset       = 0x39 // IOUT_CAL_OFFSET
	PMBusFanConfig12         = 0x3A // FAN_CONFIG_1_2
	PMBusFanCommand1         = 0x3B // FAN_COMMAND_1
	PMBusFanCommand2         = 0x3C // FAN_COMMAND_2
	PMBusFanConfig34         = 0x3D // FAN_CONFIG_3_4
	PMBusFanCommand3         = 0x3E // FAN_COMMAND_3
	PMBusFanCommand4         = 0x3F // FAN_COMMAND_4
	PMBusVoutOVFaultLimit    = 0x40 // VOUT_OV_FAULT_LIMIT
	PMBusVoutOVFaultResponse = 0x41 // VOUT_OV_FAULT_RESPONSE
	PMBusVoutOVWarnLimit     = 0x42 // VOUT_OV_WARN_LIMIT
	PMBusVoutUVWarnLimit     = 0x43 // VOUT_UV_WARN_LIMIT
	PMBusVoutUVFaultLimit    = 0x44 // VOUT_UV_FAULT_LIMIT
	PMBusVoutUVFaultResponse = 0x45 // VOUT_UV_FAULT_RESPONSE
	PMBusIoutOCFaultLimit    = 0x46 // IOUT_OC_FAULT_LIMIT
	PMBusIoutOCFaultResponse = 0x47 // IOUT_OC_FAULT_RESPONSE
	PMBusIoutOCWarnLimit     = 0x4A // IOUT_OC_WARN_LIMIT
	PMBusOTFaultLimit        = 0x4F // OT_FAULT_LIMIT
	PMBusOTFaultResponse     = 0x50 // OT_FAULT_RESPONSE
	PMBusOTWarnLimit         = 0x51 // OT_WARN_LIMIT
	PMBusUTWarnLimit         = 0x52 // UT_WARN_LIMIT
	PMBusUTFaultLimit        = 0x53 // UT_FAULT_LIMIT
	PMBusUTFaultResponse     = 0x54 // UT_FAULT_RESPONSE
	PMBusVinOVFaultLimit     = 0x55 // VIN_OV_FAULT_LIMIT
	PMBusVinOVFaultResponse  = 0x56 // VIN_OV_FAULT_RESPONSE
	PMBusVinOVWarnLimit      = 0x57 // VIN_OV_WARN_LIMIT
	PMBusVinUVWarnLimit      = 0x58 // VIN_UV_WARN_LIMIT
	PMBusVinUVFaultLimit     = 0x59 // VIN_UV_FAULT_LIMIT
	PMBusVinUVFaultResponse  = 0x5A // VIN_UV_FAULT_RESPONSE
	PMBusPinOPWarnLimit      = 0x6B // PIN_OP_WARN_LIMIT
	PMBusPoutOPFaultLimit    = 0x68 // POUT_OP_FAULT_LIMIT
	PMBusPoutOPFaultResponse = 0x69 // POUT_OP_FAULT_RESPONSE
	PMBusPoutOPWarnLimit     = 0x6A // POUT_OP_WARN_LIMIT
	PMBusStatusByte          = 0x78 // STATUS_BYTE
	PMBusStatusWord          = 0x79 // STATUS_WORD
	PMBusStatusVout          = 0x7A // STATUS_VOUT
	PMBusStatusIout          = 0x7B // STATUS_IOUT
	PMBusStatusInput         = 0x7C // STATUS_INPUT
	PMBusStatusTemperature   = 0x7D // STATUS_TEMPERATURE
	PMBusStatusCML           = 0x7E // STATUS_CML
	PMBusStatusOther         = 0x7F // STATUS_OTHER
	PMBusStatusMfrSpecific   = 0x80 // STATUS_MFR_SPECIFIC
	PMBusStatusFans12        = 0x81 // STATUS_FANS_1_2
	PMBusStatusFans34        = 0x82 // STATUS_FANS_3_4
	PMBusReadEin             = 0x86 // READ_EIN
	PMBusReadEout            = 0x87 // READ_EOUT
	PMBusReadVin             = 0x88 // READ_VIN
	PMBusReadIin             = 0x89 // READ_IIN
	PMBusReadVcap            = 0x8A // READ_VCAP
	PMBusReadVout            = 0x8B // READ_VOUT
	PMBusReadIout            = 0x8C // READ_IOUT
	PMBusReadTemperature1    = 0x8D // READ_TEMPERATURE_1
	PMBusReadTemperature2    = 0x8E // READ_TEMPERATURE_2
	PMBusReadTemperature3    = 0x8F // READ_TEMPERATURE_3
	PMBusReadFanSpeed1       = 0x90 // READ_FAN_SPEED_1
	PMBusReadFanSpeed2       = 0x91 // READ_FAN_SPEED_2
	PMBusReadFanSpeed3       = 0x92 // READ_FAN_SPEED_3
	PMBusReadFanSpeed4       = 0x93 // READ_FAN_SPEED_4
	PMBusReadDutyCycle       = 0x94 // READ_DUTY_CYCLE
	PMBusReadFrequency       = 0x95 // READ_FREQUENCY
	PMBusReadPout            = 0x96 // READ_POUT
	PMBusReadPin             = 0x97 // READ_PIN
	PMBusPMBusRevision       = 0x98 // PMBUS_REVISION
	PMBusMfrID               = 0x99 // MFR_ID
	PMBusMfrModel            = 0x9A // MFR_MODEL
	PMBusMfrRevision         = 0x9B // MFR_REVISION
	PMBusMfrLocation         = 0x9C // MFR_LOCATION
	PMBusMfrDate             = 0x9D // MFR_DATE
	PMBusMfrSerial           = 0x9E // MFR_SERIAL
	PMBusAppProfileSupport   = 0x9F // APP_PROFILE_SUPPORT
	PMBusMfrVinMin           = 0xA0 // MFR_VIN_MIN
	PMBusMfrVinMax           = 0xA1 // MFR_VIN_MAX
	PMBusMfrIinMax           = 0xA2 // MFR_IIN_MAX
	PMBusMfrPinMax           = 0xA3 // MFR_PIN_MAX
	PMBusMfrVoutMin          = 0xA4 // MFR_VOUT_MIN
	PMBusMfrVoutMax          = 0xA5 // MFR_VOUT_MAX
	PMBusMfrIoutMax          = 0xA6 // MFR_IOUT_MAX
	PMBusMfrPoutMax          = 0xA7 // MFR_POUT_MAX
	PMBusMfrTambientMax      = 0xA8 // MFR_TAMBIENT_MAX
	PMBusMfrTambientMin      = 0xA9 // MFR_TAMBIENT_MIN
	PMBusMfrEfficiencyLL     = 0xAA // MFR_EFFICIENCY_LL
	PMBusMfrEfficiencyHL     = 0xAB // MFR_EFFICIENCY_HL
	PMBusMfrPinAccuracy      = 0xAC // MFR_PIN_ACCURACY
	PMBusUserData00          = 0xB0 // USER_DATA_00
	PMBusUserData01          = 0xB1 // USER_DATA_01
	PMBusUserData02          = 0xB2 // USER_DATA_02
	PMBusUserData03          = 0xB3 // USER_DATA_03
	PMBusUserData04          = 0xB4 // USER_DATA_04
	PMBusUserData05          = 0xB5 // USER_DATA_05
	PMBusUserData06          = 0xB6 // USER_DATA_06
	PMBusUserData07          = 0xB7 // USER_DATA_07
	PMBusUserData08          = 0xB8 // USER_DATA_08
	PMBusUserData09          = 0xB9 // USER_DATA_09
	PMBusUserData10          = 0xBA // USER_DATA_10
	PMBusUserData11          = 0xBB // USER_DATA_11
	PMBusUserData12          = 0xBC // USER_DATA_12
	PMBusUserData13          = 0xBD // USER_DATA_13
	PMBusUserData14          = 0xBE // USER_DATA_14
	PMBusUserData15          = 0xBF // USER_DATA_15
)

// PMBus VOUT_MODE format constants.
const (
	VoutModeLinear = 0x00 // LINEAR format
	VoutModeDirect = 0x40 // DIRECT format
)

// PMBusLinear11 represents a PMBus LINEAR11 format value.
type PMBusLinear11 struct {
	Raw   uint16  // Raw 16-bit value
	Value float64 // Converted real value
}

// PMBusLinear16 represents a PMBus LINEAR16 format value.
type PMBusLinear16 struct {
	Raw      uint16  // Raw 16-bit value
	Exponent int8    // Exponent from VOUT_MODE
	Value    float64 // Converted real value
}

// PMBusDirect represents a PMBus DIRECT format value.
type PMBusDirect struct {
	Raw   uint16  // Raw 16-bit value
	Value float64 // Converted real value using coefficients
}

// ReadVoutMode reads the VOUT_MODE register to determine output voltage format.
func (c *Conn) ReadVoutMode() (uint8, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	return c.ReadByteData(PMBusVoutMode)
}

// ReadVin reads the input voltage using PMBus READ_VIN command.
func (c *Conn) ReadVin() (float64, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	raw, err := c.ReadWordData(PMBusReadVin)
	if err != nil {
		return 0, fmt.Errorf("%w: READ_VIN failed: %w", ErrReadFailed, err)
	}

	return c.convertLinear11ToFloat(raw), nil
}

// ReadVout reads the output voltage using PMBus READ_VOUT command.
// The format depends on the VOUT_MODE setting.
func (c *Conn) ReadVout() (float64, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	raw, err := c.ReadWordData(PMBusReadVout)
	if err != nil {
		return 0, fmt.Errorf("%w: READ_VOUT failed: %w", ErrReadFailed, err)
	}

	// Get VOUT_MODE to determine format
	voutMode, err := c.ReadVoutMode()
	if err != nil {
		return 0, fmt.Errorf("%w: failed to read VOUT_MODE: %w", ErrReadFailed, err)
	}

	switch voutMode & 0xE0 { // Check format bits
	case VoutModeLinear:
		// LINEAR16 format - exponent in bits 4:0
		exponent := int8(voutMode & 0x1F)
		if exponent > 15 {
			exponent = exponent - 32 // Convert to signed
		}
		return c.convertLinear16ToFloat(raw, exponent), nil
	case VoutModeDirect:
		// DIRECT format - use coefficients
		return c.convertDirectToFloat(raw)
	default:
		return 0, fmt.Errorf("%w: unsupported VOUT_MODE format: 0x%02x", ErrPMBusDataFormatError, voutMode)
	}
}

// ReadIin reads the input current using PMBus READ_IIN command.
func (c *Conn) ReadIin() (float64, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	raw, err := c.ReadWordData(PMBusReadIin)
	if err != nil {
		return 0, fmt.Errorf("%w: READ_IIN failed: %w", ErrReadFailed, err)
	}

	return c.convertLinear11ToFloat(raw), nil
}

// ReadIout reads the output current using PMBus READ_IOUT command.
func (c *Conn) ReadIout() (float64, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	raw, err := c.ReadWordData(PMBusReadIout)
	if err != nil {
		return 0, fmt.Errorf("%w: READ_IOUT failed: %w", ErrReadFailed, err)
	}

	if c.config.PMBusFormat == PMBusFormatDirect {
		return c.convertDirectToFloat(raw)
	}
	return c.convertLinear11ToFloat(raw), nil
}

// ReadPin reads the input power using PMBus READ_PIN command.
func (c *Conn) ReadPin() (float64, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	raw, err := c.ReadWordData(PMBusReadPin)
	if err != nil {
		return 0, fmt.Errorf("%w: READ_PIN failed: %w", ErrReadFailed, err)
	}

	return c.convertLinear11ToFloat(raw), nil
}

// ReadPout reads the output power using PMBus READ_POUT command.
func (c *Conn) ReadPout() (float64, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	raw, err := c.ReadWordData(PMBusReadPout)
	if err != nil {
		return 0, fmt.Errorf("%w: READ_POUT failed: %w", ErrReadFailed, err)
	}

	if c.config.PMBusFormat == PMBusFormatDirect {
		return c.convertDirectToFloat(raw)
	}
	return c.convertLinear11ToFloat(raw), nil
}

// ReadTemperature reads temperature from the specified sensor (1, 2, or 3).
func (c *Conn) ReadTemperature(sensor uint8) (float64, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	var command uint8
	switch sensor {
	case 1:
		command = PMBusReadTemperature1
	case 2:
		command = PMBusReadTemperature2
	case 3:
		command = PMBusReadTemperature3
	default:
		return 0, fmt.Errorf("%w: invalid temperature sensor %d (must be 1, 2, or 3)", ErrInvalidData, sensor)
	}

	raw, err := c.ReadWordData(command)
	if err != nil {
		return 0, fmt.Errorf("%w: READ_TEMPERATURE_%d failed: %w", ErrReadFailed, sensor, err)
	}

	return c.convertLinear11ToFloat(raw), nil
}

// WriteVout sets the output voltage using PMBus VOUT_COMMAND.
func (c *Conn) WriteVout(voltage float64) error {
	if c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	// Get VOUT_MODE to determine format
	voutMode, err := c.ReadVoutMode()
	if err != nil {
		return fmt.Errorf("%w: failed to read VOUT_MODE: %w", ErrPMBusDataFormatError, err)
	}

	var raw uint16
	switch voutMode & 0xE0 { // Check format bits
	case VoutModeLinear:
		// LINEAR16 format
		exponent := int8(voutMode & 0x1F)
		if exponent > 15 {
			exponent = exponent - 32 // Convert to signed
		}
		raw = c.convertFloatToLinear16(voltage, exponent)
	case VoutModeDirect:
		// DIRECT format
		raw, err = c.convertFloatToDirect(voltage)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: unsupported VOUT_MODE format: 0x%02x", ErrPMBusDataFormatError, voutMode)
	}

	if err := c.WriteWordData(PMBusVoutCommand, raw); err != nil {
		return fmt.Errorf("%w: VOUT_COMMAND failed: %w", ErrWriteFailed, err)
	}

	return nil
}

// SetVoutOVFaultLimit sets the output voltage overvoltage fault limit.
func (c *Conn) SetVoutOVFaultLimit(voltage float64) error {
	if c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	// Convert to same format as VOUT
	voutMode, err := c.ReadVoutMode()
	if err != nil {
		return fmt.Errorf("%w: failed to read VOUT_MODE: %w", ErrPMBusDataFormatError, err)
	}

	var raw uint16
	switch voutMode & 0xE0 {
	case VoutModeLinear:
		exponent := int8(voutMode & 0x1F)
		if exponent > 15 {
			exponent = exponent - 32
		}
		raw = c.convertFloatToLinear16(voltage, exponent)
	case VoutModeDirect:
		raw, err = c.convertFloatToDirect(voltage)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: unsupported VOUT_MODE format: 0x%02x", ErrPMBusDataFormatError, voutMode)
	}

	if err := c.WriteWordData(PMBusVoutOVFaultLimit, raw); err != nil {
		return fmt.Errorf("%w: VOUT_OV_FAULT_LIMIT failed: %w", ErrWriteFailed, err)
	}

	return nil
}

// ReadStatusWord reads the PMBus STATUS_WORD register.
func (c *Conn) ReadStatusWord() (uint16, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	return c.ReadWordData(PMBusStatusWord)
}

// ReadStatusByte reads the PMBus STATUS_BYTE register.
func (c *Conn) ReadStatusByte() (uint8, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	return c.ReadByteData(PMBusStatusByte)
}

// ClearFaults sends the CLEAR_FAULTS command to clear all fault conditions.
func (c *Conn) ClearFaults() error {
	if c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	return c.SendByte(PMBusClearFaults)
}

// ReadManufacturerID reads the manufacturer ID string.
func (c *Conn) ReadManufacturerID() (string, error) {
	if c.config.Protocol != ProtocolPMBus {
		return "", ErrProtocolViolation
	}

	buffer := make([]byte, 16) // Typical max length
	length, err := c.ReadBlockData(PMBusMfrID, buffer)
	if err != nil {
		return "", fmt.Errorf("%w: MFR_ID failed: %w", ErrPMBusDataFormatError, err)
	}

	return string(buffer[:length]), nil
}

// ReadManufacturerModel reads the manufacturer model string.
func (c *Conn) ReadManufacturerModel() (string, error) {
	if c.config.Protocol != ProtocolPMBus {
		return "", ErrProtocolViolation
	}

	buffer := make([]byte, 16) // Typical max length
	length, err := c.ReadBlockData(PMBusMfrModel, buffer)
	if err != nil {
		return "", fmt.Errorf("%w: MFR_MODEL failed: %w", ErrPMBusDataFormatError, err)
	}

	return string(buffer[:length]), nil
}

// ReadPMBusRevision reads the PMBus revision.
func (c *Conn) ReadPMBusRevision() (uint8, error) {
	if c.config.Protocol != ProtocolPMBus {
		return 0, ErrProtocolViolation
	}

	return c.ReadByteData(PMBusPMBusRevision)
}

// SetPage sets the PMBus page for multi-page devices.
func (c *Conn) SetPage(page uint8) error {
	if c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	return c.WriteByteData(PMBusPageCommand, page)
}

// convertLinear11ToFloat converts PMBus LINEAR11 format to float64.
// LINEAR11: 5-bit exponent (bits 15:11), 11-bit mantissa (bits 10:0).
func (c *Conn) convertLinear11ToFloat(raw uint16) float64 {
	// Extract exponent (bits 15:11)
	exponent := int8((raw >> 11) & 0x1F)
	if exponent > 15 {
		exponent = exponent - 32 // Convert to signed
	}

	// Extract mantissa (bits 10:0)
	mantissa := int16(raw & 0x7FF)
	if mantissa > 1023 {
		mantissa = mantissa - 2048 // Convert to signed
	}

	// Calculate real value: mantissa * 2^exponent
	return float64(mantissa) * math.Pow(2, float64(exponent))
}

// convertLinear16ToFloat converts PMBus LINEAR16 format to float64.
// LINEAR16: 16-bit mantissa with separate exponent from VOUT_MODE.
func (c *Conn) convertLinear16ToFloat(raw uint16, exponent int8) float64 {
	mantissa := int16(raw)
	// Calculate real value: mantissa * 2^exponent
	return float64(mantissa) * math.Pow(2, float64(exponent))
}

// convertDirectToFloat converts PMBus DIRECT format to float64 using coefficients.
// Formula: Real = (1/M) * (Y * 10^(-R) - B).
func (c *Conn) convertDirectToFloat(raw uint16) (float64, error) {
	if c.config.PMBusCoefficients == nil {
		return 0, fmt.Errorf("%w: no coefficients configured for DIRECT format", ErrPMBusCoefficientsInvalid)
	}

	coeff := c.config.PMBusCoefficients
	if coeff.M == 0 {
		return 0, fmt.Errorf("%w: M must be non-zero for DIRECT format", ErrPMBusCoefficientsInvalid)
	}
	y := float64(int16(raw)) // Convert to signed

	// Apply formula: Real = (1/M) * (Y * 10^(-R) - B)
	realVal := (1.0 / float64(coeff.M)) * (y*math.Pow(10, float64(-coeff.R)) - float64(coeff.B))
	return realVal, nil
}

// convertFloatToLinear11 converts float64 to PMBus LINEAR11 format.
func (c *Conn) convertFloatToLinear11(value float64) uint16 { //nolint:unused
	if value == 0 {
		return 0
	}

	// Find appropriate exponent
	exponent := int8(0)
	absValue := math.Abs(value)

	// Scale down if value is too large
	for absValue >= 1024 && exponent < 15 {
		absValue /= 2
		exponent++
	}

	// Scale up if value is too small
	for absValue < 1 && exponent > -16 {
		absValue *= 2
		exponent--
	}

	// Calculate mantissa
	mantissa := int16(math.Round(value / math.Pow(2, float64(exponent))))

	// Clamp mantissa to 11-bit signed range
	if mantissa > 1023 {
		mantissa = 1023
	} else if mantissa < -1024 {
		mantissa = -1024
	}

	// Encode exponent and mantissa
	encodedExponent := uint16(exponent & 0x1F)
	encodedMantissa := uint16(mantissa & 0x7FF)

	return (encodedExponent << 11) | encodedMantissa
}

// convertFloatToLinear16 converts float64 to PMBus LINEAR16 format.
func (c *Conn) convertFloatToLinear16(value float64, exponent int8) uint16 {
	// Calculate mantissa: value / 2^exponent
	m := math.Round(value / math.Pow(2, float64(exponent)))
	if m > math.MaxInt16 {
		m = math.MaxInt16
	} else if m < math.MinInt16 {
		m = math.MinInt16
	}
	mantissa := int16(m)
	return uint16(mantissa)
}

// convertFloatToDirect converts float64 to PMBus DIRECT format using coefficients.
// Formula: Y = M * Real + B, then scale by 10^R.
func (c *Conn) convertFloatToDirect(value float64) (uint16, error) {
	if c.config.PMBusCoefficients == nil {
		return 0, fmt.Errorf("%w: no coefficients configured for DIRECT format", ErrPMBusCoefficientsInvalid)
	}

	coeff := c.config.PMBusCoefficients
	if coeff.M == 0 {
		return 0, fmt.Errorf("%w: M must be non-zero for DIRECT format", ErrPMBusCoefficientsInvalid)
	}

	// Apply inverse formula: Y = (M * Real + B) * 10^R
	y := (float64(coeff.M)*value + float64(coeff.B)) * math.Pow(10, float64(coeff.R))

	// Convert to 16-bit signed integer with range check
	ry := math.Round(y)
	if ry > math.MaxInt16 || ry < math.MinInt16 {
		return 0, fmt.Errorf("%w: DIRECT value out of int16 range (%.0f)", ErrInvalidData, ry)
	}
	return uint16(int16(ry)), nil
}

// ValidatePMBusDevice performs PMBus-specific device validation.
func (c *Conn) ValidatePMBusDevice() error {
	if c.config.Protocol != ProtocolPMBus {
		return ErrProtocolViolation
	}

	// Try to read PMBus revision as a basic connectivity test
	if _, err := c.ReadPMBusRevision(); err != nil {
		return fmt.Errorf("%w: failed to read PMBus revision: %w", ErrDeviceNotResponding, err)
	}

	return nil
}

// GetPMBusCapabilities returns PMBus-specific capability information.
func (c *Conn) GetPMBusCapabilities() ([]string, error) {
	if c.config.Protocol != ProtocolPMBus {
		return nil, ErrProtocolViolation
	}

	var capabilities []string

	// Check if device responds to basic PMBus commands
	if _, err := c.ReadPMBusRevision(); err == nil {
		capabilities = append(capabilities, "PMBus Basic Commands")
	}

	if _, err := c.ReadVin(); err == nil {
		capabilities = append(capabilities, "Input Voltage Monitoring")
	}

	if _, err := c.ReadVout(); err == nil {
		capabilities = append(capabilities, "Output Voltage Monitoring")
	}

	if _, err := c.ReadIin(); err == nil {
		capabilities = append(capabilities, "Input Current Monitoring")
	}

	if _, err := c.ReadIout(); err == nil {
		capabilities = append(capabilities, "Output Current Monitoring")
	}

	if _, err := c.ReadPin(); err == nil {
		capabilities = append(capabilities, "Input Power Monitoring")
	}

	if _, err := c.ReadPout(); err == nil {
		capabilities = append(capabilities, "Output Power Monitoring")
	}

	if _, err := c.ReadTemperature(1); err == nil {
		capabilities = append(capabilities, "Temperature Monitoring")
	}

	if _, err := c.ReadStatusWord(); err == nil {
		capabilities = append(capabilities, "Status Reporting")
	}

	return capabilities, nil
}
