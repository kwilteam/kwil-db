package chain

import (
	"context"
	ccDTO "kwil/x/chain-client/dto"
	"math/big"
)

type ReturnFundsParams struct {
	Recipient     string
	CorrelationId string
	Amount        *big.Int
	Fee           *big.Int
}

type ReturnFundsResponse struct {
	TxHash string
}

func (c *chain) ReturnFunds(ctx context.Context, params *ReturnFundsParams) (*ReturnFundsResponse, error) {
	res, err := c.chainClient.ReturnFunds(ctx, &ccDTO.ReturnFundsParams{
		Recipient:     params.Recipient,
		CorrelationId: params.CorrelationId,
		Amount:        params.Amount,
		Fee:           params.Fee,
	})

	if err != nil {
		return nil, err
	}

	return &ReturnFundsResponse{
		TxHash: res.TxHash,
	}, nil
}
