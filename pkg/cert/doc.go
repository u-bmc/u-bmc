// SPDX-License-Identifier: BSD-3-Clause

// Package cert provides comprehensive X.509 certificate generation and management
// capabilities for TLS/SSL connections. It supports both self-signed certificates
// for development and testing environments, and Let's Encrypt ACME certificates
// for production use with automatic renewal.
//
// The package is designed with a flexible configuration system using the options
// pattern, allowing fine-grained control over certificate properties including
// subject information, alternative names, key sizes, validity periods, and
// certificate authority capabilities.
//
// # Certificate Types
//
// The package supports two primary certificate types:
//
//   - Self-signed certificates: Generated locally using RSA key pairs, suitable
//     for development, testing, or isolated environments where certificate
//     authority validation is not required.
//
//   - Let's Encrypt certificates: Automatically obtained from Let's Encrypt
//     using the ACME protocol, providing trusted certificates for production
//     use with automatic renewal capabilities.
//
// # Basic Usage
//
// For simple self-signed certificate generation:
//
//	cfg := cert.NewConfig(
//		cert.WithHostname("localhost"),
//		cert.WithAlternativeNames("127.0.0.1", "::1"),
//		cert.WithOrganization("My Company"),
//	)
//
//	certData, keyData, err := cert.GenerateSelfsigned(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use certData and keyData for TLS configuration
//
// For Let's Encrypt certificates:
//
//	cfg := cert.NewConfig(
//		cert.WithCertificateType(cert.CertificateTypeLetsEncrypt),
//		cert.WithHostname("example.com"),
//		cert.WithEmail("admin@example.com"),
//		cert.WithCacheDir("/var/cache/letsencrypt"),
//		cert.WithAcceptTOS(true),
//	)
//
//	tlsConfig, httpHandler, err := cert.GenerateAndSign(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use tlsConfig for HTTPS server and httpHandler for ACME challenges
//
// # Advanced Configuration
//
// The package supports extensive customization through configuration options:
//
//	cfg := cert.NewConfig(
//		cert.WithCertificateType(cert.CertificateTypeSelfSigned),
//		cert.WithHostname("bmc.example.com"),
//		cert.WithAlternativeNames(
//			"bmc.local",
//			"192.168.1.100",
//			"fd12:3456:789a:1::100",
//		),
//		cert.WithOrganization("Hardware Management Corp"),
//		cert.WithOrganizationalUnit("BMC Division"),
//		cert.WithCountry("US"),
//		cert.WithProvince("California"),
//		cert.WithLocality("San Francisco"),
//		cert.WithKeySize(4096),
//		cert.WithValidityPeriod(365*24*time.Hour), // 1 year
//		cert.WithIsCA(false),
//		cert.WithCertPath("/etc/ssl/certs/bmc.crt"),
//		cert.WithKeyPath("/etc/ssl/private/bmc.key"),
//	)
//
//	if err := cfg.Validate(); err != nil {
//		log.Fatalf("Invalid configuration: %v", err)
//	}
//
//	certData, keyData, err := cert.GenerateSelfsigned(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Persistent Certificate Management
//
// For applications that need persistent certificates with automatic generation:
//
//	cfg := cert.NewConfig(
//		cert.WithHostname("bmc.local"),
//		cert.WithCertPath("/var/lib/bmc/tls/cert.pem"),
//		cert.WithKeyPath("/var/lib/bmc/tls/key.pem"),
//	)
//
//	// This will load existing certificates or generate new ones if they don't exist
//	certData, keyData, err := cert.LoadOrGenerateCertificate(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Parse into tls.Certificate for use with HTTP servers
//	tlsCert, err := tls.X509KeyPair(certData, keyData)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	server := &http.Server{
//		Addr: ":8443",
//		TLSConfig: &tls.Config{
//			Certificates: []tls.Certificate{tlsCert},
//		},
//	}
//
// # Certificate Authority Usage
//
// The package can generate CA certificates for signing other certificates:
//
//	caCfg := cert.NewConfig(
//		cert.WithHostname("BMC Root CA"),
//		cert.WithOrganization("Hardware Management Corp"),
//		cert.WithIsCA(true),
//		cert.WithValidityPeriod(10*365*24*time.Hour), // 10 years
//		cert.WithKeySize(4096),
//	)
//
//	caCertData, caKeyData, err := cert.GenerateSelfsigned(caCfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// The CA certificate can now be used to sign other certificates
//
// # Error Handling
//
// The package defines specific error types for different failure scenarios:
//
//	certData, keyData, err := cert.GenerateSelfsigned(cfg)
//	if err != nil {
//		switch {
//		case errors.Is(err, cert.ErrInvalidCertificateOptions):
//			log.Printf("Configuration error: %v", err)
//		case errors.Is(err, cert.ErrGenerateRSAKey):
//			log.Printf("Key generation failed: %v", err)
//		case errors.Is(err, cert.ErrCreateCertificate):
//			log.Printf("Certificate creation failed: %v", err)
//		default:
//			log.Printf("Unexpected error: %v", err)
//		}
//		return
//	}
//
// # Security Considerations
//
// When using this package in production:
//
//   - Use RSA key sizes of at least 2048 bits (4096 bits recommended for CA certificates)
//   - Set appropriate file permissions (0600) for private key files
//   - Use Let's Encrypt certificates for public-facing services
//   - Regularly rotate certificates and monitor expiration dates
//   - Validate certificate configurations before use
//   - Store CA private keys securely and offline when possible
//
// # Integration with HTTP Servers
//
// Complete example of integrating with an HTTP server:
//
//	func setupHTTPSServer() error {
//		cfg := cert.NewConfig(
//			cert.WithHostname("bmc.local"),
//			cert.WithAlternativeNames("192.168.1.100"),
//			cert.WithOrganization("BMC Service"),
//		)
//
//		certData, keyData, err := cert.LoadOrGenerateCertificate(cfg)
//		if err != nil {
//			return fmt.Errorf("failed to get certificate: %w", err)
//		}
//
//		tlsCert, err := tls.X509KeyPair(certData, keyData)
//		if err != nil {
//			return fmt.Errorf("failed to parse certificate: %w", err)
//		}
//
//		mux := http.NewServeMux()
//		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//			fmt.Fprintf(w, "Secure BMC interface")
//		})
//
//		server := &http.Server{
//			Addr:    ":8443",
//			Handler: mux,
//			TLSConfig: &tls.Config{
//				Certificates: []tls.Certificate{tlsCert},
//				MinVersion:   tls.VersionTLS12,
//			},
//		}
//
//		log.Println("Starting HTTPS server on :8443")
//		return server.ListenAndServeTLS("", "")
//	}
package cert
