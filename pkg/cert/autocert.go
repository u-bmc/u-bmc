// SPDX-License-Identifier: BSD-3-Clause

package cert

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

// GenerateAndSign generates a new certificate using Let's Encrypt with the provided configuration,
// and returns a TLS configuration and an HTTP handler.
// The function uses the autocert package to automatically obtain certificates from Let's Encrypt.
func GenerateAndSign(cfg *Config) (*tls.Config, http.Handler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrInvalidCertificateOptions, err)
	}

	if cfg.Type != CertificateTypeLetsEncrypt {
		return nil, nil, fmt.Errorf("%w: configuration type must be Let's Encrypt", ErrAutocertSetup)
	}

	if cfg.Email == "" {
		return nil, nil, fmt.Errorf("%w: email is required for Let's Encrypt", ErrInvalidEmail)
	}

	if cfg.CacheDir == "" {
		return nil, nil, fmt.Errorf("%w: cache directory is required for Let's Encrypt", ErrCacheDirectory)
	}

	// Get all hostnames for the certificate
	hostnames := cfg.GetAllHostnames()
	if len(hostnames) == 0 {
		return nil, nil, fmt.Errorf("%w: at least one hostname is required", ErrInvalidHostname)
	}

	m := &autocert.Manager{
		Email:      cfg.Email,
		Cache:      autocert.DirCache(cfg.CacheDir),
		HostPolicy: autocert.HostWhitelist(hostnames...),
	}

	// Set the TOS acceptance policy based on configuration
	if cfg.AcceptTOS {
		m.Prompt = autocert.AcceptTOS
	} else {
		m.Prompt = func(tosURL string) bool {
			// Return false to require manual TOS acceptance
			return false
		}
	}

	return m.TLSConfig(), m.HTTPHandler(nil), nil
}
