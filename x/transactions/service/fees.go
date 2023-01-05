package service

import (
	"fmt"
	"math/big"
)

// checkFee takes the sent fee and the expected fee and checks if the sent fee is enough
func checkFee(sentFee, price string) (bool, error) {
	// convert the sent fee to a big int
	signedFee, ok := big.NewInt(0).SetString(sentFee, 10)
	if !ok {
		return false, fmt.Errorf("failed to convert fee to big int")
	}

	// convert the expected fee to a big int
	expectedFeeBigInt, ok := big.NewInt(0).SetString(price, 10)
	if !ok {
		return false, fmt.Errorf("failed to convert price to big int")
	}

	// check if fee is enough
	if signedFee.Cmp(expectedFeeBigInt) < 0 {
		return false, nil
	}

	return true, nil
}
