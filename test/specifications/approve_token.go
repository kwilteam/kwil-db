package specifications

import (
	"context"
	grpc "kwil/pkg/grpc/client/v1"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ApproveTokenDsl interface {
	ApproveToken(ctx context.Context, amount *big.Int) error
	GetAllowance(ctx context.Context) (*big.Int, error)
	GetServiceConfig(ctx context.Context) (*grpc.SvcConfig, error)
	GetUserAddress() string
}

func ApproveTokenSpecification(ctx context.Context, t *testing.T, approve ApproveTokenDsl) {
	t.Logf("Executing ApproveTokenSpecification")
	// @yaiba TODO: make this into args?
	//Given a user and a validator address, and an amount
	//decimals := 18
	amount := new(big.Int).Mul(big.NewInt(100), big.NewInt(1000000000000000000)) // amount here doesn't matter since we can approve any amount

	// When i approve validator to spend my tokens
	err := approve.ApproveToken(ctx, amount)

	// Then i expect success
	assert.NoError(t, err)

	// And i expect the allowance to be set
	allowance, err := approve.GetAllowance(ctx)
	assert.NoError(t, err)
	assert.Equal(t, amount, allowance)
}
