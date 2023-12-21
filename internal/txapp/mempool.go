package txapp

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
)

type mempool struct {
	accounts map[string]*accounts.Account
	mu       sync.Mutex

	accountStore   AccountReader
	validatorStore IsValidatorChecker
}

// accountInfo retrieves the account info from the mempool state or the account store.
func (m *mempool) accountInfo(ctx context.Context, acctID []byte) (*accounts.Account, error) {
	if acctInfo, ok := m.accounts[string(acctID)]; ok {
		return acctInfo, nil // there is an unconfirmed tx for this account
	}

	// get account from account store
	acct, err := m.accountStore.GetAccount(ctx, acctID)
	if err != nil {
		return nil, err
	}

	m.accounts[string(acctID)] = acct

	return acct, nil
}

// accountInfoSafe is wraps accountInfo in a mutex lock.
func (m *mempool) accountInfoSafe(ctx context.Context, acctID []byte) (*accounts.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.accountInfo(ctx, acctID)
}

// applyTransaction validates account specific info and applies valid transactions to the mempool state.
func (m *mempool) applyTransaction(ctx context.Context, tx *transactions.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// seems like maybe this should go in the switch statement below,
	// but I put it here to avoid extra db call for account info
	if tx.Body.PayloadType == transactions.PayloadTypeValidatorVoteIDs {
		isValidator, err := m.validatorStore.IsCurrent(ctx, tx.Sender)
		if err != nil {
			return err
		}

		if !isValidator {
			return fmt.Errorf("only validators can submit validator vote transactions")
		}
	}
	if tx.Body.PayloadType == transactions.PayloadTypeValidatorVoteBodies {
		// not sure if this is the right error code
		return fmt.Errorf("validator vote bodies can not enter the mempool, and can only be submitted during block proposal")
	}

	// get account info from mempool state or account store
	acct, err := m.accountInfo(ctx, tx.Sender)
	if err != nil {
		return err
	}

	// It is normally permissible to accept a transaction with the same nonce as
	// a tx already in mempool (but not in a block), however without gas we
	// would not want to allow that since there is no criteria for selecting the
	// one to mine (normally higher fee).
	if tx.Body.Nonce != uint64(acct.Nonce)+1 {
		return fmt.Errorf("%w for account %s: got %d, expected %d", transactions.ErrInvalidNonce,
			hex.EncodeToString(tx.Sender), tx.Body.Nonce, acct.Nonce+1)
	}

	spend := big.NewInt(0).Set(tx.Body.Fee) // NOTE: this could be the fee *limit*, but it depends on how the modules work

	switch tx.Body.PayloadType {
	case transactions.PayloadTypeTransfer:
		transfer := &transactions.Transfer{}
		err = transfer.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return err
		}

		amt, ok := big.NewInt(0).SetString(transfer.Amount, 10)
		if !ok {
			return transactions.ErrInvalidAmount
		}

		if amt.Cmp(&big.Int{}) < 0 {
			return errors.Join(transactions.ErrInvalidAmount, errors.New("negative transfer not permitted"))
		}

		if amt.Cmp(acct.Balance) > 0 {
			return transactions.ErrInsufficientBalance
		}

		spend.Add(spend, amt)
	}

	// We'd check balance against the total spend (fees plus value sent) if we
	// know gas is enabled. Transfers must be funded regardless of transaction
	// gas requirement:

	// if spend.Cmp(acct.balance) > 0 {
	// 	return errors.New("insufficient funds")
	// }

	// Since we're not yet operating with different policy depending on whether
	// gas is enabled for the chain, we're just going to reduce the account's
	// pending balance, but no lower than zero. Tx execution will handle it.
	if spend.Cmp(acct.Balance) > 0 {
		acct.Balance.SetUint64(0)
	} else {
		acct.Balance.Sub(acct.Balance, spend)
	}

	// Account nonces and spends tracked by mempool should be incremented only for the
	// valid transactions. This is to avoid the case where mempool rejects a transaction
	// due to insufficient balance, but the account nonce and spend are already incremented.
	// Due to which it accepts the next transaction with nonce+1, instead of nonce
	// (but Tx with nonce is never pushed to the consensus pool).
	acct.Nonce = int64(tx.Body.Nonce)

	return nil
}

// reset clears the in-memory unconfirmed account states.
// This should be done at the end of block commit.
func (m *mempool) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts = make(map[string]*accounts.Account)
}
