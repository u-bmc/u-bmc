// SPDX-License-Identifier: BSD-3-Clause

package cert

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

// CertificateOptions contains the configurable options for generating a certificate.
type CertificateOptions struct {
	Hostname     string
	Organization string
	Country      string
	Province     string
	Locality     string
	NotBefore    time.Time
	NotAfter     time.Time
	IsCA         bool
}

// GenerateSelfsigned creates a self-signed certificate for the given hostname.
// It returns the certificate and private key as byte slices, or an error if any step of the generation fails.
func GenerateSelfsigned(opts CertificateOptions) ([]byte, []byte, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	if opts.Organization == "" {
		opts.Organization = "u-bmc"
	}

	if opts.NotBefore.IsZero() {
		opts.NotBefore = time.Now().Add(-30 * time.Second)
	}

	if opts.NotAfter.IsZero() {
		opts.NotAfter = time.Now().Add(262980 * time.Hour)
	}

	subject := pkix.Name{
		CommonName:   opts.Hostname,
		Organization: []string{opts.Organization},
	}

	if opts.Country != "" {
		subject.Country = []string{opts.Country}
	}

	if opts.Province != "" {
		subject.Province = []string{opts.Province}
	}

	if opts.Locality != "" {
		subject.Locality = []string{opts.Locality}
	}

	template := &x509.Certificate{
		Subject:  subject,
		DNSNames: []string{opts.Hostname},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageKeyAgreement | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		SerialNumber:          sn,
		NotBefore:             opts.NotBefore,
		NotAfter:              opts.NotAfter,
		IsCA:                  opts.IsCA,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		return nil, nil, err
	}

	var certPem bytes.Buffer
	if err := pem.Encode(&certPem, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, nil, err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	var keyPem bytes.Buffer
	if err := pem.Encode(&keyPem, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, nil, err
	}

	return certPem.Bytes(), keyPem.Bytes(), nil
}

// LoadOrGenerateCertificate loads a certificate and key from disk if they exist,
// otherwise generates a new self-signed certificate and saves it to disk.
func LoadOrGenerateCertificate(certPath, keyPath string, opts CertificateOptions) ([]byte, []byte, error) {
	certData, certErr := os.ReadFile(certPath)
	keyData, keyErr := os.ReadFile(keyPath)

	if certErr == nil && keyErr == nil {
		return certData, keyData, nil
	}

	certData, keyData, err := GenerateSelfsigned(opts)
	if err != nil {
		return nil, nil, err
	}

	if err := os.WriteFile(certPath, certData, 0o600); err != nil {
		return nil, nil, err
	}

	if err := os.WriteFile(keyPath, keyData, 0o600); err != nil {
		return nil, nil, err
	}

	return certData, keyData, nil
}
