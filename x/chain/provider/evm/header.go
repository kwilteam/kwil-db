package evm

import (
	"context"
	"kwil/x/chain/provider/dto"
	"math/big"
)

func (c *ethClient) HeaderByNumber(ctx context.Context, number *big.Int) (*dto.Header, error) {
	h, err := c.ethclient.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	return &dto.Header{
		Height: h.Number.Int64(),
		Hash:   h.Hash().Hex(),
	}, nil
}
