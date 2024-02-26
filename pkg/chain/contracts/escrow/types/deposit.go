package types

import "math/big"

type DepositParams struct {
	Validator string
	Amount    *big.Int
}

type DepositResponse struct {
	TxHash string
}
