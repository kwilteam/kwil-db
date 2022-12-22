package evm

import (
	"fmt"
	"kwil/x/chain/provider/dto"

	ethereumclient "github.com/ethereum/go-ethereum/ethclient"
)

// client fulfills the dto.ChainProvider interface
type ethClient struct {
	ethclient *ethereumclient.Client
	chainId   int64
}

func New(endpoint string, chainId int64) (dto.ChainProvider, error) {
	client, err := ethereumclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %v", err)
	}

	return &ethClient{
		ethclient: client,
		chainId:   chainId,
	}, nil
}
