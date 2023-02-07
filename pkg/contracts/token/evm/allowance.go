package evm

import (
	"context"
	kwilCommon "kwil/pkg/contracts/common/evm"
	"kwil/pkg/contracts/token/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) Allowance(owner, spender string) (*big.Int, error) {
	return c.ctr.Allowance(nil, common.HexToAddress(owner), common.HexToAddress(spender))
}

func (c *contract) Approve(ctx context.Context, spender string, amount *big.Int) (*types.ApproveResponse, error) {
	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, c.privateKey)
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
