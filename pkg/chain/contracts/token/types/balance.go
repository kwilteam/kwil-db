package types

import "math/big"

type BalanceParams struct {
	Address string
}

type BalanceResponse struct {
	Balance *big.Int
}
