// Package admin provides an client for communicating with an authenticated
// administrative gRPC service on a running kwild instance.
package admin

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	admClient "github.com/kwilteam/kwil-db/pkg/grpc/client/admin/v0"
	"github.com/kwilteam/kwil-db/pkg/log"

	"go.uber.org/zap"
)

// Client is performs node administrative actions via the authenticated gRPC
// service on a running kwild node.
type Client struct {
	client *admClient.AdminClient
	signer crypto.Signer // for use in methods that require signing a transaction with a Kwil account
	logger log.Logger
}

// New creates a new admin client. TLS is required so the kwild TLS certificate
// is required. Authentication is done at the protocol level (mTLS), so our own
// key pair is also required. The server must have our client certificate loaded
// in it's own tls.Config.ClientCAs. This client keypair can be thought of as a
// preshared key (like a password or token), but handled automatically by the
// TLS handshake, thus requiring no application level logic such as transmitting
// a pass/token with each request.
func New(host string, kwildCertFile, clientKeyFile, clientCertFile string, opts ...ClientOpt) (c *Client, err error) {
	c = &Client{
		logger: log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(c)
	}

	tlsConfig, err := newAuthenticatedTLSConfig(kwildCertFile, clientCertFile, clientKeyFile)
	if err != nil {
		return nil, err
	}
	c.client, err = admClient.New(host, tlsConfig)
	if err != nil {
		return nil, err
	}

	c.logger = *c.logger.Named("admin").With(zap.String("host", host))

	return c, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	return c.client.Ping(ctx)
}

func (c *Client) Version(ctx context.Context) (string, error) {
	return c.client.Version(ctx)
}

/* TODO: validator actions that work via server-side transaction authoring
   rather that client-side authoring followed by broadcast via the public tx
   service.

func (c *Client) ApproveValidator(ctx context.Context, joiner []byte) ([]byte, error) {
	_, err := crypto.Ed25519PublicKeyFromBytes(joiner) if err != nil {
    	return nil, fmt.Errorf("invalid candidate validator public key: %w", err)
    }

	...
}

func (c *Client) ValidatorJoin(ctx context.Context) ([]byte, error) {
	...
}

func (c *Client) ValidatorLeave(ctx context.Context) ([]byte, error) { return
    ...
}

*/
