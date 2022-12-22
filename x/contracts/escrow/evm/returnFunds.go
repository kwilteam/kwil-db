package evm

import (
	"context"
	"kwil/x/contracts/escrow/dto"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// ReturnFunds calls the returnDeposit function
func (c *contract) ReturnFunds(ctx context.Context, params *dto.ReturnFundsParams) (*dto.ReturnFundsResponse, error) {

	txOpts, err := bind.NewKeyedTransactorWithChainID(c.key, c.cid)
	if err != nil {
		return nil, err
	}

	res, err := c.ctr.ReturnDeposit(txOpts, common.HexToAddress(params.Recipient), params.Amount, params.Fee, params.CorrelationId)
	if err != nil {
		return nil, err
	}

	return &dto.ReturnFundsResponse{
		TxHash: res.Hash().String(),
	}, nil
}
