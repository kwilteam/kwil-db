package escrow

import (
	"context"
	"crypto/ecdsa"

	"github.com/kwilteam/kwil-db/internal/chain/types"
)

func (c *EscrowContract) Deposit(ctx context.Context, params *types.DepositParams, privateKey *ecdsa.PrivateKey) (*types.DepositResponse, error) {

	auth, err := c.client.PrepareTxAuth(ctx, c.chainId, privateKey)
	if err != nil {
		return nil, err
	}

	res, err := c.ctr.Deposit(auth, params.Amount)
	if err != nil {
		return nil, err
	}
	return &types.DepositResponse{
		TxHash: res.Hash().String(),
	}, nil
}
