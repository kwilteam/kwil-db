package chainsync

import (
	"context"
	escrowTypes "kwil/x/types/contracts/escrow"
)

func (c *chain) ReturnFunds(ctx context.Context, params *escrowTypes.ReturnFundsParams) (*escrowTypes.ReturnFundsResponse, error) {
	return c.escrowContract.ReturnFunds(ctx, params)
}
