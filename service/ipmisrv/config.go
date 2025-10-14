// SPDX-License-Identifier: BSD-3-Clause

package ipmisrv

// config holds the configuration for the IPMI server service.
type config struct {
	name string
}

// Option represents a configuration option for the IPMI server service.
type Option interface {
	apply(*config)
}

type nameOption struct {
	name string
}

func (o *nameOption) apply(c *config) {
	c.name = o.name
}

func WithServiceName(name string) Option {
	return &nameOption{
		name: name,
	}
}
