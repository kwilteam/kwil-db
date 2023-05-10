package service

import (
	"context"
	"fmt"
	provider "github.com/kwilteam/kwil-db/pkg/chain/provider/dto"
	"math/big"
)

// GetLatestBlock returns the latest block number.
// If finalized is true, it will return the latest block number that has enough confirmations.
func (c *chainClient) GetLatestBlock(ctx context.Context) (*provider.Header, error) {
	// this involes 2 calls; one to get the latest block and one to get the latest finalized block
	header, err := c.provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %v", err)
	}

	lastFinalized := header.Height - c.requiredConfirmations
	if lastFinalized < 0 {
		return nil, fmt.Errorf("latest block is less than required confirmations.  latest block: %d.  required confirmations: %d: %v", header.Height, c.requiredConfirmations, err)
	}

	bigLastFinalized := big.NewInt(lastFinalized)

	finalizedHeader, err := c.provider.HeaderByNumber(ctx, bigLastFinalized)
	if err != nil {
		return nil, err
	}

	return finalizedHeader, nil
}

func (c *chainClient) setLatestBlock(ctx context.Context) error {
	latest, err := c.GetLatestBlock(ctx)
	if err != nil {
		return err
	}

	c.lastBlock = latest.Height

	return nil
}
