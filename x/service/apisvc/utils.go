package apisvc

import (
	"math/big"
)

func parseBigInt(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 10)
}
