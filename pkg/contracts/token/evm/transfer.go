package evm

import (
	"context"
	kwilCommon "kwil/pkg/contracts/common/evm"
	"kwil/pkg/contracts/token/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) Transfer(ctx context.Context, to string, amount *big.Int) (*types.TransferResponse, error) {
	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, c.privateKey)
	if err != nil {
		return nil, err
	}

	// create the transaction
	tx, err := c.ctr.Transfer(auth, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}

	return &types.TransferResponse{
		TxHash: tx.Hash().String(),
	}, nil
}
