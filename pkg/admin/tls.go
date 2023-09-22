package admin

import (
	"crypto/tls"

	"github.com/kwilteam/kwil-db/internal/pkg/transport"
)

// newAuthenticatedTLSConfig creates a new tls.Config for an
// mutually-authenticated TLS (mTLS) client. In addition to the server's
// certificate file, the client's own key pair is required to support protocol
// level client authentication.
func newAuthenticatedTLSConfig(kwildCertFile, clientCertFile, clientKeyFile string) (*tls.Config, error) {
	cfg, err := transport.NewClientTLSConfigFromFile(kwildCertFile)
	if err != nil {
		return nil, err
	}

	authCert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return nil, err
	}
	cfg.Certificates = append(cfg.Certificates, authCert)

	return cfg, nil
}
