// SPDX-License-Identifier: BSD-3-Clause

package cert

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

// GenerateAndSign generates a new certificate for the given fully qualified domain name (fqdn) and email,
// and returns a TLS configuration and an HTTP handler.
// The function uses the autocert package to automatically access certificates from Let's Encrypt.
func GenerateAndSign(fqdn, email string) (*tls.Config, http.Handler) {
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(fqdn),
		Email:      email,
		Cache:      autocert.DirCache("/var/cache/cert"),
	}

	return m.TLSConfig(), m.HTTPHandler(nil)
}
