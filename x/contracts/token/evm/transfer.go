package evm

import (
	"context"
	"kwil/x/types/contracts/token"

	"math/big"

	kwilCommon "kwil/x/contracts/common/evm"

	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) Transfer(ctx context.Context, to string, amount *big.Int) (*token.TransferResponse, error) {
	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, c.privateKey)
	if err != nil {
		return nil, err
	}

	// create the transaction
	tx, err := c.ctr.Transfer(auth, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}

	return &token.TransferResponse{
		TxHash: tx.Hash().String(),
	}, nil
}
