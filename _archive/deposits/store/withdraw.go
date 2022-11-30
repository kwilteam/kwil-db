package store

import (
	"encoding/json"
	"math/big"
	"strings"
)

/*
Withdraw will get the deposit amount for the given address.
If the amount in deposits is less than requested, then the entire amount
will be stored in withdrawals, and deposits will be set to 0 in the same transaction.

In order to handle idempotency and prevent replay attacks, the nonce must be included.

First, the function will check for nonce uniqueness.
*/

type withdrawal struct {
	Amt       *big.Int
	Fee       *big.Int
	To        string
	Processed bool
}

// Bytes marshalls the withdrawal to bytes
func (w *withdrawal) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

// Unmarshal unmarshalls the withdrawal from bytes
func (w *withdrawal) Unmarshal(b []byte) error {
	return json.Unmarshal(b, w)
}

// this can be deleted
func (ds *depositStore) Withdraw(addr, nonce, amt string) error {
	return nil
}

// remove expiry splits a nonce by the colon and returns the second half
func removeExpiry(n string) (string, error) {
	s := strings.Split(n, ":")
	if len(s) != 2 {
		return n, ErrInvalidNonce // returning the nonce for coverage
	}
	return s[1], nil
}
