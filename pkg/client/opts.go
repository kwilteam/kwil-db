package client

import (
	"github.com/kwilteam/kwil-db/pkg/auth"
	"github.com/kwilteam/kwil-db/pkg/client/types"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type Option func(*Client)

func WithLogger(logger log.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithSigner(signer auth.Signer) Option {
	return func(c *Client) {
		c.Signer = signer
	}
}

func WithTLSCert(certFile string) Option {
	return func(c *Client) {
		c.tlsCertFile = certFile
	}
}

func WithTransportClient(tc types.TransportClient) Option {
	return func(c *Client) {
		c.transportClient = tc
	}
}

type callOptions struct {
	// forceAuthenticated is used to force the client to authenticate
	// if nil, the client will use the default value
	// if false, it will not authenticate
	// if true, it will authenticate
	forceAuthenticated *bool // is pointer necessary here?
}

type CallOpt func(*callOptions)

// Authenticated can be used to force the client to authenticate (or not)
// if true, the client will authenticate. if false, it will not authenticate
// if nil, the client will decide itself
func Authenticated(shouldSign bool) CallOpt {
	return func(o *callOptions) {
		copied := shouldSign
		o.forceAuthenticated = &copied
	}
}
