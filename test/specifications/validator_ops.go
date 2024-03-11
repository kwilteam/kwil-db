package specifications

import (
	"context"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	// Ensure that the Tx is mined.
	expectTxSuccess(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// Get Request status, #approvals = 0, #board = valCount
	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	require.NoError(t, err)
	assert.Equal(t, valCount, len(joinStatus.Board))
	assert.Equal(t, 0, approvalCount(joinStatus))

	// Current validators should remain the same
	vals, err = netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, valCount, len(vals))
}

func JoinExistingValidatorSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, joiner []byte) {
	t.Log("Executing existing validator join specification")

	// Validator issues a Join request
	rec, err := netops.ValidatorNodeJoin(ctx)
	require.NoError(t, err)

	// Ensure that the Tx is mined.
	expectTxFail(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// Get Request status, #approvals = 0, #board = valCount
	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	require.Error(t, err)
	require.Nil(t, joinStatus)
}

// NonValidatorLeaveSpecification tests the validator remove process on a non-validator node
func NonValidatorLeaveSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, target []byte) {
	t.Log("Executing non validator leave specification")

	// Ensure that the validator set precondition for this spec test is met.
	preVals, err := netops.ValidatorsList(ctx)
	require.NoError(t, err)

	// non validator node tries to leave
	rec, err := netops.ValidatorNodeLeave(ctx)
	require.NoError(t, err)

	// Ensure that the Validator Leave Tx is mined.
	expectTxFail(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// ValidatorSet count should remain same
	postVals, err := netops.ValidatorsList(ctx)
	require.NoError(t, err)
	assert.Equal(t, len(preVals), len(postVals))
}

func RemoveNonValidatorSpecification(ctx context.Context, t *testing.T, netops ValidatorRemoveDsl, target []byte) {
	t.Log("Executing remove non-validator specification")

	// Ensure that the validator set precondition for this spec test is met.
	preVals, err := netops.ValidatorsList(ctx)
	require.NoError(t, err)

	// non validator node tries to remove a validator
	rec, err := netops.ValidatorNodeRemove(ctx, target)
	require.NoError(t, err)

	// Ensure that the Validator Remove Tx is mined.
	expectTxFail(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// ValidatorSet count should remain same
	postVals, err := netops.ValidatorsList(ctx)
	require.NoError(t, err)
	assert.Equal(t, len(preVals), len(postVals))
}

// ValidatorNodeRemoveSpecificationN4R1 tests the validator remove process on a
// network with 4 validators, where 3 nodes target the last.
func ValidatorNodeRemoveSpecificationV4R1(ctx context.Context, t *testing.T, n0, n1, n2 ValidatorRemoveDsl, target []byte) {
	t.Log("Executing network node remove specification")

	// Ensure that the validator set precondition for this spec test is met.
	const expectNumVals = 4
	vals, err := n0.ValidatorsList(ctx)
	assert.NoError(t, err)
	numVals := len(vals)
	t.Logf("Initial validator set size = %d", numVals)
	if numVals != expectNumVals {
		t.Fatalf("have %d validators, but require %d", numVals, expectNumVals)
	}

	// node 0 sends remove tx targeting node 3
	rec, err := n0.ValidatorNodeRemove(ctx, target)
	require.NoError(t, err)

	expectTxSuccess(t, n0, ctx, rec, defaultTxQueryTimeout)()

	// node 3 is still in the validator set
	vals, err = n0.ValidatorsList(ctx)
	assert.NoError(t, err)
	numVals = len(vals)
	t.Logf("Current validator set size = %d", numVals)
	if numVals != expectNumVals {
		t.Fatalf("have %d validators, but expected %d", numVals, expectNumVals)
	}

	// node 1 also sends remove tx
	rec, err = n1.ValidatorNodeRemove(ctx, target)
	assert.NoError(t, err)

	expectTxSuccess(t, n0, ctx, rec, defaultTxQueryTimeout)()

	// node 3 is still in the validator set (2 / 4 validators is sub-threshold)
	vals, err = n0.ValidatorsList(ctx)
	assert.NoError(t, err)
	numVals = len(vals)
	t.Logf("Current validator set size = %d", numVals)
	if numVals != expectNumVals {
		t.Fatalf("have %d validators, but expected %d", numVals, expectNumVals)
	}

	// node 2 also sends remove tx
	rec, err = n2.ValidatorNodeRemove(ctx, target)
	assert.NoError(t, err)

	expectTxSuccess(t, n0, ctx, rec, defaultTxQueryTimeout)()

	// node 3 is gone from the validator set
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
	require.NoError(t, err)
	assert.Equal(t, preCnt, len(vals))

	// Get Join Request status, #board = preCnt
	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	require.NoError(t, err)
	assert.Equal(t, preCnt, len(joinStatus.Board))
	preApprovalCnt := approvalCount(joinStatus)

	// Approval Request
	rec, err := netops.ValidatorNodeApprove(ctx, joiner)
	require.NoError(t, err)

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
		require.NoError(t, err)
		postApprovalCnt := approvalCount(joinStatus)
		assert.Equal(t, preApprovalCnt+1, postApprovalCnt)
	}

	// ValidatorSet count should be equal to postCnt
	vals, err = netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, postCnt, len(vals))
}

func ValidatorNodeSelfApproveSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl, joiner []byte) {
	// Get Join Request status, #board = preCnt
	joinStatus, err := netops.ValidatorJoinStatus(ctx, joiner)
	require.NoError(t, err)
	preApprovalCnt := approvalCount(joinStatus)

	// Approval Request
	rec, err := netops.ValidatorNodeApprove(ctx, joiner)
	require.NoError(t, err)

	// TX should fail as validator cannot approve its own join request
	expectTxFail(t, netops, ctx, rec, defaultTxQueryTimeout)()

	// Check Join Request Status: #approvals = preApprovalCnt
	joinStatus, err = netops.ValidatorJoinStatus(ctx, joiner)
	require.NoError(t, err)
	postApprovalCnt := approvalCount(joinStatus)
	assert.Equal(t, preApprovalCnt, postApprovalCnt)
}

func ValidatorNodeLeaveSpecification(ctx context.Context, t *testing.T, netops ValidatorOpsDsl) {
	t.Log("Executing network node leave specification")

	// Get current validator count
	vals, err := netops.ValidatorsList(ctx)
	assert.NoError(t, err)
	preCnt := len(vals)

	// Validator issues a Leave request
	rec, err := netops.ValidatorNodeLeave(ctx)
	require.NoError(t, err)

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
			cnt++
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
	require.NoError(t, err)
	assert.Equal(t, 0, approvalCount(joinStatus))

	// Wait for the join request to expire
	t.Logf("Waiting %v for join request to expire", expiry)
	time.Sleep(expiry)

	// join request should be expired and removed
	joinStatus, err = netops.ValidatorJoinStatus(ctx, joiner)
	assert.Equal(t, err.Error(), "no active join request for that validator")
	assert.Nil(t, joinStatus)
}
