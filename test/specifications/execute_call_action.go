package specifications

import (
	"context"
	"testing"

	"github.com/cstockton/go-conv"
	"github.com/stretchr/testify/assert"
)

func ExecuteCallSpecification(ctx context.Context, t *testing.T, caller ExecuteCallDsl) {
	t.Logf("Executing ExecuteCallSpecification")

	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := caller.DBID(db.Name)

	getPostInput := []any{1111}

	_, err := caller.Call(ctx, dbID, "get_post_authenticated", getPostInput, false)
	if err == nil {
		t.Errorf("expected error calling action without authentication")
	}

	res, err := caller.Call(ctx, dbID, "get_post_authenticated", getPostInput, true)
	if err != nil {
		t.Fatalf("error calling action: %s", err.Error())
	}

	checkGetPostResults(t, res.Export())

	res, err = caller.Call(ctx, dbID, "get_post_unauthenticated", getPostInput, false)
	if err != nil {
		t.Fatalf("error calling action: %s", err.Error())
	}
	checkGetPostResults(t, res.Export())

	// try calling mutable action, should fail
	_, err = caller.Call(ctx, dbID, "delete_user", nil, false)
	assert.Error(t, err, "expected error calling mutable action")

	// test that modifiers "public owner view" enforces authentication
	// the caller here is the correct owner, but not authenticated
	_, err = caller.Call(ctx, dbID, "owner_only", nil, false)
	assert.Error(t, err, "expected error calling owner only action without authentication")

	// and test that authenticating works
	_, err = caller.Call(ctx, dbID, "owner_only", nil, true)
	assert.NoError(t, err, "calling owner only action with authentication should succeed")
}

func checkGetPostResults(t *testing.T, results []map[string]any) {
	if len(results) != 1 {
		t.Fatalf("expected 1 statement result, got %d", len(results))
	}

	returnedPost := results[0]

	postId, _ := conv.Int32(returnedPost["id"])

	if postId != 1111 {
		t.Errorf("expected post id to be 1111, got %d", postId)
	}
}
