package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExecuteCallSpecification(ctx context.Context, t *testing.T, caller ExecuteCallDsl, visitor ExecuteCallDsl) {
	t.Logf("Executing ExecuteCallSpecification")

	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := caller.DBID(db.Name)

	getPostInput := []any{1111}

	res, err := caller.Call(ctx, dbID, "get_post", getPostInput)
	if err != nil {
		t.Fatalf("error calling action: %s", err.Error())
	}

	checkGetPostResults(t, res.Export())

	// try calling mutable action, should fail
	_, err = caller.Call(ctx, dbID, "delete_user", nil)
	assert.Error(t, err, "expected error calling mutable action")

	// test that modifiers "public owner view" enforces checks on the caller
	// the caller here is the correct owner
	_, err = caller.Call(ctx, dbID, "owner_only", nil)
	assert.NoError(t, err, "calling owner only action with owner as sender should succeed")

	// TODO: make this a separate specification
	// and test that authenticating works
	_, err = visitor.Call(ctx, dbID, "owner_only", nil)
	assert.Error(t, err, "calling owner only action with non-owner as sender should fail")
}

// ExecuteAuthnCallActionSpecification tests that kgw authn annotation action
// accepts calls with authentication
func ExecuteAuthnCallActionSpecification(ctx context.Context, t *testing.T, caller ExecuteCallDsl, dbid string) {
	t.Logf("Executing ExecuteAuthnCallActionSpecification")

	// try calling authn action, should success
	_, err := caller.Call(ctx, dbid, "authn_only_action", nil)
	assert.NoError(t, err, "expected success calling kgw authn action")
}

// ExecuteAuthnCallProcedureSpecification tests that kgw authn annotation procedure
// accepts calls with authentication
func ExecuteAuthnCallProcedureSpecification(ctx context.Context, t *testing.T, caller ExecuteCallDsl, dbid string) {
	t.Logf("Executing ExecuteAuthnCallProcedureSpecification")

	// try calling authn procedure, should success
	_, err := caller.Call(ctx, dbid, "authn_only_procedure", nil)
	assert.NoError(t, err, "expected success calling kgw authn procedure")
}

func checkGetPostResults(t *testing.T, results []map[string]any) {
	if len(results) != 1 {
		t.Fatalf("expected 1 statement result, got %d", len(results))
	}

	returnedPost := results[0]

	postId, ok := returnedPost["id"].(int64)
	require.Truef(t, ok, "expected a int64, got a %T", returnedPost["id"])

	if postId != 1111 {
		t.Errorf("expected post id to be 1111, got %d", postId)
	}
}
