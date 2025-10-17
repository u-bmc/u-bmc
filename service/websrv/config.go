// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"time"

	"github.com/u-bmc/u-bmc/pkg/cert"
)

type config struct {
	name         string
	addr         string
	webui        bool
	webuiPath    string
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	rmemMax      string
	wmemMax      string
	certConfig   *cert.Config
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

// WithServiceName sets the service name for the web server.
// This name is used for logging and identification purposes.
func WithServiceName(name string) Option {
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

type certConfigOption struct {
	certConfig *cert.Config
}

func (o *certConfigOption) apply(c *config) {
	c.certConfig = o.certConfig
}

// WithCertConfig sets the certificate configuration for TLS.
// This allows full customization of certificate generation and management.
func WithCertConfig(certConfig *cert.Config) Option {
	return &certConfigOption{
		certConfig: certConfig,
	}
}

type hostnameOption struct {
	hostname string
}

func (o *hostnameOption) apply(c *config) {
	if c.certConfig == nil {
		c.certConfig = cert.NewConfig()
	}
	c.certConfig.Hostname = o.hostname
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
	if c.certConfig == nil {
		c.certConfig = cert.NewConfig()
	}
	c.certConfig.CertPath = o.certPath
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
	if c.certConfig == nil {
		c.certConfig = cert.NewConfig()
	}
	c.certConfig.KeyPath = o.keyPath
}

// WithKeyPath sets the file path where the TLS private key is stored or will be generated.
// The private key file should be in PEM format.
func WithKeyPath(keyPath string) Option {
	return &keyPathOption{
		keyPath: keyPath,
	}
}

type certificateTypeOption struct {
	certType cert.CertificateType
}

func (o *certificateTypeOption) apply(c *config) {
	if c.certConfig == nil {
		c.certConfig = cert.NewConfig()
	}
	c.certConfig.Type = o.certType
}

// WithCertificateType sets the type of certificate to use (self-signed or Let's Encrypt).
func WithCertificateType(certType cert.CertificateType) Option {
	return &certificateTypeOption{
		certType: certType,
	}
}

type certEmailOption struct {
	email string
}

func (o *certEmailOption) apply(c *config) {
	if c.certConfig == nil {
		c.certConfig = cert.NewConfig()
	}
	c.certConfig.Email = o.email
}

// WithCertEmail sets the email address for Let's Encrypt certificate registration.
func WithCertEmail(email string) Option {
	return &certEmailOption{
		email: email,
	}
}

type alternativeNamesOption struct {
	altNames []string
}

func (o *alternativeNamesOption) apply(c *config) {
	if c.certConfig == nil {
		c.certConfig = cert.NewConfig()
	}
	c.certConfig.AlternativeNames = o.altNames
}

// WithAlternativeNames sets additional hostnames and IP addresses for the certificate.
func WithAlternativeNames(altNames ...string) Option {
	return &alternativeNamesOption{
		altNames: altNames,
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

// GetCertConfig returns the certificate configuration, creating a default one if none exists.
func (c *config) GetCertConfig() *cert.Config {
	if c.certConfig == nil {
		c.certConfig = cert.NewConfig()
	}
	return c.certConfig
}

// SetCertDefaults applies sensible defaults to the certificate configuration if not already set.
func (c *config) SetCertDefaults() {
	if c.certConfig == nil {
		c.certConfig = cert.NewConfig()
	}

	// Set default paths if not specified
	if c.certConfig.CertPath == "" {
		c.certConfig.CertPath = "/var/cache/cert/cert.pem"
	}
	if c.certConfig.KeyPath == "" {
		c.certConfig.KeyPath = "/var/cache/cert/key.pem"
	}
	if c.certConfig.CacheDir == "" {
		c.certConfig.CacheDir = "/var/cache/cert"
	}

	// Set default hostname if not specified
	if c.certConfig.Hostname == "" {
		c.certConfig.Hostname = "localhost"
	}
}
