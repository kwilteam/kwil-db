// Package admin provides a client for communicating with an authenticated
// administrative gRPC service on a running kwild instance. This is presently to
// be used by kwil-admin, but it could be made part of our public API, perhaps
// in the client (SDK) package, once fleshed out a little more.
package admin

import (
	"context"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	admClient "github.com/kwilteam/kwil-db/core/rpc/client/admin"
	types "github.com/kwilteam/kwil-db/core/types/admin"

	"go.uber.org/zap"
)

// Client is performs node administrative actions via the authenticated gRPC
// service on a running kwild node.
type Client struct {
	*client.Client
	client *admClient.AdminClient
	signer auth.Signer // for use in methods that require signing a transaction with a Kwil account
	logger log.Logger
}

// New creates a new admin TCP client. TLS is required so the kwild TLS certificate
// is required. Authentication is done at the protocol level (mTLS), so our own
// key pair is also required. The server must have our client certificate loaded
// in it's own tls.Config.ClientCAs. This client keypair can be thought of as a
// preshared key (like a password or token), but handled automatically by the
// TLS handshake, thus requiring no application level logic such as transmitting
// a pass/token with each request.
func New(ctx context.Context, host string, kwildCertFile, clientKeyFile, clientCertFile string, opts ...ClientOpt) (c *Client, err error) {
	c = &Client{
		logger: log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Client, err = client.Dial(ctx, host, client.WithLogger(c.logger), client.WithSigner(c.signer))

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

func (c *Client) Version(ctx context.Context) (string, error) {
	return c.client.Version(ctx)
}

func (c *Client) Status(ctx context.Context) (*types.Status, error) {
	return c.client.Status(ctx)
}

func (c *Client) Peers(ctx context.Context) ([]*types.PeerInfo, error) {
	return c.client.Peers(ctx)
}
