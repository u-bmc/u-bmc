// SPDX-License-Identifier: BSD-3-Clause

package cert

import (
	"fmt"
	"net"
	"net/mail"
	"slices"
	"strings"
	"time"
)

// CertificateType represents the type of certificate to generate or use.
type CertificateType int

const (
	// CertificateTypeSelfSigned generates a self-signed certificate.
	CertificateTypeSelfSigned CertificateType = iota
	// CertificateTypeLetsTencrypt uses Let's Encrypt to obtain a certificate.
	CertificateTypeLetsTencrypt
)

// Config holds the configuration for certificate generation and management.
type Config struct {
	// Type specifies the type of certificate to use
	Type CertificateType
	// Hostname is the primary hostname for the certificate
	Hostname string
	// AlternativeNames are additional hostnames and IP addresses for the certificate
	AlternativeNames []string
	// Organization is the organization name for the certificate subject
	Organization string
	// OrganizationalUnit is the organizational unit for the certificate subject
	OrganizationalUnit string
	// Country is the country code for the certificate subject
	Country string
	// Province is the state or province for the certificate subject
	Province string
	// Locality is the city or locality for the certificate subject
	Locality string
	// Email is the email address for Let's Encrypt registration or certificate contact
	Email string
	// KeySize is the RSA key size in bits
	KeySize int
	// ValidityPeriod is how long the certificate should be valid
	ValidityPeriod time.Duration
	// NotBeforeOffset is how far back the certificate validity should start (to account for clock skew)
	NotBeforeOffset time.Duration
	// IsCA indicates whether this certificate should be a Certificate Authority
	IsCA bool
	// CertPath is the file path where the certificate will be stored
	CertPath string
	// KeyPath is the file path where the private key will be stored
	KeyPath string
	// CacheDir is the directory for Let's Encrypt certificate cache
	CacheDir string
	// AutoRenew enables automatic certificate renewal for Let's Encrypt
	AutoRenew bool
	// AcceptTOS automatically accepts Let's Encrypt Terms of Service
	AcceptTOS bool
}

// Option represents a configuration option for certificate management.
type Option interface {
	apply(*Config)
}

type certificateTypeOption struct {
	certType CertificateType
}

func (o *certificateTypeOption) apply(c *Config) {
	c.Type = o.certType
}

// WithCertificateType sets the type of certificate to generate or obtain.
func WithCertificateType(certType CertificateType) Option {
	return &certificateTypeOption{
		certType: certType,
	}
}

type hostnameOption struct {
	hostname string
}

func (o *hostnameOption) apply(c *Config) {
	c.Hostname = o.hostname
}

// WithHostname sets the primary hostname for the certificate.
// This will be used as the Common Name and added to Subject Alternative Names.
func WithHostname(hostname string) Option {
	return &hostnameOption{
		hostname: hostname,
	}
}

type alternativeNamesOption struct {
	altNames []string
}

func (o *alternativeNamesOption) apply(c *Config) {
	c.AlternativeNames = o.altNames
}

// WithAlternativeNames sets additional hostnames and IP addresses for the certificate.
// These will be added to the Subject Alternative Names extension.
func WithAlternativeNames(altNames ...string) Option {
	return &alternativeNamesOption{
		altNames: altNames,
	}
}

type organizationOption struct {
	organization string
}

func (o *organizationOption) apply(c *Config) {
	c.Organization = o.organization
}

// WithOrganization sets the organization name in the certificate subject.
func WithOrganization(organization string) Option {
	return &organizationOption{
		organization: organization,
	}
}

type organizationalUnitOption struct {
	ou string
}

func (o *organizationalUnitOption) apply(c *Config) {
	c.OrganizationalUnit = o.ou
}

// WithOrganizationalUnit sets the organizational unit in the certificate subject.
func WithOrganizationalUnit(ou string) Option {
	return &organizationalUnitOption{
		ou: ou,
	}
}

type countryOption struct {
	country string
}

func (o *countryOption) apply(c *Config) {
	c.Country = o.country
}

// WithCountry sets the country code in the certificate subject (e.g., "US", "CA", "GB").
func WithCountry(country string) Option {
	return &countryOption{
		country: country,
	}
}

type provinceOption struct {
	province string
}

func (o *provinceOption) apply(c *Config) {
	c.Province = o.province
}

// WithProvince sets the state or province in the certificate subject.
func WithProvince(province string) Option {
	return &provinceOption{
		province: province,
	}
}

type localityOption struct {
	locality string
}

func (o *localityOption) apply(c *Config) {
	c.Locality = o.locality
}

// WithLocality sets the city or locality in the certificate subject.
func WithLocality(locality string) Option {
	return &localityOption{
		locality: locality,
	}
}

type emailOption struct {
	email string
}

func (o *emailOption) apply(c *Config) {
	c.Email = o.email
}

// WithEmail sets the email address for Let's Encrypt registration or certificate contact information.
func WithEmail(email string) Option {
	return &emailOption{
		email: email,
	}
}

type keySizeOption struct {
	keySize int
}

func (o *keySizeOption) apply(c *Config) {
	c.KeySize = o.keySize
}

// WithKeySize sets the RSA key size in bits. Common values are 2048, 3072, and 4096.
func WithKeySize(keySize int) Option {
	return &keySizeOption{
		keySize: keySize,
	}
}

type validityPeriodOption struct {
	period time.Duration
}

func (o *validityPeriodOption) apply(c *Config) {
	c.ValidityPeriod = o.period
}

// WithValidityPeriod sets how long the certificate should be valid.
// This only applies to self-signed certificates.
func WithValidityPeriod(period time.Duration) Option {
	return &validityPeriodOption{
		period: period,
	}
}

type notBeforeOffsetOption struct {
	offset time.Duration
}

func (o *notBeforeOffsetOption) apply(c *Config) {
	c.NotBeforeOffset = o.offset
}

// WithNotBeforeOffset sets how far back the certificate validity should start.
// This helps account for clock skew between systems.
func WithNotBeforeOffset(offset time.Duration) Option {
	return &notBeforeOffsetOption{
		offset: offset,
	}
}

type isCAOption struct {
	isCA bool
}

func (o *isCAOption) apply(c *Config) {
	c.IsCA = o.isCA
}

// WithIsCA sets whether this certificate should be a Certificate Authority.
// CA certificates can sign other certificates.
func WithIsCA(isCA bool) Option {
	return &isCAOption{
		isCA: isCA,
	}
}

type certPathOption struct {
	certPath string
}

func (o *certPathOption) apply(c *Config) {
	c.CertPath = o.certPath
}

// WithCertPath sets the file path where the certificate will be stored.
func WithCertPath(certPath string) Option {
	return &certPathOption{
		certPath: certPath,
	}
}

type keyPathOption struct {
	keyPath string
}

func (o *keyPathOption) apply(c *Config) {
	c.KeyPath = o.keyPath
}

// WithKeyPath sets the file path where the private key will be stored.
func WithKeyPath(keyPath string) Option {
	return &keyPathOption{
		keyPath: keyPath,
	}
}

type cacheDirOption struct {
	cacheDir string
}

func (o *cacheDirOption) apply(c *Config) {
	c.CacheDir = o.cacheDir
}

// WithCacheDir sets the directory for Let's Encrypt certificate cache.
func WithCacheDir(cacheDir string) Option {
	return &cacheDirOption{
		cacheDir: cacheDir,
	}
}

type autoRenewOption struct {
	autoRenew bool
}

func (o *autoRenewOption) apply(c *Config) {
	c.AutoRenew = o.autoRenew
}

// WithAutoRenew enables or disables automatic certificate renewal for Let's Encrypt.
func WithAutoRenew(autoRenew bool) Option {
	return &autoRenewOption{
		autoRenew: autoRenew,
	}
}

type acceptTOSOption struct {
	acceptTOS bool
}

func (o *acceptTOSOption) apply(c *Config) {
	c.AcceptTOS = o.acceptTOS
}

// WithAcceptTOS sets whether to automatically accept Let's Encrypt Terms of Service.
func WithAcceptTOS(acceptTOS bool) Option {
	return &acceptTOSOption{
		acceptTOS: acceptTOS,
	}
}

// NewConfig creates a new Config with sane defaults and applies the provided options.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Type:               CertificateTypeSelfSigned,
		Hostname:           "localhost",
		AlternativeNames:   []string{},
		Organization:       "u-bmc",
		OrganizationalUnit: "BMC",
		Country:            "",
		Province:           "",
		Locality:           "",
		Email:              "",
		KeySize:            2048,
		ValidityPeriod:     365 * 24 * time.Hour * 30, // 30 years
		NotBeforeOffset:    30 * time.Second,
		IsCA:               false,
		CertPath:           "/var/cache/cert/cert.pem",
		KeyPath:            "/var/cache/cert/key.pem",
		CacheDir:           "/var/cache/cert",
		AutoRenew:          true,
		AcceptTOS:          false,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	if c.Hostname == "" {
		return fmt.Errorf("%w: hostname cannot be empty", ErrInvalidCertificateOptions)
	}

	// Validate hostname format
	if !isValidHostname(c.Hostname) {
		return fmt.Errorf("%w: invalid hostname format: %s", ErrInvalidHostname, c.Hostname)
	}

	// Validate alternative names
	for _, altName := range c.AlternativeNames {
		if !isValidHostname(altName) && !isValidIP(altName) {
			return fmt.Errorf("%w: invalid alternative name: %s", ErrInvalidHostname, altName)
		}
	}

	// Validate key size
	if c.KeySize < 2048 {
		return fmt.Errorf("%w: key size must be at least 2048 bits", ErrInvalidCertificateOptions)
	}

	// Validate validity period for self-signed certificates
	if c.Type == CertificateTypeSelfSigned && c.ValidityPeriod <= 0 {
		return fmt.Errorf("%w: validity period must be positive", ErrInvalidCertificateOptions)
	}

	// Validate Let's Encrypt specific options
	if c.Type == CertificateTypeLetsTencrypt {
		if c.Email == "" {
			return fmt.Errorf("%w: email is required for Let's Encrypt", ErrInvalidEmail)
		}
		if !isValidEmail(c.Email) {
			return fmt.Errorf("%w: invalid email format: %s", ErrInvalidEmail, c.Email)
		}
		if c.CacheDir == "" {
			return fmt.Errorf("%w: cache directory is required for Let's Encrypt", ErrCacheDirectory)
		}
	}

	// Validate file paths
	if c.CertPath == "" {
		return fmt.Errorf("%w: certificate path cannot be empty", ErrInvalidCertificateOptions)
	}
	if c.KeyPath == "" {
		return fmt.Errorf("%w: key path cannot be empty", ErrInvalidCertificateOptions)
	}

	// Validate country code format (if provided)
	if c.Country != "" && len(c.Country) != 2 {
		return fmt.Errorf("%w: country code must be 2 characters", ErrInvalidCertificateOptions)
	}

	return nil
}

// GetAllHostnames returns all hostnames (primary + alternatives) for the certificate.
func (c *Config) GetAllHostnames() []string {
	hostnames := []string{c.Hostname}
	for _, altName := range c.AlternativeNames {
		// Only add if it's a hostname (not an IP) and not already in the list
		if isValidHostname(altName) && !slices.Contains(hostnames, altName) {
			hostnames = append(hostnames, altName)
		}
	}
	return hostnames
}

// GetAllIPs returns all IP addresses from the alternative names.
func (c *Config) GetAllIPs() []net.IP {
	var ips []net.IP
	for _, altName := range c.AlternativeNames {
		if ip := net.ParseIP(altName); ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips
}

// isValidHostname checks if a string is a valid hostname.
func isValidHostname(hostname string) bool {
	if hostname == "" || len(hostname) > 253 {
		return false
	}

	// Remove trailing dot if present
	hostname = strings.TrimSuffix(hostname, ".")

	// Split into labels
	for label := range strings.SplitSeq(hostname, ".") {
		if len(label) == 0 || len(label) > 63 {
			return false
		}

		// Label must start and end with alphanumeric character
		if !isAlphanumeric(label[0]) || !isAlphanumeric(label[len(label)-1]) {
			return false
		}

		// Label can only contain alphanumeric characters and hyphens
		for _, char := range label {
			if !isAlphanumeric(byte(char)) && char != '-' {
				return false
			}
		}
	}

	return true
}

// isValidIP checks if a string is a valid IP address.
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isValidEmail checks if a string is a valid email address.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// isAlphanumeric checks if a byte is alphanumeric.
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
