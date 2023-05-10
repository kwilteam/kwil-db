package evm

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	kwilCommon "github.com/kwilteam/kwil-db/pkg/chain/contracts/common/evm"
	"github.com/kwilteam/kwil-db/pkg/chain/contracts/escrow/types"
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
