package client

import (
	"context"
	"fmt"
	ccs "kwil/pkg/chain/client/service"
	escrowContracts "kwil/pkg/chain/contracts/escrow"
	tokenContracts "kwil/pkg/chain/contracts/token"
)

func (c *client) EscrowContract(ctx context.Context) (escrowContracts.EscrowContract, error) {
	if c.chainClient == nil {
		err := c.initChainClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return c.chainClient.Contracts().Escrow(c.escrowContractAddress)
}

func (c *client) TokenContract(ctx context.Context, address string) (tokenContracts.TokenContract, error) {
	if c.chainClient == nil {
		err := c.initChainClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return c.chainClient.Contracts().Token(address)
}

func (c *client) initChainClient(ctx context.Context) error {
	if c.chainRpcUrl == nil {
		return fmt.Errorf("chain rpc url is not set")
	}

	var err error
	c.chainClient, err = ccs.NewChainClient(*c.chainRpcUrl,
		ccs.WithChainCode(c.chainCode),
	)
	if err != nil {
		return fmt.Errorf("failed to create chain client: %w", err)
	}

	return nil
}
