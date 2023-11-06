package specifications

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CurrentValidatorsSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, count int) {
	t.Log("Executing network node validator set specification")
	vals, err := netops.ValidatorsList(ctx)
	require.NoError(t, err)
	require.Equal(t, count, len(vals))
}

func ValidatorNodeJoinSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, joiner []byte, valCount int) {
	t.Log("Executing network node join specification")
	// ValidatorSet count doesn't change just by issuing a Join request. Pre and Post cnt should be the same.
	vals, err := netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, valCount, len(vals))

	// Validator issues a Join request
	rec, err := netops.ValidatorNodeJoin(ctx)
	assert.NoError(t, err)

	// Ensure that the Tx is mined.
	expectTxSuccess(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// Get Request status, #approvals = 0, #board = valCount
	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	assert.NoError(t, err)
	assert.Equal(t, valCount, len(joinStatus.Board))
	assert.Equal(t, 0, approvalCount(joinStatus))

	// Current validators should remain the same
	vals, err = netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, valCount, len(vals))
}

// ValidatorNodeRemoveSpecificationN4R2 tests the validator remove process on a
// network with 4 validators, where nodes 0, 1, and 3 target node 2.
func ValidatorNodeRemoveSpecificationV4R2(ctx context.Context, t *testing.T, n0, n1, n3 ValidatorOpsDsl, n2PubKey []byte) {
	t.Log("Executing network node remove specification")
	// NOTE: In the integration tests where there are multiple nodes and each
	// has their own driver (which individually satisfy a DSL), what is the
	// right way to define these specifications? e.g.
	// TestKwildValidatorUpdatesIntegration creates an integration helper with a
	// number of validators and non-validators defined in that test function, so
	// how is that network information supposed to be conveyed to the test
	// function?  Should it just be part of the assumed environment of the
	// specification i.e. V4R2 for validators = 4, remove node = 2?
	const expectNumVals = 4

	// Ensure that the validator set precondition for this spec test is met.
	vals, err := n0.ValidatorsList(ctx)
	assert.NoError(t, err)
	numVals := len(vals)
	t.Logf("Initial validator set size = %d", numVals)
	if numVals != expectNumVals {
		t.Fatalf("have %d validators, but require %d", numVals, expectNumVals)
	}

	// node 0 sends remove tx targeting node 2
	rec, err := n0.ValidatorNodeRemove(ctx, n2PubKey)
	assert.NoError(t, err)

	expectTxSuccess(t, n0, ctx, rec, defaultTxQueryTimeout)()

	// node 2 is still in the validator set
	vals, err = n0.ValidatorsList(ctx)
	assert.NoError(t, err)
	numVals = len(vals)
	t.Logf("Current validator set size = %d", numVals)
	if numVals != expectNumVals {
		t.Fatalf("have %d validators, but expected %d", numVals, expectNumVals)
	}

	// node 1 also sends remove tx
	rec, err = n1.ValidatorNodeRemove(ctx, n2PubKey)
	assert.NoError(t, err)

	expectTxSuccess(t, n0, ctx, rec, defaultTxQueryTimeout)()

	// node 2 is still in the validator set (2 / 4 validators is sub-threshold)
	vals, err = n0.ValidatorsList(ctx)
	assert.NoError(t, err)
	numVals = len(vals)
	t.Logf("Current validator set size = %d", numVals)
	if numVals != expectNumVals {
		t.Fatalf("have %d validators, but expected %d", numVals, expectNumVals)
	}

	// node 3 also sends remove tx
	rec, err = n3.ValidatorNodeRemove(ctx, n2PubKey)
	assert.NoError(t, err)

	expectTxSuccess(t, n0, ctx, rec, defaultTxQueryTimeout)()

	// node 2 is gone from the validator set
	vals, err = n0.ValidatorsList(ctx)
	assert.NoError(t, err)
	numVals = len(vals)
	t.Logf("Current validator set size = %d", numVals)
	const expectReducedNumVals = expectNumVals - 1
	if numVals != expectReducedNumVals {
		t.Fatalf("have %d validators, but expected %d", numVals, expectReducedNumVals)
	}
}

func ValidatorNodeApproveSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, joiner []byte, preCnt int, postCnt int, approved bool) {
	t.Log("Executing network node approve specification")

	// Get current validator count, should be equal to preCnt
	vals, err := netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, preCnt, len(vals))

	// Get Join Request status, #board = preCnt
	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	assert.NoError(t, err)
	assert.Equal(t, preCnt, len(joinStatus.Board))
	preApprovalCnt := approvalCount(joinStatus)

	// Approval Request
	rec, err := netops.ValidatorNodeApprove(ctx, joiner)
	assert.NoError(t, err)

	// Ensure that the Tx is mined.
	expectTxSuccess(t, netops, ctx, rec, defaultTxQueryTimeout)()

	/*
		Check Join Request Status:
		- If Join request approved (2/3rd majority), Join request should be removed
		- If not approved, ensure that the vote is included, i.e #approvals = preApprovalCnt + 1
	*/
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

func ValidatorNodeLeaveSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl) {
	t.Log("Executing network node leave specification")

	// Get current validator count
	vals, err := netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	preCnt := len(vals)

	// Validator issues a Leave request
	rec, err := netops.ValidatorNodeLeave(ctx)
	assert.NoError(t, err)

	// Ensure that the Validator Leave Tx is mined.
	expectTxSuccess(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// ValidatorSet count should be reduced by 1
	vals, err = netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	postCnt := len(vals)
	assert.Equal(t, preCnt-1, postCnt)
}

func approvalCount(joinStatus *types.JoinRequest) int {
	cnt := 0
	for _, vote := range joinStatus.Approved {
		if vote {
			cnt += 1
		}
	}
	return cnt
}

func ValidatorJoinExpirySpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, joiner []byte, expiry time.Duration) {
	t.Log("Executing validator join expiry specification")

	// Issue a join request
	rec, err := netops.ValidatorNodeJoin(ctx)
	assert.NoError(t, err)

	// Ensure that the Tx is mined.
	expectTxSuccess(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// Get Request status, #approvals = 0
	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	assert.NoError(t, err)
	assert.Equal(t, 0, approvalCount(joinStatus))

	// Wait for the join request to expire
	t.Logf("Waiting %v for join request to expire", expiry)
	time.Sleep(expiry)

	// join request should be expired and removed
	joinStatus, err = netops.ValidatorJoinStatus(ctx, joiner)
	errors.Is(err, status.Error(codes.NotFound, "no active join request"))
	assert.Nil(t, joinStatus)
}
