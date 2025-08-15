// SPDX-License-Identifier: BSD-3-Clause

package ipc

import "github.com/nats-io/nats-server/v2/server"

type config struct {
	name       string
	serverOpts *server.Options
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

type serverOption struct {
	opts *server.Options
}

func (o *serverOption) apply(c *config) {
	c.serverOpts = o.opts
}

func WithServerOpts(opts *server.Options) Option {
	return &serverOption{
		opts: opts,
	}
}
