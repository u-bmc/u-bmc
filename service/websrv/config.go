// SPDX-License-Identifier: BSD-3-Clause

package websrv

import "time"

type config struct {
	name         string
	addr         string
	webui        bool
	hostname     string
	certPath     string
	keyPath      string
	webuiPath    string
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	rmemMax      string
	wmemMax      string
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

// WithName sets the service name for the web server.
// This name is used for logging and identification purposes.
func WithName(name string) Option {
	return &nameOption{
		name: name,
	}
}

type addrOption struct {
	addr string
}

func (o *addrOption) apply(c *config) {
	c.addr = o.addr
}

// WithAddr sets the network address for the web server to listen on.
// The address should be in the format "host:port" (e.g., ":443" or "localhost:8443").
// Both HTTP/2 (TCP) and HTTP/3 (UDP) servers will bind to this address.
func WithAddr(addr string) Option {
	return &addrOption{
		addr: addr,
	}
}

type webuiOption struct {
	webui bool
}

func (o *webuiOption) apply(c *config) {
	c.webui = o.webui
}

// WithWebUI enables or disables serving the web UI static files.
// When enabled, the server will serve static files from the configured web UI path.
func WithWebUI(webui bool) Option {
	return &webuiOption{
		webui: webui,
	}
}

type hostnameOption struct {
	hostname string
}

func (o *hostnameOption) apply(c *config) {
	c.hostname = o.hostname
}

// WithHostname sets the hostname used for TLS certificate generation.
// This hostname will be included in the certificate's Subject Alternative Names (SAN).
func WithHostname(hostname string) Option {
	return &hostnameOption{
		hostname: hostname,
	}
}

type certPathOption struct {
	certPath string
}

func (o *certPathOption) apply(c *config) {
	c.certPath = o.certPath
}

// WithCertPath sets the file path where the TLS certificate is stored or will be generated.
// The certificate file should be in PEM format.
func WithCertPath(certPath string) Option {
	return &certPathOption{
		certPath: certPath,
	}
}

type keyPathOption struct {
	keyPath string
}

func (o *keyPathOption) apply(c *config) {
	c.keyPath = o.keyPath
}

// WithKeyPath sets the file path where the TLS private key is stored or will be generated.
// The private key file should be in PEM format.
func WithKeyPath(keyPath string) Option {
	return &keyPathOption{
		keyPath: keyPath,
	}
}

type webuiPathOption struct {
	webuiPath string
}

func (o *webuiPathOption) apply(c *config) {
	c.webuiPath = o.webuiPath
}

// WithWebUIPath sets the directory path containing the web UI static files.
// This path is used when web UI serving is enabled to locate HTML, CSS, JS, and other assets.
func WithWebUIPath(webuiPath string) Option {
	return &webuiPathOption{
		webuiPath: webuiPath,
	}
}

type readTimeoutOption struct {
	readTimeout time.Duration
}

func (o *readTimeoutOption) apply(c *config) {
	c.readTimeout = o.readTimeout
}

// WithReadTimeout sets the maximum duration for reading the entire request, including the body.
// A zero or negative value means there will be no timeout.
func WithReadTimeout(readTimeout time.Duration) Option {
	return &readTimeoutOption{
		readTimeout: readTimeout,
	}
}

type writeTimeoutOption struct {
	writeTimeout time.Duration
}

func (o *writeTimeoutOption) apply(c *config) {
	c.writeTimeout = o.writeTimeout
}

// WithWriteTimeout sets the maximum duration before timing out writes of the response.
// A zero or negative value means there will be no timeout.
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return &writeTimeoutOption{
		writeTimeout: writeTimeout,
	}
}

type idleTimeoutOption struct {
	idleTimeout time.Duration
}

func (o *idleTimeoutOption) apply(c *config) {
	c.idleTimeout = o.idleTimeout
}

// WithIdleTimeout sets the maximum amount of time to wait for the next request when keep-alives are enabled.
// If IdleTimeout is zero, the value of ReadTimeout is used. If both are zero, there is no timeout.
func WithIdleTimeout(idleTimeout time.Duration) Option {
	return &idleTimeoutOption{
		idleTimeout: idleTimeout,
	}
}

type rmemMaxOption struct {
	rmemMax string
}

func (o *rmemMaxOption) apply(c *config) {
	c.rmemMax = o.rmemMax
}

// WithRmemMax sets the maximum socket receive buffer size (net.core.rmem_max sysctl).
// This kernel parameter affects QUIC/UDP performance. The value should be a string
// representing the buffer size in bytes (e.g., "7500000").
func WithRmemMax(rmemMax string) Option {
	return &rmemMaxOption{
		rmemMax: rmemMax,
	}
}

type wmemMaxOption struct {
	wmemMax string
}

func (o *wmemMaxOption) apply(c *config) {
	c.wmemMax = o.wmemMax
}

// WithWmemMax sets the maximum socket send buffer size (net.core.wmem_max sysctl).
// This kernel parameter affects QUIC/UDP performance. The value should be a string
// representing the buffer size in bytes (e.g., "7500000").
func WithWmemMax(wmemMax string) Option {
	return &wmemMaxOption{
		wmemMax: wmemMax,
	}
}
