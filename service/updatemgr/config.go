// SPDX-License-Identifier: BSD-3-Clause

package updatemgr

type config struct {
	name string
}

type Option interface {
	apply(*config)
}

type nameOption struct {
	name string
}

func (o *nameOption) apply(c *config) {
	c.name = o.name
}

func WithName(name string) Option {
	return &nameOption{
		name: name,
	}
}
