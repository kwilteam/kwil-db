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

func (ds *depositStore) Withdraw(addr, nonce string, amt *big.Int) error {

	// get the current amount in the deposit bucket
	curDepAmt, err := ds.db.Get(append(DEPOSITKEY, []byte(addr)...))
	if err != nil {
		return err
	}

	// current amt in bigint
	depAmt := new(big.Int).SetBytes(curDepAmt)

	var retAmt *big.Int
	// check if amt <= depAmt
	if amt.Cmp(depAmt) == 1 {
		retAmt = depAmt
		depAmt = big.NewInt(0)
	} else {
		retAmt = amt
		depAmt.Sub(depAmt, amt)
	}

	// get amt from spent
	spent, err := ds.db.Get(append(SPENTKEY, []byte(addr)...))
	if err != nil {
		return err
	}

	// convert spent to big int
	spentAmt := new(big.Int).SetBytes(spent)

	sn, err := removeExpiry(nonce)
	if err != nil {
		// log sn for coverage
		ds.log.Sugar().Errorf("Invalid nonce", "nonce", sn)
		return err
	}

	// create new transaction
	txn := ds.db.NewTransaction(true)
	defer txn.Discard()

	// this transaction should set the sn to a byte representation of the withdrawal with the withdrawal prefix
	// it should also set the spent to 0 with the spent prefix
	// it should also set the deposit to depAmt with the deposit prefix
	// finally, it should map the nonce to sn with the expiration prefix

	// set the new balance in the deposit bucket
	depKey := append(DEPOSITKEY, []byte(addr)...)
	err = txn.Set(depKey, depAmt.Bytes())
	if err != nil {
		return err
	}

	// set the new balance in the spent bucket
	spendKey := append(SPENTKEY, []byte(addr)...)
	err = txn.Set(spendKey, []byte{0, 0, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return err
	}

	// set the new balance in the withdrawal bucket
	withdrawKey := append(WITHDRAWALKEY, []byte(sn)...)
	wd := &withdrawal{
		Amt:       retAmt,
		Fee:       spentAmt,
		To:        addr,
		Processed: false,
	}
	b, err := wd.Marshal()
	if err != nil {
		return err
	}

	err = txn.Set(withdrawKey, b)
	if err != nil {
		return err
	}

	// set the nonce to sn
	expKey := append(EXPIRYKEY, []byte(nonce)...)
	err = txn.Set(expKey, []byte(sn))
	if err != nil {
		return err
	}

	// commit the transaction
	return txn.Commit()
}

// remove expiry splits a nonce by the colon and returns the second half
func removeExpiry(n string) (string, error) {
	s := strings.Split(n, ":")
	if len(s) != 2 {
		return n, ErrInvalidNonce // returning the nonce for coverage
	}
	return s[1], nil
}
