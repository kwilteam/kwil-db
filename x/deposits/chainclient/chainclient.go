package chainclient

import (
	"errors"

	"kwil/x/deposits/types"
	"kwil/x/logx"

	"kwil/x/deposits/chainclient/evmclient"
)

type clientBuilder struct {
	chainCode string
	endpoint  string
	logger    logx.Logger
}

type ClientBuilder interface {
	Build() (types.Client, error)
	Chain(chainCode string) ClientBuilder
	Endpoint(endpoint string) ClientBuilder
	Logger(l logx.Logger) ClientBuilder
}

var ErrChainNotSpecified = errors.New("chain not specified")

func (c *clientBuilder) Build() (types.Client, error) {
	switch c.chainCode {
	case "eth-mainnet":
		return evmclient.New(c.logger, c.endpoint, c.chainCode)
	case "eth-goerli": // for now it is the same as eth mainnet
		return evmclient.New(c.logger, c.endpoint, c.chainCode)
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

func (c *clientBuilder) Logger(l logx.Logger) ClientBuilder {
	c.logger = l
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
	Goerli           | eth-goerli
	-----------------+------

*/
