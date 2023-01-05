package evm

import (
	"fmt"
	"kwil/x/chain"

	ethereumclient "github.com/ethereum/go-ethereum/ethclient"
)

// client fulfills the dto.ChainProvider interface
type ethClient struct {
	ethclient *ethereumclient.Client
	chainCode chain.ChainCode
	endpoint  string
}

func New(endpoint string, chainCode chain.ChainCode) (*ethClient, error) {
	client, err := ethereumclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %v", err)
	}

	return &ethClient{
		ethclient: client,
		chainCode: chainCode,
		endpoint:  endpoint,
	}, nil
}

func (c *ethClient) ChainCode() chain.ChainCode {
	return c.chainCode
}

func (c *ethClient) AsEthClient() (*ethereumclient.Client, error) {
	if c.chainCode.ToChainId().Int64() == 0 {
		return nil, fmt.Errorf("unable to convert provider to ethclient: invalid chain code: %s", fmt.Sprint(c.chainCode))
	}
	return c.ethclient, nil
}

func (c *ethClient) Endpoint() string {
	return c.endpoint
}
