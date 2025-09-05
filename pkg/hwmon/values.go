// SPDX-License-Identifier: BSD-3-Clause

package hwmon

import (
	"fmt"
	"strconv"
	"strings"
)

// Value represents a sensor value with type-safe conversion methods.
type Value interface {
	// Raw returns the raw integer value as read from sysfs
	Raw() int64
	// Float returns the value as a float64 in standard units
	Float() float64
	// String returns a human-readable representation
	String() string
	// Type returns the sensor type this value represents
	Type() SensorType
	// IsValid returns true if the value is within expected ranges
	IsValid() bool
	// AsTemperature converts to temperature (if applicable)
	AsTemperature() TemperatureValue
	// AsVoltage converts to voltage (if applicable)
	AsVoltage() VoltageValue
	// AsFan converts to fan speed (if applicable)
	AsFan() FanValue
	// AsPower converts to power (if applicable)
	AsPower() PowerValue
	// AsCurrent converts to current (if applicable)
	AsCurrent() CurrentValue
	// AsHumidity converts to humidity (if applicable)
	AsHumidity() HumidityValue
	// AsPressure converts to pressure (if applicable)
	AsPressure() PressureValue
	// AsPWM converts to PWM (if applicable)
	AsPWM() PWMValue
	// AsGeneric converts to generic value
	AsGeneric() GenericValue
}

// TemperatureValue represents a temperature sensor value.
type TemperatureValue struct {
	raw int64 // millidegree Celsius
}

// NewTemperatureValue creates a new temperature value from millidegree Celsius.
func NewTemperatureValue(millidegree int64) TemperatureValue {
	return TemperatureValue{raw: millidegree}
}

// Raw returns the raw millidegree Celsius value.
func (t TemperatureValue) Raw() int64 {
	return t.raw
}

// Float returns the temperature in degrees Celsius.
func (t TemperatureValue) Float() float64 {
	return float64(t.raw) / 1000.0
}

// Celsius returns the temperature in degrees Celsius.
func (t TemperatureValue) Celsius() float64 {
	return t.Float()
}

// Fahrenheit returns the temperature in degrees Fahrenheit.
func (t TemperatureValue) Fahrenheit() float64 {
	return t.Celsius()*9.0/5.0 + 32.0
}

// Kelvin returns the temperature in Kelvin.
func (t TemperatureValue) Kelvin() float64 {
	return t.Celsius() + 273.15
}

// String returns a human-readable temperature string.
func (t TemperatureValue) String() string {
	return fmt.Sprintf("%.1f°C", t.Celsius())
}

// Type returns the sensor type.
func (t TemperatureValue) Type() SensorType {
	return SensorTypeTemperature
}

// IsValid returns true if the temperature is within reasonable bounds.
func (t TemperatureValue) IsValid() bool {
	celsius := t.Celsius()
	return celsius >= -273.15 && celsius <= 200.0
}

// AsTemperature returns itself.
func (t TemperatureValue) AsTemperature() TemperatureValue { return t }
func (t TemperatureValue) AsVoltage() VoltageValue         { return VoltageValue{} }
func (t TemperatureValue) AsFan() FanValue                 { return FanValue{} }
func (t TemperatureValue) AsPower() PowerValue             { return PowerValue{} }
func (t TemperatureValue) AsCurrent() CurrentValue         { return CurrentValue{} }
func (t TemperatureValue) AsHumidity() HumidityValue       { return HumidityValue{} }
func (t TemperatureValue) AsPressure() PressureValue       { return PressureValue{} }
func (t TemperatureValue) AsPWM() PWMValue                 { return PWMValue{} }
func (t TemperatureValue) AsGeneric() GenericValue {
	return GenericValue{raw: t.raw, sensorType: SensorTypeTemperature}
}

// VoltageValue represents a voltage sensor value.
type VoltageValue struct {
	raw int64 // millivolts
}

// NewVoltageValue creates a new voltage value from millivolts.
func NewVoltageValue(millivolts int64) VoltageValue {
	return VoltageValue{raw: millivolts}
}

// Raw returns the raw millivolt value.
func (v VoltageValue) Raw() int64 {
	return v.raw
}

// Float returns the voltage in volts.
func (v VoltageValue) Float() float64 {
	return float64(v.raw) / 1000.0
}

// Volts returns the voltage in volts.
func (v VoltageValue) Volts() float64 {
	return v.Float()
}

// Millivolts returns the voltage in millivolts.
func (v VoltageValue) Millivolts() int64 {
	return v.raw
}

// String returns a human-readable voltage string.
func (v VoltageValue) String() string {
	return fmt.Sprintf("%.3fV", v.Volts())
}

// Type returns the sensor type.
func (v VoltageValue) Type() SensorType {
	return SensorTypeVoltage
}

// IsValid returns true if the voltage is within reasonable bounds.
func (v VoltageValue) IsValid() bool {
	volts := v.Volts()
	return volts >= 0.0 && volts <= 50.0
}

func (v VoltageValue) AsTemperature() TemperatureValue { return TemperatureValue{} }
func (v VoltageValue) AsVoltage() VoltageValue         { return v }
func (v VoltageValue) AsFan() FanValue                 { return FanValue{} }
func (v VoltageValue) AsPower() PowerValue             { return PowerValue{} }
func (v VoltageValue) AsCurrent() CurrentValue         { return CurrentValue{} }
func (v VoltageValue) AsHumidity() HumidityValue       { return HumidityValue{} }
func (v VoltageValue) AsPressure() PressureValue       { return PressureValue{} }
func (v VoltageValue) AsPWM() PWMValue                 { return PWMValue{} }
func (v VoltageValue) AsGeneric() GenericValue {
	return GenericValue{raw: v.raw, sensorType: SensorTypeVoltage}
}

// FanValue represents a fan sensor value.
type FanValue struct {
	raw int64 // RPM
}

// NewFanValue creates a new fan value from RPM.
func NewFanValue(rpm int64) FanValue {
	return FanValue{raw: rpm}
}

// Raw returns the raw RPM value.
func (f FanValue) Raw() int64 {
	return f.raw
}

// Float returns the fan speed in RPM.
func (f FanValue) Float() float64 {
	return float64(f.raw)
}

// RPM returns the fan speed in RPM.
func (f FanValue) RPM() int64 {
	return f.raw
}

// String returns a human-readable fan speed string.
func (f FanValue) String() string {
	return fmt.Sprintf("%d RPM", f.raw)
}

// Type returns the sensor type.
func (f FanValue) Type() SensorType {
	return SensorTypeFan
}

// IsValid returns true if the fan speed is within reasonable bounds.
func (f FanValue) IsValid() bool {
	return f.raw >= 0 && f.raw <= 50000
}

func (f FanValue) AsTemperature() TemperatureValue { return TemperatureValue{} }
func (f FanValue) AsVoltage() VoltageValue         { return VoltageValue{} }
func (f FanValue) AsFan() FanValue                 { return f }
func (f FanValue) AsPower() PowerValue             { return PowerValue{} }
func (f FanValue) AsCurrent() CurrentValue         { return CurrentValue{} }
func (f FanValue) AsHumidity() HumidityValue       { return HumidityValue{} }
func (f FanValue) AsPressure() PressureValue       { return PressureValue{} }
func (f FanValue) AsPWM() PWMValue                 { return PWMValue{} }
func (f FanValue) AsGeneric() GenericValue {
	return GenericValue{raw: f.raw, sensorType: SensorTypeFan}
}

// PowerValue represents a power sensor value.
type PowerValue struct {
	raw int64 // microwatts
}

// NewPowerValue creates a new power value from microwatts.
func NewPowerValue(microwatts int64) PowerValue {
	return PowerValue{raw: microwatts}
}

// Raw returns the raw microwatt value.
func (p PowerValue) Raw() int64 {
	return p.raw
}

// Float returns the power in watts.
func (p PowerValue) Float() float64 {
	return float64(p.raw) / 1000000.0
}

// Watts returns the power in watts.
func (p PowerValue) Watts() float64 {
	return p.Float()
}

// Milliwatts returns the power in milliwatts.
func (p PowerValue) Milliwatts() float64 {
	return float64(p.raw) / 1000.0
}

// Microwatts returns the power in microwatts.
func (p PowerValue) Microwatts() int64 {
	return p.raw
}

// String returns a human-readable power string.
func (p PowerValue) String() string {
	if p.Watts() >= 1.0 {
		return fmt.Sprintf("%.2fW", p.Watts())
	}
	return fmt.Sprintf("%.1fmW", p.Milliwatts())
}

// Type returns the sensor type.
func (p PowerValue) Type() SensorType {
	return SensorTypePower
}

// IsValid returns true if the power is within reasonable bounds.
func (p PowerValue) IsValid() bool {
	watts := p.Watts()
	return watts >= 0.0 && watts <= 10000.0
}

func (p PowerValue) AsTemperature() TemperatureValue { return TemperatureValue{} }
func (p PowerValue) AsVoltage() VoltageValue         { return VoltageValue{} }
func (p PowerValue) AsFan() FanValue                 { return FanValue{} }
func (p PowerValue) AsPower() PowerValue             { return p }
func (p PowerValue) AsCurrent() CurrentValue         { return CurrentValue{} }
func (p PowerValue) AsHumidity() HumidityValue       { return HumidityValue{} }
func (p PowerValue) AsPressure() PressureValue       { return PressureValue{} }
func (p PowerValue) AsPWM() PWMValue                 { return PWMValue{} }
func (p PowerValue) AsGeneric() GenericValue {
	return GenericValue{raw: p.raw, sensorType: SensorTypePower}
}

// CurrentValue represents a current sensor value.
type CurrentValue struct {
	raw int64 // milliamps
}

// NewCurrentValue creates a new current value from milliamps.
func NewCurrentValue(milliamps int64) CurrentValue {
	return CurrentValue{raw: milliamps}
}

// Raw returns the raw milliamp value.
func (c CurrentValue) Raw() int64 {
	return c.raw
}

// Float returns the current in amps.
func (c CurrentValue) Float() float64 {
	return float64(c.raw) / 1000.0
}

// Amps returns the current in amps.
func (c CurrentValue) Amps() float64 {
	return c.Float()
}

// Milliamps returns the current in milliamps.
func (c CurrentValue) Milliamps() int64 {
	return c.raw
}

// String returns a human-readable current string.
func (c CurrentValue) String() string {
	if c.Amps() >= 1.0 {
		return fmt.Sprintf("%.3fA", c.Amps())
	}
	return fmt.Sprintf("%dmA", c.raw)
}

// Type returns the sensor type.
func (c CurrentValue) Type() SensorType {
	return SensorTypeCurrent
}

// IsValid returns true if the current is within reasonable bounds.
func (c CurrentValue) IsValid() bool {
	amps := c.Amps()
	return amps >= 0.0 && amps <= 1000.0
}

func (c CurrentValue) AsTemperature() TemperatureValue { return TemperatureValue{} }
func (c CurrentValue) AsVoltage() VoltageValue         { return VoltageValue{} }
func (c CurrentValue) AsFan() FanValue                 { return FanValue{} }
func (c CurrentValue) AsPower() PowerValue             { return PowerValue{} }
func (c CurrentValue) AsCurrent() CurrentValue         { return c }
func (c CurrentValue) AsHumidity() HumidityValue       { return HumidityValue{} }
func (c CurrentValue) AsPressure() PressureValue       { return PressureValue{} }
func (c CurrentValue) AsPWM() PWMValue                 { return PWMValue{} }
func (c CurrentValue) AsGeneric() GenericValue {
	return GenericValue{raw: c.raw, sensorType: SensorTypeCurrent}
}

// HumidityValue represents a humidity sensor value.
type HumidityValue struct {
	raw int64 // percentage * 1000
}

// NewHumidityValue creates a new humidity value from percentage * 1000.
func NewHumidityValue(milliPercent int64) HumidityValue {
	return HumidityValue{raw: milliPercent}
}

// Raw returns the raw value (percentage * 1000).
func (h HumidityValue) Raw() int64 {
	return h.raw
}

// Float returns the humidity as a percentage.
func (h HumidityValue) Float() float64 {
	return float64(h.raw) / 1000.0
}

// Percent returns the humidity as a percentage.
func (h HumidityValue) Percent() float64 {
	return h.Float()
}

// String returns a human-readable humidity string.
func (h HumidityValue) String() string {
	return fmt.Sprintf("%.1f%%", h.Percent())
}

// Type returns the sensor type.
func (h HumidityValue) Type() SensorType {
	return SensorTypeHumidity
}

// IsValid returns true if the humidity is within valid bounds.
func (h HumidityValue) IsValid() bool {
	percent := h.Percent()
	return percent >= 0.0 && percent <= 100.0
}

func (h HumidityValue) AsTemperature() TemperatureValue { return TemperatureValue{} }
func (h HumidityValue) AsVoltage() VoltageValue         { return VoltageValue{} }
func (h HumidityValue) AsFan() FanValue                 { return FanValue{} }
func (h HumidityValue) AsPower() PowerValue             { return PowerValue{} }
func (h HumidityValue) AsCurrent() CurrentValue         { return CurrentValue{} }
func (h HumidityValue) AsHumidity() HumidityValue       { return h }
func (h HumidityValue) AsPressure() PressureValue       { return PressureValue{} }
func (h HumidityValue) AsPWM() PWMValue                 { return PWMValue{} }
func (h HumidityValue) AsGeneric() GenericValue {
	return GenericValue{raw: h.raw, sensorType: SensorTypeHumidity}
}

// PressureValue represents a pressure sensor value.
type PressureValue struct {
	raw int64 // pascals
}

// NewPressureValue creates a new pressure value from pascals.
func NewPressureValue(pascals int64) PressureValue {
	return PressureValue{raw: pascals}
}

// Raw returns the raw pascal value.
func (p PressureValue) Raw() int64 {
	return p.raw
}

// Float returns the pressure in pascals.
func (p PressureValue) Float() float64 {
	return float64(p.raw)
}

// Pascals returns the pressure in pascals.
func (p PressureValue) Pascals() int64 {
	return p.raw
}

// Kilopascals returns the pressure in kilopascals.
func (p PressureValue) Kilopascals() float64 {
	return float64(p.raw) / 1000.0
}

// Bars returns the pressure in bars.
func (p PressureValue) Bars() float64 {
	return float64(p.raw) / 100000.0
}

// String returns a human-readable pressure string.
func (p PressureValue) String() string {
	if p.Kilopascals() >= 1.0 {
		return fmt.Sprintf("%.2f kPa", p.Kilopascals())
	}
	return fmt.Sprintf("%d Pa", p.raw)
}

// Type returns the sensor type.
func (p PressureValue) Type() SensorType {
	return SensorTypePressure
}

// IsValid returns true if the pressure is within reasonable bounds.
func (p PressureValue) IsValid() bool {
	return p.raw >= 0 && p.raw <= 1000000000
}

func (p PressureValue) AsTemperature() TemperatureValue { return TemperatureValue{} }
func (p PressureValue) AsVoltage() VoltageValue         { return VoltageValue{} }
func (p PressureValue) AsFan() FanValue                 { return FanValue{} }
func (p PressureValue) AsPower() PowerValue             { return PowerValue{} }
func (p PressureValue) AsCurrent() CurrentValue         { return CurrentValue{} }
func (p PressureValue) AsHumidity() HumidityValue       { return HumidityValue{} }
func (p PressureValue) AsPressure() PressureValue       { return p }
func (p PressureValue) AsPWM() PWMValue                 { return PWMValue{} }
func (p PressureValue) AsGeneric() GenericValue {
	return GenericValue{raw: p.raw, sensorType: SensorTypePressure}
}

// PWMValue represents a PWM output value.
type PWMValue struct {
	raw int64 // 0-255
}

// NewPWMValue creates a new PWM value (0-255).
func NewPWMValue(value int64) PWMValue {
	if value < 0 {
		value = 0
	} else if value > 255 {
		value = 255
	}
	return PWMValue{raw: value}
}

// Raw returns the raw PWM value (0-255).
func (p PWMValue) Raw() int64 {
	return p.raw
}

// Float returns the PWM value as a percentage (0.0-100.0).
func (p PWMValue) Float() float64 {
	return float64(p.raw) * 100.0 / 255.0
}

// Value returns the PWM value (0-255).
func (p PWMValue) Value() int64 {
	return p.raw
}

// Percent returns the PWM value as a percentage.
func (p PWMValue) Percent() float64 {
	return p.Float()
}

// String returns a human-readable PWM string.
func (p PWMValue) String() string {
	return fmt.Sprintf("PWM %d (%.1f%%)", p.raw, p.Percent())
}

// Type returns the sensor type.
func (p PWMValue) Type() SensorType {
	return SensorTypePWM
}

// IsValid returns true if the PWM value is within valid bounds.
func (p PWMValue) IsValid() bool {
	return p.raw >= 0 && p.raw <= 255
}

func (p PWMValue) AsTemperature() TemperatureValue { return TemperatureValue{} }
func (p PWMValue) AsVoltage() VoltageValue         { return VoltageValue{} }
func (p PWMValue) AsFan() FanValue                 { return FanValue{} }
func (p PWMValue) AsPower() PowerValue             { return PowerValue{} }
func (p PWMValue) AsCurrent() CurrentValue         { return CurrentValue{} }
func (p PWMValue) AsHumidity() HumidityValue       { return HumidityValue{} }
func (p PWMValue) AsPressure() PressureValue       { return PressureValue{} }
func (p PWMValue) AsPWM() PWMValue                 { return p }
func (p PWMValue) AsGeneric() GenericValue {
	return GenericValue{raw: p.raw, sensorType: SensorTypePWM}
}

// GenericValue represents a generic sensor value.
type GenericValue struct {
	raw        int64
	sensorType SensorType
}

// NewGenericValue creates a new generic value.
func NewGenericValue(value int64, sensorType SensorType) GenericValue {
	return GenericValue{raw: value, sensorType: sensorType}
}

// Raw returns the raw value.
func (g GenericValue) Raw() int64 {
	return g.raw
}

// Float returns the value as a float64.
func (g GenericValue) Float() float64 {
	return float64(g.raw)
}

// String returns a human-readable generic value string.
func (g GenericValue) String() string {
	return fmt.Sprintf("%d (%v)", g.raw, g.sensorType)
}

// Type returns the sensor type.
func (g GenericValue) Type() SensorType {
	return g.sensorType
}

// IsValid always returns true for generic values.
func (g GenericValue) IsValid() bool {
	return true
}

func (g GenericValue) AsTemperature() TemperatureValue { return TemperatureValue{raw: g.raw} }
func (g GenericValue) AsVoltage() VoltageValue         { return VoltageValue{raw: g.raw} }
func (g GenericValue) AsFan() FanValue                 { return FanValue{raw: g.raw} }
func (g GenericValue) AsPower() PowerValue             { return PowerValue{raw: g.raw} }
func (g GenericValue) AsCurrent() CurrentValue         { return CurrentValue{raw: g.raw} }
func (g GenericValue) AsHumidity() HumidityValue       { return HumidityValue{raw: g.raw} }
func (g GenericValue) AsPressure() PressureValue       { return PressureValue{raw: g.raw} }
func (g GenericValue) AsPWM() PWMValue                 { return PWMValue{raw: g.raw} }
func (g GenericValue) AsGeneric() GenericValue         { return g }

// ParseValue parses a string value from sysfs and returns the appropriate Value type.
func ParseValue(rawValue string, sensorType SensorType) (Value, error) {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return nil, fmt.Errorf("%w: empty value", ErrValueParseFailure)
	}

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValueParseFailure, err)
	}

	switch sensorType {
	case SensorTypeTemperature:
		return NewTemperatureValue(value), nil
	case SensorTypeVoltage:
		return NewVoltageValue(value), nil
	case SensorTypeFan:
		return NewFanValue(value), nil
	case SensorTypePower:
		return NewPowerValue(value), nil
	case SensorTypeCurrent:
		return NewCurrentValue(value), nil
	case SensorTypeHumidity:
		return NewHumidityValue(value), nil
	case SensorTypePressure:
		return NewPressureValue(value), nil
	case SensorTypePWM:
		return NewPWMValue(value), nil
	default:
		return NewGenericValue(value, sensorType), nil
	}
}

// FormatValue formats a Value for writing to sysfs.
func FormatValue(value Value) string {
	return fmt.Sprintf("%d", value.Raw())
}
