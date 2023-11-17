package abci

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

func extractDepInfo(txBody *transactions.TransactionBody) (id string, acct string, amt *big.Int, err error) {
	if txBody.PayloadType != transactions.PayloadTypeAccountCredit {
		err = errors.New("not an account credit transaction")
		return
	}
	if txBody.Fee.Cmp(big.NewInt(0)) != 0 {
		err = errors.New("account credit transaction must have fee of 0")
		return
	}
	var creditPayload transactions.AccountCredit
	err = creditPayload.UnmarshalBinary(txBody.Payload)
	if err != nil {
		return
	}
	var ok bool
	amt, ok = big.NewInt(0).SetString(creditPayload.Amount, 10)
	if !ok {
		err = fmt.Errorf("bad amount %q", creditPayload.Amount)
		return
	}
	return creditPayload.DepositEventID, creditPayload.Account, amt, nil
}

func newAccountCreditTxn(eventID string, chainID string, nonce uint64, amt *big.Int, acct string, signer auth.Signer) ([]byte, error) {
	tx, err := transactions.CreateTransaction(&transactions.AccountCredit{
		DepositEventID: eventID,
		Amount:         amt.String(),
		Account:        acct,
	}, chainID, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to create account credit transaction: %w", err)
	}
	// Fee is intentionally left zero.
	if err = tx.Sign(signer); err != nil {
		return nil, fmt.Errorf("failed to sign account credit transaction: %w", err)
	}
	rawTx, err := tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize signed account credit transaction: %w", err)
	}
	return rawTx, nil
}
