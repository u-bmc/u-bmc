// SPDX-License-Identifier: BSD-3-Clause

package websrv

type config struct {
	name  string
	addr  string
	webui bool
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
