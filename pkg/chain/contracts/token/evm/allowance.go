package evm

import (
	"context"
	"crypto/ecdsa"
	kwilCommon "kwil/pkg/chain/contracts/common/evm"
	"kwil/pkg/chain/contracts/token/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) Allowance(owner, spender string) (*big.Int, error) {
	return c.ctr.Allowance(nil, common.HexToAddress(owner), common.HexToAddress(spender))
}

func (c *contract) Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types.ApproveResponse, error) {
	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, privateKey)
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
