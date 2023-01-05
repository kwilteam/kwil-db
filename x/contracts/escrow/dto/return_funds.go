package dto

import "math/big"

type ReturnFundsParams struct {
	Recipient     string
	CorrelationId string
	Amount        *big.Int
	Fee           *big.Int
}

type ReturnFundsResponse struct {
	TxHash string
}
