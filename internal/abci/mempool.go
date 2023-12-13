package abci

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// mempoolState maintains in-memory account state to validate the transactions against.
type mempool struct {
	accountStore AccountsModule

	// in-memory account state to validate transactions against, purged at the end of commit.
	accounts map[string]*userAccount
	mu       sync.Mutex

	// TODO: noGas bool, so we can accept replacement transactions (same
	// nonce but higher fee).
}

type userAccount struct {
	nonce   int64
	balance *big.Int
}

// accountInfo retrieves the account info from the mempool state or the account store.
// If the account is not found, it returns a dummy account with nonce 0 and balance 0.
func (m *mempool) accountInfo(ctx context.Context, acctID []byte) (*userAccount, error) {
	if acctInfo, ok := m.accounts[string(acctID)]; ok {
		return acctInfo, nil // there is an unconfirmed tx for this account
	}

	// get account from account store
	acct, err := m.accountStore.Account(ctx, acctID)
	if err != nil {
		return nil, err
	}

	acctInfo := &userAccount{
		nonce:   acct.Nonce,
		balance: acct.Balance,
	}
	m.accounts[string(acctID)] = acctInfo

	return acctInfo, nil
}

// peekAccountInfo is like accountInfo, but it does not query the account store
// if there are no unconfirmed transactions for the user.
func (m *mempool) peekAccountInfo(ctx context.Context, acctID []byte) *userAccount {
	acctInfo := &userAccount{balance: &big.Int{}} // must be new instance
	m.mu.Lock()
	defer m.mu.Unlock()
	if acct, ok := m.accounts[string(acctID)]; ok {
		acctInfo.balance = big.NewInt(0).Set(acct.balance)
		acctInfo.nonce = acct.nonce
	}
	return acctInfo
}

// applyTransaction validates account specific info and applies valid transactions to the mempool state.
func (m *mempool) applyTransaction(ctx context.Context, tx *transactions.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// get account info from mempool state or account store
	acct, err := m.accountInfo(ctx, tx.Sender)
	if err != nil {
		return err
	}

	// It is normally permissible to accept a transaction with the same nonce as
	// a tx already in mempool (but not in a block), however without gas we
	// would not want to allow that since there is no criteria for selecting the
	// one to mine (normally higher fee).
	if tx.Body.Nonce != uint64(acct.nonce)+1 {
		return fmt.Errorf("%w for account %s: got %d, expected %d", transactions.ErrInvalidNonce,
			hex.EncodeToString(tx.Sender), tx.Body.Nonce, acct.nonce+1)
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

		if amt.Cmp(acct.balance) > 0 {
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
	if spend.Cmp(acct.balance) > 0 {
		acct.balance.SetUint64(0)
	} else {
		acct.balance.Sub(acct.balance, spend)
	}

	// Account nonces and spends tracked by mempool should be incremented only for the
	// valid transactions. This is to avoid the case where mempool rejects a transaction
	// due to insufficient balance, but the account nonce and spend are already incremented.
	// Due to which it accepts the next transaction with nonce+1, instead of nonce
	// (but Tx with nonce is never pushed to the consensus pool).
	acct.nonce = int64(tx.Body.Nonce)

	return nil
}

// reset clears the in-memory unconfirmed account states.
// This should be done at the end of block commit.
func (m *mempool) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts = make(map[string]*userAccount)
}

// groupTransactions groups the transactions by sender.
func groupTxsBySender(txns [][]byte) (map[string][]*transactions.Transaction, error) {
	grouped := make(map[string][]*transactions.Transaction)
	for _, tx := range txns {
		t := &transactions.Transaction{}
		err := t.UnmarshalBinary(tx)
		if err != nil {
			return nil, err
		}
		key := string(t.Sender)
		grouped[key] = append(grouped[key], t)
	}
	return grouped, nil
}

// nonceList is for debugging
func nonceList(txns []*transactions.Transaction) []uint64 {
	nonces := make([]uint64, len(txns))
	for i, tx := range txns {
		nonces[i] = tx.Body.Nonce
	}
	return nonces
}
