package abci

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// mempoolState maintains in-memory account state to validate the transactions against.
type mempool struct {
	accountStore AccountsModule
	// nonceTracker tracks the last valid nonce for each account. nonce is
	// unconfirmed if the value is greater than the nonce in the account store.
	// Key: sender's public key, Value: last valid nonce
	nonceTracker map[string]uint64

	// in-memory account state to validate transactions against, purged at the end of commit.
	accounts map[string]*userAccount
	mu       sync.Mutex
}

type userAccount struct {
	nonce   int64
	balance *big.Int
}

// accountInfo retrieves the account info from the mempool state or the account store.
// If the account is not found, it returns a dummy account with nonce 0 and balance 0.
func (m *mempool) accountInfo(ctx context.Context, pubKey []byte) (*userAccount, error) {
	if acctInfo, ok := m.accounts[string(pubKey)]; ok {
		return acctInfo, nil // there is an unconfirmed tx for this account
	}

	// get account from account store
	acct, err := m.accountStore.GetAccount(ctx, pubKey)
	if err != nil {
		return nil, err
	}

	acctInfo := &userAccount{
		nonce:   acct.Nonce,
		balance: acct.Balance,
	}
	m.accounts[string(pubKey)] = acctInfo

	return acctInfo, nil
}

// applyTransaction validates account specific info and applies valid transactions to the mempool state.
func (m *mempool) applyTransaction(ctx context.Context, tx *transactions.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// get account info from mempool state or account store
	acct, err := m.accountInfo(context.Background(), tx.Sender)
	if err != nil {
		return err
	}

	hexKey := hex.EncodeToString(tx.Sender)

	//  It is normally permissible to accept a transaction with the same nonce
	//	as a tx already in mempool (but not in a block), however without gas
	//	we would not want to allow that since there is no criteria
	//  for selecting the one to mine (normally higher fee).
	if tx.Body.Nonce != uint64(acct.nonce)+1 {
		return fmt.Errorf("invalid nonce for account %s: got %d, expected %d", hexKey, tx.Body.Nonce, acct.nonce+1)
	}

	m.updateAccount(tx.Sender, tx.Body.Nonce)
	return nil
}

// updateAccount is called post-transaction validation, so that the effects of
// the transaction is reflected in the mempool's view of the account state.
// This ensures that the subsequent transactions are validated against this
// updated view of the account state, rather than the one from the account store.
func (m *mempool) updateAccount(pubKey []byte, txNonce uint64) {
	publicKey := string(pubKey)

	m.accounts[publicKey].nonce++
	//acct.balance.Sub(acct.balance, fee)

	if txNonce > m.nonceTracker[publicKey] {
		m.nonceTracker[publicKey] = txNonce
	}
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
