package client

import (
	"crypto/ecdsa"

	chainCodes "github.com/kwilteam/kwil-db/pkg/chain/types"
)

type ClientOpt func(*Client)

func WithPrivateKey(key *ecdsa.PrivateKey) ClientOpt {
	return func(c *Client) {
		c.PrivateKey = key
	}
}

func WithChainCode(chainCode int32) ClientOpt {
	return func(c *Client) {
		c.ChainCode = chainCodes.ChainCode(chainCode)
	}
}

func WithProviderAddress(address string) ClientOpt {
	return func(c *Client) {
		c.ProviderAddress = address
	}
}

func WithPoolAddress(address string) ClientOpt {
	return func(c *Client) {
		c.PoolAddress = address
	}
}

func WithChainRpcUrl(url string) ClientOpt {
	return func(c *Client) {
		c.chainRpcUrl = url
	}
}

func WithoutProvider() ClientOpt {
	return func(c *Client) {
		c.usingProvider = false
	}
}

func WithoutServiceConfig() ClientOpt {
	return func(c *Client) {
		c.withServerConfig = false
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

func WithBcRpcUrl(url string) ClientOpt {
	return func(c *Client) {
		c.BcRpcUrl = url
	}
}

