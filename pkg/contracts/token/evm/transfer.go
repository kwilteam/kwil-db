package evm

import (
	"context"
	"crypto/ecdsa"
	kwilCommon "kwil/pkg/contracts/common/evm"
	"kwil/pkg/contracts/token/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) Transfer(ctx context.Context, to string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types.TransferResponse, error) {
	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, privateKey)
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
