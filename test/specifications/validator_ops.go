package specifications

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/validators"
	"github.com/stretchr/testify/assert"
)

// ValidatorOpsDsl is a DSL for validator set updates specification such as join, leave, approve, etc.
type ValidatorOpsDsl interface {
	TxQueryDsl
	ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) ([]byte, error)
	ValidatorNodeJoin(ctx context.Context) ([]byte, error)
	ValidatorNodeLeave(ctx context.Context) ([]byte, error)
	ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*validators.JoinRequest, error)
	ValidatorsList(ctx context.Context) ([]*validators.Validator, error)
}

func NetworkNodeValidatorSetSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, count int) {
	t.Log("Executing network node validator set specification")
	vals, err := netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, count, len(vals))
}

func NetworkNodeJoinSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, joiner []byte, valCount int) {
	t.Log("Executing network node join specification")
	// ValidatorSet count doesn't change just by issuing a Join request. Pre and Post cnt should be the same.
	vals, err := netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, valCount, len(vals))

	rec, err := netops.ValidatorNodeJoin(ctx)
	assert.NoError(t, err)

	// Ensure that the Tx is mined.
	expectTxSuccess(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// Get Request status, #approvals = 0
	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	assert.NoError(t, err)
	assert.Equal(t, valCount, len(joinStatus.Board))
	assert.Equal(t, 0, approvalCount(joinStatus))

	// Current validators should remain the same
	vals, err = netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, valCount, len(vals))
}

func NetworkNodeApproveSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, joiner []byte, preCnt int, postCnt int, approved bool) {
	t.Log("Executing network node approve specification")
	// Pre approval verification
	vals, err := netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, preCnt, len(vals))

	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	assert.NoError(t, err)
	assert.Equal(t, preCnt, len(joinStatus.Board))
	preApprovalCnt := approvalCount(joinStatus)

	// Approval Request
	rec, err := netops.ValidatorNodeApprove(ctx, joiner)
	assert.NoError(t, err)

	// Ensure that the Tx is mined.
	expectTxSuccess(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// Check Join Request Status to ensure that the vote is included
	joinStatus, err = netops.ValidatorJoinStatus(ctx, joiner)
	if approved {
		assert.Error(t, err)
		assert.Nil(t, joinStatus)
	} else {
		assert.NoError(t, err)
		postApprovalCnt := approvalCount(joinStatus)
		assert.Equal(t, preApprovalCnt+1, postApprovalCnt)
	}

	// ValidatorSet count should be equal to postCnt
	vals, err = netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, postCnt, len(vals))
}

func NetworkNodeLeaveSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl) {
	t.Log("Executing network node leave specification")

	vals, err := netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	preCnt := len(vals)

	rec, err := netops.ValidatorNodeLeave(ctx)
	assert.NoError(t, err)

	// Ensure that the Tx is mined.
	expectTxSuccess(t, netops, ctx, rec, defaultTxQueryTimeout)()

	vals, err = netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	postCnt := len(vals)
	assert.Equal(t, preCnt-1, postCnt)
}

func approvalCount(joinStatus *validators.JoinRequest) int {
	cnt := 0
	for _, vote := range joinStatus.Approved {
		if vote {
			cnt += 1
		}
	}
	return cnt
}
