package specifications

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// All the tests are performed with 0 required confirmations, so the effects are immediate.

/*
Tests:
GasCosts enabled, TokenBridge enabled

Moving pieces to be tested:
- Deposit store
- Vote Extensions
- Account store

Actions:
- Approve: Approves escrow contract as spender for certain amount of tokens
- Transfer: Transfers tokens to escrow contract and on the kwil eventually

  - Deposit store listens to this event

  - Vote extensions are created

  - Account store is updated once the threshold agrees on the transfer

    How to test this:

  - Approve

  - Transfer

  - Check deposit store for the event to disappear (#rows to be empty)

    Then check the account store for the new balance

- Transactions:

  - Issue a transaction to the kwil node and let it mine

  - Check the account store for the new balance (shld be less as tx spends)

    Test this with different tx types: deploy, drop, execute
*/

const (
// deployPrice = 10
// dropPrice   = 2
// execPrice   = 1
)

// Approve doesn't validate whether the owner have funds. Can't test failure case.
func TokenBridgeApproveSuccessSpecification(ctx context.Context, t *testing.T, spender string, bridge TokenBridgeDsl) {
	// Approve the escrow contract as spender for 100 tokens
	_, err := bridge.Approve(ctx, spender, big.NewInt(100))
	assert.NoError(t, err)

	// Check the allowance
	allowance, err := bridge.Allowance(ctx, spender)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(100), allowance)
}

// Transfer funds to the escrow contract - Success scenario
func TokenBridgeDepositSuccessSpecification(ctx context.Context, t *testing.T, spender string, bridge TokenBridgeDsl) {
	// Check the balance
	preBalance, err := bridge.BalanceOf(ctx)
	assert.NoError(t, err)

	// Approve the escrow contract as spender for 100 tokens
	_, err = bridge.Approve(ctx, spender, big.NewInt(100))
	assert.NoError(t, err)

	// Check the allowance

	allowance, err := bridge.Allowance(ctx, spender)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(100), allowance)

	// Transfer 100 tokens to the escrow contract
	_, err = bridge.Deposit(ctx, big.NewInt(100))
	assert.NoError(t, err)

	// Check the allowance
	allowance, err = bridge.Allowance(ctx, spender)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), allowance.Int64())

	// Check the balance
	diff := big.NewInt(0)
	postBalance, err := bridge.BalanceOf(ctx)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(100), diff.Sub(preBalance, postBalance))

	// Account should be credited with 100 tokens
	acct, err := bridge.GetAccount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(100), acct.Balance)
}
