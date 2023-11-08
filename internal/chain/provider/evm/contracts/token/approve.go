package token

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kwilteam/kwil-db/internal/chain/types"
)

func (c *TokenContract) Allowance(ctx context.Context, owner, spender string) (*big.Int, error) {
	return c.ctr.Allowance(nil, common.HexToAddress(owner), common.HexToAddress(spender))
}

func (c *TokenContract) Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types.ApproveResponse, error) {
	auth, err := c.client.PrepareTxAuth(ctx, c.chainId, privateKey)
	if err != nil {
		return nil, err
	}

	// create the transaction
	tx, err := c.ctr.Approve(auth, common.HexToAddress(spender), amount)
	if err != nil {
		return nil, err
	}

	return &types.ApproveResponse{
		TxHash: tx.Hash().String(),
	}, nil
}
