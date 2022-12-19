package service

import "context"

// GetLatestBlock returns the latest block number.
// If finalized is true, it will return the latest block number that has enough confirmations.
func (c *chainClient) GetLatestBlock(ctx context.Context, confirmed bool) (int64, error) {
	block, err := c.client.GetLatestBlock(ctx)
	if err != nil {
		return 0, err
	}

	if confirmed {
		block -= c.requiredConfirmations
	}

	return block, nil
}
