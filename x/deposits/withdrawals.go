package deposits

import (
	"math/big"
	"math/rand"
	"strconv"
	"strings"
)

/*
	The process for Withdrawals is as follows:

	1. The wallet seeking a withdrawal sends a request to the validator.
	2. Function checks things like unique ID, amount requested, etc.
	3. The validator then finds how much the wallet has spent.  The validator will then see how much
	   the wallet has not spent yet.  If the unspent amount is less than the amount requested, the
	   validator will return only the unspent amount, and take the fee.  If the unspent amount is greater
	   than the amount requested, the validator will return the amount requested, and take the fee.
*/

func (d *deposits) Withdraw(addr string, amt *big.Int) error {

	return nil
}

// validateNonce should check that the nonce is in the proper format (with the block expiration prepended)
// and that the block expiration is within the proper range
// e.g. 1453052:cmkr3oen3o3g4j0 (block expiration:nonce)
func validateNonce(n string, low, high int64, l uint8) bool {
	hvs := strings.Split(n, ":")
	if len(hvs) != 2 {
		return false
	}

	// check that the block expiration is a number
	nm, err := strconv.Atoi(hvs[0])
	if err != nil {
		return false
	}

	// check that the block expiration is within the proper range
	if int64(nm) < low || int64(nm) > high {
		return false
	}

	// checks hvs[1] is 5 characters long
	if uint8(len(hvs[1])) != l {
		return false
	}

	return true
}

// generateRandomString generates a random string of length l
func generateNonce(l uint8) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+=-"
	result := make([]byte, l)
	for i := uint8(0); i < l; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
