package specifications

import (
	"context"
	"github.com/stretchr/testify/assert"
	grpc "kwil/pkg/grpc/client"
	"math/big"
	"testing"
)

type ApproveTokenDsl interface {
	ApproveToken(ctx context.Context, spender string, amount *big.Int) error
	GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error)
	GetServiceConfig(ctx context.Context) (grpc.SvcConfig, error)
	GetUserAddress() string
}

func ApproveTokenSpecification(ctx context.Context, t *testing.T, approve ApproveTokenDsl) {
	t.Logf("Executing ApproveTokenSpecification")
	// @yaiba TODO: make this into args?
	//Given a user and a validator address, and an amount
	//decimals := 18
	amount := new(big.Int).Mul(big.NewInt(100), big.NewInt(1000000000000000000))
	svcCfg, err := approve.GetServiceConfig(ctx)
	assert.NoError(t, err)

	// When i approve validator to spend my tokens
	err = approve.ApproveToken(ctx, svcCfg.Funding.PoolAddress, amount)

	// Then i expect success
	assert.NoError(t, err)

	// And i expect the allowance to be set
	allowance, err := approve.GetAllowance(ctx, approve.GetUserAddress(), svcCfg.Funding.PoolAddress)
	assert.NoError(t, err)
	assert.Equal(t, amount, allowance)
}
