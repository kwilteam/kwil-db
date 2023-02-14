package evm

import (
	"context"
	"crypto/ecdsa"
	kwilCommon "kwil/pkg/contracts/common/evm"
	"kwil/pkg/contracts/escrow/types"

	"github.com/ethereum/go-ethereum/common"
)

// ReturnFunds calls the returnDeposit function
func (c *contract) ReturnFunds(ctx context.Context, params *types.ReturnFundsParams, privateKey *ecdsa.PrivateKey) (*types.ReturnFundsResponse, error) {

	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, privateKey)
	if err != nil {
		return nil, err
	}

	res, err := c.ctr.ReturnDeposit(auth, common.HexToAddress(params.Recipient), params.Amount, params.Fee, params.CorrelationId)
	if err != nil {
		return nil, err
	}

	return &types.ReturnFundsResponse{
		TxHash: res.Hash().String(),
	}, nil
}
