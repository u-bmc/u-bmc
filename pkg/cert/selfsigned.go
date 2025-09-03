// SPDX-License-Identifier: BSD-3-Clause

package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// GenerateSelfsigned creates a self-signed certificate using the provided configuration.
// It returns the certificate and private key as byte slices, or an error if any step of the generation fails.
func GenerateSelfsigned(cfg *Config) ([]byte, []byte, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrInvalidCertificateOptions, err)
	}

	// Generate RSA key pair
	priv, err := rsa.GenerateKey(rand.Reader, cfg.KeySize)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrGenerateRSAKey, err)
	}
	pub := &priv.PublicKey

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrGenerateSerialNumber, err)
	}

	notBefore := time.Now().Add(-cfg.NotBeforeOffset)
	notAfter := notBefore.Add(cfg.ValidityPeriod)

	subject := pkix.Name{
		CommonName:   cfg.Hostname,
		Organization: []string{cfg.Organization},
	}

	if cfg.OrganizationalUnit != "" {
		subject.OrganizationalUnit = []string{cfg.OrganizationalUnit}
	}

	if cfg.Country != "" {
		subject.Country = []string{cfg.Country}
	}

	if cfg.Province != "" {
		subject.Province = []string{cfg.Province}
	}

	if cfg.Locality != "" {
		subject.Locality = []string{cfg.Locality}
	}

	// Build DNS names and IP addresses
	dnsNames := cfg.GetAllHostnames()
	ipAddresses := cfg.GetAllIPs()

	// Set key usage based on whether this is a CA certificate
	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	if cfg.IsCA {
		keyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	template := &x509.Certificate{
		Subject:     subject,
		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:              keyUsage,
		BasicConstraintsValid: true,
		SerialNumber:          sn,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  cfg.IsCA,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrCreateCertificate, err)
	}

	var certPem bytes.Buffer
	if err := pem.Encode(&certPem, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrEncodeCertificatePEM, err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrMarshalPrivateKey, err)
	}

	var keyPem bytes.Buffer
	if err := pem.Encode(&keyPem, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrEncodePrivateKeyPEM, err)
	}

	return certPem.Bytes(), keyPem.Bytes(), nil
}

// LoadOrGenerateCertificate loads a certificate and key from disk if they exist,
// otherwise generates a new self-signed certificate and saves it to disk using the provided configuration.
func LoadOrGenerateCertificate(cfg *Config) ([]byte, []byte, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrInvalidCertificateOptions, err)
	}

	certData, certErr := os.ReadFile(cfg.CertPath)
	keyData, keyErr := os.ReadFile(cfg.KeyPath)

	if certErr == nil && keyErr == nil {
		return certData, keyData, nil
	}

	certData, keyData, err := GenerateSelfsigned(cfg)
	if err != nil {
		return nil, nil, err
	}

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(cfg.CertPath), 0o755); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrCreateCertificateDirectory, err)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.KeyPath), 0o755); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrCreatePrivateKeyDirectory, err)
	}

	if err := os.WriteFile(cfg.CertPath, certData, 0o600); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrWriteCertificateFile, err)
	}

	if err := os.WriteFile(cfg.KeyPath, keyData, 0o600); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrWritePrivateKeyFile, err)
	}

	return certData, keyData, nil
}
