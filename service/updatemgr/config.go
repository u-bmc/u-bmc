// SPDX-License-Identifier: BSD-3-Clause

package updatemgr

// config holds the configuration for the update manager service.
type config struct {
	name string
}

// Option represents a configuration option for the update manager service.
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
