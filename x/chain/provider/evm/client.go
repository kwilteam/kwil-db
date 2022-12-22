package evm

import (
	"fmt"
	"kwil/x/chain"
	"kwil/x/chain/provider/dto"

	ethereumclient "github.com/ethereum/go-ethereum/ethclient"
)

// client fulfills the dto.ChainProvider interface
type ethClient struct {
	ethclient *ethereumclient.Client
	chainCode chain.ChainCode
}

func New(endpoint string, chainCode chain.ChainCode) (dto.ChainProvider, error) {
	client, err := ethereumclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %v", err)
	}

	return &ethClient{
		ethclient: client,
		chainCode: chainCode,
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
