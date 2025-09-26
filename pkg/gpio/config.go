// SPDX-License-Identifier: BSD-3-Clause

//go:build linux
// +build linux

package gpio

import (
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// EdgeType represents GPIO edge detection types.
type EdgeType int

const (
	// EdgeNone disables edge detection.
	EdgeNone EdgeType = iota
	// EdgeRising enables detection of rising edges.
	EdgeRising
	// EdgeFalling enables detection of falling edges.
	EdgeFalling
	// EdgeBoth enables detection of both rising and falling edges.
	EdgeBoth
)

// Option represents a configuration option for GPIO line requests.
type Option interface {
	apply() gpiocdev.LineReqOption
}

type inputOption struct{}

func (o inputOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.AsInput
}

// AsInput configures the GPIO line as an input.
func AsInput() Option {
	return inputOption{}
}

type outputOption struct {
	initialValue int
}

func (o outputOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.AsOutput(o.initialValue)
}

// AsOutput configures the GPIO line as an output with initial value 0.
func AsOutput() Option {
	return outputOption{initialValue: 0}
}

// AsOutputValue configures the GPIO line as an output with the specified initial value.
func AsOutputValue(value int) Option {
	return outputOption{initialValue: value}
}

type pullUpOption struct{}

func (o pullUpOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.WithPullUp
}

// WithPullUp enables the internal pull-up resistor.
func WithPullUp() Option {
	return pullUpOption{}
}

type pullDownOption struct{}

func (o pullDownOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.WithPullDown
}

// WithPullDown enables the internal pull-down resistor.
func WithPullDown() Option {
	return pullDownOption{}
}

type activeLowOption struct{}

func (o activeLowOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.AsActiveLow
}

// WithActiveLow configures the line as active-low.
func WithActiveLow() Option {
	return activeLowOption{}
}

type activeHighOption struct{}

func (o activeHighOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.AsActiveHigh
}

// WithActiveHigh configures the line as active-high (default).
func WithActiveHigh() Option {
	return activeHighOption{}
}

type consumerOption struct {
	consumer string
}

func (o consumerOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.WithConsumer(o.consumer)
}

// WithConsumer sets the consumer name for the GPIO line.
func WithConsumer(consumer string) Option {
	return consumerOption{consumer: consumer}
}

type debounceOption struct {
	period time.Duration
}

func (o debounceOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.WithDebounce(o.period)
}

// WithDebounce sets the debounce period for input lines.
func WithDebounce(period time.Duration) Option {
	return debounceOption{period: period}
}

type edgeDetectionOption struct {
	edgeType EdgeType
}

func (o edgeDetectionOption) apply() gpiocdev.LineReqOption {
	switch o.edgeType {
	case EdgeRising:
		return gpiocdev.WithRisingEdge
	case EdgeFalling:
		return gpiocdev.WithFallingEdge
	case EdgeBoth:
		return gpiocdev.WithBothEdges
	default:
		return gpiocdev.WithoutEdges
	}
}

// WithEdgeDetection enables edge detection for input lines.
func WithEdgeDetection(edgeType EdgeType) Option {
	return edgeDetectionOption{edgeType: edgeType}
}

type openDrainOption struct{}

func (o openDrainOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.AsOpenDrain
}

// WithOpenDrain configures the output as open-drain.
func WithOpenDrain() Option {
	return openDrainOption{}
}

type openSourceOption struct{}

func (o openSourceOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.AsOpenSource
}

// WithOpenSource configures the output as open-source.
func WithOpenSource() Option {
	return openSourceOption{}
}

type initialValueOption struct {
	value int
}

func (o initialValueOption) apply() gpiocdev.LineReqOption {
	return gpiocdev.AsOutput(o.value)
}

// WithInitialValue sets the initial value for output lines.
func WithInitialValue(value int) Option {
	return initialValueOption{value: value}
}

// convertOptions converts our Option types to gpiocdev options.
func convertOptions(opts []Option) []gpiocdev.LineReqOption {
	var gpiocdevOpts []gpiocdev.LineReqOption
	for _, opt := range opts {
		if opt != nil {
			gpiocdevOpts = append(gpiocdevOpts, opt.apply())
		}
	}
	return gpiocdevOpts
}
