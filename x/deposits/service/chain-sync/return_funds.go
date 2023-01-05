package chainsync

import (
	"context"
	escrowDTO "kwil/x/contracts/escrow/dto"
)

func (c *chain) ReturnFunds(ctx context.Context, params *escrowDTO.ReturnFundsParams) (*escrowDTO.ReturnFundsResponse, error) {
	return c.escrowContract.ReturnFunds(ctx, params)
}
