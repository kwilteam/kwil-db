package escrow

import "math/big"

type DepositBalanceParams struct {
	Validator string
	Address   string
}

type DepositBalanceResponse struct {
	Balance *big.Int
}
