package evm

import (
	"context"
	kwilCommon "kwil/x/contracts/common/evm"
	"kwil/x/contracts/token/dto"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) Allowance(owner, spender string) (*big.Int, error) {
	return c.ctr.Allowance(nil, common.HexToAddress(owner), common.HexToAddress(spender))
}

func (c *contract) Approve(ctx context.Context, spender string, amount *big.Int) (*dto.ApproveResponse, error) {
	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, c.privateKey)
	if err != nil {
		return nil, err
	}

	// create the transaction
	tx, err := c.ctr.Approve(auth, common.HexToAddress(spender), amount)
	if err != nil {
		return nil, err
	}

	return &dto.ApproveResponse{
		TxHash: tx.Hash().String(),
	}, nil
}
