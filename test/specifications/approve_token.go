package specifications

import (
	"context"
	"crypto/ecdsa"
	"github.com/stretchr/testify/assert"
	"kwil/pkg/fund"
	"math/big"
	"testing"
)

type ApproveTokenDsl interface {
	ApproveToken(ctx context.Context, from *ecdsa.PrivateKey, spender string, amount *big.Int) error
	GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error)
	GetFundConfig() *fund.Config
}

func ApproveTokenSpecification(t *testing.T, ctx context.Context, approve ApproveTokenDsl) {
	t.Logf("Executing ApproveTokenSpecification")
	//@yaiba TODO: make this into args?
	//Given a user and a validator address, and an amount
	//decimals := 18
	amount := new(big.Int).Mul(big.NewInt(100), big.NewInt(1000000000000000000))
	chainCfg := approve.GetFundConfig()

	//When i approve validator to spend my tokens
	err := approve.ApproveToken(ctx, chainCfg.PrivateKey, chainCfg.PoolAddress, amount)

	//Then i expect success
	assert.NoError(t, err)

	//And i expect the allowance to be set
	allowance, err := approve.GetAllowance(ctx, chainCfg.GetAccountAddress(), chainCfg.PoolAddress)
	assert.NoError(t, err)
	assert.Equal(t, amount, allowance)
}
