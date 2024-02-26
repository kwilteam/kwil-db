package client

import (
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type ClientOpt func(*Client)

func WithLogger(logger log.Logger) ClientOpt {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithSigner(signer crypto.Signer) ClientOpt {
	return func(c *Client) {
		c.Signer = signer
	}
}

func WithTLSCert(certFile string) ClientOpt {
	return func(c *Client) {
		c.certFile = certFile
	}
}

type callOptions struct {
	// forceAuthenticated is used to force the client to authenticate
	// if nil, the client will use the default value
	// if false, it will not authenticate
	// if true, it will authenticate
	forceAuthenticated *bool
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
