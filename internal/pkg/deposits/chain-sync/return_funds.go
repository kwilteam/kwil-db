package chainsync

import (
	"context"
	"fmt"
	escrowTypes "kwil/pkg/chain/contracts/escrow/types"
)

func (c *chain) ReturnFunds(ctx context.Context, params *escrowTypes.ReturnFundsParams) (*escrowTypes.ReturnFundsResponse, error) {
	fmt.Println("IMPLEMENT ME")
	return nil, nil
}
