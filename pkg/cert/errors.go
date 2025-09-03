// SPDX-License-Identifier: BSD-3-Clause

package cert

import "errors"

var (
	// ErrGenerateRSAKey indicates a failure during RSA private key generation.
	ErrGenerateRSAKey = errors.New("failed to generate RSA private key")
	// ErrGenerateSerialNumber indicates a failure to generate a random serial number for the certificate.
	ErrGenerateSerialNumber = errors.New("failed to generate certificate serial number")
	// ErrCreateCertificate indicates a failure during X.509 certificate creation.
	ErrCreateCertificate = errors.New("failed to create X.509 certificate")
	// ErrEncodeCertificatePEM indicates a failure to encode the certificate in PEM format.
	ErrEncodeCertificatePEM = errors.New("failed to encode certificate to PEM")
	// ErrMarshalPrivateKey indicates a failure to marshal the private key to PKCS#8 format.
	ErrMarshalPrivateKey = errors.New("failed to marshal private key to PKCS#8")
	// ErrEncodePrivateKeyPEM indicates a failure to encode the private key in PEM format.
	ErrEncodePrivateKeyPEM = errors.New("failed to encode private key to PEM")
	// ErrReadCertificateFile indicates a failure to read the certificate file from disk.
	ErrReadCertificateFile = errors.New("failed to read certificate file")
	// ErrReadPrivateKeyFile indicates a failure to read the private key file from disk.
	ErrReadPrivateKeyFile = errors.New("failed to read private key file")
	// ErrWriteCertificateFile indicates a failure to write the certificate file to disk.
	ErrWriteCertificateFile = errors.New("failed to write certificate file")
	// ErrWritePrivateKeyFile indicates a failure to write the private key file to disk.
	ErrWritePrivateKeyFile = errors.New("failed to write private key file")
	// ErrCreateCertificateDirectory indicates a failure to create the certificate directory.
	ErrCreateCertificateDirectory = errors.New("failed to create certificate directory")
	// ErrCreatePrivateKeyDirectory indicates a failure to create the private key directory.
	ErrCreatePrivateKeyDirectory = errors.New("failed to create private key directory")
	// ErrParseTLSKeyPair indicates a failure to parse the certificate and private key into a TLS key pair.
	ErrParseTLSKeyPair = errors.New("failed to parse TLS key pair")
	// ErrInvalidCertificateOptions indicates that the provided certificate options are invalid.
	ErrInvalidCertificateOptions = errors.New("invalid certificate options")
	// ErrAutocertSetup indicates a failure during Let's Encrypt autocert configuration.
	ErrAutocertSetup = errors.New("failed to setup autocert")
	// ErrInvalidHostname indicates that the provided hostname is invalid for certificate generation.
	ErrInvalidHostname = errors.New("invalid hostname for certificate")
	// ErrInvalidEmail indicates that the provided email address is invalid for Let's Encrypt registration.
	ErrInvalidEmail = errors.New("invalid email address for Let's Encrypt registration")
	// ErrCacheDirectory indicates a failure related to the certificate cache directory.
	ErrCacheDirectory = errors.New("certificate cache directory error")
)
