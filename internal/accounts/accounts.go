package accounts

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"go.uber.org/zap"
)

// CommitRegister is an interface for registering a commit.
type CommitRegister interface {
	// Skip returns true if the commit should be skipped.
	// This isgnals that the account store should not be updated,
	// and simply return nil.
	Skip() bool

	// Register registers a commit.
	// This should be called when data is written to the database.
	Register(value []byte) error
}

type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
}

type AccountStore struct {
	db            Datastore
	log           log.Logger
	rw            sync.RWMutex
	gasEnabled    bool
	noncesEnabled bool

	committable CommitRegister
}

func NewAccountStore(ctx context.Context, datastore Datastore, committable CommitRegister, opts ...AccountStoreOpts) (*AccountStore, error) {
	ar := &AccountStore{
		log:         log.NewNoOp(),
		db:          datastore,
		committable: committable,
	}

	for _, opt := range opts {
		opt(ar)
	}

	err := ar.initTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return ar, nil
}

func (a *AccountStore) Close() error {
	return nil
}

// Credit creates an account with an address and no pubkey.
//
// spend based on pubkey will first check by pubkey, and then look for the
// corresponding address, and if it is found then it can update the pubkey
// column

// GetAccountByAddress ?

func (a *AccountStore) GetAccount(ctx context.Context, pubKey []byte) (*Account, error) {
	a.rw.RLock()
	defer a.rw.RUnlock()

	return a.getAccountReadOnly(ctx, pubKey)
}

type Spend struct {
	AccountPubKey []byte
	Amount        *big.Int
	Nonce         int64
}

func (s *Spend) bytes() []byte {
	bts := s.AccountPubKey
	bts = append(bts, s.Amount.Bytes()...)

	binary.LittleEndian.AppendUint64(bts, uint64(s.Nonce))

	return bts
}

// Spend spends an amount from an account. It blocks until the spend is written to the database.
func (a *AccountStore) Spend(ctx context.Context, spend *Spend) error {
	a.rw.Lock()
	defer a.rw.Unlock()

	if a.committable.Skip() {
		return nil
	}

	balance, nonce, err := a.checkSpend(ctx, spend)
	if err != nil {
		if !errors.Is(err, errAccountNotFound) {
			return fmt.Errorf("failed to check spend: %w", err)
		}
		if spend.Nonce != 1 {
			return errors.New("initial spend with nonce != 1")
		}
		if spend.Amount.Cmp(big.NewInt(0)) == 0 {
			// hack special case for free (governance) transactions that only require nonce check

			// if this is really free, just create the account entry for nonce purposes
			// TODO: This is probably incorrect way to do it. while testing with gas disabled and chain syncer on, which we wont be doing. But this just creates an account with nil balance, due to which further get account calls will fail. May need a better way to handle node accounts.
			if err = a.createAccount(ctx, "", spend.AccountPubKey, big.NewInt(0), 1); err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}
		} else {
			//  BAD BAD BAD BAD, this is a hack so we can work until *identifier* replaces pubkey
			addr, _ := auth.EthSecp256k1Authenticator{}.Address(spend.AccountPubKey)
			balance, err = a.getPendingAccountBalance(ctx, addr)
			if errors.Is(err, errAccountNotFound) {
				return fmt.Errorf("getPendingAccountBalance: %w", err)
			}

			a.log.Info("found pending account", zap.String("addr", addr), zap.String("balance", balance.String()))

			balance, err = (&Account{Balance: balance}).validateSpend(spend.Amount)
			if err != nil {
				return fmt.Errorf("validateSpend: %w", err)
			}
			fmt.Println("Spend creating account", spend.AccountPubKey, balance, 1)
			nonce = 1
			if err = a.createAccount(ctx, addr, spend.AccountPubKey, balance, 1); err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}
			if err = a.deletePendingAccount(ctx, addr); err != nil {
				return fmt.Errorf("failed to delete pending account: %w", err)
			}
		}
	}
	fmt.Println("Spend udpating account", spend.AccountPubKey, balance, nonce)
	err = a.updateAccount(ctx, spend.AccountPubKey, balance, nonce)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return a.committable.Register(spend.bytes())
}

func (a *AccountStore) Credit(ctx context.Context, addr string, amt *big.Int) error {
	a.rw.Lock()
	defer a.rw.Unlock()

	if a.committable.Skip() {
		return nil
	}

	// First try to find the account by address in the accounts table. If found,
	// update the balance in that table, maybe return pubkey. If not found,
	// check the new_accounts table, if found, update the balance in that table,
	// otherwise insert a new row.  ALL OF THIS GOES AWAY with identifier replacing pubkey
	account, err := a.getAccountByAddress(ctx, addr)
	if err != nil {
		if !errors.Is(err, errAccountNotFound) {
			return err
		}
		bal, err := a.getPendingAccountBalance(ctx, addr)
		if err != nil {
			if !errors.Is(err, errAccountNotFound) {
				return err
			}
			return a.createPendingAccount(ctx, addr, amt)
		}
		bal = bal.Add(bal, amt)
		err = a.updatePendingAccount(ctx, addr, bal)
		if err != nil {
			return err
		}
	} else {
		bal := new(big.Int).Add(account.Balance, amt)
		err = a.updateAccountBalance(ctx, account.PublicKey, bal)
		if err != nil {
			return fmt.Errorf("failed to update account: %w", err)
		}
	}

	b := append([]byte(addr), amt.Bytes()...)
	return a.committable.Register(b)
}

// checkSpend checks that a spend is valid.  If gas costs are enabled, it checks that the account has enough gas to pay for the spend.
// If nonces are enabled, it checks that the nonce is correct.  It returns the new balance and nonce if the spend is valid. It returns an
// error if the spend is invalid.
func (a *AccountStore) checkSpend(ctx context.Context, spend *Spend) (*big.Int, int64, error) {
	account, err := a.getAccountSynchronous(ctx, spend.AccountPubKey)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get account: %w", err)
	}

	nonce := account.Nonce
	if a.noncesEnabled {
		err = account.validateNonce(spend.Nonce)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to validate nonce: %w", err)
		}

		nonce = spend.Nonce
	}

	balance := account.Balance
	if a.gasEnabled {
		balance, err = account.validateSpend(spend.Amount)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to subtract gas: %w", err)
		}
	}

	return balance, nonce, nil
}
