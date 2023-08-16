package client

import "github.com/kwilteam/kwil-db/pkg/crypto"

type ClientOpt func(*Client)

func WithSigner(signer crypto.Signer) ClientOpt {
	return func(c *Client) {
		c.Signer = signer
	}
}

// TODO: replace this, since we should not be using cometBFT RPCs
func WithCometBftUrl(url string) ClientOpt {
	return func(c *Client) {
		c.cometBftRpcUrl = url
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
