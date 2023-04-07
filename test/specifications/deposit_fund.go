package specifications

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// DepositFundDsl is dsl for deposit fund specification
type DepositFundDsl interface {
	DepositFund(ctx context.Context, amount *big.Int) error
	GetDepositBalance(ctx context.Context) (*big.Int, error)
}

func DepositFundSpecification(ctx context.Context, t *testing.T, deposit DepositFundDsl) {
	t.Logf("Executing DepositFundSpecification")
	// Given a user and a validator address, and an amount

	amount := new(big.Int).Mul(big.NewInt(10), big.NewInt(1000000000000000000))
	depositedAmountOld, err := deposit.GetDepositBalance(ctx)
	assert.NoError(t, err)

	// When i deposit fund from user to validator
	err = deposit.DepositFund(ctx, amount)

	// Then i expect success
	assert.NoError(t, err)

	// And i expect the deposited amount to be set
	depositedAmountNew, err := deposit.GetDepositBalance(ctx)
	assert.NoError(t, err)
	assert.Equal(t, depositedAmountOld.Cmp(depositedAmountNew), -1)
}
