package client

import (
	"crypto/ecdsa"
	chainCodes "kwil/pkg/chain/types"
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
