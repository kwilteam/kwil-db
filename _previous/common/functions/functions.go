package functions

import (
	"context"
	"math/big"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/voting"
)

// Accounts is a namespace for account specific functions.
var Accounts accounting

// accounting is a namespace for account specific functions.
type accounting struct{}

// Credit credits an account with the given amount. If the account
// does not exist, it will be created. A negative amount will be
// treated as a debit. Accounts cannot have negative balances, and
// will return an error if the amount would cause the balance to go
// negative.
func (accounting) Credit(ctx context.Context, tx sql.DB, account []byte, amt *big.Int) error {
	return accounts.Credit(ctx, tx, account, amt)
}

// Transfer transfers an amount from one account to another. If the
// from account does not have enough funds to transfer the amount,
// it will fail. If the to account does not exist, it will be
// created. The amount must be greater than 0.
func (accounting) Transfer(ctx context.Context, tx sql.DB, from, to []byte, amt *big.Int) error {
	return accounts.Transfer(ctx, tx, from, to, amt)
}

// GetAccount retrieves the account with the given identifier. If the
// account does not exist, it will return an account with a balance
// of 0 and a nonce of 0.
func (accounting) GetAccount(ctx context.Context, tx sql.DB, account []byte) (*types.Account, error) {
	return accounts.GetAccount(ctx, tx, account)
}

// Validators is a namespace for validator specific functions.
var Validators validators

type validators struct{}

// GetValidatorPower retrieves the power of the given validator. If
// the validator does not exist, it will return 0.
func (validators) GetValidatorPower(ctx context.Context, tx sql.DB, validator []byte) (int64, error) {
	return voting.GetValidatorPower(ctx, tx, validator)
}

// GetValidators retrieves all validators.
func (validators) GetValidators(ctx context.Context, tx sql.DB) ([]*types.Validator, error) {
	return voting.GetValidators(ctx, tx)
}

// SetValidatorPower sets the power of a validator. If the target
// validator does not exist, it will be created with the given power.
// If set to 0, the target validator will be deleted, and will no
// longer be considered a validator. It will return an error if a
// negative power is given.
func (validators) SetValidatorPower(ctx context.Context, tx sql.DB, validator []byte, power int64) error {
	return voting.SetValidatorPower(ctx, tx, validator, power)
}
