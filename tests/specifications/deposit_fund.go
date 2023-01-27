package specifications

import (
	"context"
	"crypto/ecdsa"
	"github.com/stretchr/testify/assert"
	"kwil/x/fund"
	"math/big"
	"testing"
)

// DepositFundDsl is dsl for deposit fund specification
type DepositFundDsl interface {
	DepositFund(ctx context.Context, from *ecdsa.PrivateKey, to string, amount *big.Int) error
	GetDepositBalance(ctx context.Context, from string, to string) (*big.Int, error)
	GetFundConfig() *fund.Config
}

func DepositFundSpecification(t *testing.T, ctx context.Context, deposit DepositFundDsl) {
	t.Logf("Executing DepositFundSpecification")
	//Given a user and a validator address, and an amount

	amount := new(big.Int).Mul(big.NewInt(10), big.NewInt(1000000000000000000))
	cfg := deposit.GetFundConfig()
	depositedAmountOld, err := deposit.GetDepositBalance(ctx, cfg.GetAccount(), cfg.ValidatorAddress)

	//When i deposit fund from user to validator
	err = deposit.DepositFund(ctx, cfg.PrivateKey, cfg.ValidatorAddress, amount)

	//Then i expect success
	assert.NoError(t, err)

	//TODO: check balance
	//And i expect the deposited amount to be set
	depositedAmountNew, err := deposit.GetDepositBalance(ctx, cfg.GetAccount(), cfg.ValidatorAddress)

	assert.NoError(t, err)
	assert.Equal(t, depositedAmountOld.Cmp(depositedAmountNew), -1)
}
