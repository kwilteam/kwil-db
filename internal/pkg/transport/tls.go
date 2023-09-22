// Package transport provides transport related functions to other kwil-db
// packages to help configure and use TLS clients and services.
package transport

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/decred/dcrd/certgen"
)

// DefaultClientTLSConfig returns a client tls config using just the system's CA
// pool, if present, otherwise an empty certificate pool.
func DefaultClientTLSConfig() *tls.Config {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	return &tls.Config{
		// For proper verification of the server-provided certificate chain
		// during the TLS handshake, the root CAs, which may contain a custom
		// certificate we append, are used by the client tls.Conn. If we disable
		// verification with InsecureSkipVerify, the connection is still
		// encrypted, but we cannot ensure the server is who they claim to be.
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}
}

// NewClientTLSConfigFromFile creates a new basic tls.Config for a TLS client
// given the path to a PEM encoded certificate file to include in the root CAs.
// Provide the server's certificate or it's root certificate authority. Use
// DefaultClientTLSConfig to use just the system's CA pool.
func NewClientTLSConfigFromFile(certFile string) (*tls.Config, error) {
	pemCerts, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %w", err)
	}
	return NewClientTLSConfig(pemCerts)
}

// NewClientTLSConfig creates a new basic tls.Config for a TLS client. This is
// not appropriate for a server's config. Provide the server's certificate or
// it's root certificate authority.
func NewClientTLSConfig(pemCerts []byte) (*tls.Config, error) {
	cfg := DefaultClientTLSConfig()
	if !cfg.RootCAs.AppendCertsFromPEM(pemCerts) {
		return nil, errors.New("credentials: failed to append certificates")
	}
	return cfg, nil
}

// GenTLSKeyPair generates a key/cert pair to the paths provided.
func GenTLSKeyPair(certFile, keyFile string, org string, altDNSNames []string) error {
	validUntil := time.Now().Add(10 * 365 * 24 * time.Hour)
	cert, key, err := certgen.NewEd25519TLSCertPair(org,
		validUntil, altDNSNames)
	if err != nil {
		return err
	}

	if err = os.WriteFile(certFile, cert, 0644); err != nil {
		return err
	}
	if err = os.WriteFile(keyFile, key, 0600); err != nil {
		os.Remove(certFile)
		return err
	}

	return nil
}
