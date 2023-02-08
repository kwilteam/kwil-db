package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	kwilCommon "kwil/pkg/contracts/common/evm"
	"kwil/pkg/contracts/escrow/types"
)

// ReturnFunds calls the returnDeposit function
func (c *contract) ReturnFunds(ctx context.Context, params *types.ReturnFundsParams) (*types.ReturnFundsResponse, error) {

	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, c.privateKey)
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
