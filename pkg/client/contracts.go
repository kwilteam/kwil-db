package client

import (
	"context"
	"fmt"
	ccs "kwil/pkg/chain/client/service"
	escrowContracts "kwil/pkg/chain/contracts/escrow"
	tokenContracts "kwil/pkg/chain/contracts/token"
)

func (c *KwilClient) EscrowContract(ctx context.Context) (escrowContracts.EscrowContract, error) {
	if c.chainClient == nil {
		err := c.initChainClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return c.chainClient.Contracts().Escrow(c.EscrowContractAddress)
}

func (c *KwilClient) TokenContract(ctx context.Context) (tokenContracts.TokenContract, error) {
	if c.chainClient == nil {
		err := c.initChainClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	escrow, err := c.EscrowContract(ctx)
	if err != nil {
		return nil, err
	}

	address := escrow.TokenAddress()

	return c.chainClient.Contracts().Token(address)
}

func (c *KwilClient) initChainClient(ctx context.Context) error {
	if c.chainRpcUrl == nil {
		return fmt.Errorf("chain rpc url is not set")
	}

	var err error
	c.chainClient, err = ccs.NewChainClient(*c.chainRpcUrl,
		ccs.WithChainCode(c.ChainCode),
	)
	if err != nil {
		return fmt.Errorf("failed to create chain client: %w", err)
	}

	return nil
}
