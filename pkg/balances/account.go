package balances

import "math/big"

type Account struct {
	Address string   `json:"address"`
	Balance *big.Int `json:"balance"`
	Nonce   int64    `json:"nonce"`
}
