package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// NetworkOpsDsl is dsl for blockchain network operations e.g. validator
// join/approve/leave.
type NetworkOpsDsl interface {
	ApproveNode(ctx context.Context, joinerPubKey []byte) error
	ValidatorNodeJoin(ctx context.Context) error
	ValidatorNodeLeave(ctx context.Context) error
	// ValidatorJoinStatus(ctx context.Context, pubKey []byte) error
	ValidatorSetCount(ctx context.Context) (int, error)
}

func NetworkNodeValidatorSetSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, count int) {
	t.Log("Executing network node validator set specification")
	cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, count, cnt)
}

func NetworkNodeJoinSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl) {
	t.Log("Executing network node join specification")
	err := netops.ValidatorNodeJoin(ctx)
	assert.NoError(t, err)
}

func NetworkNodeLeaveSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl) {
	t.Log("Executing network node leave specification")
	err := netops.ValidatorNodeLeave(ctx)
	assert.NoError(t, err)
}

func NetworkNodeApproveSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, joiner []byte) {
	t.Log("Executing network node approve specification")
	err := netops.ApproveNode(ctx, joiner)
	assert.NoError(t, err)
}
