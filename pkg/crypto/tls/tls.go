package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// NewTLSConfig returns a TLS configuration from the given certificate file.
// change arg to io.Reader ?
func NewTLSConfig(pemCerts []byte) (*tls.Config, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	// NOTE: we're testing a special case of "-" meaning use TLS, but just use
	// system CAs without appending a known server certificate. We may change
	// this or formally document it.
	if pemCerts != nil {
		if !rootCAs.AppendCertsFromPEM(pemCerts) {
			return nil, fmt.Errorf("credentials: failed to append certificates")
		}
	}
	return &tls.Config{
		// For proper verification of the server-provided certificate chain
		// during the TLS handshake, the root CAs, which may contain a custom
		// certificate we appended above, are used by the client tls.Conn. If we
		// disable verification with InsecureSkipVerify, the connection is still
		// encrypted, but we cannot ensure the server is who they claim to be.
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}, nil
}
