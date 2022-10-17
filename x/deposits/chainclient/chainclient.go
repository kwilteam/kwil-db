package chainclient

import (
	"errors"

	"kwil/x/deposits/chainclient/types"

	"kwil/x/deposits/chainclient/ethclient"
)

type clientBuilder struct {
	chainCode string
	endpoint  string
}

type ClientBuilder interface {
	Build() (types.Client, error)
	Chain(chainCode string) ClientBuilder
	Endpoint(endpoint string) ClientBuilder
}

var ErrChainNotSpecified = errors.New("chain not specified")

func (c *clientBuilder) Build() (types.Client, error) {
	switch c.chainCode {
	case "eth-mainnet":
		return ethclient.New(c.endpoint, c.chainCode)
	default:
		return nil, ErrChainNotSpecified
	}
}

func (c *clientBuilder) Chain(chainCode string) ClientBuilder {
	c.chainCode = chainCode
	return c
}

func (c *clientBuilder) Endpoint(endpoint string) ClientBuilder {
	c.endpoint = endpoint
	return c
}

func Builder() ClientBuilder {
	return &clientBuilder{}
}

/*
	Below are the supported chains.  We will add more as needed.

	Supported Chains | Code
	-----------------+------
	Ethereum         | eth-mainnet
	-----------------+------

*/
