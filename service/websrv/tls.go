// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/u-bmc/u-bmc/pkg/cert"
)

// setupTLS configures TLS settings and loads or generates certificates based on the configuration.
func (s *WebSrv) setupTLS() (*tls.Config, http.Handler, error) {
	s.SetCertDefaults()
	certConfig := s.GetCertConfig()

	if err := certConfig.Validate(); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrSetupTLS, err)
	}

	switch certConfig.Type {
	case cert.CertificateTypeSelfSigned:
		return setupSelfSignedTLS(certConfig)
	case cert.CertificateTypeLetsTencrypt:
		return setupLetsTEncryptTLS(certConfig)
	default:
		return nil, nil, fmt.Errorf("%w: unsupported certificate type", ErrSetupTLS)
	}
}

// setupSelfSignedTLS configures TLS using self-signed certificates.
func setupSelfSignedTLS(certConfig *cert.Config) (*tls.Config, http.Handler, error) {
	certPem, keyPem, err := cert.LoadOrGenerateCertificate(certConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrLoadOrGenerateCertificate, err)
	}

	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParseCertificate, err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
	}

	return tlsConfig, nil, nil
}

// setupLetsTEncryptTLS configures TLS using Let's Encrypt certificates.
func setupLetsTEncryptTLS(certConfig *cert.Config) (*tls.Config, http.Handler, error) {
	tlsConfig, httpHandler, err := cert.GenerateAndSign(certConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrSetupTLS, err)
	}

	tlsConfig.MinVersion = tls.VersionTLS13
	tlsConfig.CurvePreferences = []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
		tls.CurveP384,
	}

	return tlsConfig, httpHandler, nil
}

// configureTLSForHTTP3 applies HTTP/3 specific TLS configuration.
func configureTLSForHTTP3(baseConfig *tls.Config) *tls.Config {
	config := baseConfig.Clone()
	config.NextProtos = []string{"h3"}
	return config
}

// configureTLSForHTTP2 applies HTTP/2 specific TLS configuration.
func configureTLSForHTTP2(baseConfig *tls.Config) *tls.Config {
	config := baseConfig.Clone()
	config.NextProtos = []string{"h2", "http/1.1"}
	return config
}

// validateTLSConfig performs basic validation on a TLS configuration.
func validateTLSConfig(config *tls.Config) error {
	if config == nil {
		return fmt.Errorf("TLS config cannot be nil")
	}

	if len(config.Certificates) == 0 && config.GetCertificate == nil {
		return fmt.Errorf("no certificates configured")
	}

	if config.MinVersion < tls.VersionTLS12 {
		return fmt.Errorf("minimum TLS version must be at least TLS 1.2")
	}

	return nil
}
