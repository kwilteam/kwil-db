package specifications

import (
	"context"
	"testing"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/stretchr/testify/assert"
)

type ExecuteCallDsl interface {
	DatabaseIdentifier
	Call(ctx context.Context, dbid, action string, inputs []any, opts ...client.CallOpt) (*client.Records, error)
}

func ExecuteCallSpecification(ctx context.Context, t *testing.T, caller ExecuteCallDsl) {
	t.Logf("Executing ExecuteCallSpecification")

	db := SchemaLoader.Load(t, schemaTestDB)
	dbID := caller.DBID(db.Name)

	getPostInput := []any{1111}

	res, err := caller.Call(ctx, dbID, "get_post_authenticated", getPostInput)
	if err != nil {
		t.Fatalf("error calling action: %s", err.Error())
	}

	checkGetPostResults(t, res.Export())

	_, err = caller.Call(ctx, dbID, "get_post_authenticated", getPostInput, client.Authenticated(false))
	if err == nil {
		t.Errorf("expected error calling action without authentication")
	}

	res, err = caller.Call(ctx, dbID, "get_post_unauthenticated", getPostInput, client.Authenticated(false))
	if err != nil {
		t.Fatalf("error calling action: %s", err.Error())
	}
	checkGetPostResults(t, res.Export())

	// try calling mutable action, should fail
	_, err = caller.Call(ctx, dbID, "delete_user", nil, client.Authenticated(true))
	assert.Error(t, err, "expected error calling mutable action")

	// test that modifiers "public owner view" enforces authentication
	// the caller here is the correct owner, but not authenticated
	_, err = caller.Call(ctx, dbID, "owner_only", nil, client.Authenticated(false))
	assert.Error(t, err, "expected error calling owner only action without authentication")

	// and test that authenticating works
	_, err = caller.Call(ctx, dbID, "owner_only", nil, client.Authenticated(true))
	assert.NoError(t, err, "expected no error calling owner only action with authentication")
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
